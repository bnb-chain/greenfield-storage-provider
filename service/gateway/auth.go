package gateway

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

func (g *Gateway) VerifySignature(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext := newRequestContext(r)
		requestSignature := requestContext.request.Header.Get(model.GnfdAuthorizationHeader)
		signPrefix := signaturePrefix(model.SignTypeV1, model.SignAlgorithm)
		if !strings.HasPrefix(requestSignature, signPrefix) {
			errorDescription := SignatureDoesNotMatch
			log.Infow("signature type mismatch")
			_ = errorDescription.errorResponse(w, requestContext)
			return
		}
		addr, err := requestContext.verifySignatureV3(requestSignature)
		if err != nil {
			errorDescription := SignatureDoesNotMatch
			log.Infow("failed to verify signature", "error", err)
			_ = errorDescription.errorResponse(w, requestContext)
			return
		}
		exist, err := g.chain.HasAccount(context.Background(), hex.EncodeToString(addr))
		if err != nil {
			errorDescription := SignatureDoesNotMatch
			log.Infow("check account on chain failed", "error", err)
			_ = errorDescription.errorResponse(w, requestContext)
			return
		}
		if !exist {
			errorDescription := SignatureDoesNotMatch
			log.Infow("check account on chain not exist")
			_ = errorDescription.errorResponse(w, requestContext)
			return
		}
		r.Header.Set(model.GnfdAddress, hex.EncodeToString(addr))
		f.ServeHTTP(w, r)
	}
}
