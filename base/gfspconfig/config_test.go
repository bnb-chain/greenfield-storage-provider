package gfspconfig

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var mockErr = errors.New("mock error")

func TestGfSpConfig_ApplySuccess(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	opt := func(cfg *GfSpConfig) error { return nil }
	err := cfg.Apply(opt)
	assert.Equal(t, nil, err)
}

func TestGfSpConfig_ApplyFailure(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	opt := func(cfg *GfSpConfig) error { return mockErr }
	err := cfg.Apply(opt)
	assert.Equal(t, mockErr, err)
}

func TestGfSpConfig_StringSuccess(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	result := cfg.String()
	assert.NotNil(t, result)
}

func TestGfSpConfig_StringRedactsSecrets(t *testing.T) {
	cfg := &GfSpConfig{
		Env: "mainnet",
		SpAccount: SpAccountConfig{
			SpOperatorAddress:  "0xABCD",
			OperatorPrivateKey: "secret_operator",
			FundingPrivateKey:  "secret_funding",
			SealPrivateKey:     "secret_seal",
			ApprovalPrivateKey: "secret_approval",
			GcPrivateKey:       "secret_gc",
			BlsPrivateKey:      "secret_bls",
		},
		P2P: P2PConfig{
			P2PPrivateKey: "secret_p2p",
		},
		SpDB: storeconfig.SQLDBConfig{Passwd: "secret_spdb"},
		BsDB: storeconfig.SQLDBConfig{Passwd: "secret_bsdb"},
	}

	result := cfg.String()

	secrets := []string{
		"secret_operator", "secret_funding", "secret_seal",
		"secret_approval", "secret_gc", "secret_bls",
		"secret_p2p", "secret_spdb", "secret_bsdb",
	}
	for _, s := range secrets {
		assert.False(t, strings.Contains(result, s), "output must not contain plaintext secret: %s", s)
	}

	// 9 sensitive fields should each appear as [REDACTED]
	assert.Equal(t, 9, strings.Count(result, redactedPlaceholder),
		"output must contain exactly 9 [REDACTED] placeholders")

	assert.Contains(t, result, "0xABCD", "non-secret fields must still be present")
}

func TestGfSpConfig_StringDoesNotMutateOriginal(t *testing.T) {
	cfg := &GfSpConfig{
		Env: "mainnet",
		SpAccount: SpAccountConfig{
			OperatorPrivateKey: "original_key",
		},
		P2P: P2PConfig{
			P2PPrivateKey: "original_p2p",
		},
	}

	_ = cfg.String()

	assert.Equal(t, "original_key", cfg.SpAccount.OperatorPrivateKey, "String() must not mutate original config")
	assert.Equal(t, "original_p2p", cfg.P2P.P2PPrivateKey, "String() must not mutate original config")
}

func TestGfSpConfig_StringEmptySecretsNotRedacted(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	result := cfg.String()
	assert.False(t, strings.Contains(result, redactedPlaceholder),
		"empty secret fields should not produce [REDACTED]")
}
