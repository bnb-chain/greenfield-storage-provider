package gateway

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	signer "github.com/bnb-chain/greenfield-go-sdk/keys/signer"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// requestContext is a request context.
type requestContext struct {
	requestID  string
	bucketName string
	objectName string
	request    *http.Request
	startTime  time.Time
	vars       map[string]string
}

// newRequestContext return a request context.
func newRequestContext(r *http.Request) *requestContext {
	// todo: sdk signature need ignore it, here will be deleted
	// https://github.com/minio/minio-go/blob/7aa4b0e0d1a9fdb4a99f50df715c21ec21d91753/pkg/signer/request-signature-v4.go#L60
	r.Header.Del("Accept-Encoding")

	vars := mux.Vars(r)
	return &requestContext{
		requestID:  util.GenerateRequestID(),
		bucketName: vars["bucket"],
		objectName: vars["object"],
		request:    r,
		startTime:  time.Now(),
		vars:       vars,
	}
}

// generateRequestDetail is used to log print detailed info.
func (requestContext *requestContext) generateRequestDetail() string {
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
		requestContext.requestID, requestContext.request.Host, requestContext.request.Method,
		requestContext.request.URL.String(), headerToString(requestContext.request.Header),
		getRequestIP(requestContext.request), time.Since(requestContext.startTime))
}

// signaturePrefix return supported Authorization prefix
func signaturePrefix(version, algorithm string) string {
	return version + " " + algorithm + ","
}

// verifySign used to verify request signature, return nil if check succeed
func (requestContext *requestContext) verifySignature() (sdk.AccAddress, error) {
	requestSignature := requestContext.request.Header.Get(model.GnfdAuthorizationHeader)
	v1SignaturePrefix := signaturePrefix(model.SignTypeV1, model.SignAlgorithm)
	if strings.HasPrefix(requestSignature, v1SignaturePrefix) {
		return requestContext.verifySignatureV1(requestSignature[len(v1SignaturePrefix):])
	}
	v2SignaturePrefix := signaturePrefix(model.SignTypeV2, model.SignAlgorithm)
	if strings.HasPrefix(requestSignature, v2SignaturePrefix) {
		return requestContext.verifySignatureV2(requestSignature[len(v2SignaturePrefix):])
	}
	return nil, errors.ErrUnsupportedSignType
}

// verifySignatureV1 used to verify request type v1 signature, return (address, nil) if check succeed
func (requestContext *requestContext) verifySignatureV1(requestSignature string) (sdk.AccAddress, error) {
	var (
		signedMsg string
		signature []byte
		err       error
	)
	requestSignature = strings.ReplaceAll(requestSignature, " ", "")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) < 2 {
		return nil, errors.ErrAuthorizationFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return nil, errors.ErrAuthorizationFormat
		}
		switch pair[0] {
		case model.SignedMsg:
			signedMsg = pair[1]
		case model.Signature:
			if signature, err = hex.DecodeString(pair[1]); err != nil {
				return nil, err
			}
		default:
			return nil, errors.ErrAuthorizationFormat
		}
	}

	// check request integrity
	realMsgToSign := commonhttp.GetMsgToSign(requestContext.request)
	if hex.EncodeToString(realMsgToSign) != signedMsg {
		log.Errorw("failed to check signed msg")
		return nil, errors.ErrRequestConsistent
	}

	// check signature consistent
	addr, pk, err := signer.RecoverAddr(realMsgToSign, signature)
	if err != nil {
		log.Errorw("failed to recover address")
		return nil, errors.ErrSignatureConsistent
	}
	if !secp256k1.VerifySignature(pk.Bytes(), realMsgToSign, signature[:len(signature)-1]) {
		log.Errorw("failed to verify signature")
		return nil, errors.ErrSignatureConsistent
	}
	return addr, nil
}

