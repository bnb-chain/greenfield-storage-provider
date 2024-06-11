package gater

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	commonhash "github.com/bnb-chain/greenfield-common/go/hash"
	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	modelgateway "github.com/zkMeLabs/mechain-storage-provider/model/gateway"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
	"github.com/zkMeLabs/mechain-storage-provider/util"
)

const (
	MaxExpiryAgeInSec         int32  = commonhttp.MaxExpiryAgeInSec // 7 days
	ExpiryDateFormat          string = time.RFC3339
	ExpectedEddsaPubKeyLength int    = 64
	SignedContentV2Pattern           = `(.+) wants you to sign in with your BNB Greenfield account:\n*(.+)\n*Register your identity public key (.+)\n*URI: (.+)\n*Version: (.+)\n*Chain ID: (.+)\n*Issued At: (.+)\n*Expiration Time: (.+)`
)

type RequestNonceResp struct {
	CurrentNonce     int32  `xml:"CurrentNonce"`
	NextNonce        int32  `xml:"NextNonce"`
	CurrentPublicKey string `xml:"CurrentPublicKey"`
	ExpiryDate       int64  `xml:"ExpiryDate"`
}

type UpdateUserPublicKeyResp struct {
	Result bool `xml:"Result"`
}

type DeleteUserPublicKeyV2Resp struct {
	Result bool `xml:"Result"`
}

type ListUserPublicKeyV2Resp struct {
	PublicKeys []string `xml:"Result"`
}

// requestNonceHandler handle requestNonce request
func (g *GateModular) requestNonceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      []byte
		reqCtx *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to request nonce", "req_info", reqCtx.String())
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	// ignore the error, because the requestNonce does not need signature
	reqCtx, _ = NewRequestContext(r, g)

	account := reqCtx.request.Header.Get(GnfdUserAddressHeader)
	domain := reqCtx.request.Header.Get(GnfdOffChainAuthAppDomainHeader)

	// validate account header
	if ok := common.IsHexAddress(account); !ok {
		err = ErrInvalidHeader
		log.Errorw("failed to check account address", "account_address", account, "error", err)
		return
	}

	// validate domain header
	if domain == "" {
		log.CtxErrorw(reqCtx.Context(), "failed to check GnfdOffChainAuthAppDomainHeader header")
		err = ErrInvalidDomainHeader
		return
	}

	ctx := log.Context(context.Background(), account, domain)
	currentNonce, nextNonce, currentPublicKey, expiryDate, err := g.baseApp.GfSpClient().GetAuthNonce(ctx, account, domain)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get auth nonce", "error", err)
		return
	}

	resp := &RequestNonceResp{
		CurrentNonce:     currentNonce,
		NextNonce:        nextNonce,
		CurrentPublicKey: currentPublicKey,
		ExpiryDate:       expiryDate,
	}
	b, err = xml.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal get auth nonce response", "error", err)
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	w.Write(b)
}

