package sr25519

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/crypto"
)

func TestSignAndValidateSr25519(t *testing.T) {
	privKey := GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))
	assert.True(t, pubKey.VerifySignature(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifySignature(msg, sig))
}

func TestBatchSafe(t *testing.T) {
	v := NewBatchVerifier()
	vFail := NewBatchVerifier()
	for i := 0; i <= 38; i++ {
		priv := GenPrivKey()
		pub := priv.PubKey()

		var msg []byte
		if i%2 == 0 {
			msg = []byte("easter")
		} else {
			msg = []byte("egg")
		}

		sig, err := priv.Sign(msg)
		require.NoError(t, err)

		err = v.Add(pub, msg, sig)
		require.NoError(t, err)

		switch i % 2 {
		case 0:
			err = vFail.Add(pub, msg, sig)
		case 1:
			msg[2] ^= byte(0x01)
			err = vFail.Add(pub, msg, sig)
		}
		require.NoError(t, err)
	}

	ok, valid := v.Verify()
	require.True(t, ok, "failed batch verification")
	for i, ok := range valid {
		require.Truef(t, ok, "sig[%d] should be marked valid", i)
	}

	ok, valid = vFail.Verify()
	require.False(t, ok, "succeeded batch verification (invalid batch)")
	for i, ok := range valid {
		expected := (i % 2) == 0
		require.Equalf(t, expected, ok, "sig[%d] should be %v", i, expected)
	}
}

func TestJSON(t *testing.T) {
	privKey := GenPrivKey()

	t.Run("PrivKey", func(t *testing.T) {
		b, err := json.Marshal(privKey)
		require.NoError(t, err)

		// b should be the base64 encoded MiniSecretKey, enclosed by doublequotes.
		b64 := base64.StdEncoding.EncodeToString(privKey.Bytes())
		b64 = "\"" + b64 + "\""
		require.Equal(t, []byte(b64), b)

		var privKey2 PrivKey
		err = json.Unmarshal(b, &privKey2)
		require.NoError(t, err)
		require.Len(t, privKey2.Bytes(), PrivKeySize)
		require.EqualValues(t, privKey.Bytes(), privKey2.Bytes())
	})

	// PubKeys are just []byte, so there is no special handling.
}
