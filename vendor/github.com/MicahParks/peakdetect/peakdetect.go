package peakdetect

import (
	"errors"
	"fmt"
	"math"
)

const (
	// SignalNegative indicates that a particular value is a negative peak.
	SignalNegative Signal = -1
	// SignalNeutral indicates that a particular value is not a peak.
	SignalNeutral Signal = 0
	// SignalPositive indicates that a particular value is a positive peak.
	SignalPositive Signal = 1
)

// Signal is a set of enums that indicates what type of peak, if any a particular value is.
type Signal int8

// ErrInvalidInitialValues indicates that the initial values provided are not valid to initialize a PeakDetector.
var ErrInvalidInitialValues = errors.New("the initial values provided are invalid")

type peakDetector struct {
	index            uint
	influence        float64
	lag              uint
	movingMeanStdDev *movingMeanStdDev
	prevMean         float64
	prevStdDev       float64
	prevValue        float64
	threshold        float64
}

// PeakDetector detects peaks in realtime timeseries data using z-scores.
//
// This is a Golang interface for the algorithm described by this StackOverflow answer:
// https://stackoverflow.com/a/22640362/14797322
//
// Brakel, J.P.G. van (2014). "Robust peak detection algorithm using z-scores". Stack Overflow. Available
// at: https://stackoverflow.com/questions/22583391/peak-signal-detection-in-realtime-timeseries-data/22640362#22640362
// (version: 2020-11-08).
type PeakDetector interface {
	// Initialize initializes the PeakDetector with its configuration and initialValues. The initialValues are the first
	// values to be processed by the PeakDetector. The length of these values are used to configure the PeakDetector's
	// lag (see description below). The PeakDetector will never return any signals for the initialValues.
	//
	// influence determines the influence of signals on the algorithm's detection threshold. If put at 0, signals have
	// no influence on the threshold, such that future signals are detected based on a threshold that is calculated with
	// a mean and standard deviation that is not influenced by past signals. If put at 0.5, signals have half the
	// influence of normal data points. Another way to think about this is that if you put the influence at 0, you
	// implicitly assume stationary (i.e. no matter how many signals there are, you always expect the time series to
	// return to the same average over the long term). If this is not the case, you should put the influence parameter
	// somewhere between 0 and 1, depending on the extent to which signals can systematically influence the time-varying
	// trend of the data. E.g., if signals lead to a structural break of the long-term average of the time series, the
	// influence parameter should be put high (close to 1) so the threshold can react to structural breaks quickly.
	//
	// threshold is the number of standard deviations from the moving mean above which the algorithm will classify a new
	// datapoint as being a signal. For example, if a new datapoint is 4.0 standard deviations above the moving mean and
	// the threshold parameter is set as 3.5, the algorithm will identify the datapoint as a signal. This parameter
	// should be set based on how many signals you expect. For example, if your data is normally distributed, a
	// threshold (or: z-score) of 3.5 corresponds to a signaling probability of 0.00047 (from this table), which implies
	// that you expect a signal once every 2128 datapoints (1/0.00047). The threshold therefore directly influences how
	// sensitive the algorithm is and thereby also determines how often the algorithm signals. Examine your own data and
	// choose a sensible threshold that makes the algorithm signal when you want it to (some trial-and-error might be
	// needed here to get to a good threshold for your purpose).
	//
	// lag determines how much your data will be smoothed and how adaptive the algorithm is to change in the long-term
	// average of the data. The more stationary your data is, the more lags you should include (this should improve the
	// robustness of the algorithm). If your data contains time-varying trends, you should consider how quickly you want
	// the algorithm to adapt to these trends. I.e., if you put lag at 10, it takes 10 'periods' before the algorithm's
	// threshold is adjusted to any systematic changes in the long-term average. So choose the lag parameter based on
	// the trending behavior of your data and how adaptive you want the algorithm to be.
	Initialize(influence, threshold float64, initialValues []float64) error
	// Next processes the next value and determines its signal.
	Next(value float64) Signal
	// NextBatch processes the next values and determines their signals. Their signals will be returned in a slice equal
	// to the length of the input.
	NextBatch(values []float64) []Signal
}

// NewPeakDetector creates a new PeakDetector. It must be initialized before use.
func NewPeakDetector() PeakDetector {
	return &peakDetector{
		movingMeanStdDev: &movingMeanStdDev{},
	}
}

func (p *peakDetector) Initialize(influence, threshold float64, initialValues []float64) error {
	p.lag = uint(len(initialValues))
	if p.lag == 0 {
		return fmt.Errorf("the length of the initial values is zero, the length is used as the lag for the algorithm: %w", ErrInvalidInitialValues)
	}
	p.influence = influence
	p.threshold = threshold

	p.prevMean, p.prevStdDev = p.movingMeanStdDev.initialize(initialValues)
	p.prevValue = initialValues[p.lag-1]

	return nil
}

func (p *peakDetector) Next(value float64) (signal Signal) {
	p.index++
	if p.index == p.lag {
		p.index = 0
	}

	if math.Abs(value-p.prevMean) > p.threshold*p.prevStdDev {
		if value > p.prevMean {
			signal = SignalPositive
		} else {
			signal = SignalNegative
		}
		value = p.influence*value + (1-p.influence)*p.prevValue
	} else {
		signal = SignalNeutral
	}

	p.prevMean, p.prevStdDev = p.movingMeanStdDev.next(value)
	p.prevValue = value

	return signal
}

func (p *peakDetector) NextBatch(values []float64) []Signal {
	signals := make([]Signal, len(values))
	for i, v := range values {
		signals[i] = p.Next(v)
	}
	return signals
}

// meanStdDev determines the mean and population standard deviation for the given population.
type movingMeanStdDev struct {
	cache        []float64
	cacheLen     float64
	cacheLenU    uint
	index        uint
	prevMean     float64
	prevVariance float64
}

// initialize creates the needed assets for the movingMeanStdDev. It also computes the resulting mean and population
// standard deviation using Welford's method.
//
// https://www.johndcook.com/blog/standard_deviation/
func (m *movingMeanStdDev) initialize(initialValues []float64) (mean, stdDev float64) {
	m.cacheLenU = uint(len(initialValues))
	m.cacheLen = float64(m.cacheLenU)
	m.cache = make([]float64, m.cacheLenU)
	copy(m.cache, initialValues)

	mean = initialValues[0]
	prevMean := mean
	var sumOfSquares float64
	for i := uint(2); i <= m.cacheLenU; i++ {
		value := initialValues[i-1]
		mean = prevMean + (value-prevMean)/float64(i)
		sumOfSquares = sumOfSquares + (value-prevMean)*(value-mean)
		prevMean = mean
	}

	m.prevMean = mean
	m.prevVariance = sumOfSquares / m.cacheLen
	return mean, math.Sqrt(m.prevVariance)
}

// Next computes the next mean and population standard deviation. It uses a sliding window and is based on Welford's
// method.
//
// https://stackoverflow.com/a/14638138/14797322
func (m *movingMeanStdDev) next(value float64) (mean, stdDev float64) {
	outOfWindow := m.cache[m.index]
	m.cache[m.index] = value
	m.index++
	if m.index == m.cacheLenU {
		m.index = 0
	}

	newMean := m.prevMean + (value-outOfWindow)/m.cacheLen
	m.prevVariance = m.prevVariance + (value-newMean+outOfWindow-m.prevMean)*(value-outOfWindow)/(m.cacheLen)
	m.prevMean = newMean

	return m.prevMean, math.Sqrt(m.prevVariance)
}
