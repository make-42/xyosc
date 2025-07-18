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
	FPSCounter                                   bool
	ShowFilterInfo                               bool
	ShowMPRIS                                    bool
	MPRISTextOpacity                             uint8
	TargetFPS                                    int32
	WindowWidth                                  int32
	WindowHeight                                 int32
	CaptureDeviceIndex                           int
	CaptureDeviceName                            string
	CaptureDeviceSampleRate                      int
	SampleRate                                   uint32
	AudioCaptureBufferSize                       uint32
	RingBufferSize                               uint32
	ReadBufferSize                               uint32
	ReadBufferDelay                              uint32
	BeatDetectReadBufferSize                     uint32
	BeatDetectDownSampleFactor                   uint32
	Gain                                         float32
	LineOpacity                                  uint8
	LineBrightness                               float64
	LineThickness                                float32
	LineInvSqrtOpacityControl                    bool
	Particles                                    bool
	ParticleGenPerFrameEveryXSamples             int
	ParticleMaxCount                             int
	ParticleMinSize                              float32
	ParticleMaxSize                              float32
	ParticleAcceleration                         float32
	ParticleDrag                                 float32
	DefaultToSingleChannel                       bool
	PeakDetectSeparator                          int
	OscilloscopeStartPeakDetection               bool
	PeakDetectEdgeGuardBufferSize                uint32
	SingleChannelWindow                          uint32
	PeriodCrop                                   bool
	PeriodCropCount                              int
	PeriodCropLoopOverCount                      uint32
	FFTBufferOffset                              uint32
	ForceColors                                  bool
	AccentColor                                  string
	FirstColor                                   string
	ThirdColor                                   string
	ParticleColor                                string
	BGColor                                      string
	DisableTransparency                          bool
	CopyPreviousFrame                            bool
	CopyPreviousFrameAlpha                       float32
	BeatDetect                                   bool
	BeatDetectInterval                           int64 //ms
	BeatDetectBPMCorrectionSpeed                 float64
	BeatDetectTimeCorrectionSpeed                float64
	BeatDetectMaxBPM                             float64
	ShowMetronome                                bool
	MetronomeHeight                              float64
	MetronomePadding                             float64
	MetronomeThinLineMode                        bool
	MetronomeThinLineThicknessChangeWithVelocity bool
	MetronomeThinLineThickness                   float64
	MetronomeThinLineHintThickness               float64
	ShowBPM                                      bool
	BPMTextSize                                  float64
	UseShaders                                   bool
	Shaders                                      []Shader
	CustomShaderCode                             map[string]string
}

var DefaultConfig = ConfigS{
	FPSCounter:                       false,
	ShowFilterInfo:                   true,
	ShowMPRIS:                        true,
	MPRISTextOpacity:                 255,
	TargetFPS:                        240,
	WindowWidth:                      1300,
	WindowHeight:                     1300,
	CaptureDeviceIndex:               0,
	CaptureDeviceName:                "",
	CaptureDeviceSampleRate:          0, // In case there are multiple outputs with different sample rates and you want to pick a specific one, else leave equal to 0
	SampleRate:                       192000,
	AudioCaptureBufferSize:           64, // Affects latency
	RingBufferSize:                   262144 * 16,
	ReadBufferSize:                   9600,
	ReadBufferDelay:                  32,
	BeatDetectReadBufferSize:         262144 * 16,
	BeatDetectDownSampleFactor:       4,
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
	DefaultToSingleChannel:           false,
	PeakDetectSeparator:              100,
	OscilloscopeStartPeakDetection:   true,
	PeakDetectEdgeGuardBufferSize:    100,
	SingleChannelWindow:              1200,
	PeriodCrop:                       true,
	PeriodCropCount:                  2,
	PeriodCropLoopOverCount:          1,
	FFTBufferOffset:                  3200,
	ForceColors:                      false,
	AccentColor:                      "#FF0000",
	FirstColor:                       "#FF0000",
	ThirdColor:                       "#FF0000",
	ParticleColor:                    "#FF0000",
	BGColor:                          "#222222",
	DisableTransparency:              false,
	CopyPreviousFrame:                true,
	CopyPreviousFrameAlpha:           0.4,
	BeatDetect:                       true,
	BeatDetectInterval:               100, // ms
	BeatDetectBPMCorrectionSpeed:     0.01,
	BeatDetectTimeCorrectionSpeed:    0.001,
	BeatDetectMaxBPM:                 500.0,
	ShowMetronome:                    true,
	MetronomeHeight:                  8,
	MetronomePadding:                 8,
	MetronomeThinLineMode:            true,
	MetronomeThinLineThicknessChangeWithVelocity: true,
	MetronomeThinLineThickness:                   64,
	MetronomeThinLineHintThickness:               2,
	ShowBPM:                                      true,
	BPMTextSize:                                  24,
	UseShaders:                                   true,
	Shaders: []Shader{
		{
			Name: "glow",
			Arguments: map[string]any{
				"Strength": 0.1,
			},
		}, {
			Name: "chromaticabberation",
			Arguments: map[string]any{
				"Strength": 0.01,
			},
		},
	},
	CustomShaderCode: map[string]string{
		"noise": `//go:build ignore

//kage:unit pixels

package main

var Strength float
var Time float
var Scale float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
			var clr vec4
			clr = imageSrc2At(srcPos)
			amount := abs(cos(sin(srcPos.x*Scale+Time+cos(srcPos.y*Scale+Time)*Scale)*Scale+sin(srcPos.x*Scale+Time)*Scale)) * Strength
			clr.r += amount
			clr.g += amount
			clr.b += amount
			clr.a += amount
			return clr
}
`,
	},
}

