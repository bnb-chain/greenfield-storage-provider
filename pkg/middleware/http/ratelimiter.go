package http

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	slimiter "github.com/ulule/limiter/v3"
	smemory "github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type RateLimiterCell struct {
	Key        string
	RateLimit  int
	RatePeriod string
}

type HTTPLimitConfig struct {
	On         bool
	RateLimit  int
	RatePeriod string
}

type RateLimiterConfig struct {
	HTTPLimitCfg HTTPLimitConfig
	Default      []RateLimiterCell
	Pattern      []RateLimiterCell
	APILimits    []RateLimiterCell
}

type MemoryLimiterConfig struct {
	RateLimit  int    // rate
	RatePeriod string // per period
}

type APILimiterConfig struct {
	HTTPLimitCfg HTTPLimitConfig
	Default      map[string]MemoryLimiterConfig
	APILimits    map[string]MemoryLimiterConfig // routePrefix-apiName  =>  limit config
	Pattern      map[string]MemoryLimiterConfig
}

type apiLimiter struct {
	store      slimiter.Store
	limiterMap sync.Map
	cfg        APILimiterConfig
}

var limiter *apiLimiter

func NewAPILimiter(cfg *APILimiterConfig) error {
	localStore := smemory.NewStoreWithOptions(slimiter.StoreOptions{
		Prefix:          "sp_api_rate_limiter",
		CleanUpInterval: 5 * time.Second,
	})
	limiter_ := &apiLimiter{
		store: localStore,
		cfg: APILimiterConfig{
			APILimits:    make(map[string]MemoryLimiterConfig),
			Default:      make(map[string]MemoryLimiterConfig),
			Pattern:      make(map[string]MemoryLimiterConfig),
			HTTPLimitCfg: cfg.HTTPLimitCfg,
		},
	}

	var err error
	var rate slimiter.Rate

	for k, v := range cfg.Default {
		limiter_.cfg.Default[strings.ToLower(k)] = v
	}

	for k, v := range cfg.Pattern {
		limiter_.cfg.Pattern[strings.ToLower(k)] = v
	}

	for k, v := range cfg.APILimits {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}

		limiter_.limiterMap.Store(strings.ToLower(k), slimiter.New(localStore, rate))
	}

	return nil
}

func (a *apiLimiter) findLimiter(host, prefix, key string) *slimiter.Limiter {
	newLimiter, ok := a.limiterMap.Load(key)
	if ok {
		return newLimiter.(*slimiter.Limiter)
	}
	for p, l := range a.cfg.Pattern {
		if regexp.MustCompile(p).MatchString(host) {
			rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", l.RateLimit, l.RatePeriod))
			if err != nil {
				log.Errorw("failed to new rate from formatted", "err", err)
				continue
			}
			newLimiter = slimiter.New(a.store, rate)
			return newLimiter.(*slimiter.Limiter)
		}
	}
	defaultCfg, exist := a.cfg.Default[prefix]
	if !exist {
		return nil
	}
	rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", defaultCfg.RateLimit, defaultCfg.RatePeriod))
	if err != nil {
		log.Errorw("failed to new rate from formatted", "err", err)
		return nil
	}
	newLimiter = slimiter.New(a.store, rate)
	a.limiterMap.Store(key, newLimiter)
	return newLimiter.(*slimiter.Limiter)
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
		return false
	}
	return true
}

func (t *apiLimiter) HTTPAllow(ctx context.Context, r *http.Request) bool {
	if !t.cfg.HTTPLimitCfg.On {
		return true
	}
	ipStr := GetIP(r)
	key := "ip_" + ipStr

	rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", t.cfg.HTTPLimitCfg.RateLimit, t.cfg.HTTPLimitCfg.RatePeriod))
	if err != nil {
		log.Errorw("failed to new rate from formatted", "err", err)
		return true
	}
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

func Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(context.Background(), r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		if !limiter.HTTPAllow(context.Background(), r) {
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
