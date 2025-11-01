package vu

import (
	"math"
	"slices"
	"time"
	"xyosc/config"
	"xyosc/kaiser"
	"xyosc/utils"

	"github.com/argusdusty/gofft"
)

var LoudnessLTarget = 0.0
var LoudnessLPos = 0.0
var LoudnessLVel = 0.0
var LoudnessRTarget = 0.0
var LoudnessRPos = 0.0
var LoudnessRVel = 0.0

var LoudnessLPeakTarget = 0.0
var LoudnessLPeakPos = 0.0
var LoudnessLPeakVel = 0.0
var LoudnessRPeakTarget = 0.0
var LoudnessRPeakPos = 0.0
var LoudnessRPeakVel = 0.0

var LoudnessPeakL []float64
var LoudnessPeakR []float64
var LoudnessCursorPos = 0
var startTime time.Time

func Init() {
	LoudnessPeakL = make([]float64, int(float64(config.Config.TargetFPS)*config.Config.VUPeakHistorySeconds))
	LoudnessPeakR = make([]float64, int(float64(config.Config.TargetFPS)*config.Config.VUPeakHistorySeconds))
	startTime = time.Now()
}

func CommitPeakPoint() {
	newCursorPos := int(float64(config.Config.TargetFPS) * (float64(time.Since(startTime).Nanoseconds()) / 1e9))
	newOffset := min(newCursorPos-LoudnessCursorPos, len(LoudnessPeakL))
	LoudnessCursorPos = newCursorPos
	for x := range int(newOffset) {
		idx := (newCursorPos - x + len(LoudnessPeakL)) % len(LoudnessPeakL)
		LoudnessPeakL[idx] = LoudnessLTarget
		LoudnessPeakR[idx] = LoudnessRTarget
	}
	LoudnessLPeakTarget = slices.Max(LoudnessPeakL)
	LoudnessRPeakTarget = slices.Max(LoudnessPeakR)
}

func CalcLoudness(inputArray *[]complex128, lowCutFrac, hiCutFrac float64) float64 { // in out units
	vol := 0.
	if lowCutFrac != 0 || hiCutFrac != 1.0 {
		if config.Config.BarsUseWindow {
			for i := uint32(0); i < config.Config.ReadBufferSize/2; i++ {
				(*inputArray)[i] = complex(real((*inputArray)[i])*kaiser.WindowBuffer[i], 0)
			}
		}
		err := gofft.FFT(*inputArray)
		utils.CheckError(err)
		for x := range len(*inputArray) {
			frac := float64(x+1) / float64(len(*inputArray))
			if lowCutFrac <= frac && frac < hiCutFrac {
				(*inputArray)[x] = 0
			}
		}
		err = gofft.IFFT(*inputArray)
		utils.CheckError(err)
		for x := range len(*inputArray) {
			val := real((*inputArray)[x])*real((*inputArray)[x]) + imag((*inputArray)[x])*imag((*inputArray)[x])
			vol = max(val, vol)
		}
	} else {
		for x := range len(*inputArray) {
			val := real((*inputArray)[x]) * real((*inputArray)[x])
			vol = max(val, vol)
		}
	}
	if config.Config.VULogScale {
		vol = ((max(math.Log10(vol)*10, config.Config.VULogScaleMinDB)) - config.Config.VULogScaleMinDB) / (config.Config.VULogScaleMaxDB - config.Config.VULogScaleMinDB)
	} else {
		vol = vol / config.Config.VULinScaleMax
	}
	return vol
}

func ComputeBarLayout(barIndex int, vol float64) (x float64, y float64, w float64, h float64) {
	barsWidth := (float64(config.Config.WindowWidth)-config.Config.VUPaddingBetween)/2 - config.Config.VUPaddingEdge

	if config.FiltersApplied && config.Config.ShowFilterInfo {
		return (config.Config.VUPaddingEdge) + float64(barIndex)*(barsWidth+config.Config.VUPaddingBetween), float64(config.Config.WindowHeight) - (config.Config.VUPaddingEdge) - (config.Config.FilterInfoTextSize) - (config.Config.FilterInfoTextPaddingBottom), (barsWidth), -(float64(config.Config.WindowHeight) - 2*(config.Config.BarsPaddingEdge) - (config.Config.FilterInfoTextSize) - (config.Config.FilterInfoTextPaddingBottom)) * vol
	} else {
		return (config.Config.VUPaddingEdge) + float64(barIndex)*(barsWidth+config.Config.VUPaddingBetween), float64(config.Config.WindowHeight) - (config.Config.VUPaddingEdge), (barsWidth), -(float64(config.Config.WindowHeight) - 2*(config.Config.VUPaddingEdge)) * vol
	}
}

func Interpolate(deltaTime float64) {
	if config.Config.VUInterpolate {
		LoudnessLPos += (LoudnessLTarget - LoudnessLPos) * min(1.0, deltaTime*config.Config.VUInterpolateDirect)
		LoudnessLVel += (LoudnessLTarget - LoudnessLPos) * deltaTime * config.Config.VUInterpolateAccel
		LoudnessLVel -= LoudnessLVel * min(1.0, deltaTime*config.Config.VUInterpolateDrag)
		LoudnessLPos += LoudnessLVel * deltaTime
		LoudnessRPos += (LoudnessRTarget - LoudnessRPos) * min(1.0, deltaTime*config.Config.VUInterpolateDirect)
		LoudnessRVel += (LoudnessRTarget - LoudnessRPos) * deltaTime * config.Config.VUInterpolateAccel
		LoudnessRVel -= LoudnessRVel * min(1.0, deltaTime*config.Config.VUInterpolateDrag)
		LoudnessRPos += LoudnessRVel * deltaTime
	} else {
		LoudnessLPos = LoudnessLTarget
		LoudnessRPos = LoudnessRTarget
	}
	if config.Config.VUPeak {
		if config.Config.VUPeakInterpolate {
			LoudnessLPeakPos += (LoudnessLPeakTarget - LoudnessLPeakPos) * min(1.0, deltaTime*config.Config.VUInterpolateDirect)
			LoudnessLPeakVel += (LoudnessLPeakTarget - LoudnessLPeakPos) * deltaTime * config.Config.VUInterpolateAccel
			LoudnessLPeakVel -= LoudnessLPeakVel * min(1.0, deltaTime*config.Config.VUInterpolateDrag)
			LoudnessLPeakPos += LoudnessLPeakVel * deltaTime
			LoudnessRPeakPos += (LoudnessRPeakTarget - LoudnessRPeakPos) * min(1.0, deltaTime*config.Config.VUInterpolateDirect)
			LoudnessRPeakVel += (LoudnessRPeakTarget - LoudnessRPeakPos) * deltaTime * config.Config.VUInterpolateAccel
			LoudnessRPeakVel -= LoudnessRPeakVel * min(1.0, deltaTime*config.Config.VUInterpolateDrag)
			LoudnessRPeakPos += LoudnessRPeakVel * deltaTime
		} else {
			LoudnessLPeakPos = LoudnessLPeakTarget
			LoudnessRPeakPos = LoudnessRPeakTarget
		}
	}
}
