# gofft [![GoDoc][godoc-badge]][godoc] [![Build Status][travis-ci-badge]][travis-ci] [![Report Card][report-card-badge]][report-card]
A better radix-2 fast Fourier transform in Go.

Package gofft provides an efficient radix-2 fast discrete Fourier transformation algorithm in pure Go.

This code is much faster than existing FFT implementations and uses no additional memory.

The algorithm is non-recursive, works in-place overwriting the input array, and requires O(1) additional space.

## What
I took [an existing](https://github.com/ktye/fft) FFT implementation in Go, cleaned and improved the code API and performance, and replaced the permutation step with an algorithm that works with no temp array.

Performance was more than doubled over the original code, and is consistently the fastest Go FFT library (see benchmarks below) while remaining in pure Go.

Added convolution functions `Convolve(x, y)`, `FastConvolve(x, y)`, `MultiConvolve(x...)`, `FastMultiConvolve(X)`, which implement the discrete convolution and a new hierarchical convolution algorithm that has utility in a number of CS problems. This computes the convolution of many arrays in O(n\*ln(n)<sup>2</sup>) run time, and in the case of FastMultiConvolve O(1) additional space.

Also included new utility functions: `IsPow2`, `NextPow2`, `ZeroPad`, `ZeroPadToNextPow2`, `Float64ToComplex128Array`, `Complex128ToFloat64Array`, and `RoundFloat64Array`

## Why
Most existing FFT libraries in Go allocate temporary arrays with O(N) additional space. This is less-than-ideal when you have arrays of length of 2<sup>25</sup> or more, where you quickly end up allocating gigabytes of data and dragging down the FFT calculation to a halt.

Additionally, the new convolution functions have significant utility for projects I've written or am planning.

One downside is that the FFT is not multithreaded (like go-dsp is), so for large vector size FFTs on a multi-core machine it will be slower than it could be. FFTs can be run in parallel, however, so in the case of many FFT calls it will be faster.

## How
```go
package main

import (
	"fmt"
	"github.com/argusdusty/gofft"
)

func main() {
	// Do an FFT and IFFT and get the same result
	testArray := gofft.Float64ToComplex128Array([]float64{1, 2, 3, 4, 5, 6, 7, 8})
	err := gofft.FFT(testArray)
	if err != nil {
		panic(err)
	}
	err = gofft.IFFT(testArray)
	if err != nil {
		panic(err)
	}
	result := gofft.Complex128ToFloat64Array(testArray)
	gofft.RoundFloat64Array(result)
	fmt.Println(result)

	// Do a discrete convolution of the testArray with itself
	testArray, err = gofft.Convolve(testArray, testArray)
	if err != nil {
		panic(err)
	}
	result = gofft.Complex128ToFloat64Array(testArray)
	gofft.RoundFloat64Array(result)
	fmt.Println(result)
}
```

Outputs:
```
[1 2 3 4 5 6 7 8]
[1 4 10 20 35 56 84 120 147 164 170 164 145 112 64]
```

### Benchmarks
```
gofft>go test -bench=FFT$ -benchmem -cpu=1 -benchtime=5s
goos: windows
goarch: amd64
pkg: github.com/argusdusty/gofft
BenchmarkSlowFFT/Tiny_(4)                       26778385               211 ns/op         302.94 MB/s          64 B/op          1 allocs/op
BenchmarkSlowFFT/Small_(128)                       19299            307909 ns/op           6.65 MB/s        2048 B/op          1 allocs/op
BenchmarkSlowFFT/Medium_(4096)                        20         289214125 ns/op           0.23 MB/s       65536 B/op          1 allocs/op
BenchmarkKtyeFFT/Tiny_(4)                       93711196              62.2 ns/op        1028.68 MB/s          64 B/op          1 allocs/op
BenchmarkKtyeFFT/Small_(128)                     2772037              2065 ns/op         991.94 MB/s        2048 B/op          1 allocs/op
BenchmarkKtyeFFT/Medium_(4096)                     60714             99719 ns/op         657.20 MB/s       65536 B/op          1 allocs/op
BenchmarkKtyeFFT/Large_(131072)                      615           9582652 ns/op         218.85 MB/s     2097152 B/op          1 allocs/op
BenchmarkGoDSPFFT/Tiny_(4)                       1875174              3379 ns/op          18.94 MB/s         519 B/op         13 allocs/op
BenchmarkGoDSPFFT/Small_(128)                     495807             12516 ns/op         163.63 MB/s        5591 B/op         18 allocs/op
BenchmarkGoDSPFFT/Medium_(4096)                    36092            162981 ns/op         402.11 MB/s      164364 B/op         23 allocs/op
BenchmarkGoDSPFFT/Large_(131072)                     802           7506683 ns/op         279.37 MB/s     5243448 B/op         28 allocs/op
BenchmarkGoDSPFFT/Huge_(4194304)                       6         906703233 ns/op          74.01 MB/s   167772810 B/op         33 allocs/op
BenchmarkScientificFFT/Tiny_(4)                 54239148               111 ns/op         576.40 MB/s         128 B/op          2 allocs/op
BenchmarkScientificFFT/Small_(128)               3121048              1938 ns/op        1056.98 MB/s        4096 B/op          2 allocs/op
BenchmarkScientificFFT/Medium_(4096)               77004             74428 ns/op         880.52 MB/s      131072 B/op          2 allocs/op
BenchmarkScientificFFT/Large_(131072)               1816           3206107 ns/op         654.11 MB/s     4194304 B/op          2 allocs/op
BenchmarkScientificFFT/Huge_(4194304)                 24         247888846 ns/op         270.72 MB/s   134217728 B/op          2 allocs/op
BenchmarkFFT/Tiny_(4)                          794523928              7.73 ns/op        8279.81 MB/s           0 B/op          0 allocs/op
BenchmarkFFT/Small_(128)                         5375140              1132 ns/op        1809.57 MB/s           0 B/op          0 allocs/op
BenchmarkFFT/Medium_(4096)                         97856             56821 ns/op        1153.38 MB/s           0 B/op          0 allocs/op
BenchmarkFFT/Large_(131072)                         2142           2850784 ns/op         735.64 MB/s           0 B/op          0 allocs/op
BenchmarkFFT/Huge_(4194304)                           26         223165362 ns/op         300.71 MB/s           0 B/op          0 allocs/op
```

[travis-ci-badge]:   https://api.travis-ci.org/argusdusty/gofft.svg?branch=master
[travis-ci]:         https://api.travis-ci.org/argusdusty/gofft
[godoc-badge]:       https://godoc.org/github.com/argusdusty/gofft?status.svg
[godoc]:             https://godoc.org/github.com/argusdusty/gofft
[report-card-badge]: https://goreportcard.com/badge/github.com/argusdusty/gofft
[report-card]:       https://goreportcard.com/report/github.com/argusdusty/gofft