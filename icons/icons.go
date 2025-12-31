package icons

import (
	"embed"
	"image"

	"github.com/ztrue/tracerr"

	"xyosc/utils"

	_ "image/png"
)

//go:embed assets/icon-48.png
//go:embed assets/icon-16.png
//go:embed assets/icon-32.png
var fs embed.FS

var WindowIcon48 image.Image
var WindowIcon32 image.Image
var WindowIcon16 image.Image

func Init() {
	f48, err := fs.Open("assets/icon-48.png")
	utils.CheckError(tracerr.Wrap(err))
	defer f48.Close()
	WindowIcon48, _, err = image.Decode(f48)
	utils.CheckError(tracerr.Wrap(err))
	f32, err := fs.Open("assets/icon-32.png")
	utils.CheckError(tracerr.Wrap(err))
	defer f32.Close()
	WindowIcon32, _, err = image.Decode(f32)
	utils.CheckError(tracerr.Wrap(err))
	f16, err := fs.Open("assets/icon-16.png")
	utils.CheckError(tracerr.Wrap(err))
	defer f16.Close()
	WindowIcon16, _, err = image.Decode(f16)
	utils.CheckError(tracerr.Wrap(err))
}
