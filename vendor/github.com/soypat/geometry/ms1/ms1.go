// package ms1 implements basic 1D math useful for 3D graphics applications.
// Functions in this package have their OpenGL equivalent which is usually of the same name.
package ms1

import (
	math "github.com/chewxy/math32"
	"github.com/soypat/geometry/internal"
)

// Sign returns -1, 0, or 1 for negative, zero or positive x argument, respectively, just like OpenGL's "sign" function.
func Sign(x float32) float32 {
	if x == 0 {
		return 0
	}
	return math.Copysign(1, x)
}

// Clamp returns value v clamped between Min and Max.
func Clamp(v, Min, Max float32) float32 {
	return math.Min(Max, math.Max(v, Min))
}

// Interp performs the linear interpolation between x and y, mapping with a in interval [0,1].
// This function is known as "mix" in OpenGL.
func Interp(x, y, a float32) float32 {
	return x*(1-a) + y*a
}

// InterpWrap returns the linear interpolation between x and y on the cyclic domain [0,wrapAt].
//   - x and y define interpolation domain and must be contained within cyclic domain [0,wrapAt]
//   - a is the interpolation parameter and must be contained within [0,1]. a=0 returns x, a=1 returns y.
//   - wrapAt defines the period of the cyclic domain, specifically the point at which the domain wraps back to 0.
//     Typically 2*math.Pi radians, 360 degrees or 1 for unit phases. wrapAt must be positive.
//
// If above rules are followed the result is guaranteed to lie in [x,y] which in turn is contained in [0,wrapAt].
func InterpWrap(wrapAt, x, y, a float32) float32 {
	// TODO: maybe this function can be simplified with math intrinsics like Mod or Modf? Add tests before changing.
	wrapHalf := wrapAt / 2
	dist := y - x
	if dist > wrapHalf {
		dist -= wrapAt
	} else if dist < -wrapAt {
		dist += wrapAt
	}
	v := x + a*dist
	if v > wrapAt {
		v -= wrapAt
	} else if v < 0 {
		v += wrapAt
	}
	return v
}

// SmoothStep performs smooth cubic hermite interpolation between 0 and 1 when edge0 < x < edge1.
func SmoothStep(edge0, edge1, x float32) float32 {
	t := Clamp((x-edge0)/(edge1-edge0), 0, 1)
	return t * t * (3 - 2*t)
}

// EqualWithinAbs checks if a and b are within tol of eachother.
func EqualWithinAbs(a, b, tol float32) bool {
	return math.Abs(a-b) <= tol
}

// DefaultNewtonRaphsonSolver returns a [NewtonRaphsonSolver] with recommended parameters.
func DefaultNewtonRaphsonSolver() NewtonRaphsonSolver {
	return NewtonRaphsonSolver{
		MaxIterations: 17,
		Dx:            internal.Smallfloat32,
		Tolerance:     1.49012 * internal.Smallfloat32,
	}
}

// NewtonRaphsonSolver implements Newton-Raphson root finding algorithm for an arbitrary function.
type NewtonRaphsonSolver struct {
	// MaxIterations specifies how many iterations of Newton's succesive
	// approximations to perform. Each iteration evaluates function 3 times. Parameter is required.
	MaxIterations int
	// Tolerance sets the criteria for ending the root search when f(x)/f'(x) <= Tolerance.
	Tolerance float32
	// Dx is the step with which the gradient is calculated with central-finite-differences.
	Dx float32

	// Optional parameters below:

	// Relaxation is optional parameter to avoid overshooting during gradient descent for ill conditioned functions, i.e: large gradient near root.
	Relaxation float32
	// AdaptiveDxMaxIterations sets maximum amount of changes to step (Dx) throughout root search when encountering numerical issues.
	// If not set then not used.
	AdaptiveDxMaxIterations int
	// RootLims clamps search for root to x_min=RootLims[0], x_max=RootLims[1].
	RootLims [2]float32
}

// Root solves for a root of f such that f(x)=0 by starting guessing at x0 solving using Newton-Raphson method.
// Root returns the first root found and the amount of interations before converging.
//
// If the convergence parameter returned is negative a solution was not found within the desired tolerance.
func (nra NewtonRaphsonSolver) Root(x0 float32, f func(xGuess float32) float32) (x_root float32, convergedIn int) {
	switch {
	case nra.RootLims[0] > nra.RootLims[1]:
		panic("invalid RootLims")
	case nra.MaxIterations <= 0:
		panic("invalid MaxIterations")
	case nra.Tolerance <= 0 || math.IsNaN(nra.Tolerance):
		panic("invalid Tolerance")
	case nra.Dx <= 0 || math.IsNaN(nra.Dx):
		panic("invalid Step")
	case nra.AdaptiveDxMaxIterations < 0:
		panic("invalid AdaptiveStepMaxIterations")
	case math.IsNaN(nra.Relaxation):
		panic("invalid Relaxation")
	}

	clampSol := nra.RootLims != [2]float32{}
	krelax := 1 - nra.Relaxation
	x_root = x0

	adapt := 1
	dx := nra.Dx
	dxdiv2 := dx / 2

	for i := 1; i <= nra.MaxIterations; i++ {
		// Approximate derivative f'(x) with central finite difference method.
		// Requires more evaluations but is more precise than regular finite differences.
		fxp := f(x_root + dxdiv2)
		fxn := f(x_root - dxdiv2)
		fprime := (fxp - fxn) / dx

		if fprime == 0 || math.IsNaN(fprime) {
			// Converged to a local minimum which is not a root or problem badly conditioned.
			if adapt > nra.AdaptiveDxMaxIterations {
				return x_root, -i
			}
			// Adapt step to be larger to maybe get out of badly conditioned problem.
			dx = nra.Dx * float32(int(1<<adapt))
			dxdiv2 = dx / 2
			adapt++
			continue
		}

		fx := f(x_root)
		diff := fx / fprime
		if math.Abs(diff) <= nra.Tolerance {
			// SOLUTION FOUND.
			// apply one more iteration if permitted, we have two evaluations we can make use of.
			xnew := x_root - diff*krelax
			if i < nra.MaxIterations && math.Abs(fx) > math.Abs(f(xnew)) {
				x_root = xnew
				i++
			}
			return x_root, i
		}
		x_root -= diff * krelax
		if clampSol {
			x_root = Clamp(x_root, nra.RootLims[0], nra.RootLims[1])
		}
	}
	return x_root, -nra.MaxIterations
}

// GridSubdomain returns the indexing for a 1D subdomain contained within a larger gridded domain of equally spaced nGridPoints points.
//   - iStart contains the starting index of the first point contained within the subdomain.
//   - nSub is the number of consecutive grid points contained within the subdomain.
//
// This function is primarily used for higher dimension GridSubdomain calls in 2D and 3D packages
func GridSubdomain(domMin, domMax float32, nGridPoints int, subMin, subMax float32) (iStart, nSub int) {
	if domMin > domMax || subMin > subMax || subMax > domMax || nGridPoints <= 1 {
		panic("invalid arguments to gridSubdomain")
	}
	dx := (domMax - domMin) / float32(nGridPoints-1) // The size of grid spaces.
	startOffset := subMin - domMin
	iStartf := startOffset / dx
	iStart = min(int(math.Ceil(iStartf)), nGridPoints-1)

	endOffset := subMax - domMin
	iEndf := endOffset / dx
	iEnd := min(int(math.Ceil(iEndf)), nGridPoints-1)
	nSub = iEnd - iStart
	return iStart, nSub
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
