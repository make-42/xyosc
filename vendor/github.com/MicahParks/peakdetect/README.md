[![Go Reference](https://pkg.go.dev/badge/github.com/MicahParks/peakdetect.svg)](https://pkg.go.dev/github.com/MicahParks/peakdetect) [![Go Report Card](https://goreportcard.com/badge/github.com/MicahParks/peakdetect)](https://goreportcard.com/report/github.com/MicahParks/peakdetect)
# peakdetect
Detect peaks in realtime timeseries data using z-scores. This is a Golang implementation for the algorithm described
by [this StackOverflow answer](https://stackoverflow.com/a/22640362/14797322).

Unlike some implementations, a goal is to minimize the memory footprint and allow for the processing of new data points
without reprocessing old ones.

```go
import "github.com/MicahParks/peakdetect"
```

# Configuration
`Lag` determines how much your data will be smoothed and how adaptive the algorithm is to change in the long-term
average of the data. The more stationary your data is, the more lags you should include (this should improve the
robustness of the algorithm). If your data contains time-varying trends, you should consider how quickly you want the
algorithm to adapt to these trends. I.e., if you put lag at 10, it takes 10 'periods' before the algorithm's threshold
is adjusted to any systematic changes in the long-term average. So choose the lag parameter based on the trending
behavior of your data and how adaptive you want the algorithm to be.

`Influence` determines the influence of signals on the algorithm's detection threshold. If put at 0, signals have no
influence on the threshold, such that future signals are detected based on a threshold that is calculated with a mean
and standard deviation that is not influenced by past signals. If put at 0.5, signals have half the influence of normal
data points. Another way to think about this is that if you put the influence at 0, you implicitly assume stationary (
i.e. no matter how many signals there are, you always expect the time series to return to the same average over the long
term). If this is not the case, you should put the influence parameter somewhere between 0 and 1, depending on the
extent to which signals can systematically influence the time-varying trend of the data. E.g., if signals lead to a
structural break of the long-term average of the time series, the influence parameter should be put high (close to 1) so
the threshold can react to structural breaks quickly

`Threshold` is the number of standard deviations from the moving mean above which the algorithm will classify a new
datapoint as being a signal. For example, if a new datapoint is 4.0 standard deviations above the moving mean and the
threshold parameter is set as 3.5, the algorithm will identify the datapoint as a signal. This parameter should be set
based on how many signals you expect. For example, if your data is normally distributed, a threshold (or: z-score) of
3.5 corresponds to a signaling probability of 0.00047 (from this table), which implies that you expect a signal once
every 2128 datapoints (1/0.00047). The threshold therefore directly influences how sensitive the algorithm is and
thereby also determines how often the algorithm signals. Examine your own data and choose a sensible threshold that
makes the algorithm signal when you want it to (some trial-and-error might be needed here to get to a good threshold for
your purpose)

# Usage
```go
package main

import (
	"fmt"
	"log"

	"github.com/MicahParks/peakdetect"
)

// This example is the equivalent of the R example from the algorithm's author.
// https://stackoverflow.com/a/54507329/14797322
func main() {
	data := []float64{1, 1, 1.1, 1, 0.9, 1, 1, 1.1, 1, 0.9, 1, 1.1, 1, 1, 0.9, 1, 1, 1.1, 1, 1, 1, 1, 1.1, 0.9, 1, 1.1, 1, 1, 0.9, 1, 1.1, 1, 1, 1.1, 1, 0.8, 0.9, 1, 1.2, 0.9, 1, 1, 1.1, 1.2, 1, 1.5, 1, 3, 2, 5, 3, 2, 1, 1, 1, 0.9, 1, 1, 3, 2.6, 4, 3, 3.2, 2, 1, 1, 0.8, 4, 4, 2, 2.5, 1, 1, 1}

	// Algorithm configuration from example.
	const (
		lag       = 30
		threshold = 5
		influence = 0
	)

	// Create then initialize the peak detector.
	detector := peakdetect.NewPeakDetector()
	err := detector.Initialize(influence, threshold, data[:lag]) // The length of the initial values is the lag.
	if err != nil {
		log.Fatalf("Failed to initialize peak detector.\nError: %s", err)
	}

	// Start processing new data points and determine what signal, if any they produce.
	//
	// This method, .Next(), is best for when data are being processed in a stream, but this simply iterates over a
	// slice.
	nextDataPoints := data[lag:]
	for i, newPoint := range nextDataPoints {
		signal := detector.Next(newPoint)
		var signalType string
		switch signal {
		case peakdetect.SignalNegative:
			signalType = "negative"
		case peakdetect.SignalNeutral:
			signalType = "neutral"
		case peakdetect.SignalPositive:
			signalType = "positive"
		}

		println(fmt.Sprintf("Data point at index %d has the signal: %s", i+lag, signalType))
	}

	// This method, .NextBatch(), is a helper function for processing many data points at once. It's returned slice
	// should produce the same signal outputs as the loop above.
	signals := detector.NextBatch(nextDataPoints)
	println(fmt.Sprintf("1:1 ratio of batch inputs to signal outputs: %t", len(signals) == len(nextDataPoints)))
}
```

# Testing
```
$ go test -cover -race
PASS
coverage: 100.0% of statements
ok      github.com/MicahParks/peakdetect        0.019s
```

# Performance
To further improve performance, this algorithm uses Welford's algorithm on initialization
and an adaptation of [this StackOverflow answer](https://stackoverflow.com/a/14638138/14797322) to calculate the mean
and population standard deviation for the lag period (sliding window). This appears to improve performance by more than
a factor of 10!

`v0.0.4`
```
goos: linux
goarch: amd64
pkg: github.com/MicahParks/peakdetect
cpu: AMD Ryzen 9 7950X 16-Core Processor            
BenchmarkPeakDetector_NextBatch-32      1000000000               0.0000221 ns/op
PASS
ok      github.com/MicahParks/peakdetect        0.003s
```

`v0.1.0`
```
goos: linux
goarch: amd64
pkg: github.com/MicahParks/peakdetect
cpu: AMD Ryzen 9 7950X 16-Core Processor            
BenchmarkPeakDetector_NextBatch-32      1000000000               0.0000011 ns/op
PASS
ok      github.com/MicahParks/peakdetect        0.003s
```

# References
Brakel, J.P.G. van (2014). "Robust peak detection algorithm using z-scores". Stack Overflow. Available
at: https://stackoverflow.com/questions/22583391/peak-signal-detection-in-realtime-timeseries-data/22640362#22640362
(version: 2020-11-08).

* [StackOverflow: Peak detection in realtime timeseries data](https://stackoverflow.com/a/22640362/14797322).
* [StackOverflow: sliding window for online algorithm to calculate mean and standard devation](https://stackoverflow.com/a/14638138/14797322).
* [Welford's algorithm related blog post](https://www.johndcook.com/blog/standard_deviation/).
* Yeah, I used [Wikipedia](https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance) too.
