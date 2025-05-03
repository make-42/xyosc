package utils

import "log"

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
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