func (g *GateModular) updateUserPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		b             []byte
		reqCtx        *RequestContext
		account       string
		userPublicKey string
		domain        string
		origin        string
		nonce         string
		expiryDateStr string
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey", "req_info", reqCtx.String())
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)
	// verify personal sign signature
	personalSignSignaturePrefix := commonhttp.Gnfd1EthPersonalSign + ","
	requestSignature := reqCtx.request.Header.Get(GnfdAuthorizationHeader)

	if !strings.HasPrefix(requestSignature, personalSignSignaturePrefix) {
		err = ErrUnsupportedSignType
		return
	}
	signedMsg, sigString, err := parseSignedMsgAndSigFromRequest(strings.TrimPrefix(requestSignature, personalSignSignaturePrefix))
	if err != nil {
		return
	}
	accAddress, personalSignVerifyErr := VerifyPersonalSignature(*signedMsg, *sigString)

	if personalSignVerifyErr != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", personalSignVerifyErr)
		err = personalSignVerifyErr
		return
	}
	account = accAddress.String()
	reqCtx.account = account

	domain = reqCtx.request.Header.Get(GnfdOffChainAuthAppDomainHeader)
	origin = reqCtx.request.Header.Get("Origin")
	nonce = reqCtx.request.Header.Get(GnfdOffChainAuthAppRegNonceHeader)
	userPublicKey = reqCtx.request.Header.Get(GnfdOffChainAuthAppRegPublicKeyHeader)
	expiryDateStr = reqCtx.request.Header.Get(GnfdOffChainAuthAppRegExpiryDateHeader)

	// validate headers
	if domain == "" || domain != origin {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to bad origin or domain")
		err = ErrInvalidDomainHeader
		return
	}

	if userPublicKey == "" || len(userPublicKey) != ExpectedEddsaPubKeyLength {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to bad userPublicKey")
		err = ErrInvalidPublicKeyHeader
		return
	}

	ctx := log.Context(context.Background(), account, domain)
	currentNonce, nextNonce, _, _, err := g.baseApp.GfSpClient().GetAuthNonce(ctx, account, domain)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to GetAuthNonce", "error", err)
		return
	}

	nonceInt, err := util.StringToInt32(nonce)
	if err != nil || nextNonce != nonceInt { // nonce must be the same as NextNonce
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to bad nonce")
		err = ErrInvalidRegNonceHeader
		return
	}

	expiryDate, err := time.Parse(ExpiryDateFormat, expiryDateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to InvalidExpiryDateHeader")
		err = ErrInvalidExpiryDateHeader
		return
	}
	expiryAge := int32(time.Until(expiryDate).Seconds())
	if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
		err = ErrInvalidExpiryDateHeader
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to InvalidExpiryDateHeader")
		return
	}

	err = g.verifySignedContent(*signedMsg, domain, nonce, userPublicKey, expiryDateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to bad signed content.")
		return
	}

	updateUserPublicKeyResp, err := g.baseApp.GfSpClient().UpdateUserPublicKey(ctx, account, domain, currentNonce, nonceInt, userPublicKey, expiryDate.UnixMilli())
	if err != nil {
		log.Errorw("failed to updateUserPublicKey when saving key")
		return
	}
	resp := &UpdateUserPublicKeyResp{
		Result: updateUserPublicKeyResp,
	}
	b, err = xml.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal update user public key response", "error", err)
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	w.Write(b)
}

// updateUserPublicKeyV2Handler register users' EdDSA public key into off_chain_auth_key_v2 table
func (g *GateModular) updateUserPublicKeyV2Handler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		b             []byte
		reqCtx        *RequestContext
		account       string
		userPublicKey string
		domain        string
		origin        string
		expiryDateStr string
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKeyV2", "req_info", reqCtx.String())
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)
	// verify personal sign signature
	personalSignSignaturePrefix := commonhttp.Gnfd1EthPersonalSign + ","
	requestSignature := reqCtx.request.Header.Get(GnfdAuthorizationHeader)

	if !strings.HasPrefix(requestSignature, personalSignSignaturePrefix) {
		err = ErrUnsupportedSignType
		return
	}
	signedMsg, sigString, err := parseSignedMsgAndSigFromRequest(strings.TrimPrefix(requestSignature, personalSignSignaturePrefix))
	if err != nil {
		return
	}
	accAddress, personalSignVerifyErr := VerifyPersonalSignature(*signedMsg, *sigString)

	if personalSignVerifyErr != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", personalSignVerifyErr)
		err = personalSignVerifyErr
		return
	}
	account = accAddress.String()
	reqCtx.account = account

	domain = reqCtx.request.Header.Get(GnfdOffChainAuthAppDomainHeader)
	origin = reqCtx.request.Header.Get("Origin")
	userPublicKey = reqCtx.request.Header.Get(GnfdOffChainAuthAppRegPublicKeyHeader)
	expiryDateStr = reqCtx.request.Header.Get(GnfdOffChainAuthAppRegExpiryDateHeader)

	// validate headers
	if domain == "" || domain != origin {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKeyV2 due to bad origin or domain")
		err = ErrInvalidDomainHeader
		return
	}

	if userPublicKey == "" || len(userPublicKey) != ExpectedEddsaPubKeyLength {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKeyV2 due to bad userPublicKey")
		err = ErrInvalidPublicKeyHeader
		return
	}

	expiryDate, err := time.Parse(ExpiryDateFormat, expiryDateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKeyV2 due to InvalidExpiryDateHeader")
		err = ErrInvalidExpiryDateHeader
		return
	}
	expiryAge := int32(time.Until(expiryDate).Seconds())
	if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
		err = ErrInvalidExpiryDateHeader
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKeyV2 due to InvalidExpiryDateHeader")
		return
	}

	err = g.verifySignedContentV2(*signedMsg, domain, userPublicKey, expiryDateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKeyV2 due to bad signed content.")
		return
	}

	updateUserPublicKeyResp, err := g.baseApp.GfSpClient().UpdateUserPublicKeyV2(reqCtx.ctx, account, domain, userPublicKey, expiryDate.UnixMilli())
	if err != nil {
		log.Errorw("failed to updateUserPublicKeyV2 when saving key")
		return
	}
	resp := &UpdateUserPublicKeyResp{
		Result: updateUserPublicKeyResp,
	}
	b, err = xml.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal update user public key response", "error", err)
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	w.Write(b)
}

