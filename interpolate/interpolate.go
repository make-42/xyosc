package interpolate

import (
	"math"

	"xyosc/config"
)

func Interpolate(deltaTime float64, target float64, pos *float64, vel *float64, interpolationConfig config.InterpolationConfig) {
	*pos += (target - *pos) * min(1.0, deltaTime*interpolationConfig.Direct)
	*vel += (target - *pos) * deltaTime * interpolationConfig.Accel
	*vel -= *vel * min(1.0, deltaTime*interpolationConfig.Drag)
	*pos += *vel * deltaTime
}

func InterpolateAngle(deltaTime float64, target float64, pos *float64, vel *float64, interpolationConfig config.InterpolationConfig) {
	*pos += AngleDiff(target, *pos) * min(1.0, deltaTime*interpolationConfig.Direct)
	*vel += AngleDiff(target, *pos) * deltaTime * interpolationConfig.Direct
	*vel -= *vel * min(1.0, deltaTime*interpolationConfig.Direct)
	*pos += *vel * deltaTime
	*pos = *pos - 2*math.Pi*math.Floor((*pos+math.Pi)/(2*math.Pi))

}

func AngleDiff(a, b float64) float64 {
	direct := a - b
	indirecta := a - b - 2*math.Pi
	indirectb := a - b + 2*math.Pi
	ret := direct
	if math.Abs(direct) > math.Abs(indirecta) {
		ret = indirecta
	}
	if math.Abs(indirecta) > math.Abs(indirectb) {
		ret = indirectb
	}
	return ret
} // return a-b with the smallest possible magnitude and sign in the correct direction a and b between -pi and pi
