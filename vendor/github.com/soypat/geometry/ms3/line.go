package ms3

// Line can be used to represent a line between two points
// or an infinite line.
type Line [2]Vec

// Interpolate takes a value between 0 and 1 to linearly
// interpolate a point on the line.
//
//	Interpolate(0) returns l[0]
//	Interpolate(1) returns l[1]
func (ln Line) Interpolate(t float32) Vec {
	lineDir := Sub(ln[1], ln[0])
	return Add(ln[0], Scale(t, lineDir))
}

// DistanceInfinite returns the minimum euclidean Distance of point p
// to the infinite line represented by l.
func (ln Line) DistanceInfinite(point Vec) float32 {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	num := Norm(Cross(Sub(point, ln[0]), Sub(point, ln[1])))
	return num / Norm(Sub(ln[1], ln[0]))
}

// ClosestInfinite returns the point on the infinite line closest to the argument point.
func (ln Line) ClosestInfinite(point Vec) Vec {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	t := -Dot(Sub(ln[0], point), Sub(ln[1], point)) / Norm2(Sub(ln[1], ln[0]))
	return ln.Interpolate(t)
}
