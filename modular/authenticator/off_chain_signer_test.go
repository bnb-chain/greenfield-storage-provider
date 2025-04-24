package authenticator

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func TestEd25519PrivateKeyAndVerify(t *testing.T) {
	userEddsaPublicKeyStr := "1db9bfd4f2a457f32a1c500d6d6848b843602942f1b0577d968ea378cbfacf97"
	msg := "I want to sign a message"
	sigStr := "7a72fef8af89fbdd45879b842b70c868c489484daf564a53b027bb3cd68374007a7b43a49645c82ad0dba63fb9477bc8d15988d0c0ffe1b54fb1bf8fa9a02905"

	sig, _ := hex.DecodeString(sigStr)

	pubKeyBytes, err := hex.DecodeString(userEddsaPublicKeyStr)
	if err != nil {
		t.Errorf("Decode public key failed: %v", err)
	}
	if !ed25519.Verify(pubKeyBytes, []byte(msg), sig) {
		t.Errorf("Verify failed: signature is not valid")
	}
}

func TestGenerateEddsaPrivateKey(t *testing.T) {
	sk, err := GenerateEddsaPrivateKey("testeeetgcxsaahsadcastzxbmjhgmgjhcarwewfseasdasdavacsafaewe")
	if err != nil {
		t.Fatal(err)
	}
	log.Info(new(big.Int).SetBytes(sk.Bytes()[32:64]).BitLen())
	hFunc := mimc.NewMiMC()
	msg := "use seed to sign this message"
	var frMsg fr.Element
	frMsg.SetBytes([]byte(msg))
	msgBin := frMsg.Bytes()
	signMsg, err := sk.Sign(msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}
	hFunc.Reset()
	isValid, err := sk.PublicKey.Verify(signMsg, msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}
	log.Info(isValid)
	assert.Equal(t, true, isValid)
}

func TestEddsaMIMC(t *testing.T) {

	privKey, err := GenerateEddsaPrivateKey("testeeetgcxsaahsadcastzxbmjhgmgjhcarwewfseasdasdavacsafaeweasdfasdfasdfasdfasdfasdf")
	if err != nil {
		t.Fatal(err)
	}
	pubKey := privKey.PublicKey
	hFunc := hash.MIMC_BN254.New()

	var msg = "I want to get approval from sp before creating the bucket"
	var frMsg fr.Element
	frMsg.SetBytes([]byte(msg))
	msgBin := frMsg.Bytes()
	signature, err := privKey.Sign(msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}

	// verifies correct msg
	res, err := pubKey.Verify(signature, msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}
	if !res {
		t.Fatal("Verify correct signature should return true")
	}

	// verifies wrong msg
	frMsg.SetBytes([]byte("I don't want to get approval from sp before creating the bucket"))
	msgBin = frMsg.Bytes()
	res, err = pubKey.Verify(signature, msgBin[:], hash.MIMC_BN254.New())
	if err != nil {
		t.Fatal(err)
	}
	if res {
		t.Fatal("Verify wrong signature should be false")
	}

}

// TestUserOffChainAuthSignature
func TestUserOffChainAuthSignature(t *testing.T) {
	// 1. use EDCSA (ETH personal sign) to generate a seed, which is regarded as users EDDSA private key.
	// Account information.
	privateKey, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Infof("address is: " + address.Hex())
	unSignedContent := "Sign this message to access SP 0x12345"       //
	unSignedContentHash := accounts.TextHash([]byte(unSignedContent)) // personal sign
	// Sign data.
	edcsaSig, _ := crypto.Sign(unSignedContentHash, privateKey)

	// 2. get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(string(edcsaSig))
	// This is the user eddsa public key string , stored in sp, via API "/auth/update_key"
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(string(edcsaSig))
	log.Infof("userEddsaPublicKeyStr is %s", userEddsaPublicKeyStr)
	// userEddsaPublickKey, _ := ParsePk(userEddsaPublicKeyStr)

	// 3. use EDDSA private key to sign, as the off chain auth sig.
	hFunc := hash.MIMC_BN254.New()
	msg := "I want to get approval from sp before creating the bucket."
	var frMsg fr.Element
	frMsg.SetBytes([]byte(msg))
	msgBin := frMsg.Bytes()
	sig, err := userEddsaPrivateKey.Sign(msgBin[:], hFunc)
	require.NoError(t, err)
	// 4. use public key to verify
	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte(msg))

	require.NoError(t, err)

	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte("This msg doesn't match with the sig"))
	assert.Equal(t, ErrBadSignature, err)
}

