package http

import (
	"bytes"
	"net/http"
	"strconv"
)

// responseWriterDelegator implements http.ResponseWriter and extracts the statusCode.
type responseWriterDelegator struct {
	w          http.ResponseWriter
	written    bool
	size       int
	statusCode int
	body       bytes.Buffer
}

func (wd *responseWriterDelegator) Header() http.Header {
	return wd.w.Header()
}

func (wd *responseWriterDelegator) Write(b []byte) (int, error) {
	if wd.statusCode == 0 {
		wd.statusCode = http.StatusOK
	}
	// write response body to customized body
	wd.body.Write(b)
	// write response body to http.ResponseWriter
	n, err := wd.w.Write(b)
	wd.size += n
	return n, err
}

func (wd *responseWriterDelegator) GetBody() []byte {
	return wd.body.Bytes()
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

// computeApproximateRequestSize compute HTTP request size
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s += len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