// verifySignatureV2 used to verify request type v2 signature, return (address, nil) if check succeed
func (requestContext *requestContext) verifySignatureV2(requestSignature string) (sdk.AccAddress, error) {
	var (
		signature []byte
		err       error
	)
	requestSignature = strings.ReplaceAll(requestSignature, " ", "")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) < 1 {
		return nil, errors.ErrAuthorizationFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return nil, errors.ErrAuthorizationFormat
		}
		switch pair[0] {
		case model.Signature:
			if signature, err = hex.DecodeString(pair[1]); err != nil {
				return sdk.AccAddress{}, err
			}
		default:
			return nil, errors.ErrAuthorizationFormat
		}
	}
	_ = signature
	// TODO: parse metamask signature and check timeout
	// return nil, errors.ErrUnsupportedSignType
	return sdk.AccAddress{}, nil
}

func parseRange(rangeStr string) (bool, int64, int64) {
	if rangeStr == "" {
		return false, -1, -1
	}
	rangeStr = strings.ToLower(rangeStr)
	rangeStr = strings.ReplaceAll(rangeStr, " ", "")
	if !strings.HasPrefix(rangeStr, "bytes=") {
		return false, -1, -1
	}
	rangeStr = rangeStr[len("bytes="):]
	if strings.HasSuffix(rangeStr, "-") {
		rangeStr = rangeStr[:len(rangeStr)-1]
		rangeStart, err := util.HeaderToUint64(rangeStr)
		if err != nil {
			return false, -1, -1
		}
		return true, int64(rangeStart), -1
	}
	pair := strings.Split(rangeStr, "-")
	if len(pair) == 2 {
		rangeStart, err := util.HeaderToUint64(pair[0])
		if err != nil {
			return false, -1, -1
		}
		rangeEnd, err := util.HeaderToUint64(pair[1])
		if err != nil {
			return false, -1, -1
		}
		return true, int64(rangeStart), int64(rangeEnd)
	}
	return false, -1, -1
}

// TODO: can be optimized by retirver
// checkAuthorization check addr authorization
func (g *Gateway) checkAuthorization(requestContext *requestContext, addr sdk.AccAddress) error {
	exist, err := g.chain.HasAccount(context.Background(), addr.String())
	if err != nil {
		log.Errorw("failed to check account on chain", "error", err, "address", addr.String())
		return err
	}
	if !exist {
		log.Errorw("account is not exist", "error", err, "address", addr.String())
		return fmt.Errorf("account is not exist")
	}

	switch mux.CurrentRoute(requestContext.request).GetName() {
	case putObjectRouterName:
		_, bucketExist, isInitStatus, tokenEnough, isSpBucket, ownObj, err := g.chain.AuthUploadObjectWithAccount(
			context.Background(),
			requestContext.bucketName,
			requestContext.objectName,
			addr.String(),
			g.config.StorageProvider)
		if err != nil {
			log.Errorw("failed to auth upload", "err", err,
				"bucket_name", requestContext.bucketName, "object_name", requestContext.objectName,
				"address", addr.String())
			return err
		}
		if !bucketExist || !isInitStatus || !tokenEnough || !isSpBucket || !ownObj {
			log.Errorw("failed to auth upload", "err", err,
				"bucket_name", requestContext.bucketName, "object_name", requestContext.objectName,
				"address", addr.String())
			return fmt.Errorf("account has no permission")
		}

	case getObjectRouterName:
		_, bucketExist, isServiceStatus, tokenEnough, isSpBucket, bucketID, readQuota, ownObj, err := g.chain.AuthDownloadObjectWithAccount(
			context.Background(),
			requestContext.bucketName,
			requestContext.objectName,
			addr.String(),
			g.config.StorageProvider)
		if err != nil {
			log.Errorw("failed to auth download", "err", err,
				"bucket_name", requestContext.bucketName, "object_name", requestContext.objectName,
				"address", addr.String())
			return err
		}
		if !bucketExist || !isServiceStatus || !tokenEnough || !isSpBucket || !ownObj {
			log.Errorw("failed to auth download", "err", err,
				"bucket_name", requestContext.bucketName, "object_name", requestContext.objectName,
				"address", addr.String())
			return fmt.Errorf("account has no permission")
		}
		// TODO: query read quota enough
		_, _ = bucketID, readQuota
		// TODO: update read quota
	}
	return nil
}
