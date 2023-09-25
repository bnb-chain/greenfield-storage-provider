package http

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strconv"

	modelgateway "github.com/bnb-chain/greenfield-storage-provider/model/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
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
	if wd.statusCode != http.StatusOK {
		// write response body to customized body which used for metrics
		wd.body.Write(b)
	}
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

func (wd *responseWriterDelegator) GetSPErrorCode() string {
	// get error response code if exists
	var (
		errorResp = &modelgateway.ErrorResponse{}
		errorCode string
	)
	if wd.statusCode == http.StatusOK {
		errorCode = "0" // no error
	} else {
		body := wd.GetBody()
		err := xml.Unmarshal(body, errorResp)
		if err != nil {
			log.Errorw("cannot parse gateway error response", "error", err)
			errorCode = "-1" // unknown error code
			return errorCode
		}
		errorCode = strconv.Itoa(int(errorResp.Code))
	}
	return errorCode
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