// TestUseUserPublicKeyToVerifyUserOffChainAuthSignature
func TestUseUserPublicKeyToVerifyUserOffChainAuthSignature(t *testing.T) {
	// 1. use EDCSA (ETH personal sign) to generate a seed, which is regarded as users EDDSA private key.
	// Account information.
	privateKey, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Infof("address is: " + address.Hex())
	unSignedContent := "Sign this message to access SP 0x12345"       //
	unSignedContentHash := accounts.TextHash([]byte(unSignedContent)) // personal sign
	// Sign data.
	edcsaSig, _ := crypto.Sign(unSignedContentHash, privateKey)

	// 2. get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(string(edcsaSig))
	// This is the user eddsa public key string , stored in sp, via API "/auth/update_key"
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(string(edcsaSig))
	log.Infof("userEddsaPublicKeyStr is %s", userEddsaPublicKeyStr)
	// userEddsaPublickKey, _ := ParsePk(userEddsaPublicKeyStr)

	// 3. use EDDSA private key to sign, as the off chain auth sig.
	hFunc := hash.MIMC_BN254.New()
	msg := "I want to get approval from sp before creating the bucket."
	var frMsg fr.Element
	frMsg.SetBytes([]byte(msg))
	msgBin := frMsg.Bytes()
	sig, err := userEddsaPrivateKey.Sign(msgBin[:], hFunc)
	require.NoError(t, err)
	// 4. use public key to verify
	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte(msg))

	require.NoError(t, err)

	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte("This msg doesn't match with the sig"))
	assert.Equal(t, ErrBadSignature, err)
}

func TestErrorCases(t *testing.T) {
	err := VerifyEddsaSignature("inValidPk", nil, []byte(""))
	if err != nil {
		log.Errorf("%s", err)
	}
	err = VerifyEddsaSignature("fe1d334ee593176e6b4acb2e5abd943e607c", nil, []byte("")) // len(publicKey) is too short
	if err != nil {
		log.Errorf("%s", err)
	}

}

func TestParsePK(t *testing.T) {
	sk, err := GenerateEddsaPrivateKey("testeeetgcxsaahsadcastzxbmjhgmgj")
	if err != nil {
		t.Fatal(err)
	}
	correctPK, _ := ParsePk(GetEddsaCompressedPublicKey("testeeetgcxsaahsadcastzxbmjhgmgj"))
	assert.Equal(t, sk.PublicKey.Bytes(), correctPK.Bytes())

	wrongPK, _ := ParsePk(GetEddsaCompressedPublicKey("wrongSeed"))

	hFunc := mimc.NewMiMC()
	msg := "use seed to sign this message"
	var frMsg fr.Element
	frMsg.SetString(msg)
	msgBin := frMsg.Bytes()

	sig, err := sk.Sign(msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}
	hFunc.Reset()
	isValid, err := correctPK.Verify(sig, msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}
	log.Info(isValid)
	assert.Equal(t, true, isValid)

	hFunc.Reset()
	isValid, err = wrongPK.Verify(sig, msgBin[:], hFunc)
	if err != nil {
		t.Fatal(err)
	}
	log.Info(isValid)
	assert.Equal(t, false, isValid)
}

func GetEddsaPublicKey(seed string) string {
	sk, err := GenerateEddsaPrivateKey(seed)
	if err != nil {
		return err.Error()
	}
	var buf bytes.Buffer
	buf.Write(sk.PublicKey.A.X.Marshal())
	buf.Write(sk.PublicKey.A.Y.Marshal())
	return hex.EncodeToString(buf.Bytes())
}

func GetEddsaCompressedPublicKey(seed string) string {

	sk, err := GenerateEddsaPrivateKey(seed)
	if err != nil {
		return err.Error()
	}
	var buf bytes.Buffer
	buf.Write(sk.PublicKey.Bytes())
	return hex.EncodeToString(buf.Bytes())
}

type (
	PrivateKey = eddsa.PrivateKey
)

// GenerateEddsaPrivateKey: generate eddsa private key
func GenerateEddsaPrivateKey(seed string) (sk *PrivateKey, err error) {
	//buf := make([]byte, 32)
	//copy(buf, seed)
	reader := bytes.NewReader([]byte(seed))
	sk, err = GenerateKey(reader)
	return sk, err
}

func GenerateKey(r io.Reader) (*PrivateKey, error) {
	return eddsa.GenerateKey(r)
}
