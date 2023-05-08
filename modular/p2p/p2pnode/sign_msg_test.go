package p2pnode

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield/sdk/keys"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
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
	pingMsg := &gfspp2p.GfSpPing{
		SpOperatorAddress: km.GetAddr().String(),
	}
	sigs, err := km.Sign(pingMsg.GetSignBytes())
	assert.NoError(t, err)
	pingMsg.Signature = sigs
	valid := km.PubKey().VerifySignature(pingMsg.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(pingMsg.GetSpOperatorAddress(), pingMsg.GetSignBytes(), pingMsg.GetSignature())
	assert.NoError(t, err)
}

func Test_verifyPongMsgSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	pongMsg := &gfspp2p.GfSpPong{
		Nodes: []*gfspp2p.GfSpNode{
			&gfspp2p.GfSpNode{
				NodeId:    "1234567890",
				MultiAddr: []string{"/tcp/localhost:7133"},
			},
		},
		SpOperatorAddress: km.GetAddr().String(),
	}
	sigs, err := km.Sign(pongMsg.GetSignBytes())
	assert.NoError(t, err)
	pongMsg.Signature = sigs
	valid := km.PubKey().VerifySignature(pongMsg.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(pongMsg.GetSpOperatorAddress(), pongMsg.GetSignBytes(), pongMsg.GetSignature())
	assert.NoError(t, err)
}

func Test_verifyReplicatePieceApprovalTaskSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	task := &gfsptask.GfSpReplicatePieceApprovalTask{
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(0)},
		StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 0}},
	}
	sigs, err := km.Sign(task.GetSignBytes())
	assert.NoError(t, err)
	task.SetApprovedSignature(sigs)
	task.SetApprovedSpOperatorAddress(km.GetAddr().String())
	valid := km.PubKey().VerifySignature(task.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(task.GetApprovedSpOperatorAddress(), task.GetSignBytes(), task.GetApprovedSignature())
	assert.NoError(t, err)
}

func Test_verifyReceivePieceTaskSignature(t *testing.T) {
	km, err := setupKM()
	assert.NoError(t, err)
	receive := &gfsptask.GfSpReceivePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(0)},
		StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 0}},
		PieceSize:     10,
		PieceIdx:      -1,
		ReplicateIdx:  1,
	}
	sigs, err := km.Sign(receive.GetSignBytes())
	assert.NoError(t, err)
	receive.SetSignature(sigs)
	valid := km.PubKey().VerifySignature(receive.GetSignBytes(), sigs)
	assert.True(t, valid)
	err = VerifySignature(km.GetAddr().String(), receive.GetSignBytes(), receive.GetSignature())
	assert.NoError(t, err)
}
