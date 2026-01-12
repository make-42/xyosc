package config

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/kirsle/configdir"
	"github.com/ztrue/tracerr"
	"gopkg.in/yaml.v2"

	"xyosc/utils"
)

const (
	XYMode int = iota
	SingleChannelMode
	BarsMode
	VUMode
)

type SplashConfig struct {
	Enable     bool
	StaticSecs float64
	TransSecs  float64
}

type AppConfig struct {
	TargetFPS   int32
	FPSCounter  bool
	DefaultMode int // 0 = XY, 1 = SingleChannel, 2 = Bars
	Splash      SplashConfig
}

type WindowConfig struct {
	Width     int32
	Height    int32
	Resizable bool
}

type FontConfig struct {
	UseSystemFont bool
	SystemFont    string
}

type AudioDeviceConfig struct {
	CaptureDeviceMatchIndex      int
	CaptureDeviceMatchName       string
	CaptureDeviceMatchSampleRate int
	SampleRate                   uint32
	Gain                         float32
}

type AudioBuffersConfig struct {
	AudioCaptureBufferSize       uint32
	RingBufferSize               uint32
	ReadBufferSize               uint32
	ReadBufferDelay              uint32
	XYOscilloscopeReadBufferSize uint32
	BeatDetectReadBufferSize     uint32
}

type SmoothWaveConfig struct {
	Enable                bool
	InvTau                float64
	TimeIndependent       bool
	TimeIndependentFactor float64
	MaxPeriods            uint32
}

type PeriodCropConfig struct {
	Enable        bool
	DisplayCount  int
	LoopOverCount uint32
}

type SingleChannelOscilloscopeConfig struct {
	DisplayBufferSize uint32
	PeriodCrop        PeriodCropConfig
	PeakDetect        PeakDetectionConfig
	SmoothWave        SmoothWaveConfig
	Slew              InterpolationConfig
}

type PeakDetectionConfig struct {
	Enable                                 bool
	PeakDetectSeparator                    int
	UseACF                                 bool // ACF
	ACFUseWindowFn                         bool
	UseMedian                              bool
	TriggerThroughoutWindow                bool
	UseComplexTrigger                      bool
	AlignToLastPossiblePeak                bool
	ComplexTriggerUseCorrelationToSineWave bool
	FFTBufferOffset                        uint32
	EdgeGuardBufferSize                    uint32
	QuadratureOffsetPeak                   bool
	CenterPeak                             bool
}

type InvSqrtOpacityControlConfig struct {
	Enable          bool
	UseLogDecrement bool
	LogBase         float64
	LogOffset       float64
}

type TimeDependentOpacityControlConfig struct {
	Enable bool
	Base   float64
}

type LineConfig struct {
	Opacity                     uint8
	Brightness                  float64
	ThicknessXY                 float32
	ThicknessSingleChannel      float32
	InvSqrtOpacityControl       InvSqrtOpacityControlConfig
	TimeDependentOpacityControl TimeDependentOpacityControlConfig
	OpacityAlsoAffectsThickness bool
}

type ParticleConfig struct {
	Enable           bool
	GenEveryXSamples int
	MaxCount         int
	MinSize          float32
	MaxSize          float32
	Acceleration     float32
	Drag             float32
}

type AxisConfig struct {
	TextEnable    bool
	TextSize      float64
	TextPadding   float64
	TickEnable    bool
	TickToGrid    bool
	GridThickness float32
	TickThickness float32
	TickLength    float32
	Divs          int
}

type ScaleConfig struct {
	Enable         bool
	MainAxisEnable bool
	TextOpacity    uint8

	MainAxisThickness float32

	Horz AxisConfig
	Vert AxisConfig

	HorzDivDynamicPos bool
}

type PaletteConfig struct {
	Accent   string
	First    string
	Third    string
	Particle string
	BG       string
}

type ColorConfig struct {
	UseConfigColorsInsteadOfPywal bool
	Palette                       PaletteConfig
	BGOpacity                     uint8
	DisableBGTransparency         bool
}

type ImageRetentionConfig struct {
	Enable          bool
	AlphaDecayBase  float64
	AlphaDecaySpeed float64
}

type MetronomeConfig struct {
	Enable                              bool
	Height                              float64
	Padding                             float64
	EdgeMode                            bool
	EdgeThickness                       float64
	ThinLineMode                        bool
	ThinLineThicknessChangeWithVelocity bool
	ThinLineThickness                   float64
	ThinLineHintThickness               float64
}

