package utils

import (
	"fmt"
	"log"
	"math"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
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

/*
func NextPowOf2(n uint32) uint32 {
	k := uint32(1)
	for k < n {
		k = k << 1
	}
	return k
}
*/
