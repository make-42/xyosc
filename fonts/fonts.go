package fonts

import (
	"xyosc/utils"

	"embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/SourceHanSansJP-Heavy.otf
var embedFS embed.FS

var Font *text.GoTextFaceSource

func Init() {
	// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
	f, err := embedFS.Open("assets/SourceHanSansJP-Heavy.otf")
	utils.CheckError(err)
	Font, err = text.NewGoTextFaceSource(f)
	utils.CheckError(err)
}
