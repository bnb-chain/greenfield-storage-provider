package http

import (
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"golang.org/x/time/rate"
	"sync"
)

type BandWidthLimiterConfig struct {
	Enable bool
	R      rate.Limit
	B      int
}

type BandWidthLimiter struct {
	Limiter *rate.Limiter
}

var LimiterOnce sync.Once
var BandWidthLimit *BandWidthLimiter

func NewBandWidthLimiter(r rate.Limit, b int) {
	log.Infof("config r: %v, b:%d", r, b)

	LimiterOnce.Do(func() {
		BandWidthLimit = &BandWidthLimiter{
			Limiter: rate.NewLimiter(r, b),
		}
	})
}
