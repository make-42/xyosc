// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ms3

import (
	"unsafe"

	math "github.com/chewxy/math32"
	"github.com/soypat/geometry/ms1"
)

const sizeofFloat = unsafe.Sizeof(float32(0))

var (
	_ = [1]byte{}[unsafe.Sizeof(Vec{})-4*sizeofFloat]  // Compile time check that Vec is 3 float32s long.
	_ = [1]byte{}[unsafe.Sizeof(Quat{})-4*sizeofFloat] // Compile time check that Quat is 4 float32s long.
)

// Quat represents a Quaternion, which is an extension of the imaginary numbers;
// there's all sorts of interesting theory behind it. In 3D graphics we mostly
// use it as a cheap way of representing rotation since quaternions are cheaper
// to multiply by, and easier to interpolate than matrices.
//
// A Quaternion has two parts: W, the so-called scalar component, and "V", the
// vector component. The vector component is considered to be the part in 3D
// space, while W (loosely interpreted) is its 4D coordinate.
//
// The imaginary V part is guaranteed to have an offset of zero in the Quat struct:
//
//	unsafe.Offsetof(q.V) // == 0
type Quat struct {
	// V contains I, J and K imaginary parts.
	I, J, K float32
	// W is the quaternion's real part.
	W float32
}

// IJK returns I,J,K fields of q as a vector with set fields X,Y,Z, respectively.
func (q Quat) IJK() Vec { return Vec{X: q.I, Y: q.J, Z: q.K} }

// WithIJK replaces I, J and K fields of q with X,Y and Z fields of argument Vec ijk and returns the result.
func (q Quat) WithIJK(ijk Vec) Quat {
	return Quat{
		W: q.W,
		I: ijk.X,
		J: ijk.Y,
		K: ijk.Z,
	}
}

// QuatIdent returns the quaternion identity: W=1; V=(0,0,0).
//
// As with all identities, multiplying any quaternion by this will yield the same
// quaternion you started with.
func QuatIdent() Quat {
	return Quat{W: 1.}
}

// Add adds two quaternions. It's no more complicated than
// adding their W and V components.
func (q1 Quat) Add(q2 Quat) Quat {
	return Quat{
		W: q1.W + q2.W,
		I: q1.I + q2.I,
		J: q1.J + q2.J,
		K: q1.K + q2.K,
	}
}

// Sub subtracts two quaternions. It's no more complicated than
// subtracting their W and V components.
func (q1 Quat) Sub(q2 Quat) Quat {
	return Quat{
		W: q1.W - q2.W,
		I: q1.I - q2.I,
		J: q1.J - q2.J,
		K: q1.K - q2.K,
	}
}

// Mul multiplies two quaternions. This can be seen as a rotation. Note that
// Multiplication is NOT commutative, meaning q1.Mul(q2) does not necessarily
// equal q2.Mul(q1).
func (q1 Quat) Mul(q2 Quat) Quat {
	v1 := q1.IJK()
	v2 := q2.IJK()
	m := Add(Cross(v1, v2), Scale(q1.W, v2))
	return Quat{
		W: q1.W*q2.W - Dot(v1, v2),
		I: m.X + q2.W*v1.X,
		J: m.Y + q2.W*v1.Y,
		K: m.Z + q2.W*v1.Z,
	}
}

// Scale every element of the quaternion by some constant factor.
func (q1 Quat) Scale(c float32) Quat {
	return Quat{
		W: q1.W * c,
		I: q1.I * c,
		J: q1.J * c,
		K: q1.K * c,
	}
}

// Conjugate returns the conjugate of a quaternion. Equivalent to
// Quat{q1.W, q1.V.Mul(-1)}.
func (q1 Quat) Conjugate() Quat {
	return Quat{
		W: q1.W,
		I: -q1.I,
		J: -q1.J,
		K: -q1.K,
	}
}

// Norm returns the euclidean length of the quaternion.
func (q1 Quat) Norm() float32 {
	return math.Sqrt(q1.Dot(q1))
}

