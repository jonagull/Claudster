{
  description = "Claudster — Claude Code TUI session manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = if self ? shortRev then self.shortRev else "dev";
      in
      {
        packages = rec {
          claudster = pkgs.buildGoModule {
            pname = "claudster";
            inherit version;
            src = ./.;

            vendorHash = "sha256-aJllcMJduoi8VBWMJWsxm8swXtNonYZzX8etmNZePzc=";

            ldflags = [ "-X claudster/ui.Version=${version}" ];

            meta = with pkgs.lib; {
              description = "TUI session manager for Claude Code";
              homepage = "https://github.com/jonagull/claudster";
              mainProgram = "claudster";
            };
          };

          default = claudster;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
            tmux
          ];

          shellHook = ''
            echo "claudster dev shell — $(go version)"
          '';
        };
      }
    );
}
