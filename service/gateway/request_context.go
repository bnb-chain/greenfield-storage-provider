package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type requestContext struct {
	requestID string
	bucket    string
	object    string
	r         *http.Request
	startTime time.Time
}

func generateRequestID() (string, error) {
	var uUID uuid.UUID
	var err error
	if uUID, err = uuid.NewRandom(); err != nil {
		return "", err
	}
	return strings.ReplaceAll(uUID.String(), "-", ""), nil
}

func newRequestContext(r *http.Request) *requestContext {
	requestID, err := generateRequestID()
	if err != nil {
		log.Warnw("generate request id failed", "err", err)
	}
	vars := mux.Vars(r)
	return &requestContext{
		requestID: requestID,
		bucket:    vars["bucket"],
		object:    vars["object"],
		r:         r,
		startTime: time.Now(),
	}
}

func generateRequestDetail(r *requestContext) string {
	var headerToString = func(header http.Header) string {
		var sb = strings.Builder{}
		for k := range header {
			if sb.Len() != 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("%v:[%v]", k, header.Get(k)))
		}
		return "{" + sb.String() + "}"
	}
	var getRequestIP = func(r *http.Request) string {
		IPAddress := r.Header.Get("X-Real-Ip")
		if IPAddress == "" {
			IPAddress = r.Header.Get("X-Forwarded-For")
		}
		if IPAddress == "" {
			IPAddress = r.RemoteAddr
		}
		if ok := strings.Contains(IPAddress, ":"); ok {
			IPAddress = strings.Split(IPAddress, ":")[0]
		}
		return IPAddress
	}
	return fmt.Sprintf("requestID(%v) host(%v) method(%v) url(%v) header(%v) remote(%v) cost(%v)",
		r.requestID, r.r.Host, r.r.Method, r.r.URL.String(), headerToString(r.r.Header), getRequestIP(r.r), time.Since(r.startTime))
}
