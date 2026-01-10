package beatdetect

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/goccmack/godsp/peaks"

	"xyosc/audio"
	"xyosc/config"
	"xyosc/filter"
)

var InterpolatedBPM float64 = 0.0
var CurrentBPM float64 = 0.0
var InterpolatedBeatTime time.Time = time.Now()
var CurrentBeatTime time.Time = time.Now()
var BeatDetectSampleBufferDownsampled []float64
var BeatDetectSampleBufferDownsampledComplex []complex128
var numSamples uint32
var numSampleDownsampled uint32
var readHeadPosition uint32

var isSafeToInterpolateBPMAndTiming sync.Mutex

func Start() {
	numSamples = config.Config.Buffers.BeatDetectReadBufferSize / 2
	numSampleDownsampled = numSamples / config.Config.BeatDetection.DownSampleFactor
	BeatDetectSampleBufferDownsampled = make([]float64, numSampleDownsampled)
	BeatDetectSampleBufferDownsampledComplex = make([]complex128, numSampleDownsampled)
	for {
		DetectBeat()
		time.Sleep(time.Millisecond * time.Duration(config.Config.BeatDetection.IntervalMS))
	}
}

func InterpolateBeatTime(deltaTime float64) {
	if isSafeToInterpolateBPMAndTiming.TryLock() {
		defer isSafeToInterpolateBPMAndTiming.Unlock()
		InterpolatedBPM = max(0, (InterpolatedBPM*(1-config.Config.BeatDetection.BPMCorrectionSpeed*deltaTime) + CurrentBPM*(config.Config.BeatDetection.BPMCorrectionSpeed*deltaTime)))
		InterpolatedBeatTime = InterpolatedBeatTime.Add(time.Duration(float64(CurrentBeatTime.Sub(InterpolatedBeatTime).Nanoseconds()) * (config.Config.BeatDetection.TimeCorrectionSpeed * deltaTime)))
	}
}

func DetectBeat() {
	posStartRead := (config.Config.Buffers.RingBufferSize + audio.WriteHeadPosition - numSamples*2) % config.Config.Buffers.RingBufferSize
	timeAtDump := time.Now()
	for i := uint32(0); i < numSampleDownsampled; i++ {
		BeatDetectSampleBufferDownsampledComplex[i] = 0
		for j := uint32(0); j < config.Config.BeatDetection.DownSampleFactor; j++ {
			sampleToAvg := float64((audio.SampleRingBufferUnsafe[(posStartRead+(i*config.Config.BeatDetection.DownSampleFactor+j)*2)%config.Config.Buffers.RingBufferSize]) + (audio.SampleRingBufferUnsafe[(posStartRead+(i*config.Config.BeatDetection.DownSampleFactor+j)*2+1)%config.Config.Buffers.RingBufferSize]))
			BeatDetectSampleBufferDownsampledComplex[i] += complex(sampleToAvg*sampleToAvg, 0)
		}
	}
	domains := [][2]float64{
		{40 / (float64(config.Config.Audio.SampleRate) / float64(config.Config.BeatDetection.DownSampleFactor)), 250 / (float64(config.Config.Audio.SampleRate) / float64(config.Config.BeatDetection.DownSampleFactor))},
		{1000 / (float64(config.Config.Audio.SampleRate) / float64(config.Config.BeatDetection.DownSampleFactor)), 40000 / (float64(config.Config.Audio.SampleRate) / float64(config.Config.BeatDetection.DownSampleFactor))},
	}

	filter.FilterBufferInPlaceDomains(&BeatDetectSampleBufferDownsampledComplex, domains)

	for i := uint32(0); i < numSampleDownsampled; i++ {
		BeatDetectSampleBufferDownsampled[i] = real(BeatDetectSampleBufferDownsampledComplex[i])
	}
	indices := peaks.Get(BeatDetectSampleBufferDownsampled, int(float64(config.Config.Audio.SampleRate)/float64(config.Config.BeatDetection.DownSampleFactor)*60.0/config.Config.BeatDetection.MaxBPM))
	sort.Ints(indices)
	indices = indices[1:]

	avg, avgOffset, ok := GetTiming(indices)
	if ok {
		isSafeToInterpolateBPMAndTiming.Lock()
		defer isSafeToInterpolateBPMAndTiming.Unlock()
		CurrentBPM = 60.0 * float64(config.Config.Audio.SampleRate) / float64(config.Config.BeatDetection.DownSampleFactor) / avg
		CurrentBeatTime = timeAtDump.Add(time.Duration(int64(avgOffset) * 1000000000 / int64(config.Config.Audio.SampleRate) * int64(config.Config.BeatDetection.DownSampleFactor)))
		MultFactor := 2.0
		if config.Config.BeatDetection.HalfDisplayedBPM {
			MultFactor = 4
		}
		if InterpolatedBPM != .0 {
			for CurrentBeatTime.Sub(InterpolatedBeatTime) > time.Duration(1000000000*60/InterpolatedBPM*MultFactor) { // The 2x is here to ensure that the metronome representation is accurate and doesn't flip flop comes from one side or the other
				InterpolatedBeatTime = InterpolatedBeatTime.Add(time.Duration(1000000000 * 60 / InterpolatedBPM * MultFactor))
			}
			if CurrentBeatTime.Sub(InterpolatedBeatTime) > -CurrentBeatTime.Sub(InterpolatedBeatTime.Add(time.Duration(1000000000*60/InterpolatedBPM*MultFactor))) {
				InterpolatedBeatTime = InterpolatedBeatTime.Add(time.Duration(1000000000 * 60 / InterpolatedBPM * MultFactor))
			}
		}
	}
}

func median(data []float64) float64 {
	dataCopy := make([]float64, len(data))
	copy(dataCopy, data)

	sort.Float64s(dataCopy)

	var median float64
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (dataCopy[l/2-1] + dataCopy[l/2]) / 2
	} else {
		median = dataCopy[l/2]
	}

	return median
}

func GetTiming(indices []int) (float64, float64, bool) {
	if len(indices) < 2 {
		return 0.0, 0.0, false
	}
	ok := true
	indOffsetList := []float64{}
	for index := range len(indices) - 1 {
		indOffsetList = append(indOffsetList, float64(indices[index+1]-indices[index]))
	}

	avg := median(indOffsetList)

	avgOffset := 0.0
	beatOffsetList := []float64{}
	if avg != 0.0 {
		for index := range len(indices) {
			beatOffsetList = append(beatOffsetList, math.Mod(float64(indices[index]), avg)) // maybe improve by only considering the end indices
		}
		avgOffset = median(beatOffsetList)
	} else {
		ok = false
	}
	return avg, avgOffset, ok
}
