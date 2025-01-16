package signalprocessing

import (
	"xyosc/audio"
	"xyosc/config"

	"github.com/mjibson/go-dsp/window"
)

var HannWindow []float64

func Init() {
	HannWindow = window.Hann(config.Config.ReadBufferSize / audio.SampleSizeInBytes / 2)
}
