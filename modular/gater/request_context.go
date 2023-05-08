package gater

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/mux"
)

type RequestContext struct {
	request    *http.Request
	routerName string
	bucketName string
	objectName string
	account    string // account is used to provide authentication to the sp
	vars       map[string]string

	ctx       context.Context
	cancel    func()
	err       error
	startTime time.Time
}

func NewRequestContext(r *http.Request) *RequestContext {
	vars := mux.Vars(r)
	routerName := ""
	if mux.CurrentRoute(r) != nil {
		routerName = mux.CurrentRoute(r).GetName()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &RequestContext{
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
}

func (r *RequestContext) Context() context.Context {
	return r.ctx
}

func (r *RequestContext) Cancel() {
	r.cancel()
}

func (r *RequestContext) SetError(err error) {
	r.err = err
}

func (r *RequestContext) String() string {
	var headerToString = func(header http.Header) string {
		var sb = strings.Builder{}
		for k := range header {
			if k == model.GnfdUnsignedApprovalMsgHeader {
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
	return fmt.Sprintf("action(%s) host(%v) method(%v) url(%v) header(%v) remote(%v) cost(%v) error(%s)",
		r.routerName, r.request.Host, r.request.Method, r.request.URL.String(), headerToString(r.request.Header),
		getRequestIP(r.request), time.Since(r.startTime), r.err.Error())
}

func (r *RequestContext) NeedVerifySignature() bool {
	requestSignature := r.request.Header.Get(model.GnfdAuthorizationHeader)
	v1SignaturePrefix := signaturePrefix(model.SignTypeV1, model.SignAlgorithm)
	if strings.HasPrefix(requestSignature, v1SignaturePrefix) {
		return true
	}
	return false
}

// signaturePrefix return supported Authorization prefix
func signaturePrefix(version, algorithm string) string {
	return version + " " + algorithm + ","
}

func (r *RequestContext) VerifySignature() (sdk.AccAddress, error) {
	requestSignature := r.request.Header.Get(model.GnfdAuthorizationHeader)
	v1SignaturePrefix := signaturePrefix(model.SignTypeV1, model.SignAlgorithm)
	if strings.HasPrefix(requestSignature, v1SignaturePrefix) {
		return r.verifySignatureV1(requestSignature[len(v1SignaturePrefix):])
	}
	//personalSignSignaturePrefix := signaturePrefix(model.SignTypePersonal, model.SignAlgorithm)
	//if strings.HasPrefix(requestSignature, personalSignSignaturePrefix) {
	//	return reqContext.verifyPersonalSignature(requestSignature[len(personalSignSignaturePrefix):])
	//}
	//OffChainSignaturePrefix := signaturePrefix(model.SignTypeOffChain, model.SignAlgorithmEddsa)
	//if strings.HasPrefix(requestSignature, OffChainSignaturePrefix) {
	//	return g.verifyOffChainSignature(reqContext, requestSignature[len(OffChainSignaturePrefix):])
	//}
	return nil, ErrUnsupportedSignType
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
		return nil, ErrAuthorizationFormat
	}
	for _, item := range signatureItems {
		pair := strings.Split(item, "=")
		if len(pair) != 2 {
			return nil, ErrAuthorizationFormat
		}
		switch pair[0] {
		case model.SignedMsg:
			signedMsg = pair[1]
		case model.Signature:
			if signature, err = hex.DecodeString(pair[1]); err != nil {
				return nil, err
			}
		default:
			return nil, ErrAuthorizationFormat
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
