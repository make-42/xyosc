package ms3

import (
	math "github.com/chewxy/math32"
)

// Tetra represents a tetrahedron in 3D space.
type Tetra [4]Vec

// Centroid returns the tetrahedron's centroid.
func (t Tetra) Centroid() Vec {
	return Scale(1./4., Add(t[0], Add(t[1], Add(t[2], t[3]))))
}

// Sides returns the sides as lines of the tetrahedron:
//
//	[(t[0],t[1]),(t[1],t[2]),(t[2],t[0]),
//	(t[3],t[2]),(t[0],t[3]),(t[1],t[3])]
func (t Tetra) Sides() [6]Line {
	return [6]Line{
		{t[0], t[1]}, {t[1], t[2]}, {t[2], t[0]},
		{t[3], t[2]}, {t[0], t[3]}, {t[1], t[3]},
	}
}

// Edges returns the directional vectors composing the sides of the tetrahedron as returned by [Tetra.Sides].
func (t Tetra) Edges() [6]Vec {
	return [6]Vec{
		Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2]),
		Sub(t[2], t[3]), Sub(t[3], t[0]), Sub(t[3], t[1]),
	}
}

// Volume returns the volume contained in the tetrahedron.
func (t Tetra) Volume() float32 {
	const third = 1.0 / 3.0
	base := Triangle{t[0], t[1], t[2]}
	area := base.Area()
	height := base.Plane().Distance(t[3])
	return third * area * height
}

// Aspect returns the aspect ratio of the tetrahedron calculated
// as: longestEdge/minHeight.
func (t Tetra) Aspect() float32 {
	e := t.longestEdge()
	heights := t.heights()
	hmin, _, _ := Sort(heights[0], heights[1], heights[2])
	if heights[3] < hmin {
		hmin = heights[3]
	}
	return e / hmin

}

func (t Tetra) longestEdge() float32 {
	edges := t.Edges()
	_, _, e1 := Sort(Norm2(edges[0]), Norm2(edges[1]), Norm2(edges[2]))
	_, _, e2 := Sort(Norm2(edges[3]), Norm2(edges[4]), Norm2(edges[5]))
	if e1 > e2 {
		return math.Sqrt(e1)
	}
	return math.Sqrt(e2)
}

// heights returns the "heights" of the tetrahedron,
// which are basically the shortest normal dropped from a vertex to the opposite face.
func (t Tetra) heights() (alt [4]float32) {
	for i := range t {
		j := (i + 1) % 4
		k := (i + 2) % 4
		l := (i + 3) % 4
		e1 := Sub(t[k], t[j])
		e2 := Sub(t[l], t[j])
		p := newPlane(t[l], Cross(e1, e2))
		alt[i] = p.Distance(t[i])
	}
	return alt
}
