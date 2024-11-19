package config

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"xyosc/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v2"
)

type ConfigS struct {
	FPSCounter                       bool
	ShowMPRIS                        bool
	MPRISTextOpacity                 uint8
	TargetFPS                        int32
	WindowWidth                      int32
	WindowHeight                     int32
	CaptureDeviceIndex               int
	SampleRate                       uint32
	RingBufferSize                   uint32
	ReadBufferSize                   uint32
	Gain                             float32
	LineOpacity                      uint8
	LineBrightness                   float64
	LineThickness                    float32
	LineInvSqrtOpacityControl        bool
	Particles                        bool
	ParticleGenPerFrameEveryXSamples int
	ParticleMaxCount                 int
	ParticleMinSize                  float32
	ParticleMaxSize                  float32
	ParticleAcceleration             float32
	ParticleDrag                     float32
}

var DefaultConfig = ConfigS{
	FPSCounter:                       false,
	ShowMPRIS:                        true,
	MPRISTextOpacity:                 255,
	TargetFPS:                        240,
	WindowWidth:                      1300,
	WindowHeight:                     1300,
	CaptureDeviceIndex:               0,
	SampleRate:                       96000,
	RingBufferSize:                   9600,
	ReadBufferSize:                   9600,
	Gain:                             1,
	LineOpacity:                      200,
	LineBrightness:                   1,
	LineThickness:                    3,
	LineInvSqrtOpacityControl:        false,
	Particles:                        true,
	ParticleGenPerFrameEveryXSamples: 2000,
	ParticleMaxCount:                 600,
	ParticleMinSize:                  0.2,
	ParticleMaxSize:                  2.0,
	ParticleAcceleration:             0.015,
	ParticleDrag:                     5.0,
}

var Config ConfigS

var AccentColor color.RGBA
var FirstColor color.RGBA
var ThirdColor color.RGBA
var ThirdColorAdj color.RGBA

var watcher *fsnotify.Watcher

func Init() {
	configPath := configdir.LocalConfig("ontake", "xyosc")
	err := configdir.MakePath(configPath) // Ensure it exists.
	utils.CheckError(err)

	configFile := filepath.Join(configPath, "config.yml")

	// Does the file not exist?
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		// Create the new config file.
		fh, err := os.Create(configFile)
		utils.CheckError(err)
		defer fh.Close()

		encoder := yaml.NewEncoder(fh)
		encoder.Encode(&DefaultConfig)
		Config = DefaultConfig
	} else {
		// Load the existing file.
		fh, err := os.Open(configFile)
		utils.CheckError(err)
		defer fh.Close()

		decoder := yaml.NewDecoder(fh)
		decoder.Decode(&Config)
	}

	// Get pywal accent color
	watcher, err = fsnotify.NewWatcher()
	utils.CheckError(err)
	updatePywalColors()
	walPath := configdir.LocalCache("wal")
	walFile := filepath.Join(walPath, "colors")
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					updatePywalColors()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				utils.CheckError(err)
			}
		}
	}()
	if _, err := os.Stat(walFile); os.IsNotExist(err) {
	} else {
		err = watcher.Add(walFile)
		utils.CheckError(err)
	}
}

func updatePywalColors() {
	walPath := configdir.LocalCache("wal")
	walFile := filepath.Join(walPath, "colors")
	if _, err := os.Stat(walFile); os.IsNotExist(err) {
		AccentColor = color.RGBA{255, 0, 0, Config.LineOpacity}
		FirstColor = color.RGBA{255, 120, 120, Config.LineOpacity}
		ThirdColor = color.RGBA{255, 0, 0, Config.LineOpacity}
		ThirdColorAdj = color.RGBA{uint8(255 * Config.LineBrightness), 0, 0, Config.LineOpacity}
	} else {
		fh, err := os.Open(walFile)
		utils.CheckError(err)
		defer fh.Close()
		scanner := bufio.NewScanner(fh)
		var line int
		var rgbaColor color.RGBA
		for scanner.Scan() {
			if line == 0 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(err)
				FirstColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.LineOpacity}
			}
			if line == 1 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(err)
				AccentColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.LineOpacity}
			}
			if line == 2 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(err)
				ThirdColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.LineOpacity}
				ThirdColorAdj = color.RGBA{uint8(float64(rgbaColor.R) * Config.LineBrightness), uint8(float64(rgbaColor.G) * Config.LineBrightness), uint8(float64(rgbaColor.B) * Config.LineBrightness), Config.LineOpacity}
				break
			}
			line++
		}

	}
}

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}
