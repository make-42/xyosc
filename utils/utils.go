package utils

import (
	"fmt"
	"math"

	"github.com/ztrue/tracerr"
)

func CheckError(err error) {
	if err != nil {
		tracerr.PrintSourceColor(err)
	}
}

func FormatDuration(d float64) string {
	if d == 0 {
		return "0"
	}
	switch {
	case math.Abs(d) >= 0.1:
		return fmt.Sprintf("%.2fs", d)
	case math.Abs(d) >= 0.001:
		return fmt.Sprintf("%.2fms", d*1000)
	default:
		return fmt.Sprintf("%.2fµs", d*1000000)
	}
}

func Moduint32(a, b uint32) uint32 {
	return (a%b + b) % b
}

/*
func NextPowOf2(n uint32) uint32 {
	k := uint32(1)
	for k < n {
		k = k << 1
	}
	return k
}
*/
