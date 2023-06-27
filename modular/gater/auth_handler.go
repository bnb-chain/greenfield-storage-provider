package gater

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	MaxExpiryAgeInSec int32  = 3600 * 24 * 7 // 7 days
	ExpiryDateFormat  string = time.RFC3339
)

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
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
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
	ctx := log.Context(context.Background(), account, domain)
	currentNonce, nextNonce, currentPublicKey, expiryDate, err := g.baseApp.GfSpClient().GetAuthNonce(ctx, account, domain)

	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get auth nonce", "error", err)
		return
	}

	var resp = map[string]interface{}{
		"current_nonce":      currentNonce,
		"next_nonce":         nextNonce,
		"current_public_key": currentPublicKey,
		"expiry_date":        expiryDate,
	}
	b, err = json.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal get auth nonce response")
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
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
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		// verify personal sign signature
		personalSignSignaturePrefix := signaturePrefix(SignTypePersonal, SignAlgorithm)
		requestSignature := reqCtx.request.Header.Get(GnfdAuthorizationHeader)

		if strings.HasPrefix(requestSignature, personalSignSignaturePrefix) {
			accAddress, err := verifyPersonalSignatureFromHeader(requestSignature[len(personalSignSignaturePrefix):])
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", err)
				return
			}
			account = accAddress.String()
			reqCtx.account = account
		} else {
			return
		}
	}

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

	if userPublicKey == "" {
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

	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to GetAuthNonce", "error", err)
		return
	}

	nonceInt, err := strconv.Atoi(nonce)
	if err != nil || int(nextNonce) != nonceInt { // nonce must be the same as NextNonce
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
	log.Infof("%s", time.Until(expiryDate).Seconds())
	log.Infof("%s", MaxExpiryAgeInSec)
	expiryAge := int32(time.Until(expiryDate).Seconds())
	if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
		err = ErrInvalidExpiryDateHeader
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to InvalidExpiryDateHeader")
		return
	}

	requestSignature := reqCtx.request.Header.Get(GnfdAuthorizationHeader)
	personalSignSignaturePrefix := signaturePrefix(SignTypePersonal, SignAlgorithm)
	signedMsg, _, err := parseSignedMsgAndSigFromRequest(requestSignature[len(personalSignSignaturePrefix):])
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey when parseSignedMsgAndSigFromRequest")
		return
	}
	err = g.verifySignedContent(*signedMsg, domain, nonce, userPublicKey, expiryDateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to updateUserPublicKey due to bad signed content.")
		return
	}

	updateUserPublicKeyResp, err := g.baseApp.GfSpClient().UpdateUserPublicKey(ctx, account, domain, currentNonce, int32(nonceInt), userPublicKey, expiryDate.UnixMilli())
	if err != nil {
		log.Errorw("failed to updateUserPublicKey when saving key")
		return
	}

	var resp = map[string]interface{}{
		"result": updateUserPublicKeyResp,
	}
	b, err = json.Marshal(resp)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal update user public key response")
		err = ErrDecodeMsg
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b)
}

func verifyPersonalSignatureFromHeader(requestSignature string) (sdk.AccAddress, error) {
	signedMsg, sigString, err := parseSignedMsgAndSigFromRequest(requestSignature)
	if err != nil {
		return nil, err
	}
	return VerifyPersonalSignature(*signedMsg, *sigString)
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
	addr, _, err := RecoverAddr(realMsgToSign, signature)
	if err != nil {
		log.Errorw("failed to recover address")
		return nil, ErrSignature
	}

	return addr, nil
}

func (g *GateModular) verifySignedContent(signedContent string, expectedDomain string, expectedNonce string, expectedPublicKey string, expectedExpiryDate string) error {
	pattern := `(.+) wants you to sign in with your BNB Greenfield account:\n*(.+)\n*Register your identity public key (.+)\n*URI: (.+)\n*Version: (.+)\n*Chain ID: (.+)\n*Issued At: (.+)\n*Expiration Time: (.+)\n*Resources:((?:\n- SP .+ \(name:.+\) with nonce: \d+)+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(signedContent)
	if len(match) < 4 {
		return ErrSignedMsgNotMatchTemplate
	}
	// Extract variable values
	dappDomain := match[1]
	// userAcct := match[2]  // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	publicKey := match[3]
	// eip4361URI := match[4] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	// eip4361Version := match[5] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	// eip4361ChainId := match[6] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	// eip4361IssuedAt := match[7] // unused, but keep this line here to indicate the matched details so that they could be useful in the future.
	eip4361ExpirationTime := match[8]

	spsText := match[9]
	spsPattern := `- SP (.+) \(name:(.+\S)\) with nonce: (\d+)`
	spsRe := regexp.MustCompile(spsPattern)
	spsMatch := spsRe.FindAllStringSubmatch(spsText, -1)

	var found = false
	for _, match := range spsMatch {
		spAddress := match[1]
		// spName := match[2]  // keep this line here to indicate match[2] means spName
		spNonce := match[3]
		if spAddress == g.baseApp.OperatorAddress() {
			found = true
			if expectedNonce != spNonce { // nonce doesn't match
				return ErrSignedMsgNotMatchTemplate
			}
		}
	}
	if !found { // the signed content is not for this SP  (g.config.SpOperatorAddress)
		return ErrSignedMsgNotMatchSPAddr
	}
	if dappDomain != expectedDomain {
		return ErrSignedMsgNotMatchTemplate
	}
	if publicKey != expectedPublicKey {
		return ErrSignedMsgNotMatchTemplate
	}
	if eip4361ExpirationTime != expectedExpiryDate {
		return ErrSignedMsgNotMatchTemplate
	}

	return nil
}
