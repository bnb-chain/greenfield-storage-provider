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
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// requestContext is a request context.
type requestContext struct {
	requestID  string
	bucketName string
	objectName string
	request    *http.Request
	startTime  time.Time
	vars       map[string]string
	// TODO: for auth v2 test, remove it in the future
	skipAuth   bool
	bucketInfo *storagetypes.BucketInfo
	objectInfo *storagetypes.ObjectInfo
	// accountID is used to provide authentication to the sp
	accountID string
}

// newRequestContext return a request context.
func newRequestContext(r *http.Request) *requestContext {
	vars := mux.Vars(r)
	return &requestContext{
		requestID:  util.GenerateRequestID(),
		bucketName: vars["bucket"],
		objectName: vars["object"],
		accountID:  vars["account_id"],
		request:    r,
		startTime:  time.Now(),
		vars:       vars,
	}
}

// generateRequestDetail is used to log print detailed info.
func (reqContext *requestContext) generateRequestDetail() string {
	var headerToString = func(header http.Header) string {
		var sb = strings.Builder{}
		for k := range header {
			if k == model.GnfdObjectInfoHeader || k == model.GnfdUnsignedApprovalMsgHeader {
				continue
			}
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
		reqContext.requestID, reqContext.request.Host, reqContext.request.Method,
		reqContext.request.URL.String(), headerToString(reqContext.request.Header),
		getRequestIP(reqContext.request), time.Since(reqContext.startTime))
}

// signaturePrefix return supported Authorization prefix
func signaturePrefix(version, algorithm string) string {
	return version + " " + algorithm + ","
}

// verifySign used to verify request signature, return nil if check succeed
func (reqContext *requestContext) verifySignature() (sdk.AccAddress, error) {
	requestSignature := reqContext.request.Header.Get(model.GnfdAuthorizationHeader)
	v1SignaturePrefix := signaturePrefix(model.SignTypeV1, model.SignAlgorithm)
	if strings.HasPrefix(requestSignature, v1SignaturePrefix) {
		return reqContext.verifySignatureV1(requestSignature[len(v1SignaturePrefix):])
	}
	v2SignaturePrefix := signaturePrefix(model.SignTypeV2, model.SignAlgorithm)
	if strings.HasPrefix(requestSignature, v2SignaturePrefix) {
		return reqContext.verifySignatureV2(requestSignature[len(v2SignaturePrefix):])
	}
	return nil, errors.ErrUnsupportedSignType
}

// verifySignatureV1 used to verify request type v1 signature, return (address, nil) if check succeed
func (reqContext *requestContext) verifySignatureV1(requestSignature string) (sdk.AccAddress, error) {
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
	realMsgToSign := commonhttp.GetMsgToSign(reqContext.request)
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
func (reqContext *requestContext) verifySignatureV2(requestSignature string) (sdk.AccAddress, error) {
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
	reqContext.skipAuth = true
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
		rangeStart, err := util.StringToUint64(rangeStr)
		if err != nil {
			return false, -1, -1
		}
		return true, int64(rangeStart), -1
	}
	pair := strings.Split(rangeStr, "-")
	if len(pair) == 2 {
		rangeStart, err := util.StringToUint64(pair[0])
		if err != nil {
			return false, -1, -1
		}
		rangeEnd, err := util.StringToUint64(pair[1])
		if err != nil {
			return false, -1, -1
		}
		return true, int64(rangeStart), int64(rangeEnd)
	}
	return false, -1, -1
}

// TODO: can be optimized by retriever
// checkAuthorization check addr authorization
func (g *Gateway) checkAuthorization(reqContext *requestContext, addr sdk.AccAddress) error {
	var (
		err          error
		accountExist bool
	)
	if reqContext.skipAuth {
		return nil
	}
	accountExist, err = g.chain.HasAccount(context.Background(), addr.String())
	if err != nil {
		log.Errorw("failed to check account on chain", "address", addr.String(), "error", err)
		return err
	}
	if !accountExist {
		log.Errorw("account is not exist", "address", addr.String(), "error", err)
		return errors.ErrHasNoPermission
	}

	switch mux.CurrentRoute(reqContext.request).GetName() {
	case putObjectRouterName:
		if reqContext.bucketInfo, reqContext.objectInfo, err = g.chain.QueryBucketInfoAndObjectInfo(
			context.Background(), reqContext.bucketName, reqContext.objectName); err != nil {
			log.Errorw("failed to query bucket info and object info on chain",
				"bucket_name", reqContext.bucketName, "object_name", reqContext.objectName, "error", err)
			return err
		}
		if reqContext.objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_INIT {
			log.Errorw("failed to auth due to object status is not in inited",
				"object_status", reqContext.objectInfo.GetObjectStatus())
			return errors.ErrCheckObjectState
		}
		if reqContext.objectInfo.GetOwner() != addr.String() {
			log.Errorw("failed to auth due to account is not equal to object owner",
				"object_owner", reqContext.objectInfo.GetOwner(),
				"request_address", addr.String())
			return errors.ErrHasNoPermission
		}
		if reqContext.bucketInfo.GetPrimarySpAddress() != g.config.SpOperatorAddress {
			log.Errorw("failed to auth due to bucket primary sp is not equal to current sp",
				"bucket_primary_sp", reqContext.bucketInfo.GetPrimarySpAddress(),
				"current_sp", g.config.SpOperatorAddress)
			return errors.ErrHasNoPermission
		}

	case getObjectRouterName:
		if reqContext.bucketInfo, reqContext.objectInfo, err = g.chain.QueryBucketInfoAndObjectInfo(
			context.Background(), reqContext.bucketName, reqContext.objectName); err != nil {
			log.Errorw("failed to query bucket info and object info on chain",
				"bucket_name", reqContext.bucketName, "object_name", reqContext.objectName, "error", err)
			return err
		}
		if reqContext.objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_IN_SERVICE {
			log.Errorw("object is not in service",
				"status", reqContext.objectInfo.GetObjectStatus())
			return errors.ErrCheckObjectState
		}
		if reqContext.objectInfo.GetOwner() != addr.String() {
			log.Errorw("failed to auth due to account is not equal to object owner",
				"object_owner", reqContext.objectInfo.GetOwner(),
				"request_address", addr.String())
			return errors.ErrHasNoPermission
		}
		if reqContext.bucketInfo.GetPrimarySpAddress() != g.config.SpOperatorAddress {
			log.Errorw("failed to auth due to bucket primary sp is not equal to current sp",
				"bucket_primary_sp", reqContext.bucketInfo.GetPrimarySpAddress(),
				"current_sp", g.config.SpOperatorAddress)
			return errors.ErrHasNoPermission
		}
	}
	return nil
}

// checkBilling check payment account status and traffic quota
func (g *Gateway) checkBilling(reqContext *requestContext, addr sdk.AccAddress) error {
	switch mux.CurrentRoute(reqContext.request).GetName() {
	case getObjectRouterName:
		streamRecord, err := g.chain.QueryStreamRecord(context.Background(), reqContext.bucketInfo.PaymentAddress)
		if err != nil {
			log.Errorw("failed to check billing", "error", err)
			return err
		}
		// TODO: need update to enum by the latest greenfield release tag
		if streamRecord.Status != 0 {
			log.Errorw("failed to check billing due to payment account status", "status", streamRecord.Status)
			return errors.ErrCheckBilling
		}
		// TODO: support range read size
		if err = g.spDB.CheckQuotaAndAddReadRecord(
			&sqldb.ReadRecord{
				BucketID:    reqContext.bucketInfo.Id.Uint64(),
				ObjectID:    reqContext.objectInfo.Id.Uint64(),
				UserAddress: addr.String(),
				BucketName:  reqContext.bucketInfo.GetBucketName(),
				ObjectName:  reqContext.objectInfo.GetObjectName(),
				ReadSize:    int64(reqContext.objectInfo.PayloadSize),
				ReadTime:    sqldb.GetCurrentUnixTime(),
			},
			&sqldb.BucketQuota{
				ReadQuotaSize: int64(reqContext.bucketInfo.GetReadQuota()),
			},
		); err != nil {
			log.Errorw("failed to check billing due to bucket quota", "error", err)
			return err
		}
	}
	return nil
}
