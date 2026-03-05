package ms2

import (
	math "github.com/chewxy/math32"
)

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

// DistanceInfinite returns the minimum euclidean distance of point p to the infinite line represented by l.
func (ln Line) DistanceInfinite(point Vec) float32 {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	p1 := ln[0]
	p2 := ln[1]
	num := math.Abs((p2.X-p1.X)*(p1.Y-point.Y) - (p1.X-point.X)*(p2.Y-p1.Y))
	return num / math.Hypot(p2.X-p1.X, p2.Y-p1.Y)
}

// ClosestInfinite returns the point on the infinite line closest to the argument point.
func (ln Line) ClosestInfinite(point Vec) Vec {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	t := -Dot(Sub(ln[0], point), Sub(ln[1], point)) / Norm2(Sub(ln[1], ln[0]))
	return ln.Interpolate(t)
}

// Closest returns the closest point on the line l[0]..l[1] to the argument point.
// The integer return param is 0 or 1 when closest to vertex l[0] or l[1], respectively,
// and is -1 for when the point is closest to the line segment.
func (ln Line) Closest(point Vec) (closest Vec, vertexOrSegment int8) {
	lineDir := Sub(ln[1], ln[0])
	perpendicular := Vec{-lineDir.Y, lineDir.X}
	perpend2 := Add(ln[1], perpendicular)
	e2 := Line{ln[1], perpend2}.edgeEquation(point)
	if e2 > 0 {
		return ln[1], 0
	}
	perpend1 := Add(ln[0], perpendicular)
	e1 := Line{ln[0], perpend1}.edgeEquation(point)
	if e1 < 0 {
		return ln[0], 1
	}
	e3 := ln.DistanceInfinite(point)
	toPoint := Scale(-e3, Unit(perpendicular))
	return Sub(point, toPoint), -1
}

// edgeEquation returns a signed distance of a point to
// an infinite line defined by two points
// Edge equation for a line passing through (X,Y)
// with gradient dY/dX
// E ( x; y ) =(x-X)*dY - (y-Y)*dX
func (ln Line) edgeEquation(p Vec) float32 {
	dxy := Sub(ln[1], ln[0])
	return (p.X-ln[0].X)*dxy.Y - (p.Y-ln[0].Y)*dxy.X
}