type BeatDetectConfig struct {
	Enable              bool
	ShowBPM             bool
	BPMTextSize         float64
	IntervalMS          int64
	DownSampleFactor    uint32
	BPMCorrectionSpeed  float64
	TimeCorrectionSpeed float64
	MaxBPM              float64
	HalfDisplayedBPM    bool
	Metronome           MetronomeConfig
}

type InterpolationConfig struct {
	Enable bool
	Accel  float64
	Drag   float64
	Direct float64
}

type AutoGainConfig struct {
	Enable    bool
	Speed     float64
	MinVolume float64
}

type PhaseColorsConfig struct {
	Enable      bool
	LMult       float64
	CMult       float64
	HMult       float64
	Interpolate InterpolationConfig
}

type BarsConfig struct {
	UseWindowFn             bool
	PreserveParsevalEnergy  bool
	PreventLeakageAboveFreq float64
	Width                   float64
	PaddingEdge             float64
	PaddingBetween          float64

	AutoGain    AutoGainConfig
	Interpolate InterpolationConfig

	PeakCursor  PeakCursorConfig
	PhaseColors PhaseColorsConfig
}

type PeakCursorConfig struct {
	Enable      bool
	ShowNote    bool
	RefNoteFreq float64
	TextSize    float64
	TextOpacity uint8
	TextOffset  float64
	BGWidth     float64
	BGPadding   float64

	InterpolatePos InterpolationConfig
}

type WindowFunctionConfig struct {
	UseKaiserInsteadOfHann bool
	KaiserParam            float64
}

type VUScaleConfig struct {
	Enable        bool
	TextSize      float64
	TextOffset    float64
	LogDivisions  []float64
	LinDivisions  []float64
	TicksOuter    bool
	TicksInner    bool
	TickThickness float32
	TickLength    float64
	TickPadding   float64
}

type VUPeakConfig struct {
	Enable         bool
	HistorySeconds float64
	Interpolate    InterpolationConfig
	Thickness      float32
}

type VUConfig struct {
	PaddingBetween         float64
	PaddingEdge            float64
	PreserveParsevalEnergy bool

	LogScale bool
	LogMaxDB float64
	LogMinDB float64
	LinMax   float64

	Interpolate InterpolationConfig

	Scale VUScaleConfig

	Peak VUPeakConfig
}

type FilterInfoConfig struct {
	Enable            bool
	TextSize          float64
	TextPaddingLeft   float64
	TextPaddingBottom float64
}

type ShaderConfig struct {
	Enable           bool
	ModePresetsList  [][]int
	Presets          [][]Shader
	CustomShaderCode map[string]string
}

type MPRISConfig struct {
	Enable              bool
	TextTitleYOffset    float64
	TextAlbumYOffset    float64
	TextDurationYOffset float64
	TextOpacity         uint8
}

type ConfigS struct {
	App            AppConfig
	Window         WindowConfig
	Fonts          FontConfig
	Audio          AudioDeviceConfig
	Buffers        AudioBuffersConfig
	WindowFn       WindowFunctionConfig
	Line           LineConfig
	Colors         ColorConfig
	ImageRetention ImageRetentionConfig

	SingleChannelOsc SingleChannelOscilloscopeConfig
	Scale            ScaleConfig
	Particles        ParticleConfig
	Bars             BarsConfig
	VU               VUConfig

	BeatDetection BeatDetectConfig

	FilterInfo FilterInfoConfig

	Shaders ShaderConfig
	MPRIS   MPRISConfig
}

