package http

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	slimiter "github.com/ulule/limiter/v3"
	smemory "github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	Middleware = "Middleware"
)

var (
	ErrTooManyRequest = gfsperrors.Register(Middleware, http.StatusTooManyRequests, 960001, "too many requests, please try it again later")
)

type RateLimiterCell struct {
	Key        string
	RateLimit  int
	RatePeriod string
}

type IPLimitConfig struct {
	On         bool   `comment:"optional"`
	RateLimit  int    `comment:"optional"`
	RatePeriod string `comment:"optional"`
}

type RateLimiterConfig struct {
	IPLimitCfg  IPLimitConfig
	PathPattern []RateLimiterCell `comment:"optional"`
	HostPattern []RateLimiterCell `comment:"optional"`
	APILimits   []RateLimiterCell `comment:"optional"`
}

type MemoryLimiterConfig struct {
	RateLimit  int    // rate
	RatePeriod string // per period
}

type APILimiterConfig struct {
	IPLimitCfg   IPLimitConfig
	PathPattern  map[string]MemoryLimiterConfig
	PathSequence []string
	APILimits    map[string]MemoryLimiterConfig // routePrefix-apiName  =>  limit config
	HostPattern  map[string]MemoryLimiterConfig
	HostSequence []string
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
	limiter = &apiLimiter{
		store: localStore,
		cfg: APILimiterConfig{
			APILimits:    make(map[string]MemoryLimiterConfig),
			PathPattern:  make(map[string]MemoryLimiterConfig),
			PathSequence: cfg.PathSequence,
			HostPattern:  make(map[string]MemoryLimiterConfig),
			HostSequence: cfg.HostSequence,
			IPLimitCfg:   cfg.IPLimitCfg,
		},
	}

	var err error
	var rate slimiter.Rate

	for k, v := range cfg.PathPattern {
		limiter.cfg.PathPattern[strings.ToLower(k)] = v
	}

	for k, v := range cfg.HostPattern {
		limiter.cfg.HostPattern[strings.ToLower(k)] = v
	}

	for k, v := range cfg.APILimits {
		rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
		if err != nil {
			return err
		}

		limiter.limiterMap.Store(strings.ToLower(k), slimiter.New(localStore, rate))
	}

	return nil
}

func (a *apiLimiter) findLimiter(host, path, key string) *slimiter.Limiter {
	newLimiter, ok := a.limiterMap.Load(key)
	if ok {
		return newLimiter.(*slimiter.Limiter)
	}

	for i := 0; i < len(a.cfg.HostSequence); i++ {
		hostPatternInSequence := a.cfg.HostSequence[i]
		l := a.cfg.PathPattern[hostPatternInSequence]
		if regexp.MustCompile(hostPatternInSequence).MatchString(host) {
			rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", l.RateLimit, l.RatePeriod))
			if err != nil {
				log.Errorw("failed to new rate from formatted", "err", err)
				continue
			}
			newLimiter = slimiter.New(a.store, rate)
			return newLimiter.(*slimiter.Limiter)
		}
	}

	for i := 0; i < len(a.cfg.PathPattern); i++ {
		pathPatternInSequence := a.cfg.PathSequence[i]
		l := a.cfg.PathPattern[pathPatternInSequence]
		if regexp.MustCompile(pathPatternInSequence).MatchString(path) {
			rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", l.RateLimit, l.RatePeriod))
			if err != nil {
				log.Errorw("failed to new rate from formatted", "err", err)
				continue
			}
			newLimiter = slimiter.New(a.store, rate)
			return newLimiter.(*slimiter.Limiter)
		}
	}

	return nil
}

func (t *apiLimiter) Allow(ctx context.Context, r *http.Request) bool {
	path := strings.ToLower(r.RequestURI)
	host := r.Host
	key := host + "-" + path
	key = strings.ToLower(key)

	l := t.findLimiter(host, path, key)
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
	if !t.cfg.IPLimitCfg.On {
		return true
	}
	ipStr := GetIP(r)
	key := "ip_" + ipStr

	rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", t.cfg.IPLimitCfg.RateLimit, t.cfg.IPLimitCfg.RatePeriod))
	if err != nil {
		log.Errorw("failed to new rate from formatted", "err", err)
		return true
	}
	limiterCtx, err := t.store.Increment(ctx, key, 1, rate)
	if err != nil {
		return true
	}

	if limiterCtx.Reached {
		return false
	}
	return true
}

func Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(context.Background(), r) {
			MakeLimitErrorResponse(w, ErrTooManyRequest)
			return
		}
		if !limiter.HTTPAllow(context.Background(), r) {
			MakeLimitErrorResponse(w, ErrTooManyRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func MakeLimitErrorResponse(w http.ResponseWriter, err error) {
	gfspErr := gfsperrors.MakeGfSpError(err)
	var xmlInfo = struct {
		XMLName xml.Name `xml:"Error"`
		Code    int32    `xml:"Code"`
		Message string   `xml:"Message"`
	}{
		Code:    gfspErr.GetInnerCode(),
		Message: gfspErr.GetDescription(),
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal error response", "error", gfspErr.String())
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(int(gfspErr.GetHttpStatusCode()))
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write error response", "error", gfspErr.String())
	}
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
