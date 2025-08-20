package align

import (
	"sort"
	"xyosc/config"
	"xyosc/utils"

	"github.com/argusdusty/gofft"
	"github.com/goccmack/godsp/peaks"
)

var realBuffer []float64
var realBufferUnchanged []float64

func Init() {
	realBuffer = make([]float64, config.Config.ReadBufferSize/2)
	realBufferUnchanged = make([]float64, config.Config.ReadBufferSize/2)
}

func AutoCorrelate(inputArray *[]complex128, inputArrayFlipped *[]complex128) (uint32, []int) {

	numSamples := config.Config.ReadBufferSize / 2
	for i := uint32(0); i < numSamples; i++ {
		realBufferUnchanged[(i+config.Config.FFTBufferOffset)%numSamples] = real((*inputArray)[i])
	}
	err := gofft.FastConvolve(*inputArray, *inputArrayFlipped)
	utils.CheckError(err)

	for i := uint32(0); i < numSamples; i++ {
		realBuffer[(i+config.Config.FFTBufferOffset)%numSamples] = real((*inputArray)[i])
	}
	indicesACFPeaks := peaks.Get(realBuffer, config.Config.PeakDetectSeparator)
	sort.Ints(indicesACFPeaks)
	avgPeriodSum := 0
	for i := range len(indicesACFPeaks) - 1 {
		avgPeriodSum += indicesACFPeaks[i+1] - indicesACFPeaks[i]
	}

	avgPeriod := float64(avgPeriodSum) / float64(len(indicesACFPeaks)-1)
	minVal := 0.
	offset := uint32(0)

	if avgPeriod == 0 || len(indicesACFPeaks) <= 1 {
		return 0, []int{}
	}

	for i := uint32(0); i < numSamples && i < uint32(avgPeriod); i++ {

		n := 0
		sum := 0.
		for j := uint32(0); i+uint32(float64(j)*avgPeriod) < numSamples; j++ {
			sum += realBufferUnchanged[i+uint32(float64(j)*avgPeriod)]
			n++
		}
		avg := sum / float64(n)

		if i == 0 {
			minVal = avg
		}
		if avg < minVal {
			minVal = avg
			offset = i
		}
	}
	indices := []int{}
	for j := uint32(0); offset+uint32(float64(j)*avgPeriod) < numSamples; j++ {
		indices = append(indices, int(offset)+int(float64(j)*avgPeriod))
	}
	return offset, indices
}
