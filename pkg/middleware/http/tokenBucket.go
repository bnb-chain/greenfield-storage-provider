package http

import (
	"sync"

	"golang.org/x/time/rate"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type BandwidthLimiterConfig struct {
	Enable bool       //Enable Whether to enable bandwidth limiting
	R      rate.Limit //R The speed at which tokens are generated R per second
	B      int        //B The size of the token bucket
}

type BandwidthLimiter struct {
	Limiter *rate.Limiter
}

var LimiterOnce sync.Once
var BandwidthLimit *BandwidthLimiter

func NewBandwidthLimiter(r rate.Limit, b int) {
	log.Infof("config r: %v, b:%d", r, b)

	LimiterOnce.Do(func() {
		BandwidthLimit = &BandwidthLimiter{
			Limiter: rate.NewLimiter(r, b),
		}
	})
}
