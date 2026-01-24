package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"time"

	"github.com/alltom/oklab"
	"github.com/chewxy/math32"
	"github.com/goccmack/godsp/peaks"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"xyosc/align"
	"xyosc/audio"
	"xyosc/bars"
	"xyosc/beatdetect"
	"xyosc/config"
	"xyosc/fastsqrt"
	"xyosc/filter"
	"xyosc/fonts"
	"xyosc/icons"
	"xyosc/interpolate"
	"xyosc/kaiser"
	"xyosc/media"
	"xyosc/particles"
	"xyosc/shaders"
	"xyosc/slew"
	"xyosc/splash"
	"xyosc/utils"
	"xyosc/vu"

	_ "github.com/silbinarywolf/preferdiscretegpu"
)

type Game struct {
}

var pressedKeys []ebiten.Key

func (g *Game) Update() error {
	pressedKeys = nil
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)
	return nil
}

var readHeadPosition uint32 = 0
var startTime time.Time
var prevFrame *ebiten.Image
var shaderWorkBuffer *ebiten.Image
var barCursorImageBGRectFrame *ebiten.Image
var firstFrame = true
var FFTBuffer []float64
var complexFFTBuffer []complex128
var complexFFTBufferFlipped []complex128
var lowCutOffFrac = 0.0
var highCutOffFrac = 1.0
var useRightChannel *bool
var mixChannels *bool
var beatDetectOverride *bool
var peakDetectOverride *bool
var overrideX *int
var overrideY *int

var XYComplexFFTBufferL []complex128
var XYComplexFFTBufferR []complex128

var StillSamePressFromModeToggleKey bool
var StillSamePressFromPresetToggleKey bool

var barsLastFrameTime time.Time
var beatTimeLastFrameTime time.Time
var lastFrameTime time.Time

func copyPrevFrameOp(deltaTime float64, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(float32(math.Pow(float64(config.Config.ImageRetention.AlphaDecayBase), deltaTime*config.Config.ImageRetention.AlphaDecaySpeed)))
	screen.DrawImage(prevFrame, op)
}

