package http

import "github.com/prometheus/client_golang/prometheus"

// CounterOption lets you add options to Counter metrics using With* functions.
type CounterOption func(*prometheus.CounterOpts)

type counterOptions []CounterOption

func (co counterOptions) apply(o prometheus.CounterOpts) prometheus.CounterOpts {
	for _, f := range co {
		f(&o)
	}
	return o
}

// WithCounterConstLabels allows you to add ConstLabels to Counter metrics.
func WithCounterConstLabels(labels prometheus.Labels) CounterOption {
	return func(o *prometheus.CounterOpts) {
		o.ConstLabels = labels
	}
}

// GaugeOption lets you add options to gauge metrics using With* functions.
type GaugeOption func(opts *prometheus.GaugeOpts)

type gaugeOptions []GaugeOption

func (g gaugeOptions) apply(o prometheus.GaugeOpts) prometheus.GaugeOpts {
	for _, f := range g {
		f(&o)
	}
	return o
}

// WithGaugeConstLabels allows you to add ConstLabels to Gauge metrics.
func WithGaugeConstLabels(labels prometheus.Labels) CounterOption {
	return func(o *prometheus.CounterOpts) {
		o.ConstLabels = labels
	}
}

// SummaryOption lets you add options to gauge metrics using With* functions.
type SummaryOption func(opts *prometheus.SummaryOpts)

type summaryOptions []SummaryOption

func (s summaryOptions) apply(o prometheus.SummaryOpts) prometheus.SummaryOpts {
	for _, f := range s {
		f(&o)
	}
	return o
}

// WithSummaryConstLabels allows you to add ConstLabels to Summary metrics.
func WithSummaryConstLabels(labels prometheus.Labels) CounterOption {
	return func(o *prometheus.CounterOpts) {
		o.ConstLabels = labels
	}
}

// HistogramOption lets you add options to Histogram metrics using With* functions.
type HistogramOption func(*prometheus.HistogramOpts)

type histogramOptions []HistogramOption

func (ho histogramOptions) apply(o prometheus.HistogramOpts) prometheus.HistogramOpts {
	for _, f := range ho {
		f(&o)
	}
	return o
}

// WithHistogramBuckets allows you to specify custom bucket ranges for histograms if EnableHandlingTimeHistogram is on.
func WithHistogramBuckets(buckets []float64) HistogramOption {
	return func(o *prometheus.HistogramOpts) { o.Buckets = buckets }
}

// WithHistogramConstLabels allows you to add custom ConstLabels to histograms metrics.
func WithHistogramConstLabels(labels prometheus.Labels) HistogramOption {
	return func(o *prometheus.HistogramOpts) {
		o.ConstLabels = labels
	}
}
