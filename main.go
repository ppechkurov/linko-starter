package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boot.dev/linko/internal/store"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	httpPort := flag.Int("port", 8899, "port to listen on")
	dataDir := flag.String("data", "./data", "directory to store data")
	flag.Parse()

	status := run(ctx, cancel, *httpPort, *dataDir)
	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc, httpPort int, dataDir string) int {
	logger, close, err := initializeLogger()
	if err != nil {
		fmt.Printf("initializing logger: %v\n", err)
		return 1
	}
	defer func() {
		if err = close(); err != nil {
			fmt.Printf("closing logger: %v\n", err)
		}
	}()

	st, err := store.New(dataDir, logger)
	if err != nil {
		logger.Printf("failed to create store: %v\n", err)
		return 1
	}

	s := newServer(*st, httpPort, logger, cancel)
	s.logger = logger
	var serverErr error
	go func() {
		serverErr = s.start()
	}()

	logger.Printf("Linko is running on http://localhost:%d\n", httpPort)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Printf("Linko is shutting down")
	if err := s.shutdown(shutdownCtx); err != nil {
		logger.Printf("failed to shutdown server: %v\n", err)
		return 1
	}
	if serverErr != nil {
		logger.Printf("server error: %v\n", serverErr)
		return 1
	}
	return 0
}

type closeFn func() error

func initializeLogger() (*log.Logger, closeFn, error) {
	if logFileName := os.Getenv("LINKO_LOG_FILE"); logFileName != "" {
		f, err := os.OpenFile(
			logFileName,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0o644,
		)
		buf := bufio.NewWriterSize(f, 8192)
		if err != nil {
			panic(err)
		}

		return log.New(
			io.MultiWriter(os.Stderr, buf),
			"",
			log.LstdFlags,
		), buf.Flush, nil
	}

	return log.New(
		os.Stderr,
		"",
		log.LstdFlags,
	), func() error { return nil }, nil
}
