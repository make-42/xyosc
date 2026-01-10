package bars

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/argusdusty/gofft"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/interpolate"
	"xyosc/kaiser"
	"xyosc/utils"
)

var InterpolatedBarsPos []float64
var InterpolatedBarsVel []float64
var TargetBarsPos []float64

var InterpolatedBarsPhasePos []float64
var InterpolatedBarsPhaseVel []float64
var TargetBarsPhasePos []float64

var InterpolatedMaxVolume float64 = 2000. // sane starting point

var InterpolatedPeakFreqCursorX float64
var InterpolatedPeakFreqCursorY float64
var InterpolatedPeakFreqCursorVelX float64
var InterpolatedPeakFreqCursorVelY float64
var TargetPeakFreqCursorX float64
var TargetPeakFreqCursorY float64
var PeakFreqCursorVal float64

func Init() {
	barCount := (float64(config.Config.Window.Width) - config.Config.Bars.PaddingEdge*2 + config.Config.Bars.PaddingBetween) / (config.Config.Bars.PaddingBetween + config.Config.Bars.Width)
	TargetBarsPos = make([]float64, int(barCount))
	InterpolatedBarsPos = make([]float64, int(barCount))
	InterpolatedBarsVel = make([]float64, int(barCount))
	if config.Config.Bars.PhaseColors.Enable {
		TargetBarsPhasePos = make([]float64, int(barCount))
		InterpolatedBarsPhasePos = make([]float64, int(barCount))
		InterpolatedBarsPhaseVel = make([]float64, int(barCount))
	}
}

func CalcBars(inputArray *[]complex128, lowCutOffFrac float64, highCutOffFrac float64) {
	if config.Config.Bars.UseWindowFn {
		for i := uint32(0); i < config.Config.Buffers.ReadBufferSize/2; i++ {
			(*inputArray)[i] = complex(real((*inputArray)[i])*kaiser.WindowBuffer[i], 0)
		}
	}
	err := gofft.FFT(*inputArray)
	utils.CheckError(tracerr.Wrap(err))
	numBars := float64(len(TargetBarsPos))
	complexTot := complex128(0.0)
	complexSum := complex128(0.0)
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
			if config.Config.Bars.PeakCursor.Enable {
				if val >= peakFreqVal {
					peakFreqBin = int(nthBar)
					peakFreqVal = val
					peakFreq = float64(X) / float64(len(*inputArray)) * float64(config.Config.Audio.SampleRate)
				}
			}
			if (nthBar+1)/numBars <= frac {
				for (nthBar+1)/numBars <= frac {
					if nSamples != 0 {
						TargetBarsPos[int(nthBar)] = sum / nSamples
						if config.Config.Bars.PhaseColors.Enable {
							TargetBarsPhasePos[int(nthBar)] = cmplx.Phase(complexSum)
						}
					}
					nthBar++
				}
				sum = 0.0
				complexSum = 0.0
				nSamples = 0.0
			}
		}
	}
	if nSamples != 0 {
		TargetBarsPos[int(nthBar)] = sum / nSamples
	}
	if config.Config.Bars.PeakCursor.Enable && peakFreqVal != 0 {
		x, y, w, h := ComputeBarLayout(peakFreqBin)
		TargetPeakFreqCursorX = min(max(x+w/2, 0), float64(config.Config.Window.Width)-config.Config.Bars.PeakCursor.BGWidth)
		TargetPeakFreqCursorY = min(max(y+h-config.Config.Bars.PeakCursor.TextSize-2*config.Config.Bars.PeakCursor.BGPadding, 0), float64(config.Config.Window.Height)-config.Config.Bars.PeakCursor.TextSize-2*config.Config.Bars.PeakCursor.BGPadding)
		PeakFreqCursorVal = peakFreq
	}
	if config.Config.Bars.PhaseColors.Enable {
		midPhase := cmplx.Phase(complexTot)
		for x := range len(TargetBarsPos) {
			TargetBarsPhasePos[x] = math.Mod(TargetBarsPhasePos[x]-midPhase+3*math.Pi, 2*math.Pi) - math.Pi
		}
	}
}

