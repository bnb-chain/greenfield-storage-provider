package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	AuthStrFormat    = "PersonalSign ECDSA-secp256k1,SignedMsg=%s,Signature=%s"
	SampleDAppDomain = "https://greenfield.dapp.sample.io"
	SamplePublicKey  = "a_sample_eddsa_public_key_for_off_chain_auth"
)

// TODO: Stop referencing SDK code for testing, If needed, move the mainly function to Greenfield Common
/*
func TestVerifySignatureV1(t *testing.T) {
	// mock request
	urlmap := url.Values{}
	urlmap.Add("greenfield", "storage-provider")
	parms := io.NopCloser(strings.NewReader(urlmap.Encode()))
	req, err := http.NewRequest("POST", "gnfd.test-sp.com", parms)
	require.NoError(t, err)
	req.Header.Set(model.ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Host = "testBucket.gnfd.test-sp.com"
	// mock
	privKey, _, addrInput := testdata.KeyEthSecp256k1TestPubAddr()
	keyManager, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	require.NoError(t, err)
	// sign
	spClient, err := sp.NewSpClient("example.com", sp.WithKeyManager(keyManager), sp.WithSecure(false))
	require.NoError(t, err)
	err = spClient.SignRequest(req, sp.NewAuthInfo(false, ""))
	require.NoError(t, err)
	// check sign
	rc := &requestContext{
		request: req,
	}
	addrOutput, err := gw.verifySignature(rc)
	assert.Equal(t, nil, err)
	assert.Equal(t, addrInput.String(), addrOutput.String())
}
*/

func Test_MakePersonalSignatureAndRecover(t *testing.T) {
	unSignedContent := `Register your identity of dapp http://dapp.io
    with your identity key 0x12345.
    In the following SPs:
    - SP 0X123450 (name: SP_DEV) -  Nonce: 7
    - SP 0X123451 (name: SP_QA) -  Nonce: 2
    - SP 0X123452 (name: SP_PROD) -  Nonce: 3

    The expiry date is 2023-03-25 16:34:12 GMT+08:00`

	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))

	// Account information.
	privateKey, _ := crypto.GenerateKey()

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Infof("address is " + address.Hex())
	// Sign data.
	signature, err := crypto.Sign(unSignedContentHash, privateKey)
	if err != nil {
		log.Error(err)
	}

	err = storagetypes.VerifySignature(address.Bytes(), unSignedContentHash, signature)
	assert.Equal(t, nil, err)
}

func Test_verifyPersonalSignatureFromWallet(t *testing.T) {
	unSignedContent := "Example `personal_sign` message"
	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))
	sig := "0x94a181416075908cf580d93222dbaa9b32cb73428ab88c6722e849005ca5cb301a860aa7fd6325645cc9e1e58e4dc279ce43cfd3220e54f7fbec37194127b0201b"
	sigHash, _ := hexutil.Decode(sig)

	err := storagetypes.VerifySignature(sdk.MustAccAddressFromHex("0xa64fdc3b4866cd2ac664998c7b180813fb9b06e6"), unSignedContentHash, sigHash)
	assert.Equal(t, nil, err)
}

func Test_verifyPersonalSignatureFromRequest(t *testing.T) {
	nonce := "123456"
	spAddr := "0x1234567"
	unSignedContent := UnsignedContentTemplate

	validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
	unSignedContent = fmt.Sprintf(unSignedContent, SampleDAppDomain, SamplePublicKey, spAddr, nonce, validExpiryDate)
	log.Infof("unSignedContent is: %s", unSignedContent)
	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))
	// Account information.
	privateKey, _ := crypto.GenerateKey()

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Infof("address is: " + address.Hex())
	// Sign data.
	signature, err := crypto.Sign(unSignedContentHash, privateKey)
	if err != nil {
		log.Error(err)
	}

	err = storagetypes.VerifySignature(address.Bytes(), unSignedContentHash, signature)
	assert.Equal(t, nil, err)

	// mock request
	urlmap := url.Values{}
	urlmap.Add("greenfield", "storage-provider")
	parms := io.NopCloser(strings.NewReader(urlmap.Encode()))
	req, err := http.NewRequest("POST", "example.com", parms)
	require.NoError(t, err)
	req.Header.Set(model.ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Host = "testBucket.gnfd.test-sp.com"

	authString := fmt.Sprintf(AuthStrFormat, unSignedContent, hexutil.Encode(signature))
	log.Infof("authString is: %s", authString)
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	req.Header.Set(model.GnfdUserAddressHeader, address.String())

	// check sign
	rc := &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, nil, err)

}

func Test_verifyPersonalSignatureFromRequest_Error(t *testing.T) {
	// mock request
	urlmap := url.Values{}
	urlmap.Add("greenfield", "storage-provider")
	parms := io.NopCloser(strings.NewReader(urlmap.Encode()))
	req, err := http.NewRequest("POST", "example.com", parms)
	require.NoError(t, err)
	req.Header.Set(model.ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Host = "testBucket.gnfd.test-sp.com"

	// unexpected wrong sig format
	authString := fmt.Sprintf(AuthStrFormat, "unSignedContent", "signature")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc := &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, hexutil.ErrMissingPrefix, err)

	// unexpected = char
	authString = fmt.Sprintf(AuthStrFormat, "unexpected=char", "signature")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc = &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, errors.ErrAuthorizationFormat, err)

	// unSupportedKey
	authString = `PersonalSign ECDSA-secp256k1,SignedMsg=%s,unSupportedKey=%s`
	authString = fmt.Sprintf(authString, "unSignedContent", "signature")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc = &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, errors.ErrAuthorizationFormat, err)

	// unexpected more content
	authString = `PersonalSign ECDSA-secp256k1,SignedMsg=%s,Signature=%s,wrongContent`
	authString = fmt.Sprintf(authString, "unSignedContent", "signature")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc = &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, errors.ErrAuthorizationFormat, err)

	// wrong sig length
	authString = `PersonalSign ECDSA-secp256k1,SignedMsg=%s,Signature=%s,wrongContent`
	authString = fmt.Sprintf(authString, "unSignedContent", "0x123")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc = &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, errors.ErrAuthorizationFormat, err)

	// wrong sig length
	authString = AuthStrFormat
	authString = fmt.Sprintf(authString, "unSignedContent", "0x1234")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc = &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, errors.ErrSignatureConsistent, err)

	invalidETHSig := "0x04a181416075908cf580d93222dbaa9b32cb73428ab88c6722e849005ca5cb301a860aa7fd6325645cc9e1e58e4dc279ce43cfd3220e54f7fbec37194127b0201b"
	// wrong recovery id offset
	authString = AuthStrFormat
	authString = fmt.Sprintf(authString, "unSignedContent", invalidETHSig)
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	rc = &requestContext{
		request: req,
	}
	_, err = gw.verifySignature(rc)
	assert.Equal(t, errors.ErrSignatureConsistent, err)
}

func TestParseRangeHeader(t *testing.T) {
	isRange, _, _ := parseRange("bytes=1")
	assert.Equal(t, false, isRange)

	isRange, start, end := parseRange("bytes=1-")
	assert.Equal(t, true, isRange)
	assert.Equal(t, 1, int(start))
	assert.Equal(t, -1, int(end))

	isRange, start, end = parseRange("bytes=1-100")
	assert.Equal(t, true, isRange)
	assert.Equal(t, 1, int(start))
	assert.Equal(t, 100, int(end))
}
