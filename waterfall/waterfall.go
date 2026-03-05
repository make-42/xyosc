package waterfall

import (
	"image/color"
	"math"

	"github.com/argusdusty/gofft"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/soypat/colorspace"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/interpolate"
	"xyosc/kaiser"
	"xyosc/utils"
)

var ImageA *ebiten.Image // Basically have two buffers and write sequentially to each and then swap when one is completely used up
var ImageB *ebiten.Image

var CursorPosLastDraw = 0
var CursorPos = 0.
var CursorImage = 0 // 0 for A; 1 for B

var TargetPeakVal float64
var InterpolatedPeakVal float64
var InterpolatedPeakValVel float64

func Init(width, height int) {
	ImageA = ebiten.NewImage(width, height)
	ImageB = ebiten.NewImage(width, height)
	CursorPos = 0.
	CursorImage = 0
	CursorPosLastDraw = 0
}

func colorMap(val float64) color.Color {
	if InterpolatedPeakVal <= 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	t := float32(math.Min(val/InterpolatedPeakVal, 1.))

	c1 := color.RGBA{config.ThirdColor.R, config.ThirdColor.G, config.ThirdColor.B, 0}
	c2 := color.RGBA{config.FirstColor.R, config.FirstColor.G, config.FirstColor.B, 255}

	c := colorspace.LerpOKLCH(c1, c2, t)
	r16, g16, b16, _ := c.RGBA()
	r8 := uint8(r16 >> 8)
	g8 := uint8(g16 >> 8)
	b8 := uint8(b16 >> 8)
	a8 := uint8(float32(c1.A)*(1-t) + float32(c2.A)*t)
	af := float64(a8) / 255.0
	return color.RGBA{
		R: uint8(float64(r8) * af),
		G: uint8(float64(g8) * af),
		B: uint8(float64(b8) * af),
		A: a8,
	}
}

func CalcWaterfallAndDraw(screen *ebiten.Image, inputArray *[]complex128, lowCutOffFrac float64, highCutOffFrac float64, deltaTime float64) {
	if config.Config.Bars.UseWindowFn {
		for i := uint32(0); i < config.Config.Buffers.ReadBufferSize/2; i++ {
			(*inputArray)[i] = complex(real((*inputArray)[i])*kaiser.WindowBuffer[i], 0)
		}
	}
	err := gofft.FFT(*inputArray)
	utils.CheckError(tracerr.Wrap(err))
	CursorPos += deltaTime * config.Config.Waterfall.Speed
	if int(CursorPos) != CursorPosLastDraw {
		newPxToDraw := int(CursorPos) - CursorPosLastDraw
		imgSize := ImageA.Bounds().Size()
		imgWidth, imgHeight := imgSize.X, imgSize.Y
		numBars := float64(imgWidth)
		complexTot := complex128(0.0)
		complexSum := complex128(0.0)
		sum := 0.0
		nSamples := 0.0
		nthBar := 0.0
		prevVal := -1.0
		lowCutOffFracAdj := math.Max(1/float64(len(*inputArray)), lowCutOffFrac)
		TargetPeakVal = TargetPeakVal * math.Pow(config.Config.Waterfall.TargetPeakValDecay, deltaTime)
		for x := range len(*inputArray) - 1 {
			X := x + 1 // Offset by 1 to avoid having an infinite log scale
			if X <= int(float64(len(*inputArray))*highCutOffFrac) || X >= int(float64(len(*inputArray))*lowCutOffFracAdj) {
				frac := (math.Log2(float64(X)/float64(len(*inputArray))) - math.Log2(lowCutOffFracAdj)) / (math.Log2(highCutOffFrac) - math.Log2(lowCutOffFracAdj))
				val := math.Sqrt(real((*inputArray)[X])*real((*inputArray)[X]) + imag((*inputArray)[X])*imag((*inputArray)[X]))
				complexVal := (*inputArray)[X]
				if config.Config.Bars.PreserveParsevalEnergy {
					val = val * math.Sqrt(float64(X))
					complexVal = complexVal * complex(math.Sqrt(float64(X)), 0)
				}
				if X >= int((config.Config.Bars.PreventLeakageAboveFreq/float64(config.Config.Audio.SampleRate))*float64(len(*inputArray))) {
					val = 0
					complexVal = 0
				}
				complexTot += complexVal
				complexSum += complexVal
				sum += val
				nSamples++
				if frac >= 1 {
					break
				}
				if (nthBar+1)/numBars <= frac {
					startNum := nthBar
					numToInterp := frac * numBars
					newVal := sum / nSamples
					if prevVal < 0 {
						prevVal = newVal
					}
					for (nthBar+1)/numBars <= frac {
						if nSamples != 0 {
							i := int(nthBar)
							t := (nthBar - startNum) / (numToInterp - startNum)
							val := newVal*t + prevVal*(1-t)
							if val > TargetPeakVal {
								TargetPeakVal = val
							}
							cl := colorMap(val)
							for j := 0; j < newPxToDraw; j++ {
								y := (CursorPosLastDraw + j) % imgHeight
								currImage := (CursorImage + (CursorPosLastDraw+j)/imgHeight) % 2
								if currImage == 0 {
									ImageA.Set(i, y, cl)
								} else {
									ImageB.Set(i, y, cl)
								}
							}
						}
						nthBar++
					}
					prevVal = newVal
					sum = 0.0
					complexSum = 0.0
					nSamples = 0.0
				}
			}
		}

		CursorImage = (CursorImage + (CursorPosLastDraw+newPxToDraw)/imgHeight) % 2
		CursorPosLastDraw = (CursorPosLastDraw + newPxToDraw) % imgHeight
		CursorPos = math.Mod(CursorPos, float64(imgHeight))
	}

	if config.Config.Waterfall.InterpolatePeakVal.Enable {
		interpolate.Interpolate(deltaTime, TargetPeakVal, &InterpolatedPeakVal, &InterpolatedPeakValVel, config.Config.Waterfall.InterpolatePeakVal)
	} else {
		InterpolatedPeakVal = TargetPeakVal
	}

	opA := &ebiten.DrawImageOptions{}
	opB := &ebiten.DrawImageOptions{}

	screenSize := screen.Bounds().Size()
	screenHeight := screenSize.Y
	if CursorImage == 0 {
		// ImageA is being written to, ImageB is the "older" completed buffer
		// Draw ImageA from CursorPosLastDraw downward (the recent rows)
		opA.GeoM.Translate(0, float64(screenHeight-CursorPosLastDraw))
		// Draw ImageB above it (the older rows)
		opB.GeoM.Translate(0, float64(-CursorPosLastDraw))

		screen.DrawImage(ImageB, opB)
		screen.DrawImage(ImageA, opA)
	} else {
		// ImageB is being written to, ImageA is the "older" completed buffer
		opB.GeoM.Translate(0, float64(screenHeight-CursorPosLastDraw))
		opA.GeoM.Translate(0, float64(-CursorPosLastDraw))

		screen.DrawImage(ImageA, opA)
		screen.DrawImage(ImageB, opB)
	}
}
