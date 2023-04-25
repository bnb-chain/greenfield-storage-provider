package gateway

import (
	"bytes"
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/types/s3util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/jsonpb"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

// getUserBucketsHandler handle get object request
func (gateway *Gateway) getUserBucketsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		errDescription *errorDescription
		reqContext     *requestContext
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", getUserBucketsRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", getUserBucketsRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.metadata == nil {
		log.Error("failed to get user buckets due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	if ok := common.IsHexAddress(r.Header.Get(model.GnfdUserAddressHeader)); !ok {
		log.Errorw("failed to check account id", "account_id", reqContext.accountID, "error", err)
		errDescription = InvalidAddress
		return
	}

	req := &metatypes.GetUserBucketsRequest{
		AccountId: r.Header.Get(model.GnfdUserAddressHeader),
	}
	ctx := log.Context(context.Background(), req)
	resp, err := gateway.metadata.GetUserBuckets(ctx, req)
	if err != nil {
		log.Errorf("failed to get user buckets", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to get user buckets", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listObjectsByBucketNameHandler handle list objects by bucket name request
func (gateway *Gateway) listObjectsByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err               error
		b                 bytes.Buffer
		maxKeys           uint64
		errDescription    *errorDescription
		reqContext        *requestContext
		ok                bool
		requestBucketName string
		requestMaxKeys    string
		requestStartAfter string
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", listObjectsByBucketRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", listObjectsByBucketRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.metadata == nil {
		log.Error("failed to list objects by bucket name due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	requestBucketName = reqContext.bucketName
	requestMaxKeys = reqContext.request.URL.Query().Get("max_keys")
	requestStartAfter = reqContext.request.URL.Query().Get("start_after")

	if err = s3util.CheckValidBucketName(requestBucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", requestBucketName, "error", err)
		errDescription = InvalidBucketName
		return
	}

	if requestMaxKeys != "" {
		maxKeys, err = util.StringToUint64(requestMaxKeys)
		if err != nil {
			log.Errorw("failed to parse maxKeys", "max_keys", requestMaxKeys, "error", err)
			errDescription = InvalidMaxKeys
			return
		}
		// maxKeys should > 0
		if maxKeys == 0 {
			log.Errorw("failed to check maxKeys", "max_keys", requestMaxKeys, "error", err)
			errDescription = InvalidMaxKeys
			return
		}
	}

	// startAfter is an optional input, we only check its format when user input the value
	if ok = IsHexHash(requestStartAfter); !ok && requestStartAfter != "" {
		log.Errorw("failed to check startAfter", "start_after", requestStartAfter, "error", err)
		errDescription = InvalidStartAfter
		return
	}

	req := &metatypes.ListObjectsByBucketNameRequest{
		BucketName: requestBucketName,
		MaxKeys:    maxKeys,
		StartAfter: requestStartAfter,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := gateway.metadata.ListObjectsByBucketName(ctx, req)
	if err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}
