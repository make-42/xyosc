package ms3

import math "github.com/chewxy/math32"

// Rotation creates a rotation quaternion
// that rotates an angle relative an axis (direction).
// Call [Quat.Rotate] method on Quat to apply rotation.
func Rotation(angleRadians float32, axis Vec) Quat {
	s, c := math.Sincos(0.5 * angleRadians)
	return Quat{
		W: c,
		I: axis.X * s,
		J: axis.Y * s,
		K: axis.Z * s,
	}
}

// Rotate a vector by the rotation this quaternion represents.
// This will result in a 3D vector. Strictly speaking, this is
// equivalent to q1.v.q* where the "."" is quaternion multiplication and v is interpreted
// as a quaternion with W 0 and V v. In code:
// q1.Mul(Quat{0,v}).Mul(q1.Conjugate()), and
// then retrieving the imaginary (vector) part.
//
// In practice, we hand-compute this in the general case and simplify
// to save a few operations.
func (q1 Quat) Rotate(v Vec) Vec {
	v1 := q1.IJK()
	cross := Cross(v1, v)
	// v + 2q_w * (q_v x v) + 2q_v x (q_v x v)
	finalTerm := Cross(Scale(2, v1), cross)
	x := Add(Scale(2*q1.W, cross), finalTerm)
	return Add(v, x)
}

// Rotation returns the angle and axis which would give the equivalent rotation obtained by [Rotation].
// It is effectively the inverse operation of said function.
// If quaternion is not normalized then the result is not a pure rotation.
func (q Quat) Rotation() (angleRadians float32, axis Vec) {
	angleRadians = 2 * math.Acos(q.W)
	sad2 := math.Sin(angleRadians / 2)
	if sad2 == 0 {
		axis = Vec{X: 1}
	} else {
		axis = Vec{X: q.I, Y: q.J, Z: q.K}
		axis = Scale(1/sad2, axis)
	}
	return angleRadians, axis
}

// RotationMat3FromQuat returns a 3×3 rotation matrix which applies the same rotation as q.
// It may be used to perform rotations on a 3-vector or to apply the rotation
// to a 3×n matrix of column vectors.
//
// If the receiver is not a unit quaternion, the returned matrix will not be a pure rotation.
// Rotations created with [Rotation] are unit quaternions.
func (q Quat) RotationMat3() Mat3 {
	w, i, j, k := q.W, q.I, q.J, q.K
	ii := 2 * i * i
	jj := 2 * j * j
	kk := 2 * k * k
	wi := 2 * w * i
	wj := 2 * w * j
	wk := 2 * w * k
	ij := 2 * i * j
	jk := 2 * j * k
	ki := 2 * k * i
	return mat3(
		1-(jj+kk), ij-wk, ki+wj,
		ij+wk, 1-(ii+kk), jk-wi,
		ki-wj, jk+wi, 1-(ii+jj))
}

// RotationMat4 returns an orthographic 4x4 rotation matrix (right hand rule).
func RotationMat4(angleRadians float32, axis Vec) Mat4 {
	axis = Unit(axis)
	s, c := math.Sincos(angleRadians)
	m := 1 - c
	return Mat4{
		m*axis.X*axis.X + c, m*axis.X*axis.Y - axis.Z*s, m*axis.Z*axis.X + axis.Y*s, 0,
		m*axis.X*axis.Y + axis.Z*s, m*axis.Y*axis.Y + c, m*axis.Y*axis.Z - axis.X*s, 0,
		m*axis.Z*axis.X - axis.Y*s, m*axis.Y*axis.Z + axis.X*s, m*axis.Z*axis.Z + c, 0,
		0, 0, 0, 1,
	}
}

// RotationBetweenVecs calculates the rotation between start and dest as direction vectors such that
// when the resulting rotation is applied on start, dest is returned.
func RotationBetweenVecs(start, dest Vec) Quat {
	// http://www.opengl-tutorial.org/intermediate-tutorials/tutorial-17-quaternions/#I_need_an_equivalent_of_gluLookAt__How_do_I_orient_an_object_towards_a_point__
	// https://github.com/g-truc/glm/blob/0.9.5/glm/gtx/quaternion.inl#L225
	// https://bitbucket.org/sinbad/ogre/src/d2ef494c4a2f5d6e2f0f17d3bfb9fd936d5423bb/OgreMain/include/OgreVector3.h?at=default#cl-654

	start = Unit(start)
	dest = Unit(dest)
	epsilon := float32(0.001)

	cosTheta := Dot(start, dest)
	if cosTheta < -1.0+epsilon {
		// special case when vectors in opposite directions:
		// there is no "ideal" rotation axis
		// So guess one; any will do as long as it's perpendicular to start
		axis := Cross(Vec{X: 1, Y: 0, Z: 0}, start)
		if Norm2(axis) < epsilon {
			// bad luck, they were parallel, try again!
			axis = Cross(Vec{X: 0, Y: 1, Z: 0}, start)
		}
		return Rotation(math.Pi, Unit(axis))
	}

	axis := Cross(start, dest)
	s := math.Sqrt((1.0 + cosTheta) * 2.0)

	return Quat{
		W: s * 0.5,
		I: axis.X / s,
		J: axis.Y / s,
		K: axis.Z / s,
	}
}

