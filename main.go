package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"time"
	"xyosc/align"
	"xyosc/audio"
	"xyosc/bars"
	"xyosc/beatdetect"
	"xyosc/config"
	"xyosc/fastsqrt"
	"xyosc/filter"
	"xyosc/fonts"
	"xyosc/icons"
	"xyosc/kaiser"
	"xyosc/media"
	"xyosc/particles"
	"xyosc/shaders"
	"xyosc/utils"

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

var readHeadPosition uint32 = 0
var startTime time.Time
var prevFrame *ebiten.Image
var shaderWorkBuffer *ebiten.Image
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

var StillSamePressFromToggleKey bool

var barsLastFrameTime time.Time
var beatTimeLastFrameTime time.Time
var lastFrameTime time.Time

func (g *Game) Draw(screen *ebiten.Image) {
	deltaTime := min(time.Since(lastFrameTime).Seconds(), 1.0)
	lastFrameTime = time.Now()
	var numSamples = config.Config.ReadBufferSize / 2
	if config.Config.CopyPreviousFrame {
		if !firstFrame {
			op := &ebiten.DrawImageOptions{}
			op.ColorScale.ScaleAlpha(float32(math.Pow(float64(config.Config.CopyPreviousFrameAlphaDecayBase), deltaTime*config.Config.CopyPreviousFrameAlphaDecaySpeed)))
			screen.DrawImage(prevFrame, op)
		}
	} else {
		if config.Config.DisableTransparency {
			screen.Fill(config.BGColor)
		}
	}
	scale := min(config.Config.WindowWidth, config.Config.WindowHeight) / 2
	var AX float32
	var AY float32
	var BX float32
	var BY float32
	posStartRead := (config.Config.RingBufferSize + audio.WriteHeadPosition - numSamples*2 - config.Config.ReadBufferDelay) % config.Config.RingBufferSize
	if slices.Contains(pressedKeys, ebiten.KeyF) {
		if !StillSamePressFromToggleKey {
			StillSamePressFromToggleKey = true
			config.Config.DefaultMode = (config.Config.DefaultMode + 1) % 3
		}
	} else {
		StillSamePressFromToggleKey = false
	}

	if (config.Config.DefaultMode == 0 || config.Config.DefaultMode == 1) && config.Config.ScaleEnable {
		if config.Config.ScaleMainAxisEnable {
			vector.StrokeLine(screen, 0, float32(config.Config.WindowHeight/2), float32(config.Config.WindowWidth), float32(config.Config.WindowHeight/2), config.Config.ScaleMainAxisStrokeThickness, config.ThirdColorAdj, true)
			vector.StrokeLine(screen, float32(config.Config.WindowWidth/2), 0, float32(config.Config.WindowWidth/2), float32(config.Config.WindowHeight), config.Config.ScaleMainAxisStrokeThickness, config.ThirdColorAdj, true)
		}
		if config.Config.ScaleVertTickEnable {
			for i := range config.Config.ScaleVertDiv + 1 {
				y := float32(config.Config.WindowHeight) / float32(config.Config.ScaleVertDiv) * float32(i)
				if config.Config.ScaleVertTickExpandToGrid {
					vector.StrokeLine(screen, 0, y, float32(config.Config.WindowWidth), y, config.Config.ScaleVertTickExpandToGridThickness, config.ThirdColorAdj, true)
				}
				vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)-config.Config.ScaleVertTickLength/2, y, float32(config.Config.WindowWidth/2)+config.Config.ScaleVertTickLength/2, y, config.Config.ScaleVertTickStrokeThickness, config.ThirdColorAdj, true)
				if config.Config.ScaleVertTextEnable {
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64(config.Config.WindowWidth)/2+float64(config.Config.ScaleVertTickLength/2)+config.Config.ScaleVertTextPadding, float64(y))
					op.LayoutOptions.PrimaryAlign = text.AlignStart
					op.LayoutOptions.SecondaryAlign = text.AlignCenter
					op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.ScaleTextOpacity})
					text.Draw(screen, fmt.Sprintf("%.*f", int(math.Ceil(math.Log10(float64(config.Config.ScaleVertDiv)/2))), 2/float32(config.Config.ScaleVertDiv)*float32(i)-1), &text.GoTextFace{
						Source: fonts.Font,
						Size:   config.Config.ScaleVertTextSize,
					}, op)
				}
			}
		}
	}

	if (config.Config.DefaultMode == 0) && config.Config.ScaleEnable {
		if config.Config.ScaleMainAxisEnable {
			vector.StrokeLine(screen, 0, float32(config.Config.WindowHeight/2), float32(config.Config.WindowWidth), float32(config.Config.WindowHeight/2), config.Config.ScaleMainAxisStrokeThickness, config.ThirdColorAdj, true)
			vector.StrokeLine(screen, float32(config.Config.WindowWidth/2), 0, float32(config.Config.WindowWidth/2), float32(config.Config.WindowHeight), config.Config.ScaleMainAxisStrokeThickness, config.ThirdColorAdj, true)
		}
		if config.Config.ScaleHorzTickEnable {
			for i := range config.Config.ScaleHorzDiv + 1 {
				x := float32(config.Config.WindowWidth) / float32(config.Config.ScaleHorzDiv) * float32(i)
				if config.Config.ScaleHorzTickExpandToGrid {
					vector.StrokeLine(screen, x, 0, x, float32(config.Config.WindowHeight), config.Config.ScaleHorzTickExpandToGridThickness, config.ThirdColorAdj, true)
				}
				vector.StrokeLine(screen, x, float32(config.Config.WindowHeight/2)-config.Config.ScaleHorzTickLength/2, x, float32(config.Config.WindowHeight/2)+config.Config.ScaleHorzTickLength/2, config.Config.ScaleHorzTickStrokeThickness, config.ThirdColorAdj, true)
				if config.Config.ScaleHorzTextEnable {
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64(x), float64(config.Config.WindowHeight)/2+float64(config.Config.ScaleHorzTickLength/2)+config.Config.ScaleHorzTextPadding)
					op.LayoutOptions.PrimaryAlign = text.AlignCenter
					op.LayoutOptions.SecondaryAlign = text.AlignStart
					op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.ScaleTextOpacity})
					text.Draw(screen, fmt.Sprintf("%.*f", int(math.Ceil(math.Log10(float64(config.Config.ScaleHorzDiv)/2))), 2/float32(config.Config.ScaleHorzDiv)*float32(i)-1), &text.GoTextFace{
						Source: fonts.Font,
						Size:   config.Config.ScaleHorzTextSize,
					}, op)
				}
			}
		}
	}

	if config.Config.DefaultMode == 0 {
		if config.FiltersApplied {
			for i := uint32(0); i < numSamples; i++ {
				AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.RingBufferSize]
				AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.RingBufferSize]
				XYComplexFFTBufferL[i] = complex(float64(AX), 0.0)
				XYComplexFFTBufferR[i] = complex(float64(AY), 0.0)
			}
			filter.FilterBufferInPlace(&XYComplexFFTBufferL, lowCutOffFrac, highCutOffFrac)
			filter.FilterBufferInPlace(&XYComplexFFTBufferR, lowCutOffFrac, highCutOffFrac)
			AX = float32(real(XYComplexFFTBufferL[len(XYComplexFFTBufferL)-1]))
			AY = float32(real(XYComplexFFTBufferR[len(XYComplexFFTBufferR)-1]))
		} else {
			AX = audio.SampleRingBufferUnsafe[(posStartRead)%config.Config.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+1)%config.Config.RingBufferSize]
		}
		S := float32(0)
		for i := uint32(0); i < numSamples; i++ {
			if config.FiltersApplied {
				BX = float32(real(XYComplexFFTBufferL[i]))
				BY = float32(real(XYComplexFFTBufferR[i]))
			} else {
				BX = audio.SampleRingBufferUnsafe[(posStartRead+i*2+2)%config.Config.RingBufferSize]
				BY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+3)%config.Config.RingBufferSize]
			}
			fAX := float32(AX) * config.Config.Gain * float32(scale)
			fAY := -float32(AY) * config.Config.Gain * float32(scale)
			fBX := float32(BX) * config.Config.Gain * float32(scale)
			fBY := -float32(BY) * config.Config.Gain * float32(scale)
			if i >= numSamples-config.Config.XYOscilloscopeReadBufferSize {
				if config.Config.LineInvSqrtOpacityControl || config.Config.LineTimeDependentOpacityControl {
					mult := 1.0
					if config.Config.LineInvSqrtOpacityControl {
						invStrength := min(float64(fastsqrt.FastInvSqrt32((fBX-fAX)*(fBX-fAX)+(fBY-fBY)*(fBY-fBY))), 1.0)
						if config.Config.LineInvSqrtOpacityControlUseLogDecrement {
							mult *= config.Config.LineInvSqrtOpacityControlLogDecrementOffset + math.Log(invStrength)/math.Log(config.Config.LineInvSqrtOpacityControlLogDecrementBase)
						} else {
							mult *= invStrength
						}
					}
					if config.Config.LineTimeDependentOpacityControl {
						mult *= math.Pow(config.Config.LineTimeDependentOpacityControlBase, float64(numSamples-i-1))
					}
					colorAdjusted := color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, uint8(float64(config.Config.LineOpacity) * mult)}
					if config.Config.LineOpacityControlAlsoAppliesToThickness {
						vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowHeight/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowHeight/2)+fBY, config.Config.LineThickness*float32(mult), colorAdjusted, true)
					} else {
						vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowHeight/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowHeight/2)+fBY, config.Config.LineThickness, colorAdjusted, true)
					}
				} else {
					vector.StrokeLine(screen, float32(config.Config.WindowWidth/2)+fAX, float32(config.Config.WindowHeight/2)+fAY, float32(config.Config.WindowWidth/2)+fBX, float32(config.Config.WindowHeight/2)+fBY, config.Config.LineThickness, config.ThirdColorAdj, true)
				}
			}
			S += float32(AX)*float32(AX) + float32(AY)*float32(AY)
			if config.Config.Particles {
				if rand.IntN(int(float64(config.Config.ParticleGenPerFrameEveryXSamples)/deltaTime)) == 0 {
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
			vector.DrawFilledCircle(screen, float32(config.Config.WindowWidth/2)+particle.X*float32(scale), float32(config.Config.WindowHeight/2)+particle.Y*float32(scale), particle.Size, config.ParticleColor, true)
			norm := math32.Sqrt(particles.Particles[i].X*particles.Particles[i].X + particles.Particles[i].Y*particles.Particles[i].Y)
			particles.Particles[i].X += particle.VX * float32(deltaTime)
			particles.Particles[i].Y += particle.VY * float32(deltaTime)
			speed := math32.Sqrt(particle.VX*particle.VX + particle.VY*particle.VY)
			particles.Particles[i].VX += (config.Config.ParticleAcceleration*S - speed*config.Config.ParticleDrag) * particle.X / norm * float32(deltaTime)
			particles.Particles[i].VY += (config.Config.ParticleAcceleration*S - speed*config.Config.ParticleDrag) * particle.Y / norm * float32(deltaTime)
		}
	} else if config.Config.DefaultMode == 1 {
		for i := uint32(0); i < numSamples; i++ {
			AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.RingBufferSize]
			if config.FiltersApplied || config.Config.UseBetterPeakDetectionAlgorithm {
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
			if !config.FiltersApplied || config.Config.UseBetterPeakDetectionAlgorithm {
				if *mixChannels {
					FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = (float64(AY) + float64(AX)) / 2
				} else {
					if *useRightChannel {
						FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = float64(AY)
					} else {
						FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = float64(AX)
					}
				}
			}
		}
		if config.FiltersApplied {
			filter.FilterBufferInPlace(&complexFFTBuffer, lowCutOffFrac, highCutOffFrac)
			for i := uint32(0); i < numSamples; i++ {
				FFTBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = real(complexFFTBuffer[i])
			}
		}

		indices := []int{}
		offset := uint32(0)
		freq := uint32(0)

		if config.Config.PeriodCrop || config.Config.OscilloscopeStartPeakDetection || *peakDetectOverride {
			if config.Config.UseBetterPeakDetectionAlgorithm {
				for i := range len(complexFFTBuffer) {
					complexFFTBufferFlipped[len(complexFFTBufferFlipped)-i-1] = complexFFTBuffer[i]
				}
				offset, indices, freq = align.AutoCorrelate(&complexFFTBuffer, &complexFFTBufferFlipped)

			} else {
				indices = peaks.Get(FFTBuffer, config.Config.PeakDetectSeparator)
				sort.Ints(indices)
				offset = uint32(0)
				if len(indices) != 0 {
					offset = uint32(indices[0])
				}
				if (offset+config.Config.FFTBufferOffset)%numSamples < config.Config.PeakDetectEdgeGuardBufferSize || (numSamples-((offset+config.Config.FFTBufferOffset)%numSamples)) < config.Config.PeakDetectEdgeGuardBufferSize {
					if len(indices) >= 2 {
						offset = uint32(indices[1])
					}
				}
			}
		}

		var samplesPerCrop uint32

		if config.Config.PeriodCrop && len(indices) > 1 {
			lastPeriodOffset := uint32(indices[min(len(indices)-1, config.Config.PeriodCropCount)])
			samplesPerCrop = lastPeriodOffset - offset
			if config.Config.UseBetterPeakDetectionAlgorithm {
				samplesPerCrop = freq * 2
			}
		} else {
			samplesPerCrop = numSamples
		}
		if (config.Config.DefaultMode == 1) && config.Config.ScaleEnable {
			visibleSampleCount := min(numSamples, samplesPerCrop*config.Config.PeriodCropLoopOverCount)
			timeSpanVisible := float64(visibleSampleCount) / float64(2*config.Config.SampleRate) //s
			if config.Config.ScaleMainAxisEnable {
				vector.StrokeLine(screen, 0, float32(config.Config.WindowHeight/2), float32(config.Config.WindowWidth), float32(config.Config.WindowHeight/2), config.Config.ScaleMainAxisStrokeThickness, config.ThirdColorAdj, true)
				vector.StrokeLine(screen, float32(config.Config.WindowWidth/2), 0, float32(config.Config.WindowWidth/2), float32(config.Config.WindowHeight), config.Config.ScaleMainAxisStrokeThickness, config.ThirdColorAdj, true)
			}
			scaleScale := 1.
			if config.Config.ScaleHorzDivDynamicPos {
				scaleScale = math.Pow10(int(math.Ceil(math.Log10(timeSpanVisible)))) / timeSpanVisible
			}
			if config.Config.ScaleHorzTickEnable {
				for i := range config.Config.ScaleHorzDiv + 1 {
					x := float32(config.Config.WindowWidth)/2 + float32(scaleScale)*float32(config.Config.WindowWidth)/float32(config.Config.ScaleHorzDiv)*(float32(i)-float32(config.Config.ScaleHorzDiv)/2)
					if config.Config.ScaleHorzTickExpandToGrid {
						vector.StrokeLine(screen, x, 0, x, float32(config.Config.WindowHeight), config.Config.ScaleHorzTickExpandToGridThickness, config.ThirdColorAdj, true)
					}
					vector.StrokeLine(screen, x, float32(config.Config.WindowHeight/2)-config.Config.ScaleHorzTickLength/2, x, float32(config.Config.WindowHeight/2)+config.Config.ScaleHorzTickLength/2, config.Config.ScaleHorzTickStrokeThickness, config.ThirdColorAdj, true)
					if config.Config.ScaleHorzTextEnable {
						op := &text.DrawOptions{}
						op.GeoM.Translate(float64(x), float64(config.Config.WindowHeight)/2+float64(config.Config.ScaleHorzTickLength/2)+config.Config.ScaleHorzTextPadding)
						op.LayoutOptions.PrimaryAlign = text.AlignCenter
						op.LayoutOptions.SecondaryAlign = text.AlignStart
						op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.ScaleTextOpacity})
						text.Draw(screen, fmt.Sprintf("%s", utils.FormatDuration(scaleScale*float64(timeSpanVisible/2)*(2/float64(config.Config.ScaleHorzDiv)*float64(i)-1))), &text.GoTextFace{
							Source: fonts.Font,
							Size:   config.Config.ScaleHorzTextSize,
						}, op)
					}
				}
			}
		}
		if config.Config.PeriodCrop && len(indices) > 1 {
			if config.Config.CenterPeak {
				offset -= samplesPerCrop / 2
			}
			for i := uint32(0); i < min(numSamples, samplesPerCrop*config.Config.PeriodCropLoopOverCount)-1; i++ {
				fAX := float32(FFTBuffer[(i+offset)%numSamples]) * config.Config.Gain * float32(scale)
				fBX := float32(FFTBuffer[(i+1+offset)%numSamples]) * config.Config.Gain * float32(scale)
				if (i+1+offset-config.Config.FFTBufferOffset)%numSamples != 0 {
					vector.StrokeLine(screen, float32(config.Config.WindowWidth)*float32(i%samplesPerCrop)/float32(samplesPerCrop), float32(config.Config.WindowHeight/2)+fAX, float32(config.Config.WindowWidth)*float32(i%samplesPerCrop+1)/float32(samplesPerCrop), float32(config.Config.WindowHeight/2)+fBX, config.Config.LineThickness, config.ThirdColorAdj, true)
				}
			}
		} else {
			if config.Config.CenterPeak {
				offset -= config.Config.SingleChannelWindow / 4
			}
			for i := uint32(0); i < config.Config.SingleChannelWindow/2-1; i++ {
				fAX := float32(FFTBuffer[(i+offset)%numSamples]) * config.Config.Gain * float32(scale)
				fBX := float32(FFTBuffer[(i+1+offset)%numSamples]) * config.Config.Gain * float32(scale)
				if (i+1+offset-config.Config.FFTBufferOffset)%numSamples != 0 {
					vector.StrokeLine(screen, float32(config.Config.WindowWidth)*float32(i%(config.Config.SingleChannelWindow/2))/float32((config.Config.SingleChannelWindow/2)), float32(config.Config.WindowHeight/2)+fAX, float32(config.Config.WindowWidth)*float32(i%(config.Config.SingleChannelWindow/2)+1)/float32(config.Config.SingleChannelWindow/2), float32(config.Config.WindowHeight/2)+fBX, config.Config.LineThickness, config.ThirdColorAdj, true)
				}
			}
		}
	} else {
		for i := uint32(0); i < numSamples; i++ {
			AX = audio.SampleRingBufferUnsafe[(posStartRead+i*2)%config.Config.RingBufferSize]
			AY = audio.SampleRingBufferUnsafe[(posStartRead+i*2+1)%config.Config.RingBufferSize]
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
			vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), config.ThirdColorAdj, true)
		}
		if config.Config.BarsPeakFreqCursor {
			op := &text.DrawOptions{}
			op.GeoM.Translate(bars.InterpolatedPeakFreqCursorX+config.Config.BarsPeakFreqCursorBGPadding, bars.InterpolatedPeakFreqCursorY+config.Config.BarsPeakFreqCursorBGPadding+config.Config.BarsPeakFreqCursorTextOffset)
			op.LayoutOptions.PrimaryAlign = text.AlignStart
			op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.BarsPeakFreqCursorTextOpacity})

			vector.FillRect(screen, float32(bars.InterpolatedPeakFreqCursorX), float32(bars.InterpolatedPeakFreqCursorY), float32(config.Config.BarsPeakFreqCursorBGWidth), float32(config.Config.BarsPeakFreqCursorTextSize+2*config.Config.BarsPeakFreqCursorBGPadding), config.BGColor, true)
			text.Draw(screen, fmt.Sprintf("%-6.0f Hz", bars.PeakFreqCursorVal), &text.GoTextFace{
				Source: fonts.Font,
				Size:   config.Config.BarsPeakFreqCursorTextSize,
			}, op)
		}
	}

	if config.Config.BeatDetect || *beatDetectOverride {
		beatTimeDeltaTime := min(time.Since(beatTimeLastFrameTime).Seconds(), 1.0)
		beatTimeLastFrameTime = time.Now()
		beatdetect.InterpolateBeatTime(beatTimeDeltaTime)
		layoutY := config.Config.MetronomePadding
		if config.Config.ShowMetronome {
			if beatdetect.InterpolatedBPM != 0 {
				progress := float64(time.Now().Sub(beatdetect.InterpolatedBeatTime).Nanoseconds()) / (1000000000 * 60 / beatdetect.InterpolatedBPM)
				easedProgress := math.Sin(progress * math.Pi)
				if config.Config.MetronomeThinLineMode {
					if config.Config.MetronomeThinLineThicknessChangeWithVelocity {
						easedVelocityProgress := math.Cos(progress * math.Pi)
						vector.FillRect(screen, float32(config.Config.WindowWidth)/2+float32(easedProgress)*float32(config.Config.WindowWidth)/2-float32(easedVelocityProgress*config.Config.MetronomeThinLineThickness)/2, float32(layoutY), float32(easedVelocityProgress*config.Config.MetronomeThinLineThickness), float32(config.Config.MetronomeHeight), config.ThirdColorAdj, true)
					} else {
						vector.FillRect(screen, float32(config.Config.WindowWidth)/2+float32(easedProgress)*float32(config.Config.WindowWidth)/2-float32(config.Config.MetronomeThinLineThickness)/2, float32(layoutY), float32(config.Config.MetronomeThinLineThickness), float32(config.Config.MetronomeHeight), config.ThirdColorAdj, true)
					}
					vector.FillRect(screen, float32(config.Config.WindowWidth)/2-float32(config.Config.MetronomeThinLineHintThickness)/2, float32(layoutY), float32(config.Config.MetronomeThinLineHintThickness), float32(config.Config.MetronomeHeight), config.ThirdColorAdj, true)
				} else {
					vector.FillRect(screen, float32(config.Config.WindowWidth)/2, float32(layoutY), float32(easedProgress)*float32(config.Config.WindowWidth)/2, float32(config.Config.MetronomeHeight), config.ThirdColorAdj, true)
				}
			}
			layoutY += config.Config.MetronomeHeight + config.Config.MetronomePadding
		}
		if config.Config.ShowBPM {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(config.Config.WindowWidth)/2, float64(layoutY))
			op.LayoutOptions.PrimaryAlign = text.AlignCenter
			op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRISTextOpacity})
			displayedBPM := beatdetect.InterpolatedBPM
			if config.Config.BeatDetectHalfDisplayedBPM {
				displayedBPM /= 2
			}
			text.Draw(screen, fmt.Sprintf("%0.2f BPM", displayedBPM), &text.GoTextFace{
				Source: fonts.Font,
				Size:   config.Config.BPMTextSize,
			}, op)
		}
	}

	//audio.SampleRingBuffer.Reset()

	if config.Config.CopyPreviousFrame {
		prevFrame.Clear()
		prevFrame.DrawImage(screen, nil)
		if config.Config.DisableTransparency {
			screen.Fill(config.BGColor)
			screen.DrawImage(prevFrame, nil)
		}
	}
	if config.Config.UseShaders {
		timeUniform := float32(time.Since(startTime).Milliseconds()) / 1000
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[2] = shaderWorkBuffer
		for _, shader := range shaders.ShaderRenderList {
			shaderWorkBuffer.Clear()
			shaderWorkBuffer.DrawImage(screen, nil)
			op.Uniforms = shader.Arguments
			op.Uniforms["Time"] = timeUniform * shader.TimeScale
			screen.DrawRectShader(int(config.Config.WindowWidth), int(config.Config.WindowHeight), shader.Shader, op)
		}
	}

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

	if config.FiltersApplied && config.Config.ShowFilterInfo {
		op := &text.DrawOptions{}
		loFreq := lowCutOffFrac * float64(config.Config.SampleRate)
		hiFreq := highCutOffFrac * float64(config.Config.SampleRate)

		op = &text.DrawOptions{}
		op.GeoM.Translate(config.Config.FilterInfoTextPaddingLeft, float64(config.Config.WindowHeight)-config.Config.FilterInfoTextSize-config.Config.FilterInfoTextPaddingBottom)
		op.ColorScale.ScaleWithColor(color.RGBA{config.AccentColor.R, config.AccentColor.G, config.AccentColor.B, config.Config.MPRISTextOpacity})

		text.Draw(screen, fmt.Sprintf("Lo: %0.2f Hz; Hi: %0.2f Hz", loFreq, hiFreq), &text.GoTextFace{
			Source: fonts.Font,
			Size:   config.Config.FilterInfoTextSize,
		}, op)
	}

	if firstFrame {
		firstFrame = false
		// f, _ := os.Create("image.png")
		// png.Encode(f, prevFrame)
	}

	readHeadPosition = (readHeadPosition + numSamples*8) % config.Config.RingBufferSize
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if config.Config.WindowWidth != int32(outsideWidth) || config.Config.WindowHeight != int32(outsideHeight) {
		config.Config.WindowWidth = int32(outsideWidth)
		config.Config.WindowHeight = int32(outsideHeight)
		InitBuffersAtSize(outsideWidth, outsideHeight)
		bars.Init()
	}
	return outsideWidth, outsideHeight
}

