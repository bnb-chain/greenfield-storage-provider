package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/sdk/keys"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func setupKM() (keys.KeyManager, error) {
	// the private key comes from greenfield/sdk/keys/ unit test
	privKey := "ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf"
	return keys.NewPrivateKeyManager(privKey)
}

func Test_verifyPingMsgSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	pingMsg := &Ping{
		SpOperatorAddress: km.GetAddr().String(),
	}
	sigs, err := km.GetPrivKey().Sign(pingMsg.GetSignBytes())
	assert.NoError(t, err)
	pingMsg.Signature = sigs
	valid := km.GetPrivKey().PubKey().VerifySignature(pingMsg.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(pingMsg.GetSpOperatorAddress(), pingMsg.GetSignBytes(), pingMsg.GetSignature())
	assert.NoError(t, err)
}

func Test_verifyPongMsgSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	pongMsg := &Pong{
		Nodes: []*Node{
			&Node{
				NodeId:    "1234567890",
				MultiAddr: []string{"/tcp/localhost:7133"},
			},
		},
		SpOperatorAddress: km.GetAddr().String(),
	}
	sigs, err := km.GetPrivKey().Sign(pongMsg.GetSignBytes())
	assert.NoError(t, err)
	pongMsg.Signature = sigs
	valid := km.GetPrivKey().PubKey().VerifySignature(pongMsg.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(pongMsg.GetSpOperatorAddress(), pongMsg.GetSignBytes(), pongMsg.GetSignature())
	assert.NoError(t, err)
}

func Test_verifyGetApprovalRequestMsgSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	approvalReqMsg := &GetApprovalRequest{
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:         sdkmath.NewUint(uint64(0)),
			BucketName: "test_bucket",
		},
		SpOperatorAddress: km.GetAddr().String(),
	}
	sigs, err := km.GetPrivKey().Sign(sdk.Keccak256(approvalReqMsg.GetSignBytes()))
	approvalReqMsg.Signature = sigs
	assert.NoError(t, err)
	valid := km.GetPrivKey().PubKey().VerifySignature(approvalReqMsg.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(approvalReqMsg.GetSpOperatorAddress(), approvalReqMsg.GetSignBytes(), approvalReqMsg.GetSignature())
	assert.NoError(t, err)
}

func Test_verifyGetApprovalResponseMsgSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	approvalRspMsg := &GetApprovalResponse{
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:         sdkmath.NewUint(uint64(1)),
			BucketName: "test_bucket",
		},
		SpOperatorAddress: km.GetAddr().String(),
	}
	sigs, err := km.GetPrivKey().Sign(approvalRspMsg.GetSignBytes())
	approvalRspMsg.Signature = sigs
	assert.NoError(t, err)
	valid := km.GetPrivKey().PubKey().VerifySignature(approvalRspMsg.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(approvalRspMsg.GetSpOperatorAddress(), approvalRspMsg.GetSignBytes(), approvalRspMsg.GetSignature())
	assert.NoError(t, err)
}