type Shader struct {
	Name      string
	Arguments map[string]any
	TimeScale float32
}

var Config ConfigS

var AccentColor color.RGBA
var FirstColor color.RGBA
var ThirdColor color.RGBA
var ParticleColor color.RGBA
var ThirdColorAdj color.RGBA
var BGColor color.RGBA

var watcher *fsnotify.Watcher

var SingleChannel bool = false

var HannWindow []float64

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
		Config = DefaultConfig
		// Load the existing file.
		fh, err := os.Open(configFile)
		utils.CheckError(err)
		defer fh.Close()

		decoder := yaml.NewDecoder(fh)
		decoder.Decode(&Config)
	}
	SingleChannel = Config.DefaultToSingleChannel

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
	/* This is not synced to pywal */
	BGColorParsed, err := ParseHexColor(Config.BGColor)
	utils.CheckError(err)
	BGColor = color.RGBA{BGColorParsed.R, BGColorParsed.G, BGColorParsed.B, 255}
	/* end */

	walPath := configdir.LocalCache("wal")
	walFile := filepath.Join(walPath, "colors")
	if _, err := os.Stat(walFile); os.IsNotExist(err) || Config.ForceColors {
		AccentColorParsed, err := ParseHexColor(Config.AccentColor)
		utils.CheckError(err)
		FirstColorParsed, err := ParseHexColor(Config.FirstColor)
		utils.CheckError(err)
		ThirdColorParsed, err := ParseHexColor(Config.ThirdColor)
		utils.CheckError(err)
		ParticleColorParsed, err := ParseHexColor(Config.ParticleColor)
		utils.CheckError(err)

		AccentColor = color.RGBA{AccentColorParsed.R, AccentColorParsed.G, AccentColorParsed.B, Config.LineOpacity}
		FirstColor = color.RGBA{FirstColorParsed.R, FirstColorParsed.G, FirstColorParsed.B, Config.LineOpacity}
		ThirdColor = color.RGBA{ThirdColorParsed.R, ThirdColorParsed.G, ThirdColorParsed.B, Config.LineOpacity}
		ParticleColor = color.RGBA{ParticleColorParsed.R, ParticleColorParsed.G, ParticleColorParsed.B, Config.LineOpacity}
		ThirdColorAdj = color.RGBA{uint8(float64(ThirdColorParsed.R) * Config.LineBrightness), uint8(float64(ThirdColorParsed.G) * Config.LineBrightness), uint8(float64(ThirdColorParsed.B) * Config.LineBrightness), Config.LineOpacity}
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
				ParticleColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.LineOpacity}
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
