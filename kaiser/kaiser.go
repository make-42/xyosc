package kaiser

import (
	"math"
	"xyosc/config"

	"github.com/mjibson/go-dsp/window"
)

var Prec int = 20

var WindowBuffer []float64

func factorial(n uint64) uint64 {
	p := uint64(1)
	for j := uint64(1); j < n; j++ {
		p *= (j + 1)
	}
	return p
}

func bessel0(x float64) float64 {
	s := 0.
	for j := 0; j < Prec; j++ {
		s += math.Pow(x/2, 2*float64(j)) / (float64(factorial(uint64(j))) * math.Gamma(float64(j)+1))
	}
	return s
}

func Kaiser(n int, alpha float64) []float64 {
	out := make([]float64, n)
	for i := range n {
		x := float64(i) + 1./2. - float64(n)/2
		out[i] = bessel0(math.Pi*alpha*math.Sqrt(1-(2*x/float64(n))*(2*x/float64(n)))) / (bessel0(math.Pi * alpha))
	}
	return out
}

func Init() {
	if config.Config.BarsUseWindow || config.Config.BetterPeakDetectionAlgorithmUseWindow || config.Config.ComplexTriggeringAlgorithmUseCorrelation {
		if config.Config.UseKaiserInsteadOfHannWindow {
			WindowBuffer = Kaiser(int(config.Config.ReadBufferSize/2), config.Config.KaiserWindowParam)
		} else {
			WindowBuffer = window.Hann(int(config.Config.ReadBufferSize / 2))
		}
	}
}