// EqualOrientation returns whether the quaternions represents the same orientation with a given tolerence
func (q1 Quat) EqualOrientation(q2 Quat, tol float32) bool {
	n1sq := q1.Dot(q1)
	n2sq := q2.Dot(q2)
	if n1sq == 0 || n2sq == 0 {
		return false // Degenerate quaternion.
	}
	d := q1.Dot(q2)
	return d*d >= tol*tol*(n1sq*n2sq)
}

/*
// RotationLookAt creates a rotation from an eye point to a "focus" or center point (both positions in same coordinate space).
//
// It assumes the front of the rotated object at Z- and up at Y+
func RotationLookAt(eyePosition, focusOfAttentionPosition, upDirection Vec) Quat {
	// http://www.opengl-tutorial.org/intermediate-tutorials/tutorial-17-quaternions/#I_need_an_equivalent_of_gluLookAt__How_do_I_orient_an_object_towards_a_point__
	// https://bitbucket.org/sinbad/ogre/src/d2ef494c4a2f5d6e2f0f17d3bfb9fd936d5423bb/OgreMain/src/OgreCamera.cpp?at=default#cl-161

	direction := Unit(Sub(focusOfAttentionPosition, eyePosition))

	// Find the rotation between the front of the object (that we assume towards Z-,
	// but this depends on your model) and the desired direction
	rotDir := RotationBetweenVecs(Vec{X: 0, Y: 0, Z: -1}, direction)

	// Recompute up so that it's perpendicular to the direction
	// You can skip that part if you really want to force up
	// right := direction.Cross(up)
	// up = right.Cross(direction)

	// Because of the 1st rotation, the up is probably completely screwed up.
	// Find the rotation between the "up" of the rotated object, and the desired up
	upCur := rotDir.Rotate(Vec{X: 0, Y: 1, Z: 0})
	rotUp := RotationBetweenVecs(upCur, upDirection)

	rotTarget := rotUp.Mul(rotDir) // remember, in reverse order.
	return rotTarget.Inverse()     // camera rotation should be inversed!
}

// RotatingBetweenVecsMat4 returns the rotation matrix that transforms "start" onto the same direction as "dest".
func RotatingBetweenVecsMat4(start, dest Vec) Mat4 {
	// is either vector == 0?
	const epsilon = 1e-12
	if EqualElem(start, Vec{}, epsilon) || EqualElem(dest, Vec{}, epsilon) {
		return IdentityMat4()
	}
	// normalize both vectors
	start = Unit(start)
	dest = Unit(dest)
	// are the vectors the same?
	if EqualElem(start, dest, epsilon) {
		return IdentityMat4()
	}

	// are the vectors opposite (180 degrees apart)?
	if EqualElem(Scale(-1, start), dest, epsilon) {
		return Mat4{
			-1, 0, 0, 0,
			0, -1, 0, 0,
			0, 0, -1, 0,
			0, 0, 0, 1,
		}
	}
	// general case
	// See:	https://math.stackexchange.com/questions/180418/calculate-rotation-matrix-to-align-vector-a-to-vector-b-in-3d
	v := Cross(start, dest)
	vx := Skew(v)
	k := 1. / (1. + Dot(start, dest))

	vx2 := MulMat3(vx, vx)
	vx2 = ScaleMat3(vx2, k)

	// Calculate sum of matrices.
	vx = AddMat3(vx, IdentityMat3())
	vx = AddMat3(vx, vx2)

	return vx.AsMat4()
}

// Mat4ToQuat converts a pure rotation matrix into a quaternion
func Mat4ToQuat(m Mat4) Quat {
	// http://www.euclideanspace.com/maths/geometry/rotations/conversions/matrixToQuaternion/index.htm
	if tr := m[0] + m[5] + m[10]; tr > 0 {
		s := 0.5 / math32.Sqrt(tr+1.0)
		return Quat{
			0.25 / s,
			Vec{
				(m[6] - m[9]) * s,
				(m[8] - m[2]) * s,
				(m[1] - m[4]) * s,
			},
		}
	}

	if (m[0] > m[5]) && (m[0] > m[10]) {
		s := 2.0 * math32.Sqrt(1.0+m[0]-m[5]-m[10])
		return Quat{
			(m[6] - m[9]) / s,
			Vec{
				0.25 * s,
				(m[4] + m[1]) / s,
				(m[8] + m[2]) / s,
			},
		}
	}

	if m[5] > m[10] {
		s := 2.0 * math32.Sqrt(1.0+m[5]-m[0]-m[10])
		return Quat{
			(m[8] - m[2]) / s,
			Vec{
				(m[4] + m[1]) / s,
				0.25 * s,
				(m[9] + m[6]) / s,
			},
		}

	}

	s := 2.0 * math32.Sqrt(1.0+m[10]-m[0]-m[5])
	return Quat{
		(m[1] - m[4]) / s,
		Vec{
			(m[8] + m[2]) / s,
			(m[9] + m[6]) / s,
			0.25 * s,
		},
	}
}
*/
