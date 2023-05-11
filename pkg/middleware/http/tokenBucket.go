package http

import (
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"golang.org/x/time/rate"
)

type BandwidthLimiterConfig struct {
	Enable bool   //Enable Whether to enable bandwidth limiting
	R      string //R The speed at which tokens are generated R per second
	B      int    //B The size of the token bucket
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
