# xyosc
A simple XY-oscilloscope written in Go.
<p align="middle">
  <img width="507" height="508" alt="image" src="https://github.com/user-attachments/assets/bfe95372-d7a3-46ef-ba8c-bc3eeb7fcbf2" />
  <img width="507" height="508" alt="image" src="https://github.com/user-attachments/assets/c4948f6a-0600-4c2d-a734-06aa2bcfe695" />
</p>
<p align="middle">
  <img width="507" height="508" alt="image" src="https://github.com/user-attachments/assets/8f126293-5f57-4192-a26d-530fec8c6074" />
  <img width="507" height="508" alt="image" src="https://github.com/user-attachments/assets/85c30348-bd23-47ac-84ca-fc007837a0ac" />
</p>


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
 - shader support (using ebitengine's [Kage](https://ebitengine.org/en/documents/shader.html) shader language) - you can cycle through presets with the `P` key
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
app:
  targetfps: 240
  fpscounter: false
  defaultmode: 0
  splash:
    enable: true
    staticsecs: 1
    transsecs: 1
window:
  width: 1000
  height: 1000
  resizable: false
fonts:
  usesystemfont: true
  systemfont: Maple Mono NF
audio:
  capturedevicematchindex: 2
  capturedevicematchname: "Monitor of Easy Effects Sink"
  capturedevicematchsamplerate: 0
  samplerate: 192000
  gain: 1
buffers:
  audiocapturebuffersize: 512
  ringbuffersize: 2097152
  readbuffersize: 16384
  readbufferdelay: 32
  xyoscilloscopereadbuffersize: 8192
  beatdetectreadbuffersize: 2097152
windowfn:
  usekaiserinsteadofhann: true
  kaiserparam: 8
line:
  opacity: 255
  brightness: 1
  thicknessxy: 3
  thicknesssinglechannel: 3
  invsqrtopacitycontrol:
    enable: true
    linsens: 2
		lincutofffrac: 0.02
    uselogdecrement: true
    logbase: 10
    logoffset: 1
  timedependentopacitycontrol:
    enable: true
    base: 0.9997
  opacityalsoaffectsthickness: true
colors:
  useconfigcolorsinsteadofpywal: true
  palette:
    accent: "#C1C7D2"
    first: "#C1C7D2"
    third: "#C1C7D2"
    particle: "#DDE3EE"
    bg: "#080F16"
  bgopacity: 255
  disablebgtransparency: true
imageretention:
  enable: false
  alphadecaybase: 0.1
  alphadecayspeed: 10.0
singlechannelosc:
  displaybuffersize: 8192
  iffilteringshowunfiltered: true
  periodcrop:
    enable: false
    displaycount: 2
    loopovercount: 1
  peakdetect:
    enable: true
    peakdetectseparator: 100
    useacf: true
    acfusewindowfn: true
    usemedian: true
    triggerthroughoutwindow: true
    usecomplextrigger: true
    aligntolastpossiblepeak: true
    complextriggerusecorrelationtosinewave: true
    fftbufferoffset: 0
    edgeguardbuffersize: 30
    quadratureoffsetpeak: true
    centerpeak: true
  smoothwave:
    enable: true
    invtau: 100
    timeindependent: true
    timeindependentfactor: 0.4
    maxperiods: 10
  slew:
    enable: true
    accel: 100
    drag: 20
    direct: 20
scale:
  enable: false
  mainaxisenable: true
  textopacity: 255
  mainaxisthickness: 2
  horz:
    textenable: true
    textsize: 10
    textpadding: 5
    tickenable: true
    ticktogrid: false
    gridthickness: 0.5
    tickthickness: 1
    ticklength: 10
    divs: 20
  vert:
    textenable: true
    textsize: 10
    textpadding: 5
    tickenable: true
    ticktogrid: false
    gridthickness: 0.5
    tickthickness: 1
    ticklength: 10
    divs: 20
  horzdivdynamicpos: true
particles:
  enable: false
  geneveryxsamples: 4000
  maxcount: 100
  minsize: 1
  maxsize: 3
  acceleration: 0.2
  drag: 5
bars:
  usewindowfn: true
  preserveparsevalenergy: true
  preventleakageabovefreq: 160000
  width: 0.5
  paddingedge: 4
  paddingbetween: 0
  autogain:
    enable: true
    speed: 0.5
    minvolume: 1e-09
  interpolate:
    enable: true
    accel: 20
    drag: 2
    direct: 20
  peakcursor:
    enable: false
    shownote: true
    refnotefreq: 440
    textsize: 24
    textopacity: 255
    textoffset: -4
    bgwidth: 210
    bgpadding: 2
    interpolatepos:
      enable: true
      accel: 0.1
      drag: 1
      direct: 0.01
  phasecolors:
    enable: false
    lmult: 0.8
    cmult: 3
    hmult: 1
    interpolate:
      enable: true
      accel: 2
      drag: 0.2
      direct: 2
vu:
  paddingbetween: 64
  paddingedge: 8
  preserveparsevalenergy: true
  logscale: true
  logmaxdb: 3
  logmindb: -40
  linmax: 1.1
  interpolate:
    enable: true
    accel: 2
    drag: 10
    direct: 40
  scale:
    enable: true
    textsize: 12
    textoffset: -2
    logdivisions:
    - 3
    - 2
    - 1
    - 0
    - -1
    - -2
    - -3
    - -4
    - -5
    - -6
    - -8
    - -10
    - -15
    - -20
    - -30
    - -40
    lindivisions:
    - 0
    - 0.1
    - 0.2
    - 0.3
    - 0.4
    - 0.5
    - 0.6
    - 0.7
    - 0.8
    - 0.9
    - 1
    - 1.1
    ticksouter: true
    ticksinner: true
    tickthickness: 1
    ticklength: 2
    tickpadding: 2
  peak:
    enable: true
    historyseconds: 5
    interpolate:
      enable: true
      accel: 2
      drag: 10
      direct: 40
    thickness: 2
beatdetection:
  enable: true
  showbpm: true
  bpmtextsize: 24
  intervalms: 100
  downsamplefactor: 4
  bpmcorrectionspeed: 4
  timecorrectionspeed: 0.4
  maxbpm: 500
  halfdisplayedbpm: false
  metronome:
    enable: true
    height: 8
    padding: 8
    edgemode: false
		edgethickness: 0.5
    thinlinemode: true
    thinlinethicknesschangewithvelocity: true
    thinlinethickness: 64
    thinlinehintthickness: 2
filterinfo:
  enable: true
  textsize: 16
  textpaddingleft: 16
  textpaddingbottom: 4
shaders:
  enable: true
  modepresetslist:
  - - 2
    - 4
    - 5
    - 0
  - - 3
    - 6
    - 0
  - - 3
    - 6
    - 0
  - - 3
    - 6
    - 0
  presets:
  - - name: glow
      arguments:
        Strength: 0.05
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  - - name: glow
      arguments:
        Strength: 0.05
      timescale: 0
    - name: gammacorrectionalphafriendly
      arguments:
        MidPoint: 0.1
        Strength: 2.
      timescale: 0
    - name: gammacorrectionalphafriendly
      arguments:
        MidPoint: 0.45
        Strength: 8.
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  - - name: glow
      arguments:
        Strength: 0.1
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.1
        Strength: 2.
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.45
        Strength: 10.
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  - - name: glow
      arguments:
        Strength: 0.04
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.1
        Strength: 4.
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.45
        Strength: 8.
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  - - name: crtcurve
      arguments:
        Strength: 0.5
      timescale: 0
    - name: glow
      arguments:
        Strength: 0.1
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.1
        Strength: 2.
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.45
        Strength: 10.
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  - - name: crt
      arguments: {}
      timescale: 0
    - name: glow
      arguments:
        Strength: 0.05
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.1
        Strength: 4.
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.45
        Strength: 8.
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  - - name: crtcurve
      arguments:
        Strength: 0.5
      timescale: 0
    - name: glow
      arguments:
        Strength: 0.04
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.1
        Strength: 4.
      timescale: 0
    - name: gammacorrection
      arguments:
        MidPoint: 0.45
        Strength: 8.
      timescale: 0
    - name: chromaticabberation
      arguments:
        Strength: 0.001
      timescale: 0
  customshadercode:
    noise: "//go:build ignore\n\n//kage:unit pixels\n\npackage main\n\nvar Strength
      float\nvar Time float\nvar Scale float\n\nfunc Fragment(dstPos vec4, srcPos
      vec2, color vec4) vec4 {\n\t\t\tvar clr vec4\n\t\t\tclr = imageSrc2At(srcPos)\n\t\t\tamount
      := abs(cos(sin(srcPos.x*Scale+Time+cos(srcPos.y*Scale+Time)*Scale)*Scale+sin(srcPos.x*Scale+Time)*Scale))
      * Strength\n\t\t\tclr.r += amount\n\t\t\tclr.g += amount\n\t\t\tclr.b += amount\n\t\t\tclr.a
      += amount\n\t\t\treturn clr\n}\n\t\t\t"
mpris:
  enable: false
  texttitleyoffset: 0
  textalbumyoffset: -7
  textdurationyoffset: 0
  textopacity: 255
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

https://github.com/user-attachments/assets/25415824-34aa-4e46-824f-699a99bb9cec


https://github.com/user-attachments/assets/a527369c-aa43-45d3-b790-9c0956e8d629


https://github.com/user-attachments/assets/8495da4f-dadc-44ab-9f8d-dc0331f0c421


https://github.com/user-attachments/assets/7b6c6545-5a93-4ba8-8a8a-a381d3e62672
