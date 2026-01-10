package vu

import (
	"math"
	"slices"
	"time"

	"github.com/argusdusty/gofft"
	"github.com/ztrue/tracerr"

	"xyosc/config"
	"xyosc/interpolate"
	"xyosc/kaiser"
	"xyosc/utils"
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
	LoudnessPeakL = make([]float64, int(float64(config.Config.App.TargetFPS)*config.Config.VU.Peak.HistorySeconds))
	LoudnessPeakR = make([]float64, int(float64(config.Config.App.TargetFPS)*config.Config.VU.Peak.HistorySeconds))
	startTime = time.Now()
}

func CommitPeakPoint() {
	newCursorPos := int(float64(config.Config.App.TargetFPS) * (float64(time.Since(startTime).Nanoseconds()) / 1e9))
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
		if config.Config.Bars.UseWindowFn {
			for i := uint32(0); i < config.Config.Buffers.ReadBufferSize/2; i++ {
				(*inputArray)[i] = complex(real((*inputArray)[i])*kaiser.WindowBuffer[i], 0)
			}
		}
		err := gofft.FFT(*inputArray)
		utils.CheckError(tracerr.Wrap(err))
		for x := range len(*inputArray) {
			frac := float64(x+1) / float64(len(*inputArray))
			if lowCutFrac <= frac && frac < hiCutFrac {
				(*inputArray)[x] = 0
			}
		}
		err = gofft.IFFT(*inputArray)
		utils.CheckError(tracerr.Wrap(err))
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
	if config.Config.VU.LogScale {
		vol = ((max(math.Log10(vol)*10, config.Config.VU.LogMinDB)) - config.Config.VU.LogMinDB) / (config.Config.VU.LogMaxDB - config.Config.VU.LogMinDB)
	} else {
		vol = vol / config.Config.VU.LinMax
	}
	return vol
}

func ComputeBarLayout(barIndex int, vol float64) (x float64, y float64, w float64, h float64) {
	barsWidth := (float64(config.Config.Window.Width)-config.Config.VU.PaddingBetween)/2 - config.Config.VU.PaddingEdge

	if config.FiltersApplied && config.Config.FilterInfo.Enable {
		return (config.Config.VU.PaddingEdge) + float64(barIndex)*(barsWidth+config.Config.VU.PaddingBetween), float64(config.Config.Window.Height) - (config.Config.VU.PaddingEdge) - (config.Config.FilterInfo.TextSize) - (config.Config.FilterInfo.TextPaddingBottom), (barsWidth), -(float64(config.Config.Window.Height) - 2*(config.Config.Bars.PaddingEdge) - (config.Config.FilterInfo.TextSize) - (config.Config.FilterInfo.TextPaddingBottom)) * vol
	} else {
		return (config.Config.VU.PaddingEdge) + float64(barIndex)*(barsWidth+config.Config.VU.PaddingBetween), float64(config.Config.Window.Height) - (config.Config.VU.PaddingEdge), (barsWidth), -(float64(config.Config.Window.Height) - 2*(config.Config.VU.PaddingEdge)) * vol
	}
}

func Interpolate(deltaTime float64) {
	if config.Config.VU.Interpolate.Enable {
		interpolate.Interpolate(deltaTime, LoudnessLTarget, &LoudnessLPos, &LoudnessLVel, config.Config.VU.Interpolate)
		interpolate.Interpolate(deltaTime, LoudnessRTarget, &LoudnessRPos, &LoudnessRVel, config.Config.VU.Interpolate)
	} else {
		LoudnessLPos = LoudnessLTarget
		LoudnessRPos = LoudnessRTarget
	}
	if config.Config.VU.Peak.Enable {
		if config.Config.VU.Peak.Interpolate.Enable {
			interpolate.Interpolate(deltaTime, LoudnessLPeakTarget, &LoudnessLPeakPos, &LoudnessLPeakVel, config.Config.VU.Peak.Interpolate)
			interpolate.Interpolate(deltaTime, LoudnessRPeakTarget, &LoudnessRPeakPos, &LoudnessRPeakVel, config.Config.VU.Peak.Interpolate)
		} else {
			LoudnessLPeakPos = LoudnessLPeakTarget
			LoudnessRPeakPos = LoudnessRPeakTarget
		}
	}
}
