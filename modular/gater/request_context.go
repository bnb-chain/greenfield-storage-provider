package gater

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"

	commonhash "github.com/bnb-chain/greenfield-common/go/hash"
	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// RequestContext generates from http request, it records the common info
// for handler to use.
type RequestContext struct {
	g          *GateModular
	request    *http.Request
	routerName string
	bucketName string
	objectName string
	account    string // account is used to provide authentication to the sp
	vars       map[string]string
	httpCode   int

	// ctx is the runtime context for the request
	ctx context.Context
	// the cancel func of ctx, it should be called at the end of the request
	// to make sure the connection to the back server is released.
	cancel    func()
	err       error
	startTime time.Time
}

var skipAuthRouterNames = []string{
	requestNonceRouterName,
	updateUserPublicKeyRouterName, // this will skip general auth algorithms first and use specific "personal sign" later
	downloadObjectByUniversalEndpointName,
	viewObjectByUniversalEndpointName,
	getUserBucketsRouterName,
	listObjectsByBucketRouterName,
	getObjectMetaRouterName,
	getBucketMetaRouterName,
	listBucketsByIDsRouterName,
	listObjectsByIDsRouterName,
}

// NewRequestContext returns an instance of RequestContext, and verify the
// request signature, returns the instance regardless of the success or
// failure of the verification.
func NewRequestContext(r *http.Request, g *GateModular) (*RequestContext, error) {
	vars := mux.Vars(r)
	routerName := ""
	if mux.CurrentRoute(r) != nil {
		routerName = mux.CurrentRoute(r).GetName()
	}
	ctx, cancel := context.WithCancel(context.Background())
	reqCtx := &RequestContext{
		g:          g,
		ctx:        ctx,
		cancel:     cancel,
		request:    r,
		routerName: routerName,
		bucketName: vars["bucket"],
		objectName: vars["object"],
		account:    vars["account_id"],
		vars:       vars,
		startTime:  time.Now(),
	}
	log.Infof("routerName is %s", routerName)
	if slices.Contains(skipAuthRouterNames, routerName) {
		return reqCtx, nil
	}

	err := reqCtx.CheckIfSigExpiry()
	if err != nil {
		return reqCtx, err
	}

	account, err := reqCtx.VerifySignature()
	if err != nil {
		return reqCtx, err
	}
	reqCtx.account = account
	return reqCtx, nil
}

// Context returns the RequestContext runtime context.
func (r *RequestContext) Context() context.Context {
	return r.ctx
}

// Account returns the account who send the request.
func (r *RequestContext) Account() string {
	return r.account
}

// Cancel releases the runtime context.
func (r *RequestContext) Cancel() {
	r.cancel()
}

// SetHTTPCode sets the http status code for logging and debugging.
func (r *RequestContext) SetHTTPCode(code int) {
	r.httpCode = code
}

// SetError sets the request err to RequestContext for logging and debugging.
func (r *RequestContext) SetError(err error) {
	r.err = err
}

