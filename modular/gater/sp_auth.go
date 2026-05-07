package gater

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

// SPSignable represents a message with an embedded SP operator signature.
type SPSignable interface {
	GetSignBytes() []byte
	GetSignature() []byte
}

// SPSignableWithExpiry extends SPSignable with an expiry timestamp.
type SPSignableWithExpiry interface {
	SPSignable
	GetExpireTime() int64
}

// VerifySPSignable verifies the embedded signature and returns the SP.
func (g *GateModular) VerifySPSignable(reqCtx *RequestContext, msg SPSignable) (*sptypes.StorageProvider, error) {
	spAddr, err := reqCtx.verifyTaskSignature(msg.GetSignBytes(), msg.GetSignature())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify SP operator signature", "error", err)
		return nil, err
	}
	sp, err := g.spCachePool.QuerySPByAddress(spAddr.String())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to query SP by address", "sp_addr", spAddr.String(), "error", err)
		return nil, err
	}
	return sp, nil
}

// VerifySPSignableWithExpiry verifies the signature and checks that the message has not expired.
func (g *GateModular) VerifySPSignableWithExpiry(reqCtx *RequestContext, msg SPSignableWithExpiry) (*sptypes.StorageProvider, error) {
	if msg.GetExpireTime() < time.Now().Unix() {
		log.CtxErrorw(reqCtx.Context(), "SP auth message has expired",
			"expire_time", msg.GetExpireTime(), "now", time.Now().Unix())
		return nil, ErrNoPermission
	}
	return g.VerifySPSignable(reqCtx, msg)
}

// ParseAndVerifySPAuth parses the X-Gnfd-SP-Operator-Auth header, verifies the
// embedded SP operator signature, checks expiry, and validates payload binding.
func (g *GateModular) ParseAndVerifySPAuth(reqCtx *RequestContext, r *http.Request, payloadForBinding []byte) (*sptypes.StorageProvider, error) {
	authHeader := r.Header.Get(GnfdSPOperatorAuthHeader)
	if authHeader == "" {
		log.CtxErrorw(reqCtx.Context(), "missing SP operator auth header")
		return nil, ErrNoPermission
	}

	authBytes, err := hex.DecodeString(authHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to hex decode SP operator auth header", "error", err)
		return nil, ErrDecodeMsg
	}

	var authMsg gfsptask.GfSpBucketMigrationInfo
	if err = json.Unmarshal(authBytes, &authMsg); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal SP operator auth", "error", err)
		return nil, ErrDecodeMsg
	}

	sp, err := g.VerifySPSignableWithExpiry(reqCtx, &authMsg)
	if err != nil {
		return nil, err
	}

	if len(payloadForBinding) > 0 {
		expectedHash := payloadBindingHash(payloadForBinding)
		if authMsg.GetBucketId() != expectedHash {
			log.CtxErrorw(reqCtx.Context(), "SP auth payload binding mismatch",
				"expected", expectedHash, "got", authMsg.GetBucketId())
			return nil, ErrNoPermission
		}
	}

	return sp, nil
}

// payloadBindingHash computes the first 8 bytes of Keccak256(payload) as a uint64.
func payloadBindingHash(payload []byte) uint64 {
	hash := crypto.Keccak256(payload)
	return binary.BigEndian.Uint64(hash[:8])
}
