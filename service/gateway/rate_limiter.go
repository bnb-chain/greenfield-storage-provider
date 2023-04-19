package gateway

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	slimiter "github.com/ulule/limiter/v3"
	smemory "github.com/ulule/limiter/v3/drivers/store/memory"
	"regexp"
	"time"

	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

type RateLimiterCell struct {
	Key        string
	RateLimit  int
	RatePeriod string
}

type HTTPLimitConfig struct {
	ON         bool
	RateLimit  int
	RatePeriod string
}

type RateLimiterConfig struct {
	HttpLimitConfig HTTPLimitConfig
	Default         []RateLimiterCell
	Pattern         []RateLimiterCell
	ApiLimits       []RateLimiterCell
}

type MemoryLimiterConfig struct {
	RateLimit  int    // rate
	RatePeriod string // per period
}

type ApiLimiterConfig struct {
	HttpLimitConfig HTTPLimitConfig
	Default         map[string]MemoryLimiterConfig
	ApiLimits       map[string]MemoryLimiterConfig // routePrefix-apiName  =>  limit config
	Pattern         map[string]MemoryLimiterConfig
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
			ApiLimits:       make(map[string]MemoryLimiterConfig),
			Default:         make(map[string]MemoryLimiterConfig),
			Pattern:         make(map[string]MemoryLimiterConfig),
			HttpLimitConfig: cfg.HttpLimitConfig,
		},
	}

	var err error
	var rate slimiter.Rate

	for k, v := range cfg.Default {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}
		limiter_.cfg.Default[strings.ToLower(k)] = v
	}

	for k, v := range cfg.Pattern {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}
		limiter_.cfg.Pattern[strings.ToLower(k)] = v
	}

	for k, v := range cfg.ApiLimits {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}

		limiter_.limiterMap.Store(strings.ToLower(k), slimiter.New(localStore, rate))
	}

	limiter.Store(limiter_)
	return nil
}

func (t *apiLimiter) findLimiter(host, prefix, key string) *slimiter.Limiter {
	limiter_, ok := t.limiterMap.Load(key)
	if ok {
		return limiter_.(*slimiter.Limiter)
	}
	for p, l := range t.cfg.Pattern {
		if regexp.MustCompile(p).MatchString(host) {
			rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", l.RateLimit, l.RatePeriod))
			if err != nil {
				log.Errorw("NewRateFromFormatted failed", "err", err)
				continue
			}
			limiter_ = slimiter.New(t.store, rate)
			return limiter_.(*slimiter.Limiter)
		}
	}
	defaultCfg, exist := t.cfg.Default[prefix]
	if !exist {
		return nil
	}
	rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", defaultCfg.RateLimit, defaultCfg.RatePeriod))
	if err != nil {
		log.Errorw("NewRateFromFormatted failed", "err", err)
		return nil
	}
	limiter_ = slimiter.New(t.store, rate)
	t.limiterMap.Store(key, limiter_)
	return limiter_.(*slimiter.Limiter)
}

func (t *apiLimiter) Allow(ctx context.Context, r *http.Request) bool {
	uri := strings.ToLower(r.RequestURI)
	host := r.Host
	key := host + "-" + uri

	key = strings.ToLower(key)

	l := t.findLimiter(host, uri, key)
	if l == nil {
		return true
	}
	limiterCtx, err := t.store.Increment(ctx, key, 1, l.Rate)
	if err != nil {
		return true
	}

	if limiterCtx.Reached {
		//error_utils.HttpErrorResponseWithMessage(ctx, http.StatusTooManyRequests, "Reached the total limit of this api")
		return false
	}
	return true
}

func (t *apiLimiter) HttpAllow(ctx context.Context, r *http.Request) bool {
	if !t.cfg.HttpLimitConfig.ON {
		return true
	}
	ipStr := GetIP(r)
	key := "ip_" + ipStr

	rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", t.cfg.HttpLimitConfig.RateLimit, t.cfg.HttpLimitConfig.RatePeriod))
	limiterCtx, err := t.store.Increment(ctx, key, 1, rate)
	if err != nil {
		return true
	}

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

		if !limiter.Load().HttpAllow(context.Background(), r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
