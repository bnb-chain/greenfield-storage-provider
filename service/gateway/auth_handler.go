package gateway

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/jsonpb"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
)

const (
	MaxExpiryAgeInSec int32  = 3600 * 24 * 7 // 7 days
	ExpiryDateFormat  string = "2006-01-02 15:04:05 GMT-07:00"
)

// requestNonceHandler handle requestNonce request
func (g *Gateway) requestNonceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		errDescription *errorDescription
		reqContext     *requestContext
		statusCode     = http.StatusOK
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorResponse(w, reqContext)
		}
		if statusCode == http.StatusOK || statusCode == http.StatusPartialContent {
			log.Infof("action(%v) statusCode(%v) %v", requestNonceName, statusCode, reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", requestNonceName, statusCode, reqContext.generateRequestDetail())
		}
	}()

	if g.auth == nil {
		log.Errorw("failed to request nonce due to not config auth client")
		errDescription = NotExistComponentError
		return
	}

	req := &authtypes.GetAuthNonceRequest{
		AccountId: reqContext.request.Header.Get(model.GnfdUserAddressHeader),
		Domain:    reqContext.request.Header.Get(model.GnfdOffChainAuthAppDomainHeader),
	}
	ctx := log.Context(context.Background(), req)
	resp, err := g.auth.GetAuthNonce(ctx, req)

	if err != nil {
		log.Errorw("failed to GetAuthNonce", "error", err)
		errDescription = InternalError
		return
	}
	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorw("failed to GetAuthNonce", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

func (g *Gateway) updateUserPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		errDescription *errorDescription
		reqContext     *requestContext
		userAddress    sdk.AccAddress
		userPublicKey  string
		domain         string
		origin         string
		nonce          string
		expiryDateStr  string
		statusCode     = http.StatusOK
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorResponse(w, reqContext)
		}
		if statusCode == http.StatusOK || statusCode == http.StatusPartialContent {
			log.Infof("action(%v) statusCode(%v) %v", updateUserPublicKey, statusCode, reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", updateUserPublicKey, statusCode, reqContext.generateRequestDetail())
		}
	}()

	if g.auth == nil {
		log.Errorw("failed to updateUserPublicKeyHandler due to not config auth client")
		errDescription = NotExistComponentError
		return
	}

	if userAddress, err = g.verifySignature(reqContext); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = SignatureNotMatch
		return
	}

	domain = reqContext.request.Header.Get(model.GnfdOffChainAuthAppDomainHeader)
	origin = reqContext.request.Header.Get("Origin")
	nonce = reqContext.request.Header.Get(model.GnfdOffChainAuthAppRegNonceHeader)
	userPublicKey = reqContext.request.Header.Get(model.GnfdOffChainAuthAppRegPublicKeyHeader)
	expiryDateStr = reqContext.request.Header.Get(model.GnfdOffChainAuthAppRegExpiryDateHeader)

	// validate headers
	if domain == "" || domain != origin {
		log.Errorw("failed to updateUserPublicKey due to bad origin or domain")
		errDescription = InvalidHeader
		return
	}

	if userPublicKey == "" {
		log.Errorw("failed to updateUserPublicKey due to bad userPublicKey")
		errDescription = InvalidHeader
		return
	}

	req := &authtypes.GetAuthNonceRequest{
		AccountId: userAddress.String(),
		Domain:    domain,
	}
	ctx := log.Context(context.Background(), req)
	getAuthNonceResp, err := g.auth.GetAuthNonce(ctx, req)

	if err != nil {
		log.Errorw("failed to GetAuthNonce", "error", err)
		errDescription = InternalError
		return
	}

	nonceInt, err := strconv.Atoi(nonce)
	if err != nil || int(getAuthNonceResp.NextNonce) != nonceInt { // nonce must be the same as NextNonce
		log.Errorw("failed to updateUserPublicKey due to bad nonce")
		errDescription = InvalidRegNonceHeader
		return
	}

	expiryDate, err := time.Parse(ExpiryDateFormat, expiryDateStr)
	if err != nil {
		log.Errorw("failed to updateUserPublicKey due to InvalidExpiryDateHeader")
		errDescription = InvalidExpiryDateHeader
		return
	}
	log.Infof("%s", time.Until(expiryDate).Seconds())
	log.Infof("%s", MaxExpiryAgeInSec)
	expiryAge := int32(time.Until(expiryDate).Seconds())
	if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
		errDescription = InvalidExpiryDateHeader
		log.Errorw("failed to updateUserPublicKey due to InvalidExpiryDateHeader")
		return
	}

	requestSignature := reqContext.request.Header.Get(model.GnfdAuthorizationHeader)
	personalSignSignaturePrefix := signaturePrefix(model.SignTypePersonal, model.SignAlgorithm)
	signedMsg, _, err := parseSignedMsgAndSigFromRequest(requestSignature[len(personalSignSignaturePrefix):])

	errDescription = g.verifySignedContent(*signedMsg, domain, nonce, userPublicKey, expiryDateStr)
	if errDescription != nil {
		log.Errorw("failed to updateUserPublicKey due to bad signed content.")
		return
	}

	updateUserPublicKeyReq := &authtypes.UpdateUserPublicKeyRequest{
		AccountId:     userAddress.String(),
		Domain:        domain,
		CurrentNonce:  getAuthNonceResp.CurrentNonce,
		Nonce:         int32(nonceInt),
		UserPublicKey: userPublicKey,
		ExpiryDate:    expiryDate.UnixMilli(),
	}
	updateUserPublicKeyResp, err := g.auth.UpdateUserPublicKey(ctx, updateUserPublicKeyReq)
	if err != nil {
		log.Errorw("failed to updateUserPublicKey when saving key")
		errDescription = InternalError
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, updateUserPublicKeyResp); err != nil {
		log.Errorw("failed to UpdateUserPublicKey", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

func (g *Gateway) verifySignedContent(signedContent string, expectedDomain string, expectedNonce string, expectedPublicKey string, expectedExpiryDate string) *errorDescription {
	pattern := `Register your identity of dapp (.+)\n*with your identity key (.+)\nIn the following SPs:((?:\n- SP .+ \(name:.+\) with nonce: \d+)+)\n\nThe expiry date is (.+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(signedContent)
	if len(match) < 4 {
		return SignedMsgNotMatchTemplate
	}
	// Extract variable values
	dappDomain := match[1]
	publicKey := match[2]
	spsText := match[3]
	spsPattern := `- SP (.+) \(name:(.+\S)\) with nonce: (\d+)`
	spsRe := regexp.MustCompile(spsPattern)
	spsMatch := spsRe.FindAllStringSubmatch(spsText, -1)

	var spsAddresses, spsNames, spsNonces []string
	var found = false
	for _, match := range spsMatch {
		spsAddresses = append(spsAddresses, match[1])

		spsNames = append(spsNames, match[2])
		spsNonces = append(spsNonces, match[3])

		if match[1] == g.config.SpOperatorAddress {
			found = true
			if expectedNonce != match[3] { // nonce doesn't match
				return SignedMsgNotMatchHeaders
			}
		}
	}

	expiryDate := match[4]
	if !found { // the signed content is not for this SP  (g.config.SpOperatorAddress)
		return SignedMsgNotMatchSPAddr
	}
	if dappDomain != expectedDomain {
		return SignedMsgNotMatchHeaders
	}
	if publicKey != expectedPublicKey {
		return SignedMsgNotMatchHeaders
	}
	if expiryDate != expectedExpiryDate {
		return SignedMsgNotMatchHeaders
	}

	return nil
}