func InitBuffersAtSize(width int, height int) {
	prevFrame = ebiten.NewImage(width, height)
	if config.Config.UseShaders {
		shaderWorkBuffer = ebiten.NewImage(width, height)
	}
}

func Init() {
	numSamples := config.Config.ReadBufferSize / 2
	FFTBuffer = make([]float64, numSamples)
	lowCutOff := flag.Float64("lo", 0.0, "low frequency cutoff fraction (discernable details are around 0.001 increments for a 4096 buffer size and 192kHz sample rate)")
	highCutOff := flag.Float64("hi", 1.0, "high frequency cutoff fraction (discernable details are around 0.001 increments for a 4096 buffer size and 192kHz sample rate)")
	useRightChannel = flag.Bool("right", false, "Use the right channel instead of the left for the single axis oscilloscope")
	mixChannels = flag.Bool("mix", false, "Mix channels instead of just using a single channel for the single axis oscilloscope")
	beatDetectOverride = flag.Bool("beatdetect", false, "Enable beat detection (bypassing configuration)")
	peakDetectOverride = flag.Bool("peakdetect", false, "Enable peak detection (bypassing configuration)")

	overrideGain := flag.Float64("gain", -1.0, "override gain multiplier")

	overrideWidth := flag.Int("width", int(config.Config.WindowWidth), "override window width")
	overrideHeight := flag.Int("height", int(config.Config.WindowHeight), "override window height")

	overrideX = flag.Int("x", -1, "override starting x coordinate of the center of the window; x=0 corresponds to the center of the screen")
	overrideY = flag.Int("y", -1, "override starting y coordinate of the center of the window; y=0 corresponds to the center of the screen")

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
	if config.Config.UseBetterPeakDetectionAlgorithm {
		complexFFTBufferFlipped = make([]complex128, numSamples)
		align.Init()
	}
	if config.FiltersApplied {
		XYComplexFFTBufferL = make([]complex128, numSamples)
		XYComplexFFTBufferR = make([]complex128, numSamples)
	}
	config.Config.WindowWidth = int32(*overrideWidth)
	config.Config.WindowHeight = int32(*overrideHeight)

	if *overrideGain != -1.0 {
		config.Config.Gain = float32(*overrideGain)
	}
}

