package http

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	slimiter "github.com/ulule/limiter/v3"
	smemory "github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	Middleware      = "Middleware"
	MethodSeparator = "-"
)

var (
	ErrTooManyRequest = gfsperrors.Register(Middleware, http.StatusTooManyRequests, 960001, "too many requests, please try it again later")
)

type KeyToRateLimiterNameCell struct {
	Key    string
	Method string
	Names  []string
}

type IPLimitConfig struct {
	On         bool   `comment:"optional"`
	RateLimit  int    `comment:"optional"`
	RatePeriod string `comment:"optional"`
}

type RateLimiterConfig struct {
	IPLimitCfg  IPLimitConfig
	PathPattern []KeyToRateLimiterNameCell `comment:"optional"`
	HostPattern []KeyToRateLimiterNameCell `comment:"optional"`
	APILimits   []KeyToRateLimiterNameCell `comment:"optional"`
	NameToLimit []MemoryLimiterConfig      `comment:"optional"`
}

type MemoryLimiterConfig struct {
	Name       string // limiter name
	RateLimit  int    // rate
	RatePeriod string // per period
}

type APILimiterConfig struct {
	IPLimitCfg   IPLimitConfig
	PathPattern  map[string][]MemoryLimiterConfig
	PathSequence []string
	APILimits    map[string][]MemoryLimiterConfig // routePrefix-apiName  =>  limit config
	HostPattern  map[string][]MemoryLimiterConfig
	HostSequence []string
}

type rateLimiterWithName struct {
	name        string
	rateLimiter *slimiter.Limiter
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
			APILimits:    make(map[string][]MemoryLimiterConfig),
			PathPattern:  make(map[string][]MemoryLimiterConfig),
			PathSequence: cfg.PathSequence,
			HostPattern:  make(map[string][]MemoryLimiterConfig),
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

	for k, vs := range cfg.APILimits {
		for _, v := range vs {
			rate, err = slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", v.RateLimit, v.RatePeriod))
			if err != nil {
				return err
			}

			limiter.limiterMap.Store(strings.ToLower(k), slimiter.New(localStore, rate))
		}
	}

	return nil
}

func (a *apiLimiter) findLimiter(host, path, key string, virtualHost bool, method string) []rateLimiterWithName {
	var result []rateLimiterWithName
	newLimiter, ok := a.limiterMap.Load(key)
	if ok {
		result = append(result, rateLimiterWithName{
			name:        "",
			rateLimiter: newLimiter.(*slimiter.Limiter),
		})
	}

	for i := 0; i < len(a.cfg.HostSequence); i++ {
		hostPatternInSequence := a.cfg.HostSequence[i]
		ls := a.cfg.HostPattern[hostPatternInSequence]
		if regexp.MustCompile(hostPatternInSequence).MatchString(host) {
			for _, l := range ls {
				rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", l.RateLimit, l.RatePeriod))
				if err != nil {
					log.Errorw("failed to new rate from formatted", "err", err)
					continue
				}
				newLimiter = slimiter.New(a.store, rate)
				result = append(result, rateLimiterWithName{
					name:        l.Name,
					rateLimiter: newLimiter.(*slimiter.Limiter),
				})
			}
		}
	}

	// letters before first dot index is the bucket name for sp virtual host style
	firstDotIndex := strings.Index(host, ".")

	// if it is virtual host style, we need to add the bucket name from host to path for complete path to match pattern
	if virtualHost && firstDotIndex >= 0 {
		bucketName := host[:firstDotIndex]
		path = "/" + bucketName + path
	}
	for i := 0; i < len(a.cfg.PathPattern); i++ {
		pathPatternInSequence := a.cfg.PathSequence[i]
		// get limiters match pattern
		ls := a.cfg.PathPattern[strings.ToLower(pathPatternInSequence)]
		methodSeparatorIndex := strings.Index(pathPatternInSequence, MethodSeparator)
		// if methodSeparatorIndex not exist or methodSeparatorIndex is last character, pathPatternInSequence is invalid and shall be ignored
		if methodSeparatorIndex < 0 || methodSeparatorIndex == len(pathPatternInSequence)-1 {
			continue
		}
		pathMethod := pathPatternInSequence[:methodSeparatorIndex]
		pathPattern := pathPatternInSequence[methodSeparatorIndex+1:]

		// to find a limiter, the request path and request method shall both match the pathPatternInSequence
		if regexp.MustCompile(pathPattern).MatchString(path) && strings.EqualFold(pathMethod, method) {
			for _, l := range ls {
				rate, err := slimiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", l.RateLimit, l.RatePeriod))
				if err != nil {
					log.Errorw("failed to new rate from formatted", "err", err)
					continue
				}
				newLimiter = slimiter.New(a.store, rate)
				result = append(result, rateLimiterWithName{
					name:        l.Name,
					rateLimiter: newLimiter.(*slimiter.Limiter),
				})
			}
			// we only need to include the rate limiters of the first matching pattern
			break
		}
	}

	return result
}

func (t *apiLimiter) Allow(ctx context.Context, r *http.Request, domain string) bool {
	path := strings.ToLower(r.RequestURI)
	host := r.Host
	key := host + "-" + path
	key = strings.ToLower(key)

	var virtualHost bool
	if !strings.EqualFold(domain, r.Host) {
		virtualHost = true
	} else {
		virtualHost = false
	}

	rateLimiterWithNames := t.findLimiter(host, path, key, virtualHost, r.Method)
	if len(rateLimiterWithNames) == 0 {
		return true
	}

	allow := true
	// iterate through all map component, if any one reached limit, record false and continue, so all counters get increased
	for _, rateLimiterWName := range rateLimiterWithNames {
		limiterCtx, _ := t.store.Increment(ctx, rateLimiterWName.name, 1, rateLimiterWName.rateLimiter.Rate)

		if limiterCtx.Reached {
			allow = false
		}
	}
	return allow
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

func Limit(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(context.Background(), r, domain) {
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
		log.Errorw("failed to marshal error response", "gfsp_error", gfspErr.String(), "error", err)
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(int(gfspErr.GetHttpStatusCode()))
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write error response", "gfsp_error", gfspErr.String(), "error", err)
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