// listUserPublicKeyV2Handler list user public keys from off_chain_auth_key_v2 table
func (g *GateModular) listUserPublicKeyV2Handler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      []byte
		reqCtx *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to listUserPublicKeyV2", "req_info", reqCtx.String())
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	// ignore the error, because the listUserPublicKeyV2 does not need signature
	reqCtx, _ = NewRequestContext(r, g)

	account := reqCtx.request.Header.Get(GnfdUserAddressHeader)
	domain := reqCtx.request.Header.Get(GnfdOffChainAuthAppDomainHeader)

	// validate account header
	if ok := common.IsHexAddress(account); !ok {
		err = ErrInvalidHeader
		log.Errorw("failed to check account address", "account_address", account, "error", err)
		return
	}

	// validate domain header
	if domain == "" {
		log.CtxErrorw(reqCtx.Context(), "failed to check GnfdOffChainAuthAppDomainHeader header")
		err = ErrInvalidDomainHeader
		return
	}

	userPublicKeys, err := g.baseApp.GfSpClient().ListAuthKeysV2(reqCtx.ctx, account, domain)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to listUserPublicKeyV2", "error", err)
		return
	}

	resp := &ListUserPublicKeyV2Resp{
		PublicKeys: userPublicKeys,
	}
	b, err = xml.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal listUserPublicKeyV2 response", "error", err)
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	w.Write(b)
}

func (g *GateModular) deleteUserPublicKeyV2Handler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      []byte
		reqCtx *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to deleteUserPublicKeyV2", "req_info", reqCtx.String())
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	// deleteUserPublicKeyV2Handler requires the authentication
	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}

	account := reqCtx.request.Header.Get(GnfdUserAddressHeader)
	domain := reqCtx.request.Header.Get(GnfdOffChainAuthAppDomainHeader)

	// validate account header
	if ok := common.IsHexAddress(account); !ok {
		err = ErrInvalidHeader
		log.Errorw("failed to check account address", "account_address", account, "error", err)
		return
	}
	if account != reqCtx.account {
		err = ErrNoPermission
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to read publicKeys from request body", "error", err)
		err = ErrExceptionStream
		return
	}

	publicKeys := strings.Split(string(data), ",")

	// validate domain header
	if domain == "" {
		log.CtxErrorw(reqCtx.Context(), "failed to check GnfdOffChainAuthAppDomainHeader header")
		err = ErrInvalidDomainHeader
		return
	}

	result, err := g.baseApp.GfSpClient().DeleteAuthKeysV2(reqCtx.ctx, account, domain, publicKeys)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to deleteUserPublicKeyV2", "error", err)
		return
	}

	resp := &DeleteUserPublicKeyV2Resp{
		Result: result,
	}
	b, err = xml.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal DeleteUserPublicKeyV2Resp", "error", err)
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	w.Write(b)
}

