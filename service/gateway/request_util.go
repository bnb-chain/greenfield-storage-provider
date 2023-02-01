package gateway

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-sdk-go/pkg/signer"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/mux"
)

// requestContext is a request context.
type requestContext struct {
	requestID  string
	bucketName string
	objectName string
	request    *http.Request
	startTime  time.Time

	// admin fields
	actionName string
}

// newRequestContext return a request context.
func newRequestContext(r *http.Request) *requestContext {
	vars := mux.Vars(r)
	// todo: sdk signature need ignore it, here will be deleted
	// https://github.com/minio/minio-go/blob/7aa4b0e0d1a9fdb4a99f50df715c21ec21d91753/pkg/signer/request-signature-v4.go#L60
	r.Header.Del("Accept-Encoding")
	// admin router
	if mux.CurrentRoute(r).GetName() == "GetAuthentication" {
		var (
			bucket string
			object string
		)
		bucket = r.Header.Get(model.GnfdResourceHeader)
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
			request:    r,
			startTime:  time.Now(),
		}
	}

	// bucket router
	return &requestContext{
		requestID:  util.GenerateRequestID(),
		bucketName: vars["bucket"],
		objectName: vars["object"],
		request:    r,
		startTime:  time.Now(),
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

const (
	SignedMsg = "SignedMsg"
	Signature = "Signature"
)

var (
	MetaMaskStr = hex.EncodeToString([]byte("MetaMaskStr"))
)

var (
	ErrAuthFormat          = errors.New("authorization format error")
	ErrRequestConsistent   = errors.New("request consistent check failed")
	ErrSignatureConsistent = errors.New("signature consistent check failed")
	ErrUnsupportedSignType = errors.New("not support signature type")
)

// verifySign used to verify request signature, return nil if check succeed
func (requestContext *requestContext) verifySign() error {
	requestSignature := requestContext.request.Header.Get(model.GnfdAuthorizationHeader)
	if strings.HasPrefix(requestSignature, model.SignTypeV1+" "+model.SignAlgorithm+",") {
		return requestContext.verifySignV1(requestSignature[len(model.SignTypeV1)+1+len(model.SignAlgorithm)+1:])
	}
	if strings.HasPrefix(requestSignature, model.SignTypeV2+" "+model.SignAlgorithm+",") {
		return requestContext.verifySignV2(requestSignature[len(model.SignTypeV1)+1+len(model.SignAlgorithm)+1:])
	}
	return ErrUnsupportedSignType
}

// verifySignV1 used to verify request type v1 signature, return nil if check succeed
func (requestContext *requestContext) verifySignV1(requestSignature string) error {
	var (
		signedMsg string
		signature []byte
		err       error
	)
	requestSignature = strings.ReplaceAll(requestSignature, " ", "")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) < 2 {
		return ErrAuthFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return ErrAuthFormat
		}
		switch pair[0] {
		case SignedMsg:
			signedMsg = pair[1]
		case Signature:
			if signature, err = hex.DecodeString(pair[1]); err != nil {
				return err
			}
		default:
			return ErrAuthFormat
		}
	}

	// check request integrity
	if hex.EncodeToString(signer.GetMsgToSign(*requestContext.request)) != signedMsg {
		return ErrRequestConsistent
	}

	// check signature consistent
	signedRequestHash := crypto.Keccak256([]byte(signedMsg))
	_, pk, err := signer.RecoverAddr(signedRequestHash, signature)
	if err != nil {
		return ErrSignatureConsistent
	}
	if !secp256k1.VerifySignature(pk.Bytes(), signedRequestHash, signature[:len(signature)-1]) {
		return ErrSignatureConsistent
	}
	return nil
}

// verifySignV2 used to verify request type v2 signature, return nil if check succeed
func (requestContext *requestContext) verifySignV2(requestSignature string) error {
	var (
		signature []byte
		err       error
	)
	requestSignature = strings.ReplaceAll(requestSignature, " ", "")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) < 1 {
		return ErrAuthFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return ErrAuthFormat
		}
		switch pair[0] {
		case Signature:
			if signature, err = hex.DecodeString(pair[1]); err != nil {
				return err
			}
		default:
			return ErrAuthFormat
		}
	}
	_ = signature
	// todo: parse metamask signature and check timeout
	/*
		// check signature consistent
		signedRequestHash := crypto.Keccak256([]byte(MetaMaskStr))
		_, pk, err := signer.RecoverAddr(signedRequestHash, signature)
		if err != nil {
			return ErrSignatureConsistent
		}
		if !secp256k1.VerifySignature(pk.Bytes(), signedRequestHash, signature[:len(signature)-1]) {
			return ErrSignatureConsistent
		}
	*/
	return nil
}

// redundancyType can be EC or Replica, if != EC, default is Replica
func redundancyTypeToEnum(redundancyType string) ptypes.RedundancyType {
	if redundancyType == model.ReplicaRedundancyTypeHeaderValue {
		return ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE
	}
	return ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED
}
