package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// this file is used to register some extra metrics in logic

var (
	PanicsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	}, []string{"grpc_type", "grpc_service", "grpc_method"})
)
