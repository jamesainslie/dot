{
  description = "dot: a safe, deterministic dotfile symlink manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      # Read from the release-please manifest, which is the file release-please
      # updates on every release, so the flake cannot drift from the released
      # tag. It is the derivation version and the string compiled into
      # main.version.
      version = (builtins.fromJSON (builtins.readFile ./.release-please-manifest.json)).".";

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);

      pkgsFor = system: nixpkgs.legacyPackages.${system};
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;

          dot = pkgs.buildGoModule {
            pname = "dot";
            inherit version;

            src = self;

            vendorHash = "sha256-IHnitXddgagFpk0dK85hbZyqdIlXkndjgB7r/a1X5iY=";

            subPackages = [ "cmd/dot" ];

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
              "-X main.commit=${self.shortRev or self.dirtyShortRev or "nix"}"
              # A fixed timestamp keeps the build reproducible; the real build
              # date is carried by the release artifacts, not by the Nix build.
              "-X main.date=1970-01-01T00:00:00Z"
            ];

            # The suite drives the real filesystem and expects a writable HOME
            # plus network access for a few integration paths, neither of which
            # the Nix sandbox provides. Tests run in CI (make check) instead.
            doCheck = false;

            meta = {
              description = "Safe, deterministic dotfile symlink manager";
              homepage = "https://github.com/yaklabco/dot";
              license = pkgs.lib.licenses.mit;
              mainProgram = "dot";
              platforms = systems;
            };
          };
        in
        {
          inherit dot;
          default = dot;
        }
      );

      apps = forAllSystems (system: rec {
        dot = {
          type = "app";
          program = "${self.packages.${system}.dot}/bin/dot";
          meta.description = "Run the dot dotfile symlink manager";
        };
        default = dot;
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.gopls
              pkgs.golangci-lint
              pkgs.gnumake
              pkgs.git
            ];
          };
        }
      );

      formatter = forAllSystems (system: (pkgsFor system).nixfmt-tree);
    };
}
