package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"log"
	"xyosc/audio"
	"xyosc/config"
	"xyosc/fastsqrt"
	"xyosc/fonts"
	"xyosc/icons"
	"xyosc/media"

	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	scale := min(config.Config.WindowWidth, config.Config.WindowHeight) / 2
	var AX float32
	var AY float32
	var BX float32
	var BY float32
	binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
	binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
	for i := uint32(0); i < config.Config.ReadBufferSize/audio.SampleSizeInBytes/2; i++ {
		binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BX)
		binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BY)
		fAX := float32(AX) * config.Config.Gain * float32(scale)
		fAY := -float32(AY) * config.Config.Gain * float32(scale)
		fBX := float32(BX) * config.Config.Gain * float32(scale)
		fBY := -float32(BY) * config.Config.Gain * float32(scale)
		if config.Config.LineInvSqrtOpacityControl {

			inv := fastsqrt.FastInvSqrt32((fBX-fAX)*(fBX-fAX) + (fBY-fBY)*(fBY-fBY))
			colorAdjusted := color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, uint8(float32(config.Config.LineOpacity) * inv)}
			vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowWidth/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowWidth/2)+fBY, config.Config.LineThickness, colorAdjusted, true)
		} else {
			vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowWidth/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowWidth/2)+fBY, config.Config.LineThickness, config.ThirdColorAdj, true)

		}
		AX = BX
		AY = BY
	}
	//audio.SampleRingBuffer.Reset()
	if config.Config.FPSCounter {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
	}

	if config.Config.ShowMPRIS {
		op := &text.DrawOptions{}
		op.GeoM.Translate(16, 16)
		op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRISTextOpacity})
		text.Draw(screen, media.PlayingMediaInfo.Artist+" - "+media.PlayingMediaInfo.Title, &text.GoTextFace{
			Source: fonts.Font,
			Size:   32,
		}, op)

		op = &text.DrawOptions{}
		op.GeoM.Translate(16, 64)
		op.ColorScale.ScaleWithColor(color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, config.Config.MPRISTextOpacity})

		text.Draw(screen, media.PlayingMediaInfo.Album, &text.GoTextFace{
			Source: fonts.Font,
			Size:   16,
		}, op)

		op = &text.DrawOptions{}
		op.GeoM.Translate(16, 80)
		op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRISTextOpacity})

		text.Draw(screen, media.FmtDuration(media.PlayingMediaInfo.Position)+" / "+media.FmtDuration(media.PlayingMediaInfo.Duration), &text.GoTextFace{
			Source: fonts.Font,
			Size:   32,
		}, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(config.Config.WindowWidth), int(config.Config.WindowHeight)
}

func main() {
	config.Init()
	audio.Init()
	fonts.Init()
	icons.Init()
	go audio.Start()
	go media.Start()
	ebiten.SetWindowIcon([]image.Image{icons.WindowIcon48, icons.WindowIcon32, icons.WindowIcon16})
	ebiten.SetWindowSize(int(config.Config.WindowWidth), int(config.Config.WindowHeight))
	ebiten.SetWindowTitle("xyosc")
	ebiten.SetWindowMousePassthrough(true)
	ebiten.SetTPS(int(config.Config.TargetFPS))
	ebiten.SetWindowDecorated(false)
	screenW, screenH := ebiten.Monitor().Size()
	ebiten.SetWindowPosition(screenW/2-int(config.Config.WindowWidth)/2, screenH/2-int(config.Config.WindowHeight)/2)
	ebiten.SetVsyncEnabled(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
