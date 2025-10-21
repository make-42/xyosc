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
 - XY mode, single channel mode (L/R/Mix modes), bars - can be toggled with the `F` key
 - particles
 - MPRIS support
 - frequency separation
 - theming support
 - shader support (using ebitengine's [Kage](https://ebitengine.org/en/documents/shader.html) shader language)
 - (soon: waterfall mode + keyboard vis (with fading anims))

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
usesystemfonts: true
systemfont: "Maple Mono NF"
fpscounter: false
showfilterinfo: true
filterinfotextsize: 16
filterinfotextpaddingleft: 16
filterinfotextpaddingbottom: 4
showmpris: false
mpristextopacity: 255
targetfps: 240
windowwidth: 1000
windowheight: 1000
windowresizable: false
capturedeviceindex: 0
capturedevicename: ""
capturedevicesamplerate: 0
samplerate: 192000
audiocapturebuffersize: 512
ringbuffersize: 2097152
readbuffersize: 16384
xyoscilloscopereadbuffersize: 2048
readbufferdelay: 32
beatdetectreadbuffersize: 2097152
beatdetectdownsamplefactor: 4
gain: 1
lineopacity: 200
linebrightness: 1
linethickness: 3
lineinvsqrtopacitycontrol: true
lineinvsqrtopacitycontroluselogdecrement: true
lineinvsqrtopacitycontrollogdecrementbase: 200.0
lineinvsqrtopacitycontrollogdecrementoffset: 0.99
linetimedependentopacitycontrol: true
linetimedependentopacitycontrolbase: 0.999
lineopacitycontrolalsoappliestothickness: true
particles: false
particlegenperframeeveryxsamples: 4000
particlemaxcount: 100
particleminsize: 1.0
particlemaxsize: 3.0
particleacceleration: 0.2
particledrag: 5.0
defaultmode: 0
scaleenable: true
scaletextopacity: 255
scalemainaxisenable: true
scaleverttextenable: true
scalehorztextenable: true
scaleverttextsize: 10
scalehorztextsize: 10
scaleverttextpadding: 5
scalehorztextpadding: 5
scaleverttickenable: true
scalehorztickenable: true
scaleverttickexpandtogrid: false
scalehorztickexpandtogrid: false
scalemainaxisstrokethickness: 2
scaleverttickexpandtogridthickness: 0.5
scalehorztickexpandtogridthickness: 0.5
scaleverttickstrokethickness: 1
scalehorztickstrokethickness: 1
scalevertticklength: 10
scalehorzticklength: 10
scalevertdiv: 20
scalehorzdiv: 20
scalehorzdivdynamicpos: true
peakdetectseparator: 100
oscilloscopestartpeakdetection: true
usebetterpeakdetectionalgorithm: true
betterpeakdetectionalgorithmusewindow: true
triggerthroughoutwindow: true
usecomplextriggeringalgorithm: true
complextriggeringalgorithmusecorrelation: true
frequencydetectionusemedian: true
centerpeak: true
quadratureoffset: true
peakdetectedgeguardbuffersize: 30
singlechannelwindow: 8192
periodcrop: false
periodcropcount: 2
periodcroploopovercount: 1
fftbufferoffset: 0
forcecolors: true
accentcolor: "#E7BDB9"
firstcolor: "#E7BDB9"
thirdcolor: "#E7BDB9"
particlecolor: "#F9DCD9"
bgcolor: "#2B1C1A"
disabletransparency: false
copypreviousframe: true
copypreviousframealphadecaybase: 0.0000001
copypreviousframealphadecayspeed: 2.0
beatdetect: false
beatdetectinterval: 100
beatdetectbpmcorrectionspeed: 4
beatdetecttimecorrectionspeed: 0.4
beatdetectmaxbpm: 500.0
beatdetecthalfdisplayedbpm: false
showmetronome: true
metronomeheight: 8
metronomepadding: 8
metronomethinlinemode: true
metronomethinlinethicknesschangewithvelocity: true
metronomethinlinethickness: 64
metronomethinlinehintthickness: 2
showbpm: true
bpmtextsize: 24
barsusewindow: true
barspreserveparsevalenergy: true
barspreventspectralleakageabovefreq: 170000
barswidth: 4
barspaddingedge: 4
barspaddingbetween: 4
barsautogain: true
barsautogainspeed: 0.5
barsautogainminvolume: 0.000000001
barsinterpolatepos: true
barsinterpolateaccel: 20
barsinterpolatedrag: 2
barsinterpolatedirect: 20
barspeakfreqcursor: false
barspeakfreqcursortextdisplaynote: true
barspeakfreqcursortextdisplaynotereffreq: 440
barspeakfreqcursortextsize: 24
barspeakfreqcursortextopacity: 255
barspeakfreqcursortextoffset: -4
barspeakfreqcursorbgwidth: 210
barspeakfreqcursorbgpadding: 2
barspeakfreqcursorinterpolatepos: true
barspeakfreqcursorinterpolatedirect: 1
barspeakfreqcursorinterpolateaccel: 5
barspeakfreqcursorinterpolatedrag: 20
usekaiserinsteadofhannwindow: true
kaiserwindowparam: 8
useshaders: true
shaders:
- name: glow
  arguments:
    Strength: 0.02
- name: glow
  arguments:
    Strength: 0.02
- name: chromaticabberation2
  arguments:
    Strength: 0.005
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
