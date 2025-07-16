package shaders

import (
	_ "embed"
	"log"
	"xyosc/config"
	"xyosc/utils"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/glow.go
var glowCode []byte

//go:embed assets/chromaticabberation.go
var chromaticabberationCode []byte

//go:embed assets/chromaticabberation2.go
var chromaticabberation2Code []byte

//go:embed assets/noise.go
var noiseCode []byte

var Shaders map[string]*ebiten.Shader

type ShaderRenderStep struct {
	Shader    *ebiten.Shader
	Arguments map[string]any
	TimeScale float32
}

var ShaderRenderList = []ShaderRenderStep{}

func Init() {
	var err error
	Shaders = make(map[string]*ebiten.Shader)
	// Built-in shaders
	Shaders["glow"], err = ebiten.NewShader([]byte(glowCode))
	utils.CheckError(err)
	Shaders["chromaticabberation"], err = ebiten.NewShader([]byte(chromaticabberationCode))
	utils.CheckError(err)
	Shaders["chromaticabberation2"], err = ebiten.NewShader([]byte(chromaticabberation2Code))
	utils.CheckError(err)
	Shaders["noise"], err = ebiten.NewShader([]byte(noiseCode))
	utils.CheckError(err)

	// Custom shaders
	for shaderName, shaderCode := range config.Config.CustomShaderCode {
		Shaders["custom/"+shaderName], err = ebiten.NewShader([]byte(shaderCode))
		if err != nil {
			log.Println("Couldn't compile", shaderName, "shader.")
			log.Fatal(err)
		}
	}
	for _, shader := range config.Config.Shaders {
		ShaderRenderList = append(ShaderRenderList, ShaderRenderStep{
			Shader:    Shaders[shader.Name],
			Arguments: shader.Arguments,
			TimeScale: shader.TimeScale,
		})
	}
}
