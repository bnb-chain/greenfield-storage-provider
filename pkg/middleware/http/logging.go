package http

import (
	"net/http"
	"strconv"
)

// func LoggingHTTPResponse(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		wd := &responseWriterDelegator{w: w}
// 		next.ServeHTTP(wd, r)
//
// 		method := r.Method
// 		code := wd.StatusCode()
// 		handlerName := mux.CurrentRoute(r).GetName()
//
// 		if code != http.StatusOK {
// 			log.Errorf("action(%v) statusCode(%v) %v", getBucketReadQuotaRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
// 		}
// 	})
// }

// responseWriterDelegator implements http.ResponseWriter and extracts the statusCode.
type responseWriterDelegator struct {
	w          http.ResponseWriter
	written    bool
	size       int
	statusCode int
}

func (wd *responseWriterDelegator) Header() http.Header {
	return wd.w.Header()
}

func (wd *responseWriterDelegator) Write(bytes []byte) (int, error) {
	if wd.statusCode == 0 {
		wd.statusCode = http.StatusOK
	}
	n, err := wd.w.Write(bytes)
	wd.size += n
	return n, err
}

func (wd *responseWriterDelegator) WriteHeader(statusCode int) {
	wd.written = true
	wd.statusCode = statusCode
	wd.w.WriteHeader(statusCode)
}

func (wd *responseWriterDelegator) StatusCode() int {
	if !wd.written {
		return http.StatusOK
	}
	return wd.statusCode
}

func (wd *responseWriterDelegator) Status() string {
	return strconv.Itoa(wd.StatusCode())
}
