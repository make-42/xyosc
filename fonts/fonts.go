package fonts

import (
	"os"
	"xyosc/utils"

	"github.com/flopp/go-findfont"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var Font *text.GoTextFaceSource

func Init() {
	// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
	fontPath, err := findfont.Find("SourceHanSansJP-Heavy.otf")
	utils.CheckError(err)
	f, err := os.Open(fontPath)
	utils.CheckError(err)
	Font, err = text.NewGoTextFaceSource(f)
	utils.CheckError(err)
}
