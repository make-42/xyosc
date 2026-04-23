package slew

import (
	"image/color"

	"xyosc/config"
)

var InterpolationPosBuffer []float64
var InterpolationVelBuffer []float64
var BufferSize uint32

func Init() {
	BufferSize = config.Config.SingleChannelOsc.DisplayBufferSize / 2 / config.Config.SingleChannelOsc.DisplayDownsample
	InterpolationPosBuffer = make([]float64, BufferSize*uint32(len(config.Config.SingleChannelOsc.Traces)))
	InterpolationVelBuffer = make([]float64, BufferSize*uint32(len(config.Config.SingleChannelOsc.Traces)))
}

var TraceColors []color.RGBA

func InitTraceColors() {
	TraceColors = make([]color.RGBA, len(config.Config.SingleChannelOsc.Traces))
	for traceIdx := range config.Config.SingleChannelOsc.Traces {
		TraceColors[traceIdx] = color.RGBA{uint8(float64(config.ThirdColorAdj.R) * config.Config.SingleChannelOsc.Traces[traceIdx].Opacity),
			uint8(float64(config.ThirdColorAdj.G) * config.Config.SingleChannelOsc.Traces[traceIdx].Opacity),
			uint8(float64(config.ThirdColorAdj.B) * config.Config.SingleChannelOsc.Traces[traceIdx].Opacity),
			uint8(float64(config.ThirdColorAdj.A) * config.Config.SingleChannelOsc.Traces[traceIdx].Opacity),
		}
	}
}
