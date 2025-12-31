package fonts

import (
	"embed"
	"log"
	"os"

	"github.com/go-text/typesetting/fontscan"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/utils"
)

//go:embed assets/SourceHanSansJP-Heavy.otf
var embedFS embed.FS

var FontA *text.GoTextFaceSource
var FontB *text.GoTextFaceSource
var MPRISBigTextFace *text.MultiFace
var MPRISSmallTextFace *text.MultiFace

func Init() {
	foundFont := false
	if config.Config.UseSystemFonts {
		fontMap := fontscan.NewFontMap(log.Default())
		fontMap.UseSystemFonts("")
		loc, found := fontMap.FindSystemFont(config.Config.SystemFont)
		foundFont = found
		f, err := os.Open(loc.File)
		FontA, err = text.NewGoTextFaceSource(f)
		utils.CheckError(tracerr.Wrap(err))
	}
	// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
	f, err := embedFS.Open("assets/SourceHanSansJP-Heavy.otf")
	utils.CheckError(tracerr.Wrap(err))
	FontB, err = text.NewGoTextFaceSource(f)
	utils.CheckError(tracerr.Wrap(err))

	if foundFont {
		MPRISBigTextFace, err = text.NewMultiFace(&text.GoTextFace{
			Source: FontA,
			Size:   32,
		}, &text.GoTextFace{
			Source: FontB,
			Size:   32,
		})
		utils.CheckError(tracerr.Wrap(err))
		MPRISSmallTextFace, err = text.NewMultiFace(&text.GoTextFace{
			Source: FontA,
			Size:   16,
		}, &text.GoTextFace{
			Source: FontB,
			Size:   16,
		})
		utils.CheckError(tracerr.Wrap(err))
	} else {
		MPRISBigTextFace, err = text.NewMultiFace(&text.GoTextFace{
			Source: FontB,
			Size:   32,
		})
		utils.CheckError(tracerr.Wrap(err))
		MPRISSmallTextFace, err = text.NewMultiFace(&text.GoTextFace{
			Source: FontB,
			Size:   16,
		})
		utils.CheckError(tracerr.Wrap(err))
		FontA = FontB
	}

}
