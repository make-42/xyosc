package gofft

import (
	"fmt"
)

// InputSizeError represents an error when an input vector's size is not a power of 2.
type InputSizeError struct {
	Context     string
	Requirement string
	Size        int
}

func (e *InputSizeError) Error() string {
	return fmt.Sprintf("Size of %s must be %s, is: %d", e.Context, e.Requirement, e.Size)
}

// checkLength checks that the length of x is a valid power of 2
func checkLength(Context string, N int) error {
	if !IsPow2(N) {
		return &InputSizeError{Context: Context, Requirement: "power of 2", Size: N}
	}
	return nil
}

// checkLength checks that the length of x is a valid power of 2
func checkZero(Context string, N int) error {
	if N != 0 {
		return &InputSizeError{Context: Context, Requirement: "zero", Size: N}
	}
	return nil
}
