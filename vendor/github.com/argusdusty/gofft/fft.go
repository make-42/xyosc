// Package gofft provides a fast discrete Fourier transformation algorithm.
//
// Implemented is the 1-dimensional DFT of complex input data
// for with input lengths which are powers of 2.
//
// The algorithm is non-recursive, works in-place overwriting
// the input array, and requires O(1) additional space.
package gofft

import (
	"math/bits"
	"math/cmplx"
)

// Prepare precomputes values used for FFT on a vector of length N.
// N must be a perfect power of 2, otherwise this will return an error.
//
// Deprecated: This no longer has any functionality
func Prepare(N int) error {
	return checkLength("FFT Input", N)
}

// FFT implements the fast Fourier transform.
// This is done in-place (modifying the input array).
// Requires O(1) additional memory.
// len(x) must be a perfect power of 2, otherwise this will return an error.
func FFT(x []complex128) error {
	if err := checkLength("FFT Input", len(x)); err != nil {
		return err
	}
	fft(x)
	return nil
}

// IFFT implements the inverse fast Fourier transform.
// This is done in-place (modifying the input array).
// Requires O(1) additional memory.
// len(x) must be a perfect power of 2, otherwise this will return an error.
func IFFT(x []complex128) error {
	if err := checkLength("IFFT Input", len(x)); err != nil {
		return err
	}
	ifft(x)
	return nil
}

// fft does the actual work for FFT
func fft(x []complex128) {
	N := len(x)
	// Handle small N quickly
	switch N {
	case 1:
		return
	case 2:
		x[0], x[1] = x[0]+x[1], x[0]-x[1]
		return
	case 4:
		f := complex(imag(x[1])-imag(x[3]), real(x[3])-real(x[1]))
		x[0], x[1], x[2], x[3] = x[0]+x[1]+x[2]+x[3], x[0]-x[2]+f, x[0]-x[1]+x[2]-x[3], x[0]-x[2]-f
		return
	}
	// Reorder the input array.
	permute(x)
	// Butterfly
	// First 2 steps
	for i := 0; i < N; i += 4 {
		f := complex(imag(x[i+2])-imag(x[i+3]), real(x[i+3])-real(x[i+2]))
		x[i], x[i+1], x[i+2], x[i+3] = x[i]+x[i+1]+x[i+2]+x[i+3], x[i]-x[i+1]+f, x[i]-x[i+2]+x[i+1]-x[i+3], x[i]-x[i+1]-f
	}
	// Remaining steps
	w := complex(0, -1)
	for n := 4; n < N; n <<= 1 {
		w = cmplx.Sqrt(w)
		for o := 0; o < N; o += (n << 1) {
			wj := complex(1, 0)
			for k := 0; k < n; k++ {
				i := k + o
				f := wj * x[i+n]
				x[i], x[i+n] = x[i]+f, x[i]-f
				wj *= w
			}
		}
	}
}

// ifft does the actual work for IFFT
func ifft(x []complex128) {
	N := len(x)
	// Reverse the input vector
	for i := 1; i < N/2; i++ {
		j := N - i
		x[i], x[j] = x[j], x[i]
	}

	// Do the transform.
	fft(x)

	// Scale the output by 1/N
	invN := complex(1.0/float64(N), 0)
	for i := 0; i < N; i++ {
		x[i] *= invN
	}
}

// permutate permutes the input vector using bit reversal.
// Uses an in-place algorithm that runs in O(N) time and O(1) additional space.
func permute(x []complex128) {
	N := len(x)
	// Handle small N quickly
	switch N {
	case 1, 2:
		return
	case 4:
		x[1], x[2] = x[2], x[1]
		return
	case 8:
		x[1], x[3], x[4], x[6] = x[4], x[6], x[1], x[3]
		return
	}
	shift := 64 - uint64(bits.Len64(uint64(N-1)))
	N2 := N >> 1
	for i := 0; i < N; i += 2 {
		ind := int(bits.Reverse64(uint64(i)) >> shift)
		// Skip cases where low bit isn't set while high bit is
		// This eliminates 25% of iterations
		if i < N2 {
			if ind > i {
				x[i], x[ind] = x[ind], x[i]
			}
		}
		ind |= N2 // Fast way to get int(bits.Reverse64(uint64(i+1)) >> shift) here
		if ind > i+1 {
			x[i+1], x[ind] = x[ind], x[i+1]
		}
	}
}
