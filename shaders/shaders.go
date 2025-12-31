package shaders

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/utils"

	_ "embed"
)

//go:embed assets/glow.go
var glowCode []byte

//go:embed assets/chromaticabberation.go
var chromaticabberationCode []byte

//go:embed assets/chromaticabberation2.go
var chromaticabberation2Code []byte

//go:embed assets/noise.go
var noiseCode []byte

//go:embed assets/gammacorrection.go
var gammaCorrectionCode []byte

//go:embed assets/gammacorrectionalphafriendly.go
var gammaCorrectionAlphaFriendlyCode []byte

//go:embed assets/crt.go
var crtCode []byte

//go:embed assets/crtcurve.go
var crtCurveCode []byte

var Shaders map[string]*ebiten.Shader

type ShaderRenderStep struct {
	Shader    *ebiten.Shader
	Arguments map[string]any
	TimeScale float32
}

var ShaderRenderList = []ShaderRenderStep{}
var ModeLastShaderRenderListGenerated = -1
var PresetLastShaderRenderListGenerated = -1
var SelectedPreset = [4]int{0, 0, 0, 0}

func GenShaderRenderList() {
	if config.Config.DefaultMode != ModeLastShaderRenderListGenerated || SelectedPreset[config.Config.DefaultMode] != PresetLastShaderRenderListGenerated {
		ModeLastShaderRenderListGenerated = config.Config.DefaultMode
		PresetLastShaderRenderListGenerated = SelectedPreset[config.Config.DefaultMode]
		ShaderRenderList = []ShaderRenderStep{}
		SelectedPreset[config.Config.DefaultMode] = SelectedPreset[config.Config.DefaultMode] % len(config.Config.ModeShaders[config.Config.DefaultMode])
		for _, shader := range config.Config.Shaders[config.Config.ModeShaders[config.Config.DefaultMode][SelectedPreset[config.Config.DefaultMode]]] {
			ShaderRenderList = append(ShaderRenderList, ShaderRenderStep{
				Shader:    Shaders[shader.Name],
				Arguments: shader.Arguments,
				TimeScale: shader.TimeScale,
			})
		}
	}
}

func Init() {
	var err error
	Shaders = make(map[string]*ebiten.Shader)
	// Built-in shaders
	Shaders["glow"], err = ebiten.NewShader([]byte(glowCode))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["chromaticabberation"], err = ebiten.NewShader([]byte(chromaticabberationCode))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["chromaticabberation2"], err = ebiten.NewShader([]byte(chromaticabberation2Code))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["noise"], err = ebiten.NewShader([]byte(noiseCode))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["gammacorrection"], err = ebiten.NewShader([]byte(gammaCorrectionCode))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["gammacorrectionalphafriendly"], err = ebiten.NewShader([]byte(gammaCorrectionAlphaFriendlyCode))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["crt"], err = ebiten.NewShader([]byte(crtCode))
	utils.CheckError(tracerr.Wrap(err))
	Shaders["crtcurve"], err = ebiten.NewShader([]byte(crtCurveCode))
	utils.CheckError(tracerr.Wrap(err))

	// Custom shaders
	for shaderName, shaderCode := range config.Config.CustomShaderCode {
		Shaders["custom/"+shaderName], err = ebiten.NewShader([]byte(shaderCode))
		if err != nil {
			log.Println("Couldn't compile", shaderName, "shader.")
			log.Fatal(err)
		}
	}

}
