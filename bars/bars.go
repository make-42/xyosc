package bars

import (
	"fmt"
	"math"
	"xyosc/config"
	"xyosc/kaiser"
	"xyosc/utils"

	"github.com/argusdusty/gofft"
)

var InterpolatedBarsPos []float64
var InterpolatedBarsVel []float64
var TargetBarsPos []float64
var InterpolatedMaxVolume float64 = 2000. // sane starting point

var InterpolatedPeakFreqCursorX float64
var InterpolatedPeakFreqCursorY float64
var InterpolatedPeakFreqCursorVelX float64
var InterpolatedPeakFreqCursorVelY float64
var TargetPeakFreqCursorX float64
var TargetPeakFreqCursorY float64
var PeakFreqCursorVal float64

func Init() {
	barCount := (float64(config.Config.WindowWidth) - config.Config.BarsPaddingEdge*2 + config.Config.BarsPaddingBetween) / (config.Config.BarsPaddingBetween + config.Config.BarsWidth)
	TargetBarsPos = make([]float64, int(barCount))
	InterpolatedBarsPos = make([]float64, int(barCount))
	InterpolatedBarsVel = make([]float64, int(barCount))
}

func CalcBars(inputArray *[]complex128, lowCutOffFrac float64, highCutOffFrac float64) {
	if config.Config.BarsUseWindow {
		for i := uint32(0); i < config.Config.ReadBufferSize/2; i++ {
			(*inputArray)[i] = complex(real((*inputArray)[i])*kaiser.WindowBuffer[i], 0)
		}
	}
	err := gofft.FFT(*inputArray)
	utils.CheckError(err)
	numBars := float64(len(TargetBarsPos))
	sum := 0.0
	nSamples := 0.0
	nthBar := 0.0
	lowCutOffFracAdj := math.Max(1/float64(len(*inputArray)), lowCutOffFrac)
	peakFreq, peakFreqBin, peakFreqVal := 0., 0, 0.
	for x := range len(*inputArray) - 1 {
		X := x + 1 // Offset by 1 to avoid having an infinite log scale
		if X <= int(float64(len(*inputArray))*highCutOffFrac) || X >= int(float64(len(*inputArray))*lowCutOffFracAdj) {
			frac := (math.Log2(float64(X)/float64(len(*inputArray))) - math.Log2(lowCutOffFracAdj)) / (math.Log2(highCutOffFrac) - math.Log2(lowCutOffFracAdj))
			val := math.Sqrt(real((*inputArray)[X])*real((*inputArray)[X]) + imag((*inputArray)[X])*imag((*inputArray)[X]))
			if config.Config.BarsPreserveParsevalEnergy {
				val = val * math.Sqrt(float64(X))
			}
			if X >= int((config.Config.BarsPreventSpectralLeakageAboveFreq/float64(config.Config.SampleRate))*float64(len(*inputArray))) {
				val = 0
			}
			sum += val
			nSamples++
			if frac >= 1 {
				break
			}
			if config.Config.BarsPeakFreqCursor {
				if val >= peakFreqVal {
					peakFreqBin = int(nthBar)
					peakFreqVal = val
					peakFreq = float64(X) / float64(len(*inputArray)) * float64(config.Config.SampleRate)
				}
			}
			if (nthBar+1)/numBars <= frac {
				for (nthBar+1)/numBars <= frac {
					if nSamples != 0 {
						TargetBarsPos[int(nthBar)] = sum / nSamples
					}
					nthBar++
				}
				sum = 0.0
				nSamples = 0.0
			}
		}
	}
	if nSamples != 0 {
		TargetBarsPos[int(nthBar)] = sum / nSamples
	}
	if config.Config.BarsPeakFreqCursor && peakFreqVal != 0 {
		x, y, w, h := ComputeBarLayout(peakFreqBin)
		TargetPeakFreqCursorX = min(max(x+w/2, 0), float64(config.Config.WindowWidth)-config.Config.BarsPeakFreqCursorBGWidth)
		TargetPeakFreqCursorY = min(max(y+h-config.Config.BarsPeakFreqCursorTextSize-2*config.Config.BarsPeakFreqCursorBGPadding, 0), float64(config.Config.WindowHeight)-config.Config.BarsPeakFreqCursorTextSize-2*config.Config.BarsPeakFreqCursorBGPadding)
		PeakFreqCursorVal = peakFreq
	}
}

