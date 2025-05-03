{
  description = "A simple XY-oscilloscope written in Go.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
      };

      xyosc = pkgs.buildGoModule rec {
        pname = "xyosc";
        version = "0.0.1";

        src = ./.; # Use local source

        subPackages = ["."];

        vendorHash = null; # Replace with hash after running once (or use `nix develop` to vendor)

        buildInputs = with pkgs; [
          gcc
          go
          glfw
          pkg-config
          xorg.libX11.dev
          xorg.libXrandr.dev
          xorg.libXcursor.dev
          xorg.libXinerama.dev
          xorg.libXi.dev
          xorg.libXxf86vm.dev
          libglvnd
          libxkbcommon
          libpulseaudio
          alsa-lib
          libjack2
        ];

        nativeBuildInputs = with pkgs; [pkg-config makeWrapper];

        meta = with pkgs.lib; {
          description = "A simple XY-oscilloscope written in Go.";
          homepage = "https://github.com/make-42/xyosc";
          license = licenses.gpl3;
          platforms = platforms.linux;
        };

        postInstall = ''
          wrapProgram "$out/bin/xyosc" \
            --prefix LD_LIBRARY_PATH : ${pkgs.lib.makeLibraryPath [
            pkgs.glfw
            pkgs.pkg-config
            pkgs.xorg.libX11.dev
            pkgs.xorg.libXrandr.dev
            pkgs.xorg.libXcursor.dev
            pkgs.xorg.libXinerama.dev
            pkgs.xorg.libXi.dev
            pkgs.xorg.libXxf86vm.dev
            pkgs.libxkbcommon
            pkgs.libglvnd
            pkgs.libpulseaudio
            pkgs.alsa-lib
            pkgs.libjack2
          ]}
          install -Dm644 $src/icons/assets/icon.svg $out/share/icons/hicolor/scalable/apps/xyosc.svg
          install -Dm644 $src/xyosc.desktop $out/share/applications/xyosc.desktop
        '';
      };
    in {
      packages.default = xyosc;

      apps.default = flake-utils.lib.mkApp {
        drv = xyosc;
      };

      devShells.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go
          pkg-config
          makeWrapper
        ];
        buildInputs = with pkgs; [
          glfw
          xorg.libX11.dev
          xorg.libXrandr.dev
          xorg.libXcursor.dev
          xorg.libXinerama.dev
          xorg.libXi.dev
          xorg.libXxf86vm.dev
          libglvnd
          libxkbcommon
          libpulseaudio
          alsa-lib
          libjack2
        ];
      };
    });
}