func ComputeBarLayout(barIndex int) (x float64, y float64, w float64, h float64) {
	if config.FiltersApplied && config.Config.FilterInfo.Enable {
		return (config.Config.Bars.PaddingEdge) + float64(barIndex)*(config.Config.Bars.Width+config.Config.Bars.PaddingBetween), float64(config.Config.Window.Height) - (config.Config.Bars.PaddingEdge) - (config.Config.FilterInfo.TextSize) - (config.Config.FilterInfo.TextPaddingBottom), (config.Config.Bars.Width), -(float64(config.Config.Window.Height) - 2*(config.Config.Bars.PaddingEdge) - (config.Config.FilterInfo.TextSize) - (config.Config.FilterInfo.TextPaddingBottom)) * (InterpolatedBarsPos[barIndex]) / (InterpolatedMaxVolume)
	} else {
		return (config.Config.Bars.PaddingEdge) + float64(barIndex)*(config.Config.Bars.Width+config.Config.Bars.PaddingBetween), float64(config.Config.Window.Height) - (config.Config.Bars.PaddingEdge), (config.Config.Bars.Width), -(float64(config.Config.Window.Height) - 2*(config.Config.Bars.PaddingEdge)) * (InterpolatedBarsPos[barIndex]) / (InterpolatedMaxVolume)
	}
}

func InterpolateBars(deltaTime float64) {
	if config.Config.Bars.AutoGain.Enable {
		max := 0.0
		if len(TargetBarsPos) != 0 {
			max = TargetBarsPos[0]
		}

		for _, value := range TargetBarsPos {
			max = math.Max(max, value)
		}
		InterpolatedMaxVolume += (max - InterpolatedMaxVolume) * deltaTime * config.Config.Bars.AutoGain.Speed
		InterpolatedMaxVolume = math.Max(config.Config.Bars.AutoGain.MinVolume, InterpolatedMaxVolume)
	}

	if config.Config.Bars.PeakCursor.InterpolatePos.Enable && config.Config.Bars.PeakCursor.Enable {
		interpolate.Interpolate(deltaTime, TargetPeakFreqCursorX, &InterpolatedPeakFreqCursorX, &InterpolatedPeakFreqCursorVelX, config.Config.Bars.PeakCursor.InterpolatePos)
		interpolate.Interpolate(deltaTime, TargetPeakFreqCursorY, &InterpolatedPeakFreqCursorY, &InterpolatedPeakFreqCursorVelY, config.Config.Bars.PeakCursor.InterpolatePos)
	} else {
		InterpolatedPeakFreqCursorX = TargetPeakFreqCursorX
		InterpolatedPeakFreqCursorY = TargetPeakFreqCursorY
	}
	InterpolatedPeakFreqCursorY = min(max(InterpolatedPeakFreqCursorY, 0), float64(config.Config.Window.Height)-2*config.Config.Bars.PeakCursor.BGPadding)
	InterpolatedPeakFreqCursorX = min(max(InterpolatedPeakFreqCursorX, 0), float64(config.Config.Window.Width)-config.Config.Bars.PeakCursor.BGWidth)

	if config.Config.Bars.Interpolate.Enable {
		for i := range TargetBarsPos {
			interpolate.Interpolate(deltaTime, TargetBarsPos[i], &InterpolatedBarsPos[i], &InterpolatedBarsVel[i], config.Config.Bars.Interpolate)
		}
	} else {
		copy(InterpolatedBarsPos, TargetBarsPos)
	}

	if config.Config.Bars.PhaseColors.Enable {
		if config.Config.Bars.PhaseColors.Interpolate.Enable {
			for i := range TargetBarsPhasePos {
				interpolate.InterpolateAngle(deltaTime, TargetBarsPhasePos[i], &InterpolatedBarsPhasePos[i], &InterpolatedBarsPhaseVel[i], config.Config.Bars.PhaseColors.Interpolate)
			}
		} else {
			copy(InterpolatedBarsPhasePos, TargetBarsPhasePos)
		}
	}
}

func CalcNote(freq float64) int {
	return int(math.Round(12*math.Log2(freq/config.Config.Bars.PeakCursor.RefNoteFreq))) - 3
}

func NoteDisplayName(note int) string {
	octave := 4 + (note / 12)
	name := ([]string{"C ", "C#", "D ", "D#", "E ", "F ", "F#", "G ", "G#", "A ", "A#", "B "})[(note-(note/12-1)*12)%12]
	return fmt.Sprintf("%s%d", name, octave)
}
