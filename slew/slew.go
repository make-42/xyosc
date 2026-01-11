package slew

import "xyosc/config"

var InterpolationPosBuffer []float64
var InterpolationVelBuffer []float64

func Init() {
	InterpolationPosBuffer = make([]float64, config.Config.SingleChannelOsc.DisplayBufferSize/2)
	InterpolationVelBuffer = make([]float64, config.Config.SingleChannelOsc.DisplayBufferSize/2)
}
