package bars

import (
	"math"
	"xyosc/config"
	"xyosc/utils"

	"github.com/argusdusty/gofft"
)

var InterpolatedBarsPos []float64
var InterpolatedBarsVel []float64
var TargetBarsPos []float64

func Init() {
	barCount := (float64(config.Config.WindowWidth) - config.Config.BarsPaddingEdge*2 + config.Config.BarsPaddingBetween) / (config.Config.BarsPaddingBetween + config.Config.BarsWidth)
	TargetBarsPos = make([]float64, int(barCount))
	InterpolatedBarsPos = make([]float64, int(barCount))
	InterpolatedBarsVel = make([]float64, int(barCount))
}

func CalcBars(inputArray *[]complex128, lowCutOffFrac float64, highCutOffFrac float64) {
	err := gofft.FFT(*inputArray)
	utils.CheckError(err)
	numBars := float64(len(TargetBarsPos))
	sum := 0.0
	nSamples := 0.0
	nthBar := 0.0
	lowCutOffFracAdj := math.Max(1/float64(len(*inputArray)), lowCutOffFrac)
	for x := range len(*inputArray) - 1 {
		X := x + 1 // Offset by 1 to avoid having an infinite log scale
		if X <= int(float64(len(*inputArray))*highCutOffFrac) || X >= int(float64(len(*inputArray))*lowCutOffFracAdj) {
			frac := (math.Log2(float64(X)/float64(len(*inputArray))) - math.Log2(lowCutOffFracAdj)) / (math.Log2(highCutOffFrac) - math.Log2(lowCutOffFracAdj))
			sum += math.Sqrt(real((*inputArray)[X])*real((*inputArray)[X]) + imag((*inputArray)[X])*imag((*inputArray)[X]))
			nSamples++
			if frac >= 1 {
				break
			}
			if (nthBar+1)/numBars <= frac {
				if nSamples != 0 {
					TargetBarsPos[int(nthBar)] = sum / nSamples
				}
				nthBar++
				sum = 0.0
				nSamples = 0.0
			}
		}
	}
	if nSamples != 0 {
		TargetBarsPos[int(nthBar)] = sum / nSamples
	}
}
