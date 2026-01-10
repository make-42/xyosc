package splash

import (
	"embed"
	"image"
	"time"

	"github.com/fogleman/ease"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/utils"
)

//go:embed assets/icon-48-trans.png
var f embed.FS

var splashLogo *ebiten.Image
var startTime time.Time
var splashStaticDuration time.Duration
var splashTransDuration time.Duration

func Init() {
	if config.Config.App.Splash.Enable {
		data, err := f.Open("assets/icon-48-trans.png")
		utils.CheckError(tracerr.Wrap(err))
		splashLogoUncasted, _, err := image.Decode(data)
		utils.CheckError(tracerr.Wrap(err))
		splashLogo = ebiten.NewImageFromImage(splashLogoUncasted)
		data.Close()
		startTime = time.Now()
		splashStaticDuration = time.Duration(1e9*config.Config.App.Splash.StaticSecs) * time.Nanosecond
		splashTransDuration = time.Duration(1e9*config.Config.App.Splash.TransSecs) * time.Nanosecond
	} else {
		SplashShowing = false
	}
}

var SplashShowing = true

func DrawSplash(screen *ebiten.Image) {
	timeSince := time.Since(startTime)
	if timeSince > (splashTransDuration + splashStaticDuration) {
		SplashShowing = false
		return
	}

	alpha := float32(1)
	if timeSince > (splashStaticDuration) {
		t := float64((timeSince - splashStaticDuration).Nanoseconds()) / float64(splashTransDuration.Nanoseconds())
		alpha = float32(1 - ease.InOutQuint(t))
	}
	pos := ebiten.GeoM{}
	sideLength := float64(min(config.Config.Window.Width, config.Config.Window.Height))
	scale := sideLength / float64(48)
	pos.Scale(scale, scale)
	pos.Translate((float64(config.Config.Window.Width)-sideLength)/2, (float64(config.Config.Window.Height)-sideLength)/2)
	colorScale := ebiten.ColorScale{}
	colorScale.ScaleAlpha(alpha)
	screen.DrawImage(splashLogo, &ebiten.DrawImageOptions{
		GeoM:       pos,
		ColorScale: colorScale,
	})
}
