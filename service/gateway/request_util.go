package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	pbPkg "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/gorilla/mux"
)

// requestContext is a request context.
type requestContext struct {
	requestID  string
	bucketName string
	objectName string
	r          *http.Request
	startTime  time.Time

	// admin fields
	actionName string
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
		bucket = r.Header.Get(model.BFSResourceHeader)
		fields := strings.Split(bucket, "/")
		if len(fields) >= 2 {
			bucket = fields[0]
			object = strings.Join(fields[1:], "/")
		}
		return &requestContext{
			requestID:  util.GenerateRequestID(),
			bucketName: bucket,
			objectName: object,
			actionName: vars["action"],
			r:          r,
			startTime:  time.Now(),
		}
	}

	// bucket router
	return &requestContext{
		requestID:  util.GenerateRequestID(),
		bucketName: vars["bucket"],
		objectName: vars["object"],
		r:          r,
		startTime:  time.Now(),
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

// redundancyType can be EC or Replica, if != EC, default is Replica
func redundancyTypeToEnum(redundancyType string) pbPkg.RedundancyType {
	if redundancyType == model.ReplicaRedundancyTypeHeaderValue {
		return pbPkg.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE
	}
	return pbPkg.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED

}
