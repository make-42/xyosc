package config

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"xyosc/utils"

	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v2"
)

type ConfigS struct {
	FPSCounter         bool
	TargetFPS          int32
	WindowWidth        int32
	WindowHeight       int32
	WindowOpacity      float32
	CaptureDeviceIndex int
	SampleRate         uint32
	RingBufferSize     uint32
	ReadBufferSize     uint32
	Gain               float32
	LineOpacity        uint8
	LineThickness      float32
}

var DefaultConfig = ConfigS{
	FPSCounter:         false,
	TargetFPS:          60,
	WindowWidth:        800,
	WindowHeight:       800,
	WindowOpacity:      0.9,
	CaptureDeviceIndex: 0,
	SampleRate:         192000,
	RingBufferSize:     192000,
	ReadBufferSize:     192000,
	Gain:               1,
	LineOpacity:        50,
	LineThickness:      2,
}

var Config ConfigS

var AccentColor color.RGBA

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
	walPath := configdir.LocalCache("wal")
	walFile := filepath.Join(walPath, "colors")
	if _, err = os.Stat(walFile); os.IsNotExist(err) {
		AccentColor = color.RGBA{255, 0, 0, Config.LineOpacity}
	} else {
		fh, err := os.Open(walFile)
		utils.CheckError(err)
		defer fh.Close()
		scanner := bufio.NewScanner(fh)
		var line int
		var rgbaColor color.RGBA
		for scanner.Scan() {
			if line == 1 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(err)
				break
			}
			line++
		}

		AccentColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.LineOpacity}
	}
	fmt.Println(AccentColor)
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
