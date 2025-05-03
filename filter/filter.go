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
