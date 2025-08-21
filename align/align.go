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
	avgPeriod := 0.
	if config.Config.FrequencyDetectionUseMedian {
		periods := []int{}
		for i := range len(indicesACFPeaks) - 1 {
			periods = append(periods, indicesACFPeaks[i+1]-indicesACFPeaks[i])
		}
		sort.Ints(periods)
		avgPeriodSum := 0
		n := 0
		for j := range len(periods) {
			pos := float64(j+1) / float64(len(periods)+1)
			if (pos >= 1./3. && pos <= 2./3.) || len(periods) <= 3 {
				avgPeriodSum += periods[j]
				n++
			}
		}
		avgPeriod = float64(avgPeriodSum) / float64(n)
	} else {
		avgPeriodSum := 0
		for i := range len(indicesACFPeaks) - 1 {
			avgPeriodSum += indicesACFPeaks[i+1] - indicesACFPeaks[i]
		}
		avgPeriod = float64(avgPeriodSum) / float64(len(indicesACFPeaks)-1)
	}

	offset := uint32(0)
	if avgPeriod == 0 || len(indicesACFPeaks) <= 1 {
		return 0, []int{}
	}
	if config.Config.UseComplexTriggeringAlgorithm {
		minVal := 0.
		for i := uint32(0); i < numSamples && (config.Config.TriggerThroughoutWindow || (i < uint32(avgPeriod))); i++ {
			n := 0
			sum := 0.
			for j := uint32(0); i+uint32((float64(j)+0.75)*avgPeriod) < numSamples; j++ {
				sum += realBufferUnchanged[i+uint32(float64(j)*avgPeriod)]
				sum += realBufferUnchanged[i+uint32((float64(j)+0.25)*avgPeriod)] * realBufferUnchanged[i+uint32((float64(j)+0.25)*avgPeriod)]
				sum -= realBufferUnchanged[i+uint32((float64(j)+0.5)*avgPeriod)]
				sum += realBufferUnchanged[i+uint32((float64(j)+0.75)*avgPeriod)] * realBufferUnchanged[i+uint32((float64(j)+0.75)*avgPeriod)]
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
	} else {
		minVal := 0.
		for i := uint32(0); i < numSamples && (config.Config.TriggerThroughoutWindow || (i < uint32(avgPeriod))); i++ {
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
	}
	if config.Config.QuadratureOffset {
		offset += uint32(0.75 * avgPeriod)
	}
	indices := []int{}
	for j := uint32(0); offset+uint32(float64(j)*avgPeriod) < numSamples; j++ {
		indices = append(indices, int(offset)+int(float64(j)*avgPeriod))
	}
	return offset, indices
}
