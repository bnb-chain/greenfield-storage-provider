package metrics

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: model.GatewayService,
		Subsystem: "server",
		Name:      "requests_total",
		Help:      "Compute how many HTTP requests are processed, partitioned by status code and HTTP method",
	}, []string{"method", "handler", "host", "name"})
)