func (g *Game) Draw(screen *ebiten.Image) {
	deltaTime := min(time.Since(lastFrameTime).Seconds(), 1.0)
	lastFrameTime = time.Now()
	var numSamples = config.Config.Buffers.ReadBufferSize / 2
	if config.Config.ImageRetention.Enable {
		if !firstFrame && !(config.Config.App.DefaultMode == config.BarsMode && config.Config.Bars.PeakCursor.Enable) { // leave to post background draw
			copyPrevFrameOp(deltaTime, screen)
		}
	} else {
		if config.Config.Colors.DisableBGTransparency {
			screen.Fill(config.BGColor)
		}
	}
	if config.Config.MPRIS.Enable {
		media.Interpolate()
	}
	scale := min(config.Config.Window.Width, config.Config.Window.Height) / 2
	var AX float32
	var AY float32
	var BX float32
	var BY float32
	posStartRead := (config.Config.Buffers.RingBufferSize + audio.WriteHeadPosition - numSamples*2 - config.Config.Buffers.ReadBufferDelay) % config.Config.Buffers.RingBufferSize
	if slices.Contains(pressedKeys, ebiten.KeyF) {
		if !StillSamePressFromModeToggleKey {
			StillSamePressFromModeToggleKey = true
			config.Config.App.DefaultMode = (config.Config.App.DefaultMode + 1) % 4
		}
	} else {
		StillSamePressFromModeToggleKey = false
	}
	if slices.Contains(pressedKeys, ebiten.KeyP) {
		if !StillSamePressFromPresetToggleKey {
			StillSamePressFromPresetToggleKey = true
			shaders.SelectedPreset[config.Config.App.DefaultMode] += 1
		}
	} else {
		StillSamePressFromPresetToggleKey = false
	}

	if (config.Config.App.DefaultMode == config.XYMode || config.Config.App.DefaultMode == config.SingleChannelMode) && config.Config.Scale.Enable {
		if config.Config.Scale.MainAxisEnable {
			vector.StrokeLine(screen, 0, float32(config.Config.Window.Height/2), float32(config.Config.Window.Width), float32(config.Config.Window.Height/2), config.Config.Scale.MainAxisThickness, config.ThirdColorAdj, true)
			vector.StrokeLine(screen, float32(config.Config.Window.Width/2), 0, float32(config.Config.Window.Width/2), float32(config.Config.Window.Height), config.Config.Scale.MainAxisThickness, config.ThirdColorAdj, true)
		}
		if config.Config.Scale.Vert.TickEnable {
			for i := range config.Config.Scale.Vert.Divs + 1 {
				y := float32(config.Config.Window.Height) - float32(config.Config.Window.Height)/float32(config.Config.Scale.Vert.Divs)*float32(i)
				if config.Config.Scale.Vert.TickToGrid {
					vector.StrokeLine(screen, 0, y, float32(config.Config.Window.Width), y, config.Config.Scale.Vert.GridThickness, config.ThirdColorAdj, true)
				}
				vector.StrokeLine(screen, float32(config.Config.Window.Width/2)-config.Config.Scale.Vert.TickLength/2, y, float32(config.Config.Window.Width/2)+config.Config.Scale.Vert.TickLength/2, y, config.Config.Scale.Vert.TickThickness, config.ThirdColorAdj, true)
				if config.Config.Scale.Vert.TextEnable {
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64(config.Config.Window.Width)/2+float64(config.Config.Scale.Vert.TickLength/2)+config.Config.Scale.Vert.TextPadding, float64(y))
					op.LayoutOptions.PrimaryAlign = text.AlignStart
					op.LayoutOptions.SecondaryAlign = text.AlignCenter
					op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.Scale.TextOpacity})
					text.Draw(screen, fmt.Sprintf("%.*f", int(math.Ceil(math.Log10(float64(config.Config.Scale.Vert.Divs)/2))), 2/float32(config.Config.Scale.Vert.Divs)*float32(i)-1), &text.GoTextFace{
						Source: fonts.FontA,
						Size:   config.Config.Scale.Vert.TextSize,
					}, op)
				}
			}
		}
	}

	if (config.Config.App.DefaultMode == config.XYMode) && config.Config.Scale.Enable {
		if config.Config.Scale.MainAxisEnable {
			vector.StrokeLine(screen, 0, float32(config.Config.Window.Height/2), float32(config.Config.Window.Width), float32(config.Config.Window.Height/2), config.Config.Scale.MainAxisThickness, config.ThirdColorAdj, true)
			vector.StrokeLine(screen, float32(config.Config.Window.Width/2), 0, float32(config.Config.Window.Width/2), float32(config.Config.Window.Height), config.Config.Scale.MainAxisThickness, config.ThirdColorAdj, true)
		}
		if config.Config.Scale.Horz.TickEnable {
			for i := range config.Config.Scale.Horz.Divs + 1 {
				x := float32(config.Config.Window.Width) / float32(config.Config.Scale.Horz.Divs) * float32(i)
				if config.Config.Scale.Horz.TickToGrid {
					vector.StrokeLine(screen, x, 0, x, float32(config.Config.Window.Height), config.Config.Scale.Horz.GridThickness, config.ThirdColorAdj, true)
				}
				vector.StrokeLine(screen, x, float32(config.Config.Window.Height/2)-config.Config.Scale.Horz.TickLength/2, x, float32(config.Config.Window.Height/2)+config.Config.Scale.Horz.TickLength/2, config.Config.Scale.Horz.TickThickness, config.ThirdColorAdj, true)
				if config.Config.Scale.Horz.TextEnable {
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64(x), float64(config.Config.Window.Height)/2+float64(config.Config.Scale.Horz.TickLength/2)+config.Config.Scale.Horz.TextPadding)
					op.LayoutOptions.PrimaryAlign = text.AlignCenter
					op.LayoutOptions.SecondaryAlign = text.AlignStart
					op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.Scale.TextOpacity})
					text.Draw(screen, fmt.Sprintf("%.*f", int(math.Ceil(math.Log10(float64(config.Config.Scale.Horz.Divs)/2))), 2/float32(config.Config.Scale.Horz.Divs)*float32(i)-1), &text.GoTextFace{
						Source: fonts.FontA,
						Size:   config.Config.Scale.Horz.TextSize,
					}, op)
				}
			}
		}
	}

	if config.Config.App.DefaultMode == config.XYMode {
		if config.FiltersApplied {
			for i := uint32(0); i < numSamples; i++ {
				AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.Buffers.RingBufferSize]
				AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.Buffers.RingBufferSize]
				XYComplexFFTBufferL[i] = complex(float64(AX), 0.0)
				XYComplexFFTBufferR[i] = complex(float64(AY), 0.0)
			}
			filter.FilterBufferInPlace(&XYComplexFFTBufferL, lowCutOffFrac, highCutOffFrac)
			filter.FilterBufferInPlace(&XYComplexFFTBufferR, lowCutOffFrac, highCutOffFrac)
			AX = float32(real(XYComplexFFTBufferL[len(XYComplexFFTBufferL)-1]))
			AY = float32(real(XYComplexFFTBufferR[len(XYComplexFFTBufferR)-1]))
		} else {
			AX = audio.SampleRingBufferUnsafe[(posStartRead)%config.Config.Buffers.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+1)%config.Config.Buffers.RingBufferSize]
		}
		S := float32(0)
		for i := uint32(0); i < numSamples; i++ {
			if config.FiltersApplied {
				BX = float32(real(XYComplexFFTBufferL[i]))
				BY = float32(real(XYComplexFFTBufferR[i]))
			} else {
				BX = audio.SampleRingBufferUnsafe[(posStartRead+i*2+2)%config.Config.Buffers.RingBufferSize]
				BY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+3)%config.Config.Buffers.RingBufferSize]
			}
			fAX := float32(AX) * config.Config.Audio.Gain * float32(scale)
			fAY := -float32(AY) * config.Config.Audio.Gain * float32(scale)
			fBX := float32(BX) * config.Config.Audio.Gain * float32(scale)
			fBY := -float32(BY) * config.Config.Audio.Gain * float32(scale)
			if i >= numSamples-config.Config.Buffers.XYOscilloscopeReadBufferSize {
				if config.Config.Line.InvSqrtOpacityControl.Enable || config.Config.Line.TimeDependentOpacityControl.Enable {
					mult := 1.0
					if config.Config.Line.InvSqrtOpacityControl.Enable {
						invStrength := min(config.Config.Line.InvSqrtOpacityControl.LinSens*max(float64(fastsqrt.FastInvSqrt32((fBX-fAX)*(fBX-fAX)+(fBY-fBY)*(fBY-fBY))), config.Config.Line.InvSqrtOpacityControl.LinCutoffFrac)-config.Config.Line.InvSqrtOpacityControl.LinCutoffFrac, 1)
						if config.Config.Line.InvSqrtOpacityControl.UseLogDecrement {
							mult *= config.Config.Line.InvSqrtOpacityControl.LogOffset + math.Log(invStrength)/math.Log(config.Config.Line.InvSqrtOpacityControl.LogBase)
						} else {
							mult *= invStrength
						}
					}
					if config.Config.Line.TimeDependentOpacityControl.Enable {
						mult *= math.Pow(config.Config.Line.TimeDependentOpacityControl.Base, float64(numSamples-i-1))
					}
					colorAdjusted := color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, uint8(float64(config.Config.Line.Opacity) * mult)}
					if config.Config.Line.OpacityAlsoAffectsThickness {
						vector.StrokeLine(screen, float32(config.Config.Window.Width/2)+fAX, float32(config.Config.Window.Height/2)+fAY, float32(config.Config.Window.Width/2)+fBX, float32(config.Config.Window.Height/2)+fBY, config.Config.Line.ThicknessXY*float32(mult), colorAdjusted, true)
					} else {
						vector.StrokeLine(screen, float32(config.Config.Window.Width/2)+fAX, float32(config.Config.Window.Height/2)+fAY, float32(config.Config.Window.Width/2)+fBX, float32(config.Config.Window.Height/2)+fBY, config.Config.Line.ThicknessXY, colorAdjusted, true)
					}
				} else {
					vector.StrokeLine(screen, float32(config.Config.Window.Width/2)+fAX, float32(config.Config.Window.Height/2)+fAY, float32(config.Config.Window.Width/2)+fBX, float32(config.Config.Window.Height/2)+fBY, config.Config.Line.ThicknessXY, config.ThirdColorAdj, true)
				}
			}
			S += float32(AX)*float32(AX) + float32(AY)*float32(AY)
			if config.Config.Particles.Enable {
				if rand.IntN(int(float64(config.Config.Particles.GenEveryXSamples)/deltaTime)) == 0 {
					if len(particles.Particles) >= config.Config.Particles.MaxCount {
						particles.Particles = particles.Particles[1:]
					}
					particles.Particles = append(particles.Particles, particles.Particle{
						X:    float32(AX) * config.Config.Audio.Gain,
						Y:    -float32(AY) * config.Config.Audio.Gain,
						VX:   0,
						VY:   0,
						Size: rand.Float32()*(config.Config.Particles.MaxSize-config.Config.Particles.MinSize) + config.Config.Particles.MinSize,
					})
				}
			}

			AX = BX
			AY = BY
		}

		for i, particle := range particles.Particles {
			vector.DrawFilledCircle(screen, float32(config.Config.Window.Width/2)+particle.X*float32(scale), float32(config.Config.Window.Height/2)+particle.Y*float32(scale), particle.Size, config.ParticleColor, true)
			norm := math32.Sqrt(particles.Particles[i].X*particles.Particles[i].X + particles.Particles[i].Y*particles.Particles[i].Y)
			particles.Particles[i].X += particle.VX * float32(deltaTime)
			particles.Particles[i].Y += particle.VY * float32(deltaTime)
			speed := math32.Sqrt(particle.VX*particle.VX + particle.VY*particle.VY)
			particles.Particles[i].VX += (config.Config.Particles.Acceleration*S - speed*config.Config.Particles.Drag) * particle.X / norm * float32(deltaTime)
			particles.Particles[i].VY += (config.Config.Particles.Acceleration*S - speed*config.Config.Particles.Drag) * particle.Y / norm * float32(deltaTime)
		}
	} else if config.Config.App.DefaultMode == config.SingleChannelMode {
		for i := uint32(0); i < numSamples; i++ {
			AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.Buffers.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.Buffers.RingBufferSize]
			if config.FiltersApplied || config.Config.SingleChannelOsc.PeakDetect.UseACF {
				if *mixChannels {
					complexFFTBuffer[i] = complex((float64(AY)+float64(AX))/2, 0.0)
				} else {
					if *useRightChannel {
						complexFFTBuffer[i] = complex(float64(AY), 0.0)
					} else {
						complexFFTBuffer[i] = complex(float64(AX), 0.0)
					}
				}
			}
			if !config.FiltersApplied || config.Config.SingleChannelOsc.PeakDetect.UseACF {
				if *mixChannels {
					FFTBuffer[(i+config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples] = (float64(AY) + float64(AX)) / 2
				} else {
					if *useRightChannel {
						FFTBuffer[(i+config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples] = float64(AY)
					} else {
						FFTBuffer[(i+config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples] = float64(AX)
					}
				}
			}
		}
		if config.FiltersApplied {
			filter.FilterBufferInPlace(&complexFFTBuffer, lowCutOffFrac, highCutOffFrac)
			for i := uint32(0); i < numSamples; i++ {
				FFTBuffer[(i+config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples] = real(complexFFTBuffer[i])
			}
		}

		indices := []int{}
		offset := uint32(0)
		freq := uint32(0)

		if config.Config.SingleChannelOsc.PeriodCrop.Enable || config.Config.SingleChannelOsc.PeakDetect.Enable || *peakDetectOverride {
			if config.Config.SingleChannelOsc.PeakDetect.UseACF {
				for i := range len(complexFFTBuffer) {
					complexFFTBufferFlipped[len(complexFFTBufferFlipped)-i-1] = complexFFTBuffer[i]
				}
				offset, indices, freq = align.AutoCorrelate(&complexFFTBuffer, &complexFFTBufferFlipped)

			} else {
				indices = peaks.Get(FFTBuffer, config.Config.SingleChannelOsc.PeakDetect.PeakDetectSeparator)
				sort.Ints(indices)
				offset = uint32(0)
				if len(indices) != 0 {
					offset = uint32(indices[0])
				}
				if (offset+config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples < config.Config.SingleChannelOsc.PeakDetect.EdgeGuardBufferSize || (numSamples-((offset+config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples)) < config.Config.SingleChannelOsc.PeakDetect.EdgeGuardBufferSize {
					if len(indices) >= 2 {
						offset = uint32(indices[1])
					}
				}
			}
		}

		var samplesPerCrop uint32
		var samplesPerPeriod uint32

		if (config.Config.SingleChannelOsc.PeakDetect.Enable && config.Config.SingleChannelOsc.SmoothWave.Enable) && len(indices) > 1 {
			lastPeriodOffset := uint32(indices[min(len(indices)-1, config.Config.SingleChannelOsc.PeriodCrop.DisplayCount)])
			samplesPerPeriod = lastPeriodOffset - offset
			if config.Config.SingleChannelOsc.PeakDetect.UseACF {
				samplesPerPeriod = freq * 2
			}
		} else {
			samplesPerPeriod = numSamples
		}

		if config.Config.SingleChannelOsc.PeriodCrop.Enable && len(indices) > 1 {
			lastPeriodOffset := uint32(indices[min(len(indices)-1, config.Config.SingleChannelOsc.PeriodCrop.DisplayCount)])
			samplesPerCrop = lastPeriodOffset - offset
			if config.Config.SingleChannelOsc.PeakDetect.UseACF {
				samplesPerCrop = freq * 2
			}
		} else {
			samplesPerCrop = numSamples
		}
		if config.Config.Scale.Enable {
			visibleSampleCount := min(numSamples, samplesPerCrop*config.Config.SingleChannelOsc.PeriodCrop.LoopOverCount)
			timeSpanVisible := float64(visibleSampleCount) / float64(2*config.Config.Audio.SampleRate) //s
			if config.Config.Scale.MainAxisEnable {
				vector.StrokeLine(screen, 0, float32(config.Config.Window.Height/2), float32(config.Config.Window.Width), float32(config.Config.Window.Height/2), config.Config.Scale.MainAxisThickness, config.ThirdColorAdj, true)
				vector.StrokeLine(screen, float32(config.Config.Window.Width/2), 0, float32(config.Config.Window.Width/2), float32(config.Config.Window.Height), config.Config.Scale.MainAxisThickness, config.ThirdColorAdj, true)
			}
			scaleScale := 1.
			if config.Config.Scale.HorzDivDynamicPos {
				scaleScale = math.Pow10(int(math.Ceil(math.Log10(timeSpanVisible)))) / timeSpanVisible
			}
			if config.Config.Scale.Horz.TickEnable {
				for i := range config.Config.Scale.Horz.Divs + 1 {
					x := float32(config.Config.Window.Width)/2 + float32(scaleScale)*float32(config.Config.Window.Width)/float32(config.Config.Scale.Horz.Divs)*(float32(i)-float32(config.Config.Scale.Horz.Divs)/2)
					if config.Config.Scale.Horz.TickToGrid {
						vector.StrokeLine(screen, x, 0, x, float32(config.Config.Window.Height), config.Config.Scale.Horz.GridThickness, config.ThirdColorAdj, true)
					}
					vector.StrokeLine(screen, x, float32(config.Config.Window.Height/2)-config.Config.Scale.Horz.TickLength/2, x, float32(config.Config.Window.Height/2)+config.Config.Scale.Horz.TickLength/2, config.Config.Scale.Horz.TickThickness, config.ThirdColorAdj, true)
					if config.Config.Scale.Horz.TextEnable {
						op := &text.DrawOptions{}
						op.GeoM.Translate(float64(x), float64(config.Config.Window.Height)/2+float64(config.Config.Scale.Horz.TickLength/2)+config.Config.Scale.Horz.TextPadding)
						op.LayoutOptions.PrimaryAlign = text.AlignCenter
						op.LayoutOptions.SecondaryAlign = text.AlignStart
						op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.Scale.TextOpacity})
						text.Draw(screen, fmt.Sprintf("%s", utils.FormatDuration(scaleScale*float64(timeSpanVisible/2)*(2/float64(config.Config.Scale.Horz.Divs)*float64(i)-1))), &text.GoTextFace{
							Source: fonts.FontA,
							Size:   config.Config.Scale.Horz.TextSize,
						}, op)
					}
				}
			}
		}
		if config.Config.SingleChannelOsc.PeriodCrop.Enable && len(indices) > 1 {
			if config.Config.SingleChannelOsc.PeakDetect.CenterPeak {
				offset -= samplesPerCrop / 2
			}
			its := min(numSamples, samplesPerCrop*config.Config.SingleChannelOsc.PeriodCrop.LoopOverCount) - 1
			for i := uint32(0); i < its; i++ {
				fAX := float32(FFTBuffer[(i+offset)%numSamples]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
				fBX := float32(FFTBuffer[(i+1+offset)%numSamples]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
				if config.Config.SingleChannelOsc.SmoothWave.Enable {
					smoothPeriods := min((numSamples/its)-1, config.Config.SingleChannelOsc.SmoothWave.MaxPeriods)
					if smoothPeriods > 0 {
						var q float64
						if config.Config.SingleChannelOsc.SmoothWave.TimeIndependent {
							q = config.Config.SingleChannelOsc.SmoothWave.TimeIndependentFactor
						} else {
							q = math.Exp(-float64(its) / float64(config.Config.Audio.SampleRate) * config.Config.SingleChannelOsc.SmoothWave.InvTau)
						}
						rescale := float32(1.)
						for k := range smoothPeriods {
							fact := float32(math.Pow(q, float64(k+1)))
							fAX += fact * float32(FFTBuffer[utils.Moduint32((i+offset-its*(k+1)), numSamples)]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
							fBX += fact * float32(FFTBuffer[utils.Moduint32((i+1+offset-its*(k+1)), numSamples)]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
							rescale += fact
						}
						fAX /= rescale
						fBX /= rescale
					}
				}
				if (i+1+offset-config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples != 0 {
					vector.StrokeLine(screen, float32(config.Config.Window.Width)*float32(i%samplesPerCrop)/float32(samplesPerCrop), float32(config.Config.Window.Height/2)-fAX, float32(config.Config.Window.Width)*float32(i%samplesPerCrop+1)/float32(samplesPerCrop), float32(config.Config.Window.Height/2)-fBX, config.Config.Line.ThicknessSingleChannel, config.ThirdColorAdj, true)
				}
			}
		} else {
			if config.Config.SingleChannelOsc.PeakDetect.CenterPeak {
				offset -= config.Config.SingleChannelOsc.DisplayBufferSize / 4
			}
			for i := uint32(0); i < config.Config.SingleChannelOsc.DisplayBufferSize/2-1; i++ {
				fAX := float32(FFTBuffer[(i+offset)%numSamples]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
				fBX := float32(FFTBuffer[(i+1+offset)%numSamples]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
				if config.Config.SingleChannelOsc.PeakDetect.Enable || *peakDetectOverride {
					its := samplesPerPeriod
					if config.Config.SingleChannelOsc.SmoothWave.Enable {
						smoothPeriods := min((numSamples/its)-1, config.Config.SingleChannelOsc.SmoothWave.MaxPeriods)
						if smoothPeriods > 0 {
							var q float64
							if config.Config.SingleChannelOsc.SmoothWave.TimeIndependent {
								q = config.Config.SingleChannelOsc.SmoothWave.TimeIndependentFactor
							} else {
								q = math.Exp(-float64(its) / float64(config.Config.Audio.SampleRate) * config.Config.SingleChannelOsc.SmoothWave.InvTau)
							}
							rescale := float32(1.)
							for k := range smoothPeriods {
								fact := float32(math.Pow(q, float64(k+1)))
								fAX += fact * float32(FFTBuffer[utils.Moduint32((i+offset-its*(k+1)), numSamples)]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
								fBX += fact * float32(FFTBuffer[utils.Moduint32((i+1+offset-its*(k+1)), numSamples)]) * config.Config.Audio.Gain * float32(config.Config.Window.Height) / 2
								rescale += fact
							}
							fAX /= rescale
							fBX /= rescale
						}
					}
				}
				if config.Config.SingleChannelOsc.Slew.Enable {
					if i == 0 {
						interpolate.Interpolate(deltaTime, float64(fAX), &slew.InterpolationPosBuffer[i], &slew.InterpolationVelBuffer[i], config.Config.SingleChannelOsc.Slew)
					}
					interpolate.Interpolate(deltaTime, float64(fBX), &slew.InterpolationPosBuffer[i+1], &slew.InterpolationVelBuffer[i+1], config.Config.SingleChannelOsc.Slew)
				}
				if (i+1+offset-config.Config.SingleChannelOsc.PeakDetect.FFTBufferOffset)%numSamples != 0 {
					if config.Config.SingleChannelOsc.Slew.Enable {
						vector.StrokeLine(screen, float32(config.Config.Window.Width)*float32(i)/float32((config.Config.SingleChannelOsc.DisplayBufferSize/2)), float32(config.Config.Window.Height/2)-float32(slew.InterpolationPosBuffer[i]), float32(config.Config.Window.Width)*float32(i+1)/float32(config.Config.SingleChannelOsc.DisplayBufferSize/2), float32(config.Config.Window.Height/2)-float32(slew.InterpolationPosBuffer[i+1]), config.Config.Line.ThicknessSingleChannel, config.ThirdColorAdj, true)
					} else {
						vector.StrokeLine(screen, float32(config.Config.Window.Width)*float32(i)/float32((config.Config.SingleChannelOsc.DisplayBufferSize/2)), float32(config.Config.Window.Height/2)-fAX, float32(config.Config.Window.Width)*float32(i+1)/float32(config.Config.SingleChannelOsc.DisplayBufferSize/2), float32(config.Config.Window.Height/2)-fBX, config.Config.Line.ThicknessSingleChannel, config.ThirdColorAdj, true)
					}
				}
			}
		}
	} else if config.Config.App.DefaultMode == config.BarsMode {
		for i := uint32(0); i < numSamples; i++ {
			AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.Buffers.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.Buffers.RingBufferSize]
			if *mixChannels {
				complexFFTBuffer[i] = complex((float64(AY)+float64(AX))/2, 0.0)
			} else {
				if *useRightChannel {
					complexFFTBuffer[i] = complex(float64(AY), 0.0)
				} else {
					complexFFTBuffer[i] = complex(float64(AX), 0.0)
				}
			}
		}
		if config.FiltersApplied {
			bars.CalcBars(&complexFFTBuffer, lowCutOffFrac, highCutOffFrac)
		} else {
			bars.CalcBars(&complexFFTBuffer, 0.0, 1.0)
		}
		barsDeltaTime := min(time.Since(barsLastFrameTime).Seconds(), 1.0)
		barsLastFrameTime = time.Now()
		bars.InterpolateBars(barsDeltaTime)
		for i := range bars.TargetBarsPos {
			x, y, w, h := bars.ComputeBarLayout(i)
			barColor := config.ThirdColorAdj
			if config.Config.Bars.PhaseColors.Enable {
				barColoroklch := oklab.OklabModel.Convert(barColor).(oklab.Oklab).Oklch()
				barColoroklch.H = math.Mod(barColoroklch.H+bars.InterpolatedBarsPhasePos[i]/(2*math.Pi)*config.Config.Bars.PhaseColors.HMult+1, 1)
				barColoroklch.L *= config.Config.Bars.PhaseColors.LMult
				barColoroklch.C *= config.Config.Bars.PhaseColors.CMult
				barr, barg, barb, _ := barColoroklch.RGBA()
				barColor = color.RGBA{uint8(barr >> 8), uint8(barg >> 8), uint8(barb >> 8), config.ThirdColorAdj.A}
			}
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), barColor, true)
		}
		if config.Config.Bars.PeakCursor.Enable {
			op := &text.DrawOptions{}
			op.GeoM.Translate(bars.InterpolatedPeakFreqCursorX+config.Config.Bars.PeakCursor.BGPadding, bars.InterpolatedPeakFreqCursorY+config.Config.Bars.PeakCursor.BGPadding+config.Config.Bars.PeakCursor.TextOffset)
			op.LayoutOptions.PrimaryAlign = text.AlignStart
			op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.Bars.PeakCursor.TextOpacity})
			opBG := &ebiten.DrawImageOptions{Blend: ebiten.BlendClear}
			opBG.GeoM.Translate(bars.InterpolatedPeakFreqCursorX, bars.InterpolatedPeakFreqCursorY)
			screen.DrawImage(barCursorImageBGRectFrame, opBG)
			if config.Config.ImageRetention.Enable && !firstFrame {
				copyPrevFrameOp(deltaTime, screen)
			}
			if config.Config.Bars.PeakCursor.ShowNote {
				text.Draw(screen, fmt.Sprintf("%-6.0f Hz %s", bars.PeakFreqCursorVal, bars.NoteDisplayName(bars.CalcNote(bars.PeakFreqCursorVal))), &text.GoTextFace{
					Source: fonts.FontA,
					Size:   config.Config.Bars.PeakCursor.TextSize,
				}, op)
			} else {
				text.Draw(screen, fmt.Sprintf("%-6.0f Hz", bars.PeakFreqCursorVal), &text.GoTextFace{
					Source: fonts.FontA,
					Size:   config.Config.Bars.PeakCursor.TextSize,
				}, op)
			}
		}
	} else {
		for i := uint32(0); i < numSamples; i++ {
			AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.Buffers.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.Buffers.RingBufferSize]
			XYComplexFFTBufferL[i] = complex(float64(AX), 0.0)
			XYComplexFFTBufferR[i] = complex(float64(AY), 0.0)
		}
		barsDeltaTime := min(time.Since(barsLastFrameTime).Seconds(), 1.0)
		barsLastFrameTime = time.Now()
		vu.Interpolate(barsDeltaTime)

		if config.FiltersApplied {
			vu.LoudnessLTarget = vu.CalcLoudness(&XYComplexFFTBufferL, lowCutOffFrac, highCutOffFrac)
			vu.LoudnessRTarget = vu.CalcLoudness(&XYComplexFFTBufferR, lowCutOffFrac, highCutOffFrac)
		} else {
			vu.LoudnessLTarget = vu.CalcLoudness(&XYComplexFFTBufferL, 0, 1)
			vu.LoudnessRTarget = vu.CalcLoudness(&XYComplexFFTBufferR, 0, 1)
		}
		if config.Config.VU.Peak.Enable {
			vu.CommitPeakPoint()
		}
		xl, yl, wl, hl := vu.ComputeBarLayout(0, vu.LoudnessLPos)
		xr, yr, wr, hr := vu.ComputeBarLayout(1, vu.LoudnessRPos)
		vector.DrawFilledRect(screen, float32(xl), float32(yl), float32(wl), float32(hl), config.ThirdColorAdj, true)
		vector.DrawFilledRect(screen, float32(xr), float32(yr), float32(wr), float32(hr), config.ThirdColorAdj, true)
		if config.Config.VU.Scale.Enable {
			if config.Config.VU.LogScale {
				xl, _, wl, _ := vu.ComputeBarLayout(0, 0)

				for _, div := range config.Config.VU.Scale.LogDivisions {
					op := &text.DrawOptions{}
					xr, yr, wr, hr := vu.ComputeBarLayout(1, (div-config.Config.VU.LogMinDB)/(config.Config.VU.LogMaxDB-config.Config.VU.LogMinDB))
					op.GeoM.Translate(float64(config.Config.Window.Width)/2, yr+hr-config.Config.VU.Scale.TextSize/2+config.Config.VU.Scale.TextOffset)
					op.LayoutOptions.PrimaryAlign = text.AlignCenter
					op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.Bars.PeakCursor.TextOpacity})
					text.Draw(screen, fmt.Sprintf("%3.0f dB", div), &text.GoTextFace{
						Source: fonts.FontA,
						Size:   config.Config.VU.Scale.TextSize,
					}, op)
					if config.Config.VU.Scale.TicksOuter {
						vector.StrokeLine(screen, float32(xr+wr+config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xr+wr+config.Config.VU.Scale.TickPadding+config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
						vector.StrokeLine(screen, float32(xl-config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xl-config.Config.VU.Scale.TickPadding-config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
					}
					if config.Config.VU.Scale.TicksInner {
						vector.StrokeLine(screen, float32(xr-config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xr-config.Config.VU.Scale.TickPadding-config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
						vector.StrokeLine(screen, float32(xl+wl+config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xl+wl+config.Config.VU.Scale.TickPadding+config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
					}
				}
			} else {
				for _, div := range config.Config.VU.Scale.LinDivisions {
					op := &text.DrawOptions{}
					_, yr, _, hr := vu.ComputeBarLayout(1, div/config.Config.VU.LinMax)
					op.GeoM.Translate(float64(config.Config.Window.Width)/2, yr+hr-config.Config.VU.Scale.TextSize/2+config.Config.VU.Scale.TextOffset)
					op.LayoutOptions.PrimaryAlign = text.AlignCenter
					op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.Bars.PeakCursor.TextOpacity})
					text.Draw(screen, fmt.Sprintf("%.1f", div), &text.GoTextFace{
						Source: fonts.FontA,
						Size:   config.Config.VU.Scale.TextSize,
					}, op)
					if config.Config.VU.Scale.TicksOuter {
						vector.StrokeLine(screen, float32(xr+wr+config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xr+wr+config.Config.VU.Scale.TickPadding+config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
						vector.StrokeLine(screen, float32(xl-config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xl-config.Config.VU.Scale.TickPadding-config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
					}
					if config.Config.VU.Scale.TicksInner {
						vector.StrokeLine(screen, float32(xr-config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xr-config.Config.VU.Scale.TickPadding-config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
						vector.StrokeLine(screen, float32(xl+wl+config.Config.VU.Scale.TickPadding), float32(yr+hr), float32(xl+wl+config.Config.VU.Scale.TickPadding+config.Config.VU.Scale.TickLength), float32(yr+hr), config.Config.VU.Scale.TickThickness, config.ThirdColorAdj, true)
					}
				}
			}
			if config.Config.VU.Peak.Enable {
				xl, yl, wl, hl = vu.ComputeBarLayout(0, vu.LoudnessLPeakPos)
				xr, yr, wr, hr = vu.ComputeBarLayout(1, vu.LoudnessRPeakPos)
				vector.StrokeLine(screen, float32(xl), float32(yl+hl), float32(xl+wl), float32(yl+hl), config.Config.VU.Peak.Thickness, config.ThirdColorAdj, true)
				vector.StrokeLine(screen, float32(xr), float32(yr+hr), float32(xr+wr), float32(yr+hr), config.Config.VU.Peak.Thickness, config.ThirdColorAdj, true)
			}
		}

	}

	if config.Config.BeatDetection.Enable || *beatDetectOverride {
		beatTimeDeltaTime := min(time.Since(beatTimeLastFrameTime).Seconds(), 1.0)
		beatTimeLastFrameTime = time.Now()
		beatdetect.InterpolateBeatTime(beatTimeDeltaTime)
		layoutY := config.Config.BeatDetection.Metronome.Padding
		if config.Config.BeatDetection.Metronome.Enable {
			if beatdetect.InterpolatedBPM != 0 {
				progress := float64(time.Now().Sub(beatdetect.InterpolatedBeatTime).Nanoseconds()) / (1000000000 * 60 / beatdetect.InterpolatedBPM)

				if config.Config.BeatDetection.Metronome.EdgeMode {
					easedProgress := 1 + math.Cos(progress*2*math.Pi)
					vector.DrawFilledRect(screen, 0, 0, float32(easedProgress*config.Config.BeatDetection.Metronome.EdgeThickness), float32(config.Config.Window.Height), config.ThirdColorAdj, true)
					vector.DrawFilledRect(screen, 0, 0, float32(config.Config.Window.Width), float32(easedProgress*config.Config.BeatDetection.Metronome.EdgeThickness), config.ThirdColorAdj, true)
					vector.DrawFilledRect(screen, float32(config.Config.Window.Width), 0, -float32(easedProgress*config.Config.BeatDetection.Metronome.EdgeThickness), float32(config.Config.Window.Height), config.ThirdColorAdj, true)
					vector.DrawFilledRect(screen, 0, float32(config.Config.Window.Height), float32(config.Config.Window.Width), -float32(easedProgress*config.Config.BeatDetection.Metronome.EdgeThickness), config.ThirdColorAdj, true)

				} else {
					easedProgress := math.Sin(progress * math.Pi)
					if config.Config.BeatDetection.Metronome.ThinLineMode {
						if config.Config.BeatDetection.Metronome.ThinLineThicknessChangeWithVelocity {
							easedVelocityProgress := math.Cos(progress * math.Pi)
							vector.DrawFilledRect(screen, float32(config.Config.Window.Width)/2+float32(easedProgress)*float32(config.Config.Window.Width)/2-float32(easedVelocityProgress*config.Config.BeatDetection.Metronome.ThinLineThickness)/2, float32(layoutY), float32(easedVelocityProgress*config.Config.BeatDetection.Metronome.ThinLineThickness), float32(config.Config.BeatDetection.Metronome.Height), config.ThirdColorAdj, true)
						} else {
							vector.DrawFilledRect(screen, float32(config.Config.Window.Width)/2+float32(easedProgress)*float32(config.Config.Window.Width)/2-float32(config.Config.BeatDetection.Metronome.ThinLineThickness)/2, float32(layoutY), float32(config.Config.BeatDetection.Metronome.ThinLineThickness), float32(config.Config.BeatDetection.Metronome.Height), config.ThirdColorAdj, true)
						}
						vector.DrawFilledRect(screen, float32(config.Config.Window.Width)/2-float32(config.Config.BeatDetection.Metronome.ThinLineHintThickness)/2, float32(layoutY), float32(config.Config.BeatDetection.Metronome.ThinLineHintThickness), float32(config.Config.BeatDetection.Metronome.Height), config.ThirdColorAdj, true)
					} else {
						vector.DrawFilledRect(screen, float32(config.Config.Window.Width)/2, float32(layoutY), float32(easedProgress)*float32(config.Config.Window.Width)/2, float32(config.Config.BeatDetection.Metronome.Height), config.ThirdColorAdj, true)
					}
				}
			}
			if !config.Config.BeatDetection.Metronome.EdgeMode {
				layoutY += config.Config.BeatDetection.Metronome.Height + config.Config.BeatDetection.Metronome.Padding
			}
		}

		if config.Config.BeatDetection.ShowBPM {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(config.Config.Window.Width)/2, float64(layoutY))
			op.LayoutOptions.PrimaryAlign = text.AlignCenter
			op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRIS.TextOpacity})
			displayedBPM := beatdetect.InterpolatedBPM
			if config.Config.BeatDetection.HalfDisplayedBPM {
				displayedBPM /= 2
			}
			text.Draw(screen, fmt.Sprintf("%0.2f BPM", displayedBPM), &text.GoTextFace{
				Source: fonts.FontA,
				Size:   config.Config.BeatDetection.BPMTextSize,
			}, op)
		}
	}

	//audio.SampleRingBuffer.Reset()

	if config.Config.ImageRetention.Enable {
		prevFrame.Clear()
		prevFrame.DrawImage(screen, nil)
		if config.Config.Colors.DisableBGTransparency {
			screen.Fill(config.BGColor)
			screen.DrawImage(prevFrame, nil)
		}
	}
	if config.Config.Shaders.Enable {
		timeUniform := float32(time.Since(startTime).Milliseconds()) / 1000
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[2] = shaderWorkBuffer
		shaders.GenShaderRenderList()
		for _, shader := range shaders.ShaderRenderList {
			shaderWorkBuffer.Clear()
			shaderWorkBuffer.DrawImage(screen, nil)
			op.Uniforms = shader.Arguments
			op.Uniforms["Time"] = timeUniform * shader.TimeScale
			screen.Clear()
			screen.DrawRectShader(int(config.Config.Window.Width), int(config.Config.Window.Height), shader.Shader, op)
		}
	}

	if config.Config.App.FPSCounter {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
	}

	if config.Config.MPRIS.Enable {
		op := &text.DrawOptions{}
		op.GeoM.Translate(16, 16+config.Config.MPRIS.TextTitleYOffset)
		op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRIS.TextOpacity})
		text.Draw(screen, media.PlayingMediaInfo.Artist+" - "+media.PlayingMediaInfo.Title, fonts.MPRISBigTextFace, op)

		op = &text.DrawOptions{}
		op.GeoM.Translate(16, 64+config.Config.MPRIS.TextAlbumYOffset)
		op.ColorScale.ScaleWithColor(color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, config.Config.MPRIS.TextOpacity})

		text.Draw(screen, media.PlayingMediaInfo.Album, fonts.MPRISSmallTextFace, op)

		op = &text.DrawOptions{}
		op.GeoM.Translate(16, 80+config.Config.MPRIS.TextDurationYOffset)
		op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRIS.TextOpacity})

		text.Draw(screen, media.FmtDuration(media.PlayingMediaInfo.Position)+" / "+media.FmtDuration(media.PlayingMediaInfo.Duration), fonts.MPRISBigTextFace, op)
	}

	if config.FiltersApplied && config.Config.FilterInfo.Enable {
		op := &text.DrawOptions{}
		loFreq := lowCutOffFrac * float64(config.Config.Audio.SampleRate)
		hiFreq := highCutOffFrac * float64(config.Config.Audio.SampleRate)

		op = &text.DrawOptions{}
		op.GeoM.Translate(config.Config.FilterInfo.TextPaddingLeft, float64(config.Config.Window.Height)-config.Config.FilterInfo.TextSize-config.Config.FilterInfo.TextPaddingBottom)
		op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRIS.TextOpacity})

		text.Draw(screen, fmt.Sprintf("Lo: %0.2f Hz; Hi: %0.2f Hz", loFreq, hiFreq), &text.GoTextFace{
			Source: fonts.FontA,
			Size:   config.Config.FilterInfo.TextSize,
		}, op)
	}

	if splash.SplashShowing {
		splash.DrawSplash(screen)
	}

	if firstFrame {
		firstFrame = false
		// f, _ := os.Create("image.png")
		// png.Encode(f, prevFrame)
	}

	readHeadPosition = (readHeadPosition + numSamples*8) % config.Config.Buffers.RingBufferSize
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if config.Config.Window.Width != int32(outsideWidth) || config.Config.Window.Height != int32(outsideHeight) {
		config.Config.Window.Width = int32(outsideWidth)
		config.Config.Window.Height = int32(outsideHeight)
		InitBuffersAtSize(outsideWidth, outsideHeight)
		bars.Init()
	}
	return outsideWidth, outsideHeight
}

func InitBuffersAtSize(width int, height int) {
	prevFrame = ebiten.NewImage(width, height)
	if config.Config.Shaders.Enable {
		shaderWorkBuffer = ebiten.NewImage(width, height)
	}
}

func Init() {
	numSamples := config.Config.Buffers.ReadBufferSize / 2
	FFTBuffer = make([]float64, numSamples)
	lowCutOff := flag.Float64("lo", 0.0, "low frequency cutoff fraction (discernable details are around 0.001 increments for a 4096 buffer size and 192kHz sample rate)")
	highCutOff := flag.Float64("hi", 1.0, "high frequency cutoff fraction (discernable details are around 0.001 increments for a 4096 buffer size and 192kHz sample rate)")
	useRightChannel = flag.Bool("right", false, "Use the right channel instead of the left for the single axis oscilloscope")
	mixChannels = flag.Bool("mix", false, "Mix channels instead of just using a single channel for the single axis oscilloscope")
	beatDetectOverride = flag.Bool("beatdetect", false, "Enable beat detection (bypassing configuration)")
	peakDetectOverride = flag.Bool("peakdetect", false, "Enable peak detection (bypassing configuration)")

	overrideGain := flag.Float64("gain", -1.0, "override gain multiplier")

	overrideWidth := flag.Int("width", int(config.Config.Window.Width), "override window width")
	overrideHeight := flag.Int("height", int(config.Config.Window.Height), "override window height")

	overrideX = flag.Int("x", -1, "override starting x coordinate of the center of the window; x=0 corresponds to the center of the screen")
	overrideY = flag.Int("y", -1, "override starting y coordinate of the center of the window; y=0 corresponds to the center of the screen")

	overrideMode := flag.Int("mode", -1, "override default mode")

	overrideResizable := flag.Bool("resizable", false, "override if window is resizable (toggles the value defined in the config)")

	flag.Parse()
	if *lowCutOff != 0.0 {
		config.FiltersApplied = true
		lowCutOffFrac = *lowCutOff
	}
	if *highCutOff != 1.0 {
		config.FiltersApplied = true
		highCutOffFrac = *highCutOff
	}

	complexFFTBuffer = make([]complex128, numSamples)
	if config.Config.SingleChannelOsc.PeakDetect.UseACF {
		complexFFTBufferFlipped = make([]complex128, numSamples)
		align.Init()
	}
	XYComplexFFTBufferL = make([]complex128, numSamples)
	XYComplexFFTBufferR = make([]complex128, numSamples)
	config.Config.Window.Width = int32(*overrideWidth)
	config.Config.Window.Height = int32(*overrideHeight)

	if *overrideGain != -1.0 {
		config.Config.Audio.Gain = float32(*overrideGain)
	}

	if *overrideMode != -1 {
		config.Config.App.DefaultMode = *overrideMode
	}

	if *overrideResizable {
		config.Config.Window.Resizable = !config.Config.Window.Resizable
	}
}

func main() {
	config.Init()
	audio.Init()
	Init()
	fonts.Init()
	icons.Init()
	go audio.Start()
	if config.Config.BeatDetection.Enable || *beatDetectOverride {
		go beatdetect.Start()
	}
	if config.Config.MPRIS.Enable {
		go media.Start()
	}
	if config.Config.Shaders.Enable {
		shaders.Init()
	}
	bars.Init()
	kaiser.Init()
	if config.Config.VU.Peak.Enable {
		vu.Init()
	}
	if config.Config.SingleChannelOsc.Slew.Enable {
		slew.Init()
	}
	splash.Init()
	ebiten.SetWindowIcon([]image.Image{icons.WindowIcon48, icons.WindowIcon32, icons.WindowIcon16})
	ebiten.SetWindowSize(int(config.Config.Window.Width), int(config.Config.Window.Height))
	ebiten.SetWindowTitle("xyosc")
	ebiten.SetWindowMousePassthrough(true)
	if config.Config.Window.Resizable {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	}
	ebiten.SetTPS(int(config.Config.App.TargetFPS))
	ebiten.SetWindowDecorated(false)
	screenW, screenH := ebiten.Monitor().Size()
	if *overrideX == -1 {
		*overrideX = screenW/2 - int(config.Config.Window.Width)/2
	} else {
		*overrideX = screenW/2 + *overrideX - int(config.Config.Window.Width)/2
	}
	if *overrideY == -1 {
		*overrideY = screenH/2 - int(config.Config.Window.Height)/2
	} else {
		*overrideY = screenH/2 + *overrideY - int(config.Config.Window.Height)/2
	}
	ebiten.SetWindowPosition(*overrideX, *overrideY)
	ebiten.SetVsyncEnabled(true)
	if config.Config.Bars.PeakCursor.Enable {
		barCursorImageBGRectFrame = ebiten.NewImage(int(config.Config.Bars.PeakCursor.BGWidth), int(config.Config.Bars.PeakCursor.TextSize+2*config.Config.Bars.PeakCursor.BGPadding))
	}
	InitBuffersAtSize(int(config.Config.Window.Width), int(config.Config.Window.Height))
	gameOptions := ebiten.RunGameOptions{SingleThread: true, ScreenTransparent: true}
	if config.Config.Colors.DisableBGTransparency {
		gameOptions.ScreenTransparent = false
	}

	startTime = time.Now()
	if err := ebiten.RunGameWithOptions(&Game{}, &gameOptions); err != nil {
		log.Fatal(err)
	}
}
