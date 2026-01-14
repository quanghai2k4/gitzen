{
  description = "gitzen dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            git
            goreleaser
            gh
          ];

          env = {
            CGO_ENABLED = "0";
          };

          shellHook = ''
            export GOPATH="$PWD/.go"
            export GOMODCACHE="$PWD/.go/pkg/mod"
            export GOCACHE="$PWD/.go/cache"
            mkdir -p "$GOMODCACHE" "$GOCACHE" "$GOPATH/bin"
            echo "devShell: $(go version)"
          '';

        };
      });
}
