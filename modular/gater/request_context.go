package gater

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/mux"

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

// SetHttpCode sets the http status code for logging and debugging.
func (r *RequestContext) SetHttpCode(code int) {
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
			if k == GnfdUnsignedApprovalMsgHeader || k == GnfdReplicatePieceApprovalHeader || k == GnfdReceiveMsgHeader {
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

// SkipVerifyAuthentication is temporary to Compatible SignatureV2
func (r *RequestContext) SkipVerifyAuthentication() bool {
	v2SignaturePrefix := signaturePrefix(SignTypeV2, SignAlgorithm)
	if strings.HasPrefix(r.request.Header.Get(GnfdAuthorizationHeader), v2SignaturePrefix) {
		return true
	}
	return false
}

// signaturePrefix return supported Authentication prefix
func signaturePrefix(version, algorithm string) string {
	return version + " " + algorithm + ","
}

func (r *RequestContext) VerifySignature() (string, error) {
	requestSignature := r.request.Header.Get(GnfdAuthorizationHeader)
	v1SignaturePrefix := signaturePrefix(SignTypeV1, SignAlgorithm)
	if strings.HasPrefix(requestSignature, v1SignaturePrefix) {
		accAddress, err := r.verifySignatureV1(requestSignature[len(v1SignaturePrefix):])
		if err != nil {
			return "", err
		}
		return accAddress.String(), nil
	}
	v2SignaturePrefix := signaturePrefix(SignTypeV2, SignAlgorithm)
	if strings.HasPrefix(requestSignature, v2SignaturePrefix) {
		return "", nil
	}

	OffChainSignaturePrefix := signaturePrefix(SignTypeOffChain, SignAlgorithmEddsa)
	if strings.HasPrefix(requestSignature, OffChainSignaturePrefix) {
		accAddress, err := r.verifyOffChainSignature(requestSignature[len(OffChainSignaturePrefix):])
		if err != nil {
			return "", err
		}
		return accAddress.String(), nil
	}
	return "", ErrUnsupportedSignType
}

// verifySignatureV1 used to verify request type v1 signature, return (address, nil) if check succeed
func (r *RequestContext) verifySignatureV1(requestSignature string) (sdk.AccAddress, error) {
	var (
		signedMsg string
		signature []byte
		err       error
	)
	requestSignature = strings.ReplaceAll(requestSignature, " ", "")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) < 2 {
		return nil, ErrAuthorizationHeaderFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return nil, ErrAuthorizationHeaderFormat
		}
		switch pair[0] {
		case SignedMsg:
			signedMsg = pair[1]
		case Signature:
			if signature, err = hex.DecodeString(pair[1]); err != nil {
				return nil, err
			}
		default:
			return nil, ErrAuthorizationHeaderFormat
		}
	}

	// check request integrity
	realMsgToSign := commonhttp.GetMsgToSign(r.request)
	if hex.EncodeToString(realMsgToSign) != signedMsg {
		log.CtxErrorw(r.ctx, "failed to check signed msg")
		return nil, ErrRequestConsistent
	}

	// check signature consistent
	addr, pk, err := RecoverAddr(realMsgToSign, signature)
	if err != nil {
		log.CtxErrorw(r.ctx, "failed to recover address")
		return nil, ErrRequestConsistent
	}
	if !secp256k1.VerifySignature(pk.Bytes(), realMsgToSign, signature[:len(signature)-1]) {
		log.CtxErrorw(r.ctx, "failed to verify signature")
		return nil, ErrRequestConsistent
	}
	return addr, nil
}

// RecoverAddr recovers the sender address from msg and signature
// TODO: move it to greenfield-common
func RecoverAddr(msg []byte, sig []byte) (sdk.AccAddress, ethsecp256k1.PubKey, error) {
	pubKeyByte, err := secp256k1.RecoverPubkey(msg, sig)
	if err != nil {
		return nil, ethsecp256k1.PubKey{}, err
	}
	pubKey, _ := ethcrypto.UnmarshalPubkey(pubKeyByte)
	pk := ethsecp256k1.PubKey{
		Key: ethcrypto.CompressPubkey(pubKey),
	}
	recoverAcc := sdk.AccAddress(pk.Address().Bytes())
	return recoverAcc, pk, nil
}

// verifyOffChainSignature used to verify off-chain-auth signature, return (address, nil) if check succeed
func (r *RequestContext) verifyOffChainSignature(requestSignature string) (sdk.AccAddress, error) {
	var (
		signedMsg *string
		err       error
	)
	signedMsg, sigString, err := parseSignedMsgAndSigFromRequest(requestSignature)
	if err != nil {
		return nil, err
	}

	account := r.request.Header.Get(GnfdUserAddressHeader)
	domain := r.request.Header.Get(GnfdOffChainAuthAppDomainHeader)
	offChainSig := *sigString
	realMsgToSign := *signedMsg

	verifyOffChainSignatureResp, err := r.g.baseApp.GfSpClient().VerifyOffChainSignature(r.Context(), account, domain, offChainSig, realMsgToSign)
	if err != nil {
		log.Errorf("failed to verify off chain signature", "error", err)
		return nil, err
	}
	if verifyOffChainSignatureResp {
		userAddress, _ := sdk.AccAddressFromHexUnsafe(r.request.Header.Get(GnfdUserAddressHeader))
		return userAddress, nil
	} else {
		return nil, err
	}
}

func parseSignedMsgAndSigFromRequest(requestSignature string) (*string, *string, error) {
	var (
		signedMsg string
		signature string
	)
	requestSignature = strings.ReplaceAll(requestSignature, "\\n", "\n")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) != 2 {
		return nil, nil, ErrAuthorizationHeaderFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return nil, nil, ErrAuthorizationHeaderFormat
		}
		switch pair[0] {
		case SignedMsg:
			signedMsg = pair[1]
		case Signature:
			signature = pair[1]
		default:
			return nil, nil, ErrAuthorizationHeaderFormat
		}
	}

	return &signedMsg, &signature, nil
}
