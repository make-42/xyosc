# xyosc
A simple XY-oscilloscope written in Go.

# Instalation

## Arch

Install `ontake-xyosc-git`

## NixOS

Use this repo as a flake. You can test `xyosc` out with `nix run github:make-42/xyosc` for example

Or, install permanantly with flakes:

Minimal `flake.nix`

```nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    xyosc.url = "github:make-42/xyosc";
  };
  # ...

  outputs = inputs @ {
    nixpkgs,
    xyosc,
    ...
  }: let
    system = "x86_64-linux"; # change to whatever your system should be
  in {
    nixosConfigurations."${host}" = nixpkgs.lib.nixosSystem {
      specialArgs = {
        inherit system;
        inherit inputs;
      };
    };
  };
}
```

And to install
```nix
{  inputs, pkgs, ... }:
environment.systemPackages = [
    inputs.xyosc.packages.${pkgs.system}.default
];
```
And the `xyosc` binary should be available

# Configuration
The configuration file can be found at `~/.config/ontake/xyosc/config.yml`

Note: `xyosc` might not chose the right device to get audio from by default. When you run xyosc it displays a list of capture devices with indices, change the `capturedeviceindex` option to the right index if it isn't the right one by default.

# Features
 - XY mode and single channel mode (L/R/Mix modes) - can be toggled with the `F` key
 - particles
 - MPRIS support
 - frequency separation
 - theming support
 - shader support (using ebitengine's [Kage](https://ebitengine.org/en/documents/shader.html) shader language)

## NixOS

Here is an example of how to configure `xyosc` using Nix.

```nix
{ pkgs, lib, ... }:

{
  xdg.configFile.xyosc = {
    enable = true;
    target = "ontake/xyosc/config.yml";
    text = pkgs.lib.generators.toYAML { } {
      fpscounter = false;
      showfilterinfo = true;
      showmpris = false;
      mpristextopacity = 255;
      targetfps = 120;
      accentcolor = "#FF0000";
      # rest of config below
    };
  };
}
```

# Typical configuration file
```yaml
fpscounter: false
showfilterinfo: true
showmpris: false
mpristextopacity: 255
targetfps: 240
windowwidth: 1000
windowheight: 1000
capturedeviceindex: 0
samplerate: 192000
audiocapturebuffersize: 64
ringbuffersize: 4194304
readbuffersize: 2048
beatdetectreadbuffersize: 4194304
beatdetectdownsamplefactor: 4
gain: 1
lineopacity: 200
linebrightness: 1
linethickness: 3
lineinvsqrtopacitycontrol: false
particles: false
particlegenperframeeveryxsamples: 4000
particlemaxcount: 100
particleminsize: 1.0
particlemaxsize: 3.0
particleacceleration: 0.2
particledrag: 5.0
defaulttosinglechannel: true
peakdetectseparator: 200
singlechannelwindow: 1200
periodcrop: true
periodcropcount: 2
periodcroploopovercount: 1
fftbufferoffset: 2000
forcecolors: true
accentcolor: "#C7C4DD"
firstcolor: "#C7C4DD"
thirdcolor: "#C7C4DD"
particlecolor: "#E4E0EF"
bgcolor: "#1F1F29"
disabletransparency: false
copypreviousframe: true
copypreviousframealpha: 0.1
beatdetect: true
beatdetectinterval: 100
beatdetectbpmcorrectionspeed: 0.01
beatdetecttimecorrectionspeed: 0.001
beatdetectmaxbpm: 500.0
showmetronome: true
metronomeheight: 8
metronomepadding: 8
metronomethinlinemode: true
metronomethinlinethicknesschangewithvelocity: true
metronomethinlinethickness: 64
metronomethinlinehintthickness: 2
showbpm: true
bpmtextsize: 24
useshaders: true
shaders:
- name: glow
  arguments:
    Strength: 0.05
- name: glow
  arguments:
    Strength: 0.05
- name: chromaticabberation
  arguments:
    Strength: 0.005
- name: custom/noise
  arguments:
    Strength: 0.1
    Scale: 1000.0
  timescale: 4.0
customshadercode:
  noise: "//go:build ignore\n\n//kage:unit pixels\n\npackage main\n\nvar Strength
      float\nvar Time float\nvar Scale float\n\nfunc Fragment(dstPos vec4, srcPos vec2,
      color vec4) vec4 {\n\t\t\tvar clr vec4\n\t\t\tclr = imageSrc2At(srcPos)\n\t\t\tamount
      := abs(cos(sin(srcPos.x*Scale+Time+cos(srcPos.y*Scale+Time)*Scale)*Scale+sin(srcPos.x*Scale+Time)*Scale))
      * Strength\n\t\t\tclr.r += amount\n\t\t\tclr.g += amount\n\t\t\tclr.b += amount\n\t\t\tclr.a
      += amount\n\t\t\treturn clr\n}\n"
```

# Example aliases for running the setup seen in the screenshots
```
xyosc-s-1 = "xyosc -mix -lo=0 -hi=0.0015 -width=650 -height=650 -x=-1332 -y=0 -gain=2 &";
xyosc-s-2 = "xyosc -mix -lo=0.0015 -hi=0.0045 -width=650 -height=650 -x=-666 -y=0 -gain=3 &";
xyosc-s-3 = "xyosc -mix -lo=0.0045 -hi=0.0135 -width=650 -height=650 -x=0 -y=-333 -gain=3 &";
xyosc-s-4 = "xyosc -mix -lo=0.0135 -hi=0.1875 -width=650 -height=650 -x=666 -y=0 -gain=2 &";
xyosc-s-5 = "xyosc -mix -lo=0.1875 -hi=0.9375 -width=650 -height=650 -x=1332 -y=0 -gain=4 &";
xyosc-s-6 = "xyosc -mix -width=650 -height=650 -x=0 -y=333 &";
xyosc-sep = "xyosc-s-1 & xyosc-s-2 & xyosc-s-3 & xyosc-s-4 & xyosc-s-5 & xyosc-s-6";
```

# Screenshots
https://github.com/user-attachments/assets/a527369c-aa43-45d3-b790-9c0956e8d629


https://github.com/user-attachments/assets/8495da4f-dadc-44ab-9f8d-dc0331f0c421


https://github.com/user-attachments/assets/7b6c6545-5a93-4ba8-8a8a-a381d3e62672
