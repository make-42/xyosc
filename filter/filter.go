package filter

import (
	"xyosc/utils"

	"github.com/argusdusty/gofft"
)

func FilterBufferInPlace(inputArray *[]complex128, lowCutOffFrac float64, highCutOffFrac float64) {
	err := gofft.FFT(*inputArray)
	utils.CheckError(err)
	for x := range len(*inputArray) {
		if x >= int(float64(len(*inputArray))*highCutOffFrac) || x < int(float64(len(*inputArray))*lowCutOffFrac) {
			(*inputArray)[x] = 0
		}
	}
	err = gofft.IFFT(*inputArray)
	utils.CheckError(err)
}

func FilterBufferInPlaceDomains(inputArray *[]complex128, domains [][2]float64) {
	err := gofft.FFT(*inputArray)
	utils.CheckError(err)
	for x := range len(*inputArray) {
		keep := false
		for _, domain := range domains {
			if x > int(float64(len(*inputArray))*domain[0]) && x <= int(float64(len(*inputArray))*domain[1]) {
				keep = true
				break
			}
		}
		if !keep {
			(*inputArray)[x] = 0
		}
	}
	err = gofft.IFFT(*inputArray)
	utils.CheckError(err)
}
