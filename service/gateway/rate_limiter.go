package gateway

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	slimiter "github.com/ulule/limiter/v3"
	smemory "github.com/ulule/limiter/v3/drivers/store/memory"
	"time"

	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

type MemoryLimiterConfig struct {
	Prefix     string
	RateLimit  int    // rate
	RatePeriod string // per period
}

type ApiLimiterConfig struct {
	Default   map[string]MemoryLimiterConfig
	ApiLimits map[string]MemoryLimiterConfig // routePrefix-apiName  =>  limit config
}
type apiLimiter struct {
	store      slimiter.Store
	limiterMap sync.Map
	cfg        ApiLimiterConfig
}

var limiter atomic.Pointer[apiLimiter]

func NewApiLimiter(cfg *ApiLimiterConfig) error {
	localStore := smemory.NewStoreWithOptions(slimiter.StoreOptions{
		Prefix:          "sp_api_rate_limiter",
		CleanUpInterval: 5 * time.Second,
	})
	limiter_ := &apiLimiter{
		store: localStore,
		cfg: ApiLimiterConfig{
			ApiLimits: make(map[string]MemoryLimiterConfig),
			Default:   make(map[string]MemoryLimiterConfig),
		},
	}

	var err error
	var rate slimiter.Rate
	cfgMap := make(map[string]slimiter.Rate)

	//defer func() {
	//	if err == nil {
	//		metrics.ApiRateLimitValue.Reset()
	//		metrics.ApiRateLimitPeriod.Reset()
	//		for k, v := range cfgMap {
	//			index := strings.LastIndex(k, "-")
	//			if index == -1 {
	//				continue
	//			}
	//			pkg := k[:index]
	//			method := k[index+1:]
	//			metrics.ApiRateLimitValue.WithLabelValues(pkg, method).Set(float64(v.Limit))
	//			metrics.ApiRateLimitPeriod.WithLabelValues(pkg, method).Set(v.Period.Seconds())
	//		}
	//	}
	//}()

	for k, v := range cfg.Default {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}
		cfgMap[fmt.Sprintf("%s-default", strings.ToLower(k))] = rate
		limiter_.cfg.Default[strings.ToLower(k)] = v
	}

	for k, v := range cfg.ApiLimits {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}
		cfgMap[k] = rate

		limiter_.limiterMap.Store(strings.ToLower(k), slimiter.New(localStore, rate))
	}

	limiter.Store(limiter_)
	return nil
}

func (t *apiLimiter) Allow(ctx context.Context, r *http.Request) bool {

	prefixKey := strings.ToLower(r.RequestURI)
	method := r.Method
	key := prefixKey + "-" + method

	key = strings.ToLower(key)
	log.Debugf("key: %s", key)
	limiter_, ok := t.limiterMap.Load(key)
	if !ok {
		defaultCfg, exist := t.cfg.Default[prefixKey]
		if !exist {
			return true
		}
		rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", defaultCfg.RateLimit, defaultCfg.RatePeriod))
		if err != nil {
			//error_utils.HttpErrorResponseWithMessage(ctx, http.StatusInternalServerError, "internal error")
			return false
		}
		limiter_ = slimiter.New(t.store, rate)
		t.limiterMap.Store(key, limiter_)
	}

	l := limiter_.(*slimiter.Limiter)

	//limiterCtx, err := limiter.Increment(ctx, key, 1)
	limiterCtx, err := t.store.Increment(ctx, key, 1, l.Rate)
	log.Debugf("limiterCtx: %d", limiterCtx.Limit)
	if err != nil {
		return true
	}

	log.Debugf("api limit %v, %v", key, limiterCtx.Remaining)

	if limiterCtx.Reached {
		//error_utils.HttpErrorResponseWithMessage(ctx, http.StatusTooManyRequests, "Reached the total limit of this api")
		return false
	}
	return true
}

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Load().Allow(context.Background(), r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			
			return
		}

		next.ServeHTTP(w, r)
	})
}