// String shows the detail result of the request for logging and debugging.
func (r *RequestContext) String() string {
	var headerToString = func(header http.Header) string {
		var sb = strings.Builder{}
		for k := range header {
			if k == GnfdUnsignedApprovalMsgHeader || k == GnfdReplicatePieceApprovalHeader || k == GnfdReceiveMsgHeader ||
				k == GnfdRecoveryMsgHeader || k == GnfdMigratePieceMsgHeader || k == commonhttp.HTTPHeaderAuthorization {
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
	return fmt.Sprintf("HttpStatusCode[%d] action[%s] host[%v] method[%v] url[%v] header[%v] remote[%v] cost[%v] error[%v]",
		r.httpCode, r.routerName, r.request.Host, r.request.Method, r.request.URL.String(), headerToString(r.request.Header),
		getRequestIP(r.request), time.Since(r.startTime), r.err)
}

func (r *RequestContext) VerifySignature() (string, error) {
	// check sig
	requestSignature := r.request.Header.Get(GnfdAuthorizationHeader)

	// GNFD1-ECDSA
	gnfd1EcdsaSignaturePrefix := commonhttp.Gnfd1Ecdsa + ","
	if strings.HasPrefix(requestSignature, gnfd1EcdsaSignaturePrefix) {
		accAddress, err := r.verifySignatureForGNFD1Ecdsa(requestSignature[len(gnfd1EcdsaSignaturePrefix):])
		if err != nil {
			return "", err
		}
		return accAddress.String(), nil
	}

	// GNFD1-EDDSA
	gnfd1EddsaSignaturePrefix := commonhttp.Gnfd1Eddsa + ","
	if strings.HasPrefix(requestSignature, gnfd1EddsaSignaturePrefix) {
		accAddress, err := r.verifySignatureForGNFD1Eddsa(requestSignature[len(gnfd1EddsaSignaturePrefix):])
		if err != nil {
			return "", err
		}
		return accAddress.String(), nil
	}

	return "", ErrUnsupportedSignType

}

func (r *RequestContext) CheckIfSigExpiry() error {
	// check expiry header
	requestExpiredTimestamp := r.request.Header.Get(commonhttp.HTTPHeaderExpiryTimestamp)
	if requestExpiredTimestamp == "" {
		requestExpiredTimestamp = r.request.URL.Query().Get(commonhttp.HTTPHeaderExpiryTimestamp)
	}
	expiryDate, parseErr := time.Parse(ExpiryDateFormat, requestExpiredTimestamp)
	if parseErr != nil {
		return ErrInvalidExpiryDateHeader
	}
	expiryAge := int32(time.Until(expiryDate).Seconds())
	if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
		return ErrInvalidExpiryDateHeader
	}
	return nil
}

// verifySignatureForGNFD1Ecdsa used to verify request type GNFD1_ECDSA, return (address, nil) if check succeed
func (r *RequestContext) verifySignatureForGNFD1Ecdsa(requestSignature string) (sdk.AccAddress, error) {
	var (
		signature []byte
		err       error
	)
	sigStr, err := parseSignatureFromRequest(requestSignature)
	if err != nil {
		return nil, err
	}
	if signature, err = hex.DecodeString(sigStr); err != nil {
		return nil, err
	}
	// check request integrity
	realMsgToSign := commonhttp.GetMsgToSignInGNFD1Auth(r.request)

	// check signature consistent
	addr, _, err := commonhash.RecoverAddr(realMsgToSign, signature)
	if err != nil {
		log.CtxErrorw(r.ctx, "failed to recover address")
		return nil, ErrRequestConsistent
	}
	return addr, nil
}

// verifyTaskSignature verify the task signature and return the sender address
func (r *RequestContext) verifyTaskSignature(taskMsgBytes []byte, taskSignature []byte) (sdk.AccAddress, error) {
	if len(taskMsgBytes) != 32 {
		taskMsgBytes = crypto.Keccak256(taskMsgBytes)
	}

	addr, pk, err := commonhash.RecoverAddr(taskMsgBytes, taskSignature)
	if err != nil {
		log.CtxErrorw(r.Context(), "failed to recover address", "error", err)
		return nil, err
	}
	if !secp256k1.VerifySignature(pk.Bytes(), taskMsgBytes, taskSignature[:len(taskSignature)-1]) {
		log.CtxErrorw(r.ctx, "failed to verify task signature")
		return nil, err
	}
	return addr, nil
}

// verifySignatureForGNFD1Eddsa used to verify off-chain-auth signature, return (address, nil) if check succeed
func (r *RequestContext) verifySignatureForGNFD1Eddsa(requestSignature string) (sdk.AccAddress, error) {
	var err error
	offChainSig, err := parseSignatureFromRequest(requestSignature)
	if err != nil {
		return nil, err
	}

	// check request integrity
	realMsgToSign := commonhttp.GetMsgToSignInGNFD1Auth(r.request)
	account := r.request.Header.Get(GnfdUserAddressHeader)
	domain := r.request.Header.Get(GnfdOffChainAuthAppDomainHeader)

	_, err = r.g.baseApp.GfSpClient().VerifyGNFD1EddsaSignature(r.Context(), account, domain, offChainSig, realMsgToSign)
	if err != nil {
		log.Errorf("failed to verify signature for GNFD1-Eddsa", "error", err)
		return nil, err
	} else {
		userAddress, _ := sdk.AccAddressFromHexUnsafe(r.request.Header.Get(GnfdUserAddressHeader))
		return userAddress, nil
	}
}

// verifyOffChainSignatureFromPreSignedURL used to verify off-chain-auth signature, return (address, nil) if check succeed. The auth information will be parsed from URL.
func (r *RequestContext) verifyGNFD1EddsaSignatureFromPreSignedURL(authenticationStr string, account string, domain string) (sdk.AccAddress, error) {
	var err error
	offChainSig, err := parseSignatureFromRequest(authenticationStr)
	if err != nil {
		return nil, err
	}

	// check request integrity
	realMsgToSign := commonhttp.GetMsgToSignInGNFD1AuthForPreSignedURL(r.request)
	_, err = r.g.baseApp.GfSpClient().VerifyGNFD1EddsaSignature(r.Context(), account, domain, offChainSig, realMsgToSign)
	if err != nil {
		log.Errorf("failed to verify off chain signature", "error", err)
		return nil, err
	} else {
		userAddress, _ := sdk.AccAddressFromHexUnsafe(account)
		return userAddress, nil
	}
}

// parseSignatureFromRequest get sig for both ECDSA and EDDSA auth, it expects the auth string should look like "Signature=xxxxx".
func parseSignatureFromRequest(requestSignature string) (string, error) {
	var signature string
	requestSignature = strings.ReplaceAll(requestSignature, " ", "")
	pair := strings.Split(requestSignature, "=")
	if len(pair) != 2 {
		return "", ErrAuthorizationHeaderFormat
	}
	switch pair[0] {
	case Signature:
		signature = pair[1]
		return signature, nil
	default:
		return "", ErrAuthorizationHeaderFormat
	}
}
