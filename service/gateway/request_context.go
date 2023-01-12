package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/inscription-storage-provider/util"
	"github.com/gorilla/mux"
)

// requestContext is a request context.
type requestContext struct {
	requestID string
	bucket    string
	object    string
	r         *http.Request
	startTime time.Time

	// admin
	action string
}

// newRequestContext return a request context.
func newRequestContext(r *http.Request) *requestContext {
	vars := mux.Vars(r)
	// admin router
	if mux.CurrentRoute(r).GetName() == "GetAuthentication" {
		var (
			bucket string
			object string
		)
		bucket = r.Header.Get(BFSResourceHeader)
		fields := strings.Split(bucket, "/")
		if len(fields) >= 2 {
			bucket = fields[0]
			object = strings.Join(fields[1:], "/")
		}
		return &requestContext{
			requestID: util.GenerateRequestID(),
			bucket:    bucket,
			object:    object,
			action:    vars["action"],
			r:         r,
			startTime: time.Now(),
		}
	}

	// bucket router
	return &requestContext{
		requestID: util.GenerateRequestID(),
		bucket:    vars["bucket"],
		object:    vars["object"],
		r:         r,
		startTime: time.Now(),
	}
}

// generateRequestDetail is used to log print detailed info.
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
