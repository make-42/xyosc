package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"log"
	"math/rand/v2"
	"sort"
	"xyosc/audio"
	"xyosc/config"
	"xyosc/fastsqrt"
	"xyosc/fonts"
	"xyosc/icons"
	"xyosc/media"
	"xyosc/particles"

	"fmt"

	"github.com/chewxy/math32"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/goccmack/godsp/peaks"
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
	var numSamples = config.Config.ReadBufferSize / audio.SampleSizeInBytes * 4
	var FFTBuffer = make([]float64, numSamples)
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		config.SingleChannel = !config.SingleChannel
	}
	if !config.SingleChannel {
		binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
		binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
		S := float32(0)
		for i := uint32(0); i < numSamples; i++ {
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BX)
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BY)
			fAX := float32(AX) * config.Config.Gain * float32(scale)
			fAY := -float32(AY) * config.Config.Gain * float32(scale)
			fBX := float32(BX) * config.Config.Gain * float32(scale)
			fBY := -float32(BY) * config.Config.Gain * float32(scale)
			if config.Config.LineInvSqrtOpacityControl {
				inv := fastsqrt.FastInvSqrt32((fBX-fAX)*(fBX-fAX) + (fBY-fBY)*(fBY-fBY))
				colorAdjusted := color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, uint8(float32(config.Config.LineOpacity) * inv)}
				vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowHeight/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowHeight/2)+fBY, config.Config.LineThickness, colorAdjusted, true)
			} else {
				vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowHeight/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowHeight/2)+fBY, config.Config.LineThickness, config.ThirdColorAdj, true)
			}
			S += float32(AX)*float32(AX) + float32(AY)*float32(AY)
			if config.Config.Particles {
				if rand.IntN(config.Config.ParticleGenPerFrameEveryXSamples) == 0 {
					if len(particles.Particles) >= config.Config.ParticleMaxCount {
						particles.Particles = particles.Particles[1:]
					}
					particles.Particles = append(particles.Particles, particles.Particle{
						X:    float32(AX) * config.Config.Gain,
						Y:    -float32(AY) * config.Config.Gain,
						VX:   0,
						VY:   0,
						Size: rand.Float32()*(config.Config.ParticleMaxSize-config.Config.ParticleMinSize) + config.Config.ParticleMinSize,
					})
				}
			}

			AX = BX
			AY = BY
		}

		for i, particle := range particles.Particles {
			vector.DrawFilledCircle(screen, float32(config.Config.WindowWidth/2)+particle.X*float32(scale), float32(config.Config.WindowHeight/2)+particle.Y*float32(scale), particle.Size, config.ThirdColor, true)
			norm := math32.Sqrt(particles.Particles[i].X*particles.Particles[i].X + particles.Particles[i].Y*particles.Particles[i].Y)
			particles.Particles[i].X += particle.VX / float32(ebiten.ActualTPS())
			particles.Particles[i].Y += particle.VY / float32(ebiten.ActualTPS())
			speed := math32.Sqrt(particle.VX*particle.VX + particle.VY*particle.VY)
			particles.Particles[i].VX += (config.Config.ParticleAcceleration*S - speed*config.Config.ParticleDrag) * particle.X / norm / float32(ebiten.ActualTPS())
			particles.Particles[i].VY += (config.Config.ParticleAcceleration*S - speed*config.Config.ParticleDrag) * particle.Y / norm / float32(ebiten.ActualTPS())
		}
	} else {
		for i := uint32(0); i < numSamples; i++ {
			FFTBuffer[i] = float64(AX)
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
		}

		indices := peaks.Get(FFTBuffer, config.Config.PeakDetectSeparator)
		sort.Ints(indices)
		offset := uint32(0)
		if len(indices) != 0 {
			offset = uint32(indices[0])
		}
		if config.Config.PeriodCrop && len(indices) > 1 {
			lastPeriodOffset := uint32(indices[min(len(indices)-1, config.Config.PeriodCropCount)])
			samplesPerCrop := lastPeriodOffset - offset
			for i := uint32(0); i < min(numSamples, samplesPerCrop*config.Config.PeriodCropLoopOverCount)-1; i++ {
				fAX := float32(FFTBuffer[(i+offset)%numSamples]) * config.Config.Gain * float32(scale)
				fBX := float32(FFTBuffer[(i+1+offset)%numSamples]) * config.Config.Gain * float32(scale)
				vector.StrokeLine(screen, float32(config.Config.WindowWidth)*float32(i%samplesPerCrop)/float32(samplesPerCrop), float32(config.Config.WindowHeight/2)+fAX, float32(config.Config.WindowWidth)*float32(i%samplesPerCrop+1)/float32(samplesPerCrop), float32(config.Config.WindowHeight/2)+fBX, config.Config.LineThickness, config.ThirdColorAdj, true)
			}
		} else {
			for i := uint32(0); i < numSamples-1; i++ {
				fAX := float32(FFTBuffer[(i+offset)%numSamples]) * config.Config.Gain * float32(scale)
				fBX := float32(FFTBuffer[(i+1+offset)%numSamples]) * config.Config.Gain * float32(scale)
				vector.StrokeLine(screen, float32(config.Config.WindowWidth)*float32(i%config.Config.SingleChannelWindow)/float32(config.Config.SingleChannelWindow), float32(config.Config.WindowHeight/2)+fAX, float32(config.Config.WindowWidth)*float32(i%config.Config.SingleChannelWindow+1)/float32(config.Config.SingleChannelWindow), float32(config.Config.WindowHeight/2)+fBX, config.Config.LineThickness, config.ThirdColorAdj, true)
			}
		}
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
	if err := ebiten.RunGameWithOptions(&Game{}, &ebiten.RunGameOptions{ScreenTransparent: true}); err != nil {
		log.Fatal(err)
	}
}
