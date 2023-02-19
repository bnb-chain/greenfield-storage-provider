package gateway

import (
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/bnb-chain/greenfield-sdk-go/pkg/signer"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_verifySignatureV1(t *testing.T) {
	// mock request
	urlmap := url.Values{}
	urlmap.Add("greenfield", "storage-provider")
	parms := io.NopCloser(strings.NewReader(urlmap.Encode()))
	req, err := http.NewRequest("POST", "gnfd.nodereal.com", parms)
	require.NoError(t, err)
	req.Header.Set(model.ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Host = "testBucket.gnfd.nodereal.com"
	req.Header.Set(model.GnfdDateHeader, "11:10")
	// mock pk
	privKey, _, _ := testdata.KeyEthSecp256k1TestPubAddr()

	// sign
	err = signer.SignRequest(req, privKey, signer.AuthInfo{
		SignType:        model.SignTypeV1,
		MetaMaskSignStr: "",
	})
	require.NoError(t, err)
	// check sign
	rc := &requestContext{
		request: req,
	}
	assert.Equal(t, nil, rc.verifySignature())
}

func Test_verifySignatureV2(t *testing.T) {
	// mock request
	urlmap := url.Values{}
	urlmap.Add("greenfield", "storage-provider")
	parms := io.NopCloser(strings.NewReader(urlmap.Encode()))
	req, err := http.NewRequest("POST", "gnfd.nodereal.com", parms)
	require.NoError(t, err)
	req.Header.Set(model.ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Host = "testBucket.gnfd.nodereal.com"
	req.Header.Set(model.GnfdDateHeader, "11:10")
	// mock pk
	privKey, _, _ := testdata.KeyEthSecp256k1TestPubAddr()

	// sign
	err = signer.SignRequest(req, privKey, signer.AuthInfo{
		SignType:        model.SignTypeV2,
		MetaMaskSignStr: hex.EncodeToString([]byte("123")),
	})
	require.NoError(t, err)
	// check sign
	rc := &requestContext{
		request: req,
	}
	assert.Equal(t, nil, rc.verifySignature())
}

func Test_parseRangeHeader(t *testing.T) {
	isRange, start, end := parseRange("bytes=1")
	assert.Equal(t, false, isRange)

	isRange, start, end = parseRange("bytes=1-")
	assert.Equal(t, true, isRange)
	assert.Equal(t, 1, int(start))
	assert.Equal(t, -1, int(end))

	isRange, start, end = parseRange("bytes=1-100")
	assert.Equal(t, true, isRange)
	assert.Equal(t, 1, int(start))
	assert.Equal(t, 100, int(end))
}