var DefaultConfig = ConfigS{
	App: AppConfig{
		TargetFPS:   240,
		FPSCounter:  false,
		DefaultMode: XYMode, // 0 = XY-Oscilloscope, 1 = SingleChannel-Oscilloscope, 2 = Bars 3 = VU
		Splash: SplashConfig{
			Enable:     true,
			StaticSecs: 1,
			TransSecs:  1,
		},
	},
	Window: WindowConfig{
		Width:     1000,
		Height:    1000,
		Resizable: false,
	},
	Fonts: FontConfig{
		UseSystemFont: true,
		SystemFont:    "Maple Mono NF",
	},
	Audio: AudioDeviceConfig{
		CaptureDeviceMatchIndex:      0,
		CaptureDeviceMatchName:       "",
		CaptureDeviceMatchSampleRate: 0,
		SampleRate:                   192000,
		Gain:                         1.0,
	},
	Buffers: AudioBuffersConfig{
		AudioCaptureBufferSize:       512, // Affects latency
		RingBufferSize:               2097152,
		ReadBufferSize:               16384,
		ReadBufferDelay:              32,
		XYOscilloscopeReadBufferSize: 2048,
		BeatDetectReadBufferSize:     2097152,
	},
	WindowFn: WindowFunctionConfig{
		UseKaiserInsteadOfHann: true,
		KaiserParam:            8.,
	},
	Line: LineConfig{
		Opacity:                200,
		Brightness:             1,
		ThicknessXY:            3,
		ThicknessSingleChannel: 3,
		InvSqrtOpacityControl: InvSqrtOpacityControlConfig{
			Enable:          true,
			UseLogDecrement: true,
			LogBase:         200.0,
			LogOffset:       0.99,
		},
		TimeDependentOpacityControl: TimeDependentOpacityControlConfig{
			Enable: true,
			Base:   0.999,
		},
		OpacityAlsoAffectsThickness: true,
	},
	Colors: ColorConfig{
		UseConfigColorsInsteadOfPywal: true,
		Palette: PaletteConfig{
			Accent:   "#E7BDB9",
			First:    "#E7BDB9",
			Third:    "#E7BDB9",
			Particle: "#F9DCD9",
			BG:       "#2B1C1A",
		},
		BGOpacity:             200,
		DisableBGTransparency: false,
	},
	ImageRetention: ImageRetentionConfig{
		Enable:          true,
		AlphaDecayBase:  0.0000001, // these two options are redundent at infinite FPS (but they should yield different results at low FPS)
		AlphaDecaySpeed: 2.0,
	},
	SingleChannelOsc: SingleChannelOscilloscopeConfig{
		DisplayBufferSize: 8192,
		PeriodCrop: PeriodCropConfig{
			Enable:        false,
			DisplayCount:  2,
			LoopOverCount: 1,
		},
		PeakDetect: PeakDetectionConfig{
			Enable:                                 true,
			PeakDetectSeparator:                    100,
			UseACF:                                 true, // ACF
			ACFUseWindowFn:                         true,
			UseMedian:                              true,
			TriggerThroughoutWindow:                true,
			UseComplexTrigger:                      true,
			AlignToLastPossiblePeak:                true,
			ComplexTriggerUseCorrelationToSineWave: true,
			FFTBufferOffset:                        0,
			EdgeGuardBufferSize:                    30,
			QuadratureOffsetPeak:                   true,
			CenterPeak:                             true},
		SmoothWave: SmoothWaveConfig{
			Enable:                true,
			InvTau:                100, //s^-1
			TimeIndependent:       true,
			TimeIndependentFactor: 0.4,
			MaxPeriods:            10,
		},
		Slew: InterpolationConfig{
			Enable: true,
			Accel:  100,
			Drag:   20,
			Direct: 20, // 1,10,20 gives really smooth results (100,20,20 sounds like a good compromise)
		},
	},
	Scale: ScaleConfig{
		Enable:         true,
		MainAxisEnable: true,
		TextOpacity:    255,

		MainAxisThickness: 2,

		Horz: AxisConfig{
			TextEnable:    true,
			TextSize:      10.,
			TextPadding:   5.,
			TickEnable:    true,
			TickToGrid:    false,
			GridThickness: 0.5,
			TickThickness: 1.0,
			Divs:          20,
			TickLength:    10,
		},
		Vert: AxisConfig{
			TextEnable:    true,
			TextSize:      10.,
			TextPadding:   5.,
			TickEnable:    true,
			TickToGrid:    false,
			GridThickness: 0.5,
			TickThickness: 1.0,
			TickLength:    10,
			Divs:          20,
		},
		HorzDivDynamicPos: true,
	},
	Particles: ParticleConfig{
		Enable:           false,
		GenEveryXSamples: 4000,
		MaxCount:         100,
		MinSize:          1.0,
		MaxSize:          3.0,
		Acceleration:     0.2,
		Drag:             5.0,
	},
	Bars: BarsConfig{
		UseWindowFn:             true,
		PreserveParsevalEnergy:  true,
		PreventLeakageAboveFreq: 170000, //Hz
		Width:                   4,
		PaddingEdge:             4,
		PaddingBetween:          4,

		AutoGain: AutoGainConfig{
			Enable:    true,
			Speed:     0.5,
			MinVolume: 0.000000001,
		},
		Interpolate: InterpolationConfig{
			Enable: true,
			Accel:  20.,
			Drag:   2.,
			Direct: 20.,
		},
		PeakCursor: PeakCursorConfig{
			Enable:      false,
			ShowNote:    true,
			RefNoteFreq: 440.,
			TextSize:    24,
			TextOpacity: 255,
			TextOffset:  -4,
			BGWidth:     210,
			BGPadding:   2,
			InterpolatePos: InterpolationConfig{
				Enable: true,
				Direct: 1,
				Accel:  5,
				Drag:   20,
			},
		},
		PhaseColors: PhaseColorsConfig{
			Enable: false,
			LMult:  0.8,
			CMult:  3.0,
			HMult:  1.0,
			Interpolate: InterpolationConfig{
				Enable: true,
				Accel:  2.,
				Drag:   .2,
				Direct: 2.,
			},
		},
	},
	VU: VUConfig{
		PaddingBetween:         64,
		PaddingEdge:            8,
		PreserveParsevalEnergy: true,

		LogScale: true,
		LogMaxDB: 3,
		LogMinDB: -40,
		LinMax:   1.1,

		Interpolate: InterpolationConfig{
			Enable: true,
			Accel:  2,
			Drag:   10,
			Direct: 40,
		},

		Scale: VUScaleConfig{
			Enable:        true,
			TextSize:      12,
			TextOffset:    -2,
			LogDivisions:  []float64{3., 2.0, 1.0, 0, -1, -2, -3, -4, -5, -6, -8, -10., -15, -20, -30, -40},
			LinDivisions:  []float64{0., 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1},
			TicksOuter:    true,
			TicksInner:    true,
			TickThickness: 1,
			TickLength:    2,
			TickPadding:   2,
		},

		Peak: VUPeakConfig{
			Enable:         true,
			HistorySeconds: 5.,
			Interpolate: InterpolationConfig{
				Enable: true,
				Direct: 40,
				Accel:  2,
				Drag:   10,
			},
			Thickness: 2,
		},
	},
	BeatDetection: BeatDetectConfig{
		Enable:              true,
		ShowBPM:             true,
		BPMTextSize:         24,
		IntervalMS:          100,
		DownSampleFactor:    4,
		BPMCorrectionSpeed:  4,
		TimeCorrectionSpeed: 0.4,
		MaxBPM:              500.0,
		HalfDisplayedBPM:    false,
		Metronome: MetronomeConfig{
			Enable:                              true,
			Height:                              8,
			Padding:                             8,
			EdgeMode:                            false,
			EdgeThickness:                       0.5,
			ThinLineMode:                        true,
			ThinLineThicknessChangeWithVelocity: true,
			ThinLineThickness:                   64,
			ThinLineHintThickness:               2,
		},
	},
	FilterInfo: FilterInfoConfig{
		Enable:            true,
		TextSize:          16,
		TextPaddingLeft:   16,
		TextPaddingBottom: 4,
	},
	Shaders: ShaderConfig{
		Enable: true,
		ModePresetsList: [][]int{
			{2, 4, 5, 0}, {3, 6, 0}, {3, 6, 0}, {3, 6, 0},
		},
		Presets: [][]Shader{
			{
				{
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.05,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
				},
			}, {
				{
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.05,
					},
				}, {
					Name: "gammacorrectionalphafriendly",
					Arguments: map[string]any{
						"Strength": 2.,
						"MidPoint": 0.1,
					},
				}, {
					Name: "gammacorrectionalphafriendly",
					Arguments: map[string]any{
						"Strength": 8.,
						"MidPoint": 0.45,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
				},
			}, {
				{
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.10,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 2.,
						"MidPoint": 0.1,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 10.,
						"MidPoint": 0.45,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
				},
			}, {
				{
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.04,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 4.,
						"MidPoint": 0.1,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 8.,
						"MidPoint": 0.45,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
				},
			}, {
				{
					Name: "crtcurve",
					Arguments: map[string]any{
						"Strength": 0.5,
					},
				},
				{
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.10,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 2.,
						"MidPoint": 0.1,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 10.,
						"MidPoint": 0.45,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
				},
			}, {
				{
					Name:      "crt",
					Arguments: map[string]any{},
				}, {
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.05,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 4.,
						"MidPoint": 0.1,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 8.,
						"MidPoint": 0.45,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
				},
			}, {
				{
					Name: "crtcurve",
					Arguments: map[string]any{
						"Strength": 0.5,
					},
				}, {
					Name: "glow",
					Arguments: map[string]any{
						"Strength": 0.04,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 4.,
						"MidPoint": 0.1,
					},
				}, {
					Name: "gammacorrection",
					Arguments: map[string]any{
						"Strength": 8.,
						"MidPoint": 0.45,
					},
				}, {
					Name: "chromaticabberation",
					Arguments: map[string]any{
						"Strength": 0.001,
					},
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
	},
	MPRIS: MPRISConfig{
		Enable:              false,
		TextTitleYOffset:    0,
		TextAlbumYOffset:    -7,
		TextDurationYOffset: 0,
		TextOpacity:         255,
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

var FiltersApplied bool

var HannWindow []float64

func Init() {
	configPath := configdir.LocalConfig("ontake", "xyosc")
	err := configdir.MakePath(configPath) // Ensure it exists.
	utils.CheckError(tracerr.Wrap(err))

	configFile := filepath.Join(configPath, "config.yml")

	// Does the file not exist?
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		// Create the new config file.
		fh, err := os.Create(configFile)
		utils.CheckError(tracerr.Wrap(err))
		defer fh.Close()

		encoder := yaml.NewEncoder(fh)
		encoder.Encode(&DefaultConfig)
		Config = DefaultConfig
	} else {
		Config = DefaultConfig
		// Load the existing file.
		fh, err := os.Open(configFile)
		utils.CheckError(tracerr.Wrap(err))
		defer fh.Close()

		decoder := yaml.NewDecoder(fh)
		decoder.Decode(&Config)
	}

	// Get pywal accent color
	watcher, err = fsnotify.NewWatcher()
	utils.CheckError(tracerr.Wrap(err))
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
				utils.CheckError(tracerr.Wrap(err))
			}
		}
	}()
	if _, err := os.Stat(walFile); os.IsNotExist(err) {
	} else {
		err = watcher.Add(walFile)
		utils.CheckError(tracerr.Wrap(err))
	}
}

func updatePywalColors() {
	/* This is not synced to pywal */
	BGColorParsed, err := ParseHexColor(Config.Colors.Palette.BG)
	utils.CheckError(tracerr.Wrap(err))
	BGColor = color.RGBA{BGColorParsed.R, BGColorParsed.G, BGColorParsed.B, Config.Colors.BGOpacity}
	/* end */

	walPath := configdir.LocalCache("wal")
	walFile := filepath.Join(walPath, "colors")
	if _, err := os.Stat(walFile); os.IsNotExist(err) || Config.Colors.UseConfigColorsInsteadOfPywal {
		AccentColorParsed, err := ParseHexColor(Config.Colors.Palette.Accent)
		utils.CheckError(tracerr.Wrap(err))
		FirstColorParsed, err := ParseHexColor(Config.Colors.Palette.First)
		utils.CheckError(tracerr.Wrap(err))
		ThirdColorParsed, err := ParseHexColor(Config.Colors.Palette.Third)
		utils.CheckError(tracerr.Wrap(err))
		ParticleColorParsed, err := ParseHexColor(Config.Colors.Palette.Particle)
		utils.CheckError(tracerr.Wrap(err))

		AccentColor = color.RGBA{AccentColorParsed.R, AccentColorParsed.G, AccentColorParsed.B, Config.Line.Opacity}
		FirstColor = color.RGBA{FirstColorParsed.R, FirstColorParsed.G, FirstColorParsed.B, Config.Line.Opacity}
		ThirdColor = color.RGBA{ThirdColorParsed.R, ThirdColorParsed.G, ThirdColorParsed.B, Config.Line.Opacity}
		ParticleColor = color.RGBA{ParticleColorParsed.R, ParticleColorParsed.G, ParticleColorParsed.B, Config.Line.Opacity}
		ThirdColorAdj = color.RGBA{uint8(float64(ThirdColorParsed.R) * Config.Line.Brightness), uint8(float64(ThirdColorParsed.G) * Config.Line.Brightness), uint8(float64(ThirdColorParsed.B) * Config.Line.Brightness), Config.Line.Opacity}
	} else {
		fh, err := os.Open(walFile)
		utils.CheckError(tracerr.Wrap(err))
		defer fh.Close()
		scanner := bufio.NewScanner(fh)
		var line int
		var rgbaColor color.RGBA
		for scanner.Scan() {
			if line == 0 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(tracerr.Wrap(err))
				FirstColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.Line.Opacity}
			}
			if line == 1 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(tracerr.Wrap(err))
				AccentColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.Line.Opacity}
			}
			if line == 2 {
				rgbaColor, err = ParseHexColor(scanner.Text())
				utils.CheckError(tracerr.Wrap(err))
				ThirdColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.Line.Opacity}
				ThirdColorAdj = color.RGBA{uint8(float64(rgbaColor.R) * Config.Line.Brightness), uint8(float64(rgbaColor.G) * Config.Line.Brightness), uint8(float64(rgbaColor.B) * Config.Line.Brightness), Config.Line.Opacity}
				ParticleColor = color.RGBA{rgbaColor.R, rgbaColor.G, rgbaColor.B, Config.Line.Opacity}
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
