package ms2

import (
	math "github.com/chewxy/math32"
	"github.com/soypat/geometry/internal"
)

// Triangle represents a triangle in 2D space and
// is composed by 3 vectors corresponding to the position
// of each of the vertices. Ordering of these vertices
// decides the "normal" direction.
// Inverting ordering of two vertices inverts the resulting direction.
type Triangle [3]Vec

// Centroid returns the intersection of the three medians of the triangle
// as a point in space.
func (t Triangle) Centroid() Vec {
	return Scale(1.0/3.0, Add(Add(t[0], t[1]), t[2]))
}

// Sides returns the triangle's sides as lines:
//
//	[(t[0],t[1]), (t[1],t[2]), (t[2],t[0])]
func (t Triangle) Sides() [3]Line {
	return [3]Line{
		{t[0], t[1]},
		{t[1], t[2]},
		{t[2], t[0]},
	}
}

// Area returns the surface area of the triangle.
func (t Triangle) Area() float32 {
	// Heron's Formula, see https://en.wikipedia.org/wiki/Heron%27s_formula.
	// Also see William M. Kahan (24 March 2000). "Miscalculating Area and Angles of a Needle-like Triangle"
	// for more discussion. https://people.eecs.berkeley.edu/~wkahan/Triangle.pdf.
	a, b, c := t.orderedLengths()
	A := (c + (b + a)) * (a - (c - b))
	A *= (a + (c - b)) * (c + (b - a))
	return math.Sqrt(A) / 4
}

// longIdx returns index of the longest side. The sides
// of the triangles are are as follows:
//   - Side 0 formed by vertices 0 and 1
//   - Side 1 formed by vertices 1 and 2
//   - Side 2 formed by vertices 0 and 2
func (t Triangle) longIdx() int {
	sides := [3]Vec{Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2])}
	len2 := [3]float32{Norm2(sides[0]), Norm2(sides[1]), Norm2(sides[2])}
	longLen := len2[0]
	longIdx := 0
	if len2[1] > longLen {
		longLen = len2[1]
		longIdx = 1
	}
	if len2[2] > longLen {
		longIdx = 2
	}
	return longIdx
}

// IsDegenerate returns true if all of triangle's vertices are
// within tol distance of its longest side.
func (t Triangle) IsDegenerate(tol float32) bool {
	sides := [3]Vec{Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2])}
	len2 := [3]float32{Norm2(sides[0]), Norm2(sides[1]), Norm2(sides[2])}
	longLen := len2[0]
	longIdx := 0
	if len2[1] > longLen {
		longLen = len2[1]
		longIdx = 1
	}
	if len2[2] > longLen {
		longIdx = 2
	}
	// calculate vertex distance from longest side
	ln := Line{t[longIdx], t[(longIdx+1)%3]}
	dist := ln.DistanceInfinite(t[(longIdx+2)%3])
	return dist <= tol
}

// sort performs the sort-3 algorithm and returns
// l1, l2, l3 such that l1 ≤ l2 ≤ l3.
func sort(a, b, c float32) (l1, l2, l3 float32) {
	// sort-3
	if l2 < l1 {
		l1, l2 = l2, l1
	}
	if l3 < l2 {
		l2, l3 = l3, l2
		if l2 < l1 {
			l1, l2 = l2, l1
		}
	}
	return l1, l2, l3
}

// orderedLengths returns the lengths of the sides of the triangle such that
// a ≤ b ≤ c.
func (t Triangle) orderedLengths() (a, b, c float32) {
	s1, s2, s3 := t.edges()
	l1 := Norm(s1)
	l2 := Norm(s2)
	l3 := Norm(s3)
	return sort(l1, l2, l3)
}

// edges returns vectors for each of the edges of t.
func (t Triangle) edges() (Vec, Vec, Vec) {
	return Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2])
}

// Contains returns true if point is contained within the triangle's surface.
func (t Triangle) Contains(point Vec) bool {
	d1 := d2Sign(point, t[0], t[1])
	d2 := d2Sign(point, t[1], t[2])
	d3 := d2Sign(point, t[2], t[0])
	has_neg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	has_pos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(has_neg && has_pos)
}

func d2Sign(p1, p2, p3 Vec) float32 {
	// TODO: replace with CopyOrientation?
	return (p1.X-p3.X)*(p2.Y-p3.Y) - (p2.X-p3.X)*(p1.Y-p3.Y)
}

// Closest returns the point on the triangle closest to the argument point p.
// side and vertex are non negative to flag the point is closest the non-negative side or vertex index.
// If the point lies on the triangle it returns the same point and side=vertex=-1. side and vertex cannot be both non-negative.
func (t Triangle) Closest(p Vec) (closest Vec, side int8, vertex int8) {
	if t.Contains(p) {
		return p, -1, -1
	}
	minDist := internal.Largefloat32
	for j := range t {
		nxt := (j + 1) % 3
		edge := Line{{X: t[j].X, Y: t[j].Y}, {X: t[nxt].X, Y: t[nxt].Y}}
		pointOnTriangle, maybeVertex := edge.Closest(p)
		d2 := Norm2(Sub(p, pointOnTriangle))
		if d2 < minDist {
			if vertex < 0 {
				vertex = -1
				side = int8(j)
			} else {
				side = -1
				vertex = (int8(j) + maybeVertex) % 3
			}
			minDist = d2
			closest = pointOnTriangle
		}
	}
	return closest, side, vertex
}
