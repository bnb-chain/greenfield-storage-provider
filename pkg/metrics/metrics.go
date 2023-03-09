package metrics

import "time"

// Counter holds an int64 value that can be incremented and decremented.
type Counter interface {
	Clear()
	Count() int64
	Inc(int64)
	Dec(int64)
	Snapshot() Counter
}

// EthGauge holds an int64 value that can be set arbitrarily.
type EthGauge interface {
	Update(int64)
	Dec(int64)
	Inc(int64)
	Value() int64
}

// EthMeter counts events to produce exponentially-weighted moving average rates
// at one-, five-, and fifteen-minutes and a mean rate.
type EthMeter interface {
	Count() int64
	Mark(int64)
	Rate1() float64
	Rate5() float64
	Rate15() float64
	RateMean() float64
	Snapshot() EthMeter
	Stop()
}

// EthHistogram calculates distribution statistics from a series of int64 values.
type EthHistogram interface {
	Clear()
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Sample() EthSample
	Snapshot() EthHistogram
	StdDev() float64
	Sum() int64
	Update(int64)
	Variance() float64
}

// EthSample maintains a statistically-significant selection of values from a stream.
type EthSample interface {
	Clear()
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Size() int
	Snapshot() EthSample
	StdDev() float64
	Sum() int64
	Update(int64)
	Values() []int64
	Variance() float64
}

// EthTimer captures the duration and rate of events
type EthTimer interface {
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Rate1() float64
	Rate5() float64
	Rate15() float64
	RateMean() float64
	Snapshot() EthTimer
	StdDev() float64
	Stop()
	Sum() int64
	Time(func())
	Update(time.Duration)
	UpdateSince(time.Time)
	Variance() float64
}

// PromCounter holds a float64 value that can be incremented and decremented in prometheus
type PromCounter interface {
	With(lvs ...string) PromCounter
	Inc()
	Add(delta float64)
}

// PromGauge holds a float64 value that can be set arbitrarily in prometheus
type PromGauge interface {
	With(lvs ...string) PromGauge
	Set(value float64)
	Inc()
	Dec()
	Add(delta float64)
	Sub(delta float64)
}

// PromObserver is used in prometheus histogram and summary
type PromObserver interface {
	With(lvs ...string) PromObserver
	Observe(float64)
}
