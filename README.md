# xyosc
A simple XY-oscilloscope written in Go.

# Instalation
 - Arch: `ontake-xyosc-git`
 - NixOS: Use this repo as a flake. You can test `xyosc` out with `nix run github:make-42/xyosc` for example

# Configuration
The configuration file can be found at ~/.config/ontake/xyosc/config.yml`

Note: `xyosc` might not chose the right device to get audio from by default. When you run xyosc it displays a list of capture devices with indices, change the `capturedeviceindex` option to the right index if it isn't the right one by default.

# Features
 - XY mode and single channel mode (L/R/Mix modes) - can be toggled with the `F` key
 - particles
 - MPRIS support
 - frequency separation
 - theming support

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

https://github.com/user-attachments/assets/8495da4f-dadc-44ab-9f8d-dc0331f0c421


https://github.com/user-attachments/assets/7b6c6545-5a93-4ba8-8a8a-a381d3e62672


