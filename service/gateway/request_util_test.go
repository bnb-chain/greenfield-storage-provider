package gateway

import (
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

func Test_verifySignV1(t *testing.T) {
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
	privKey, _, addr := testdata.KeyEthSecp256k1TestPubAddr()

	// sign
	req, err = signer.SignRequest(*req, addr, privKey, signer.AuthInfo{
		SignType:        model.SignTypeV1,
		MetaMaskSignStr: "",
	})
	require.NoError(t, err)
	// check sign
	rc := &requestContext{
		request: req,
	}
	assert.Equal(t, nil, rc.verifySign())
}

func Test_verifySignV2(t *testing.T) {
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
	privKey, _, addr := testdata.KeyEthSecp256k1TestPubAddr()

	// sign
	req, err = signer.SignRequest(*req, addr, privKey, signer.AuthInfo{
		SignType:        model.SignTypeV2,
		MetaMaskSignStr: MetaMaskStr,
	})
	require.NoError(t, err)
	// check sign
	rc := &requestContext{
		request: req,
	}
	assert.Equal(t, nil, rc.verifySign())
}