// Normalize the quaternion, returning its versor (unit quaternion).
//
// This is the same as normalizing it as a Vec4.
func (q1 Quat) Unit() Quat {
	length := q1.Norm()

	if math.Abs(1-length) < 1e-8 {
		return q1
	}
	if length == 0 {
		return QuatIdent()
	}
	if math.IsInf(length, 0) {
		length = math.Copysign(math.MaxFloat32, length)
	}
	inv := 1. / length
	return q1.Scale(inv)
}

// Inverse of a quaternion. The inverse is equivalent
// to the conjugate divided by the square of the length.
//
// This method computes the square norm by directly adding the sum
// of the squares of all terms instead of actually squaring q1.Len(),
// both for performance and precision.
func (q1 Quat) Inverse() Quat {
	return q1.Conjugate().Scale(1 / q1.Dot(q1))
}

// Mat4 returns the homogeneous 3D rotation matrix corresponding to the
// quaternion.
// func (q1 Quat) Mat4() Mat4 {
// 	w, x, y, z := q1.W, q1.V[0], q1.V[1], q1.V[2]
// 	return Mat4{
// 		1 - 2*y*y - 2*z*z, 2*x*y + 2*w*z, 2*x*z - 2*w*y, 0,
// 		2*x*y - 2*w*z, 1 - 2*x*x - 2*z*z, 2*y*z + 2*w*x, 0,
// 		2*x*z + 2*w*y, 2*y*z - 2*w*x, 1 - 2*x*x - 2*y*y, 0,
// 		0, 0, 0, 1,
// 	}
// }

// Dot product between two quaternions, equivalent to if this was a Vec4.
func (q1 Quat) Dot(q2 Quat) float32 {
	return q1.W*q2.W + q1.I*q2.I + q1.J*q2.J + q1.K*q2.K
}

// QuatSlerp is Spherical Linear intERPolation, a method of interpolating
// between two quaternions. This always takes the straightest path on the sphere between
// the two quaternions, and maintains constant velocity.
//
// However, it's expensive and QuatSlerp(q1,q2) is not the same as QuatSlerp(q2,q1)
func QuatSlerp(q1, q2 Quat, amount float32) Quat {
	q1, q2 = q1.Unit(), q2.Unit()
	dot := q1.Dot(q2)

	// If the inputs are too close for comfort, linearly interpolate and normalize the result.
	if dot > 0.9995 {
		return QuatNlerp(q1, q2, amount)
	}

	// This is here for precision errors, I'm perfectly aware that *technically* the dot is bound [-1,1], but since Acos will freak out if it's not (even if it's just a liiiiitle bit over due to normal error) we need to clamp it
	dot = math.Max(-1, math.Min(1, dot))

	theta := math.Acos(dot) * amount

	s, c := math.Sincos(theta)
	rel := q2.Sub(q1.Scale(dot)).Unit()

	return q1.Scale(c).Add(rel.Scale(s))
}

// QuatLerp is a *L*inear Int*erp*olation between two Quaternions, cheap and simple.
//
// Not excessively useful, but uses can be found.
func QuatLerp(q1, q2 Quat, amount float32) Quat {
	return q1.Add(q2.Sub(q1).Scale(amount))
}

// QuatNlerp is a *Normalized* *L*inear Int*erp*olation between two Quaternions. Cheaper than Slerp
// and usually just as good. This is literally Lerp with Normalize() called on it.
//
// Unlike Slerp, constant velocity isn't maintained, but it's much faster and
// Nlerp(q1,q2) and Nlerp(q2,q1) return the same path. You should probably
// use this more often unless you're suffering from choppiness due to the
// non-constant velocity problem.
func QuatNlerp(q1, q2 Quat, amount float32) Quat {
	return QuatLerp(q1, q2, amount).Unit()
}

// EqualQuat returns true if elements of q1 and q2 are equal within the given tolerance.
func EqualQuat(q1, q2 Quat, tol float32) bool {
	return ms1.EqualWithinAbs(q1.I, q2.I, tol) &&
		ms1.EqualWithinAbs(q1.J, q2.J, tol) &&
		ms1.EqualWithinAbs(q1.K, q2.K, tol) &&
		ms1.EqualWithinAbs(q1.W, q2.W, tol)
}
