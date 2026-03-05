package ms3

import (
	"github.com/soypat/geometry/ms2"
)

func (t Triangle) Closest(p Vec) (closest Vec, side int8, vertex int8) {
	// We solve the 3D problem by transforming triangle
	// and point to 2D space.
	Tform := canalisTransform(t)
	pxy := Tform.MulPosition(p)
	txy := t
	for i := range txy {
		txy[i] = Tform.MulPosition(txy[i])
	}

	var closest2 ms2.Vec
	closest2, side, vertex = txy.lower().Closest(pxy.lower())
	inv := Tform.Inverse()
	closest = inv.MulPosition(Vec{X: closest2.X, Y: closest2.Y, Z: 0})
	return closest, side, vertex
}

// canalisTransform courtesy of Agustin Canalis (acanalis).
func canalisTransform(t Triangle) Mat4 {
	u2 := Sub(t[1], t[0])
	u3 := Sub(t[2], t[0])

	xc := Unit(u2)
	yc := Sub(u3, Scale(Dot(xc, u3), xc)) // t[2] but no X component
	yc = Unit(yc)
	zc := Cross(xc, yc)

	T := Mat4{
		xc.X, xc.Y, xc.Z, 0,
		yc.X, yc.Y, yc.Z, 0,
		zc.X, zc.Y, zc.Z, 0,
		0, 0, 0, 1}

	t0T := T.MulPosition(t[0])
	// Add to translation component.
	T.x03 += t0T.X
	T.x13 += t0T.Y
	T.x23 += t0T.Z
	return T
}

func (t Triangle) lower() ms2.Triangle {
	return ms2.Triangle{t[0].lower(), t[1].lower(), t[2].lower()}
}

func (v Vec) lower() ms2.Vec {
	return ms2.Vec{X: v.X, Y: v.Y}
}
