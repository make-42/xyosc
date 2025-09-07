package fonts

import (
	"log"
	"os"
	"xyosc/config"
	"xyosc/utils"

	"embed"

	"github.com/go-text/typesetting/fontscan"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/SourceHanSansJP-Heavy.otf
var embedFS embed.FS

var Font *text.GoTextFaceSource

func Init() {
	foundFont := false
	if config.Config.UseSystemFonts {
		fontMap := fontscan.NewFontMap(log.Default())
		fontMap.UseSystemFonts("")
		loc, found := fontMap.FindSystemFont(config.Config.SystemFont)
		foundFont = found
		f, err := os.Open(loc.File)
		Font, err = text.NewGoTextFaceSource(f)
		utils.CheckError(err)
	}
	if !foundFont {
		// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
		f, err := embedFS.Open("assets/SourceHanSansJP-Heavy.otf")
		utils.CheckError(err)
		Font, err = text.NewGoTextFaceSource(f)
		utils.CheckError(err)
	}
}