func ComputeBarLayout(barIndex int) (x float64, y float64, w float64, h float64) {
	if config.FiltersApplied && config.Config.ShowFilterInfo {
		return (config.Config.BarsPaddingEdge) + float64(barIndex)*(config.Config.BarsWidth+config.Config.BarsPaddingBetween), float64(config.Config.WindowHeight) - (config.Config.BarsPaddingEdge) - (config.Config.FilterInfoTextSize) - (config.Config.FilterInfoTextPaddingBottom), (config.Config.BarsWidth), -(float64(config.Config.WindowHeight) - 2*(config.Config.BarsPaddingEdge) - (config.Config.FilterInfoTextSize) - (config.Config.FilterInfoTextPaddingBottom)) * (InterpolatedBarsPos[barIndex]) / (InterpolatedMaxVolume)
	} else {
		return (config.Config.BarsPaddingEdge) + float64(barIndex)*(config.Config.BarsWidth+config.Config.BarsPaddingBetween), float64(config.Config.WindowHeight) - (config.Config.BarsPaddingEdge), (config.Config.BarsWidth), -(float64(config.Config.WindowHeight) - 2*(config.Config.BarsPaddingEdge)) * (InterpolatedBarsPos[barIndex]) / (InterpolatedMaxVolume)
	}
}

func InterpolateBars(deltaTime float64) {
	if config.Config.BarsAutoGain {
		max := 0.0
		if len(TargetBarsPos) != 0 {
			max = TargetBarsPos[0]
		}

		for _, value := range TargetBarsPos {
			max = math.Max(max, value)
		}
		InterpolatedMaxVolume += (max - InterpolatedMaxVolume) * deltaTime * config.Config.BarsAutoGainSpeed
		InterpolatedMaxVolume = math.Max(config.Config.BarsAutoGainMinVolume, InterpolatedMaxVolume)
	}

	if config.Config.BarsPeakFreqCursorInterpolatePos {
		InterpolatedPeakFreqCursorX += (TargetPeakFreqCursorX - InterpolatedPeakFreqCursorX) * min(1.0, deltaTime*config.Config.BarsPeakFreqCursorInterpolateDirect)
		InterpolatedPeakFreqCursorVelX += (TargetPeakFreqCursorX - InterpolatedPeakFreqCursorX) * deltaTime * config.Config.BarsPeakFreqCursorInterpolateAccel
		InterpolatedPeakFreqCursorVelX -= InterpolatedPeakFreqCursorVelX * min(1.0, deltaTime*config.Config.BarsPeakFreqCursorInterpolateDrag)
		InterpolatedPeakFreqCursorX += InterpolatedPeakFreqCursorVelX * deltaTime
		InterpolatedPeakFreqCursorY += (TargetPeakFreqCursorY - InterpolatedPeakFreqCursorY) * min(1.0, deltaTime*config.Config.BarsPeakFreqCursorInterpolateDirect)
		InterpolatedPeakFreqCursorVelY += (TargetPeakFreqCursorY - InterpolatedPeakFreqCursorY) * deltaTime * config.Config.BarsPeakFreqCursorInterpolateAccel
		InterpolatedPeakFreqCursorVelY -= InterpolatedPeakFreqCursorVelY * min(1.0, deltaTime*config.Config.BarsPeakFreqCursorInterpolateDrag)
		InterpolatedPeakFreqCursorY += InterpolatedPeakFreqCursorVelY * deltaTime
	} else {
		InterpolatedPeakFreqCursorX = TargetPeakFreqCursorX
		InterpolatedPeakFreqCursorY = TargetPeakFreqCursorY
	}
	InterpolatedPeakFreqCursorY = min(max(InterpolatedPeakFreqCursorY, 0), float64(config.Config.WindowHeight)-2*config.Config.BarsPeakFreqCursorBGPadding)
	InterpolatedPeakFreqCursorX = min(max(InterpolatedPeakFreqCursorX, 0), float64(config.Config.WindowWidth)-config.Config.BarsPeakFreqCursorBGWidth)

	if config.Config.BarsInterpolatePos && config.Config.BarsPeakFreqCursor {
		for i := range TargetBarsPos {
			InterpolatedBarsPos[i] += (TargetBarsPos[i] - InterpolatedBarsPos[i]) * min(1.0, deltaTime*config.Config.BarsInterpolateDirect)
			InterpolatedBarsVel[i] += (TargetBarsPos[i] - InterpolatedBarsPos[i]) * deltaTime * config.Config.BarsInterpolateAccel
			InterpolatedBarsVel[i] -= InterpolatedBarsVel[i] * min(1.0, deltaTime*config.Config.BarsInterpolateDrag)
			InterpolatedBarsPos[i] += InterpolatedBarsVel[i] * deltaTime
		}
	} else {
		copy(InterpolatedBarsPos, TargetBarsPos)
	}
}

func CalcNote(freq float64) int {
	return int(math.Round(12*math.Log2(freq/config.Config.BarsPeakFreqCursorTextDisplayNoteRefFreq))) - 3
}

func NoteDisplayName(note int) string {
	octave := 4 + (note / 12)
	name := ([]string{"C ", "C#", "D ", "D#", "E ", "F ", "F#", "G ", "G#", "A ", "A#", "B "})[(note-(note/12-1)*12)%12]
	return fmt.Sprintf("%s%d", name, octave)
}
