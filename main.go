package main

import (
	"encoding/binary"
	"flag"
	"image"
	"image/color"
	"log"
	"math/rand/v2"
	"slices"
	"sort"
	"xyosc/audio"
	"xyosc/config"
	"xyosc/fastsqrt"
	"xyosc/filter"
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

var pressedKeys []ebiten.Key

func (g *Game) Update() error {
	pressedKeys = nil
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)
	return nil
}

var prevFrame *ebiten.Image
var firstFrame = true
var FFTBuffer []float64
var ComplexFFTBuffer []complex128
var FiltersApplied = false
var LowCutOffFrac = 0.0
var HighCutOffFrac = 1.0
var UseRightChannel *bool
var MixChannels *bool

var XYComplexFFTBufferL []complex128
var XYComplexFFTBufferR []complex128

func (g *Game) Draw(screen *ebiten.Image) {
	if config.Config.CopyPreviousFrame {
		if !firstFrame {
			op := &ebiten.DrawImageOptions{}
			op.ColorScale.ScaleAlpha(config.Config.CopyPreviousFrameAlpha)
			screen.DrawImage(prevFrame, op)
		}
	}
	scale := min(config.Config.WindowWidth, config.Config.WindowHeight) / 2
	var AX float32
	var AY float32
	var BX float32
	var BY float32
	var numSamples = config.Config.ReadBufferSize / audio.SampleSizeInBytes * 4

	if slices.Contains(pressedKeys, ebiten.KeyF) {
		config.SingleChannel = !config.SingleChannel
	}
	if !config.SingleChannel {
		if FiltersApplied {
			for i := uint32(0); i < numSamples; i++ {
				XYComplexFFTBufferL[i] = complex(float64(AX), 0.0)
				XYComplexFFTBufferR[i] = complex(float64(AY), 0.0)
				binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
				binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
			}
			filter.FilterBufferInPlace(&XYComplexFFTBufferL, LowCutOffFrac, HighCutOffFrac)
			filter.FilterBufferInPlace(&XYComplexFFTBufferR, LowCutOffFrac, HighCutOffFrac)
			AX = float32(real(XYComplexFFTBufferL[len(XYComplexFFTBufferL)-1]))
			AY = float32(real(XYComplexFFTBufferR[len(XYComplexFFTBufferR)-1]))
		} else {
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
		}
		S := float32(0)
		for i := uint32(0); i < numSamples; i++ {
			if FiltersApplied {
				BX = float32(real(XYComplexFFTBufferL[i]))
				BY = float32(real(XYComplexFFTBufferR[i]))
			} else {
				binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BX)
				binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &BY)
			}
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
			if FiltersApplied {
				if *MixChannels {
					ComplexFFTBuffer[i] = complex((float64(AY)+float64(AX))/2, 0.0)
				} else {
					if *UseRightChannel {
						ComplexFFTBuffer[i] = complex(float64(AY), 0.0)
					} else {
						ComplexFFTBuffer[i] = complex(float64(AX), 0.0)
					}
				}
			} else {
				if *MixChannels {
					FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = (float64(AY) + float64(AX)) / 2
				} else {
					if *UseRightChannel {
						FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = float64(AY)
					} else {
						FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = float64(AX)
					}
				}
			}
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AX)
			binary.Read(audio.SampleRingBuffer, binary.NativeEndian, &AY)
		}
		if FiltersApplied {
			filter.FilterBufferInPlace(&ComplexFFTBuffer, LowCutOffFrac, HighCutOffFrac)
			for i := uint32(0); i < numSamples; i++ {
				FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = real(ComplexFFTBuffer[i])
			}
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

	if config.Config.CopyPreviousFrame {
		prevFrame.Clear()
		prevFrame.DrawImage(screen, nil)
	}
	if firstFrame {
		firstFrame = false
		// f, _ := os.Create("image.png")
		// png.Encode(f, prevFrame)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(config.Config.WindowWidth), int(config.Config.WindowHeight)
}

func Init() {
	numSamples := config.Config.ReadBufferSize / audio.SampleSizeInBytes * 4
	FFTBuffer = make([]float64, numSamples)
	lowCutOff := flag.Float64("lo", 0.0, "low frequency cutoff fraction (discernable details are around 0.001 increments for a 4096 buffer size and 192kHz sample rate)")
	highCutOff := flag.Float64("hi", 1.0, "high frequency cutoff fraction (discernable details are around 0.001 increments for a 4096 buffer size and 192kHz sample rate)")
	UseRightChannel = flag.Bool("right", false, "Use the right channel instead of the left for the single axis oscilloscope")
	MixChannels = flag.Bool("mix", false, "Mix channels instead of just using a single channel for the single axis oscilloscope")

	overrideWidth := flag.Int("width", int(config.Config.WindowWidth), "override window width")
	overrideHeight := flag.Int("height", int(config.Config.WindowHeight), "override window height")

	flag.Parse()
	if *lowCutOff != 0.0 {
		FiltersApplied = true
		LowCutOffFrac = *lowCutOff
	}
	if *highCutOff != 1.0 {
		FiltersApplied = true
		HighCutOffFrac = *highCutOff
	}
	if FiltersApplied {
		ComplexFFTBuffer = make([]complex128, numSamples)
		XYComplexFFTBufferL = make([]complex128, numSamples)
		XYComplexFFTBufferR = make([]complex128, numSamples)
	}
	config.Config.WindowWidth = int32(*overrideWidth)
	config.Config.WindowHeight = int32(*overrideHeight)

}

func main() {
	config.Init()
	audio.Init()
	Init()
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
	prevFrame = ebiten.NewImage(int(config.Config.WindowWidth), int(config.Config.WindowHeight))
	if err := ebiten.RunGameWithOptions(&Game{}, &ebiten.RunGameOptions{ScreenTransparent: true}); err != nil {
		log.Fatal(err)
	}
}
