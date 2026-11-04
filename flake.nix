{
  description = "Linko flake development environment";

  # Flake inputs
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  # Flake outputs
  outputs =
    {
      self,
      nixpkgs,
      utils,
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      with pkgs;
      {
        # Development environment output
        devShells = {
          default = mkShell {
            # The Nix packages provided in the environment
            packages = [
              bootdev-cli
              gnumake
              go-task
              go_1_27
              gofumpt
              golangci-lint
              golangci-lint-langserver
              gopls
              gotestsum
              gotools
              nodejs_24
              watchexec
            ];
          };
        };
      }
    );
}