// parseSignedMsgAndSigFromRequest get sig for personal auth, it expects the auth string should look like "Signature=xxxxx,SignedMsg=xxx".
func parseSignedMsgAndSigFromRequest(requestSignature string) (*string, *string, error) {
	var (
		signedMsg string
		signature string
	)
	requestSignature = strings.ReplaceAll(requestSignature, "\\n", "\n")
	signatureItems := strings.Split(requestSignature, ",")
	if len(signatureItems) != 2 { // requestSignature should be "Signature=xxxxx,SignedMsg=xxxxx"
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

func VerifyPersonalSignature(signedMsg string, sigString string) (sdk.AccAddress, error) {
	var (
		signature []byte
		err       error
	)
	if err != nil {
		return nil, err
	}
	signature, err = hexutil.Decode(sigString)
	if err != nil {
		return nil, err
	}

	realMsgToSign := accounts.TextHash([]byte(signedMsg))

	if len(signature) != crypto.SignatureLength {
		log.Errorw("signature length (actual: %d) doesn't match typical [R||S||V] signature 65 bytes")
		return nil, ErrSignature
	}
	if signature[crypto.RecoveryIDOffset] == 27 || signature[crypto.RecoveryIDOffset] == 28 {
		signature[crypto.RecoveryIDOffset] -= 27
	}

	// check signature consistent
	addr, _, err := commonhash.RecoverAddr(realMsgToSign, signature)
	if err != nil {
		log.Errorw("failed to recover address")
		return nil, ErrSignature
	}

	return addr, nil
}

func (g *GateModular) verifySignedContent(signedContent string, expectedDomain string, expectedNonce string, expectedPublicKey string, expectedExpiryDate string) error {
	pattern := `(.+) wants you to sign in with your BNB Greenfield account:\n*(.+)\n*Register your identity public key (.+)\n*URI: (.+)\n*Version: (.+)\n*Chain ID: (.+)\n*Issued At: (.+)\n*Expiration Time: (.+)\n*Resources:((?:\n- SP .+ \(name:.+\) with nonce: \d+)+)`

	re := regexp.MustCompile(pattern)
	patternMatches := re.FindStringSubmatch(signedContent)
	if len(patternMatches) < 10 {
		return ErrSignedMsgNotMatchTemplate
	}
	// Extract variable values
	dappDomain := patternMatches[1]
	// userAcct := patternMatches[2]  // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	publicKey := patternMatches[3]
	// eip4361URI := patternMatches[4] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	// eip4361Version := patternMatches[5] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	// eip4361ChainId := patternMatches[6] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	// eip4361IssuedAt := patternMatches[7] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	eip4361ExpirationTime := patternMatches[8]

	spsText := patternMatches[9]
	spsPattern := `- SP (.+) \(name:(.+\S)\) with nonce: (\d+)`
	spsRe := regexp.MustCompile(spsPattern)
	spsMatch := spsRe.FindAllStringSubmatch(spsText, -1)

	found := false
	for _, spInfoMatches := range spsMatch {
		if len(spInfoMatches) < 4 {
			return ErrSignedMsgNotMatchTemplate
		}
		spAddress := spInfoMatches[1]
		// spName := spInfoMatches[2]  // keep this line here to indicate spInfoMatches[2] means spName
		spNonce := spInfoMatches[3]
		if spAddress == g.baseApp.OperatorAddress() {
			found = true
			if expectedNonce != spNonce { // nonce doesn't match
				return ErrSignedMsgNotMatchSPNonce
			}
		}
	}
	if !found { // the signed content is not for this SP  (g.config.SpOperatorAddress)
		return ErrSignedMsgNotMatchSPAddr
	}
	if dappDomain != expectedDomain {
		return ErrSignedMsgNotMatchDomain
	}
	if publicKey != expectedPublicKey {
		return ErrSignedMsgNotMatchPubKey
	}
	if eip4361ExpirationTime != expectedExpiryDate {
		return ErrSignedMsgNotMatchExpiry
	}

	return nil
}

func (g *GateModular) verifySignedContentV2(signedContent string, expectedDomain string, expectedPublicKey string, expectedExpiryDate string) error {
	re := regexp.MustCompile(SignedContentV2Pattern)
	patternMatches := re.FindStringSubmatch(signedContent)
	if len(patternMatches) < 9 {
		return ErrSignedMsgNotMatchTemplate
	}
	// Extract variable values
	dappDomain := patternMatches[1]
	publicKey := patternMatches[3]
	eip4361ExpirationTime := patternMatches[8]

	if dappDomain != expectedDomain {
		return ErrSignedMsgNotMatchDomain
	}
	if publicKey != expectedPublicKey {
		return ErrSignedMsgNotMatchPubKey
	}
	if eip4361ExpirationTime != expectedExpiryDate {
		return ErrSignedMsgNotMatchExpiry
	}

	return nil
}