func main() {
	config.Init()
	audio.Init()
	Init()
	fonts.Init()
	icons.Init()
	go audio.Start()
	if config.Config.BeatDetect || *beatDetectOverride {
		go beatdetect.Start()
	}
	if config.Config.ShowMPRIS {
		go media.Start()
	}
	if config.Config.UseShaders {
		shaders.Init()
	}
	bars.Init()
	kaiser.Init()
	ebiten.SetWindowIcon([]image.Image{icons.WindowIcon48, icons.WindowIcon32, icons.WindowIcon16})
	ebiten.SetWindowSize(int(config.Config.WindowWidth), int(config.Config.WindowHeight))
	ebiten.SetWindowTitle("xyosc")
	ebiten.SetWindowMousePassthrough(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(int(config.Config.TargetFPS))
	ebiten.SetWindowDecorated(false)
	screenW, screenH := ebiten.Monitor().Size()
	if *overrideX == -1 {
		*overrideX = screenW/2 - int(config.Config.WindowWidth)/2
	} else {
		*overrideX = screenW/2 + *overrideX - int(config.Config.WindowWidth)/2
	}
	if *overrideY == -1 {
		*overrideY = screenH/2 - int(config.Config.WindowHeight)/2
	} else {
		*overrideY = screenH/2 + *overrideY - int(config.Config.WindowHeight)/2
	}
	ebiten.SetWindowPosition(*overrideX, *overrideY)
	ebiten.SetVsyncEnabled(true)
	InitBuffersAtSize(int(config.Config.WindowWidth), int(config.Config.WindowHeight))
	gameOptions := ebiten.RunGameOptions{SingleThread: true, ScreenTransparent: !config.Config.DisableTransparency}
	if config.Config.DisableTransparency {
		gameOptions.ScreenTransparent = false
	}

	startTime = time.Now()
	if err := ebiten.RunGameWithOptions(&Game{}, &gameOptions); err != nil {
		log.Fatal(err)
	}
}
