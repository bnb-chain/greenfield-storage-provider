package auth

/*
import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"math/big"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

func TestGenerateEddsaPrivateKey(t *testing.T) {
	sk, err := GenerateEddsaPrivateKey("testeeetgcxsaahsadcastzxbmjhgmgjhcarwewfseasdasdavacsafaewe")
	if err != nil {
		t.Fatal(err)
	}
	log.Info(new(big.Int).SetBytes(sk.Bytes()[32:64]).BitLen())
	hFunc := mimc.NewMiMC()
	msg := "use seed to sign this message"
	signMsg, err := sk.Sign([]byte(msg), hFunc)
	if err != nil {
		t.Fatal(err)
	}
	hFunc.Reset()
	isValid, err := sk.PublicKey.Verify(signMsg, []byte(msg), hFunc)
	if err != nil {
		t.Fatal(err)
	}
	log.Info(isValid)
	assert.Equal(t, true, isValid)
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
	hFunc := mimc.NewMiMC()
	msg := "I want to get approval from sp before creating the bucket."
	sig, err := userEddsaPrivateKey.Sign([]byte(msg), hFunc)
	require.NoError(t, err)
	// 4. use public key to verify
	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte(msg))

	require.NoError(t, err)

	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte("This msg doesn't match with the sig"))
	assert.Equal(t, "invalid signature", err.Error())
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
	hFunc := mimc.NewMiMC()
	msg := "I want to get approval from sp before creating the bucket."
	sig, err := userEddsaPrivateKey.Sign([]byte(msg), hFunc)
	require.NoError(t, err)
	// 4. use public key to verify
	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte(msg))

	require.NoError(t, err)

	err = VerifyEddsaSignature(userEddsaPublicKeyStr, sig, []byte("This msg doesn't match with the sig"))
	assert.Equal(t, "invalid signature", err.Error())
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

// todo -- remove it before commit
func TestForDebug(t *testing.T) {
	seed := []byte{15, 128, 7, 103, 18, 78, 41, 63, 47, 120, 41, 247, 209, 152, 27, 43, 225, 223, 242, 203, 250, 160, 164, 252, 210, 82, 236, 192, 196, 78, 97, 41, 78, 38, 181, 77, 175, 253, 155, 97, 176, 146, 76, 158, 41, 102, 57, 177, 37, 177, 201, 227, 174, 159, 83, 80, 175, 202, 121, 136, 211, 114, 126, 199, 28}
	sk, _ := GenerateEddsaPrivateKey(string(seed))
	correctPK, _ := ParsePk(GetEddsaCompressedPublicKey(string(seed)))

	hFunc := mimc.NewMiMC()
	msg := "I want to get approval_1681436379000"
	sig, _ := sk.Sign([]byte(msg), hFunc)
	log.Infof("sig is %s", hex.EncodeToString(sig))
	isValid, err := correctPK.Verify(sig, []byte(msg), hFunc)
	if err != nil {
		t.Fatal(err)
	}
	log.Info(isValid)
	assert.Equal(t, true, isValid)
}

func TestParsePK(t *testing.T) {
	sk, err := GenerateEddsaPrivateKey("testeeetgcxsaahsadcastzxbmjhgmgj")
	if err != nil {
		t.Fatal(err)
	}
	correctPK, _ := ParsePk(GetEddsaCompressedPublicKey("testeeetgcxsaahsadcastzxbmjhgmgj"))
	wrongPK, _ := ParsePk(GetEddsaCompressedPublicKey("wrongSeed"))

	hFunc := mimc.NewMiMC()
	msg := "use seed to sign this message"
	sig, err := sk.Sign([]byte(msg), hFunc)
	if err != nil {
		t.Fatal(err)
	}
	hFunc.Reset()
	isValid, err := correctPK.Verify(sig, []byte(msg), hFunc)
	if err != nil {
		t.Fatal(err)
	}
	log.Info(isValid)
	assert.Equal(t, true, isValid)

	hFunc.Reset()
	isValid, err = wrongPK.Verify(sig, []byte(msg), hFunc)
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
	buf := make([]byte, 32)
	copy(buf, seed)
	reader := bytes.NewReader(buf)
	sk, err = GenerateKey(reader)
	return sk, err
}

const (
	sizeFr = fr.Bytes
)

func GenerateKey(r io.Reader) (*PrivateKey, error) {

	c := twistededwards.GetEdwardsCurve()

	var (
		randSrc = make([]byte, 32)
		scalar  = make([]byte, 32)
		pub     PublicKey
	)

	// hash(h) = private_key || random_source, on 32 bytes each
	seed := make([]byte, 32)
	_, err := r.Read(seed)
	if err != nil {
		return nil, err
	}
	h := blake2b.Sum512(seed[:])
	for i := 0; i < 32; i++ {
		randSrc[i] = h[i+32]
	}

	// prune the key
	// https://tools.ietf.org/html/rfc8032#section-5.1.5, key generation

	h[0] &= 0xF8
	h[31] &= 0x7F
	h[31] |= 0x40


	// 0xFC = 1111 1100
	// convert 256 bits to 254 bits supporting bn254 curve

	h[31] &= 0xFC

	// reverse first bytes because setBytes interpret stream as big endian
	// but in eddsa specs s is the first 32 bytes in little endian
	for i, j := 0, sizeFr-1; i < sizeFr; i, j = i+1, j-1 {
		scalar[i] = h[j]
	}

	a := new(big.Int).SetBytes(scalar[:])
	for i := 253; i < 256; i++ {
		a.SetBit(a, i, 0)
	}

	copy(scalar[:], a.FillBytes(make([]byte, 32)))

	var bscalar big.Int
	bscalar.SetBytes(scalar[:])
	pub.A.ScalarMul(&c.Base, &bscalar)

	var res [sizeFr * 3]byte
	pubkBin := pub.A.Bytes()
	subtle.ConstantTimeCopy(1, res[:sizeFr], pubkBin[:])
	subtle.ConstantTimeCopy(1, res[sizeFr:2*sizeFr], scalar[:])
	subtle.ConstantTimeCopy(1, res[2*sizeFr:], randSrc[:])

	var sk = &PrivateKey{}
	// make sure sk is not nil

	_, err = sk.SetBytes(res[:])

	return sk, err
}
*/
