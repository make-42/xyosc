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

      # Cross-compilation packages for Windows
      pkgsCross = import nixpkgs {
        inherit system;
        crossSystem = {
          config = "i686-w64-mingw32";
        };
      };

      xyosc = pkgs.buildGoModule rec {
        pname = "xyosc";
        version = "0.0.2";

        src = ./.; # Use local source

        subPackages = ["."];

        vendorHash = null; # Replace with hash after running once (or use `nix develop` to vendor)

        buildInputs = with pkgs; [
          gcc
          go
          glfw
          pkg-config
          libx11
          libxrandr
          libxcursor
          libxinerama
          libxi
          libxxf86vm
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
          platforms = platforms.linux ++ platforms.darwin;
        };

        postInstall = ''
          wrapProgram "$out/bin/xyosc" \
            --prefix LD_LIBRARY_PATH : ${pkgs.lib.makeLibraryPath [
            pkgs.glfw
            pkgs.pkg-config
            pkgs.libx11
            pkgs.libxrandr
            pkgs.libxcursor
            pkgs.libxinerama
            pkgs.libxi
            pkgs.libxxf86vm
            pkgs.libxkbcommon
            pkgs.libglvnd
            pkgs.libpulseaudio
            pkgs.alsa-lib
            pkgs.libjack2
          ]}
          install -Dm644 $src/icons/assets/icon.svg $out/share/icons/hicolor/scalable/apps/xyosc.svg
          install -Dm644 $src/xyosc.desktop $out/share/applications/xyosc.desktop
          install -Dm644 $src/xyosc.1 $out/share/man/man1/xyosc.1
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
          pkgsCross.stdenv.cc
          pkgsCross.buildPackages.gcc
        ];
        shellHook = ''
          export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath [
            pkgs.glfw
            pkgs.pkg-config
            pkgs.libx11
            pkgs.libxrandr
            pkgs.libxcursor
            pkgs.libxinerama
            pkgs.libxi
            pkgs.libxxf86vm
            pkgs.libxkbcommon
            pkgs.libglvnd
            pkgs.libpulseaudio
            pkgs.alsa-lib
            pkgs.libjack2
            pkgsCross.stdenv.cc
            pkgsCross.buildPackages.gcc
          ]}:$LD_LIBRARY_PATH
        '';
        buildInputs = with pkgs; [
          glfw
          libx11
          libxrandr
          libxcursor
          libxinerama
          libxi
          libxxf86vm
          libglvnd
          libxkbcommon
          libpulseaudio
          alsa-lib
          libjack2
          pkgsCross.stdenv.cc
          pkgsCross.buildPackages.gcc
        ];
      };
    });
}
