package gfspapp

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	mockSig    = []byte("mockSig")
	mockTxHash = "mockTxHash"
)

func TestGfSpBaseApp_GfSpSignSuccess1(t *testing.T) {
	t.Log("Success case description: sign create bucket info")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignCreateBucketApproval(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CreateBucketInfo{CreateBucketInfo: mockCreateBucketInfo}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess2(t *testing.T) {
	t.Log("Success case description: sign migrate bucket info")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_MigrateBucketInfo{
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			Operator:       "mockOperator",
			BucketName:     "mockBucketName",
			DstPrimarySpId: 1,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess3(t *testing.T) {
	t.Log("Success case description: sign create object info")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignCreateObjectApproval(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CreateObjectInfo{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName: "mockObjectName",
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess4(t *testing.T) {
	t.Log("Success case description: sign seal object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SealObject(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SealObjectInfo{
		SealObjectInfo: &storagetypes.MsgSealObject{
			ObjectName: "mockObjectName",
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess5(t *testing.T) {
	t.Log("Success case description: sign seal object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().RejectUnSealObject(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_RejectObjectInfo{
		RejectObjectInfo: &storagetypes.MsgRejectSealObject{
			ObjectName: "mockObjectName",
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess6(t *testing.T) {
	t.Log("Success case description: sign discontinue bucket")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().DiscontinueBucket(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_DiscontinueBucketInfo{
		DiscontinueBucketInfo: &storagetypes.MsgDiscontinueBucket{
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess7(t *testing.T) {
	t.Log("Success case description: sign secondary bls signature")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignSecondarySealBls(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SignSecondarySealBls{
		SignSecondarySealBls: &gfspserver.GfSpSignSecondarySealBls{
			ObjectId: 1,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess8(t *testing.T) {
	t.Log("Success case description: sign p2p ping msg")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignP2PPingMsg(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_PingMsg{
		PingMsg: &gfspp2p.GfSpPing{
			SpOperatorAddress: "mockSpOperatorAddress",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess9(t *testing.T) {
	t.Log("Success case description: sign p2p pong msg")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignP2PPongMsg(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_PongMsg{
		PongMsg: &gfspp2p.GfSpPong{
			SpOperatorAddress: "mockSpOperatorAddress",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess10(t *testing.T) {
	t.Log("Success case description: sign receive piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignReceivePieceTask(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspReceivePieceTask{
		GfspReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess11(t *testing.T) {
	t.Log("Success case description: sign replicate piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignReplicatePieceApproval(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask{
		GfspReplicatePieceApprovalTask: &gfsptask.GfSpReplicatePieceApprovalTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess12(t *testing.T) {
	t.Log("Success case description: sign create global virtual group")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CreateGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	coin := sdk.NewCoin("mock", sdkmath.NewInt(1))
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup{
		CreateGlobalVirtualGroup: &gfspserver.GfSpCreateGlobalVirtualGroup{
			VirtualGroupFamilyId: 1,
			Deposit:              &coin,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess13(t *testing.T) {
	t.Log("Success case description: sign recover piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignRecoveryPieceTask(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspRecoverPieceTask{
		GfspRecoverPieceTask: &gfsptask.GfSpRecoverPieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess14(t *testing.T) {
	t.Log("Success case description: sign migrate gvg task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignMigrateGVG(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspMigrateGvgTask{}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess15(t *testing.T) {
	t.Log("Success case description: sign complete migrate bucket")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CompleteMigrateBucket(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CompleteMigrateBucket{
		CompleteMigrateBucket: &storagetypes.MsgCompleteMigrateBucket{
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess16(t *testing.T) {
	t.Log("Success case description: sign secondary sp bls migration bucket")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignSecondarySPMigrationBucket(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SignSecondarySpMigrationBucket{
		SignSecondarySpMigrationBucket: &storagetypes.SecondarySpMigrationBucketSignDoc{
			BucketId: sdk.NewUint(1),
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess17(t *testing.T) {
	t.Log("Success case description: sign swap out")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SwapOut(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SwapOut{
		SwapOut: &virtual_types.MsgSwapOut{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess18(t *testing.T) {
	t.Log("Success case description: sign swap out approval")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignSwapOut(gomock.Any(), gomock.Any()).Return(mockSig, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SignSwapOut{
		SignSwapOut: &virtual_types.MsgSwapOut{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockSig, result.GetSignature())
}

func TestGfSpBaseApp_GfSpSignSuccess19(t *testing.T) {
	t.Log("Success case description: sign complete swap out")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CompleteSwapOut(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CompleteSwapOut{
		CompleteSwapOut: &virtual_types.MsgCompleteSwapOut{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess20(t *testing.T) {
	t.Log("Success case description: sign sp exit")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SPExit(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SpExit{
		SpExit: &virtual_types.MsgStorageProviderExit{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess21(t *testing.T) {
	t.Log("Success case description: sign complete sp exit")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CompleteSPExit(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CompleteSpExit{
		CompleteSpExit: &virtual_types.MsgCompleteStorageProviderExit{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignSuccess22(t *testing.T) {
	t.Log("Success case description: sign update sp price")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().UpdateSPPrice(gomock.Any(), gomock.Any()).Return(mockTxHash, nil).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SpStoragePrice{
		SpStoragePrice: &sptypes.MsgUpdateSpStoragePrice{
			SpAddress: "mockSpAddress",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockTxHash, result.GetTxHash())
}

func TestGfSpBaseApp_GfSpSignFailure1(t *testing.T) {
	t.Log("Failure case description: failed to sign seal object")
	g := setup(t)
	result, err := g.GfSpSign(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrSingTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_GfSpSignFailure2(t *testing.T) {
	t.Log("Failure case description: failed to sign create bucket approval")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignCreateBucketApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CreateBucketInfo{CreateBucketInfo: mockCreateBucketInfo}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure3(t *testing.T) {
	t.Log("Failure case description: failed to sign migrate bucket info")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_MigrateBucketInfo{
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			Operator:       "mockOperator",
			BucketName:     "mockBucketName",
			DstPrimarySpId: 1,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure4(t *testing.T) {
	t.Log("Failure case description: failed to sign create object info")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignCreateObjectApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CreateObjectInfo{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName: "mockObjectName",
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure5(t *testing.T) {
	t.Log("Failure case description: failed to sign seal object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SealObject(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SealObjectInfo{
		SealObjectInfo: &storagetypes.MsgSealObject{
			ObjectName: "mockObjectName",
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure6(t *testing.T) {
	t.Log("Failure case description: failed to sign seal object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().RejectUnSealObject(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_RejectObjectInfo{
		RejectObjectInfo: &storagetypes.MsgRejectSealObject{
			ObjectName: "mockObjectName",
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure7(t *testing.T) {
	t.Log("Failure case description: failed to sign discontinue bucket")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().DiscontinueBucket(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_DiscontinueBucketInfo{
		DiscontinueBucketInfo: &storagetypes.MsgDiscontinueBucket{
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure8(t *testing.T) {
	t.Log("Failure case description: failed to sign secondary bls signature")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignSecondarySealBls(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SignSecondarySealBls{
		SignSecondarySealBls: &gfspserver.GfSpSignSecondarySealBls{
			ObjectId: 1,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure9(t *testing.T) {
	t.Log("Failure case description: failed to sign p2p ping msg")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignP2PPingMsg(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_PingMsg{
		PingMsg: &gfspp2p.GfSpPing{
			SpOperatorAddress: "mockSpOperatorAddress",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure10(t *testing.T) {
	t.Log("Failure case description: failed to sign p2p pong msg")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignP2PPongMsg(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_PongMsg{
		PongMsg: &gfspp2p.GfSpPong{
			SpOperatorAddress: "mockSpOperatorAddress",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure11(t *testing.T) {
	t.Log("Failure case description: failed to sign receive piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignReceivePieceTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspReceivePieceTask{
		GfspReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure12(t *testing.T) {
	t.Log("Failure case description: failed to sign replicate piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignReplicatePieceApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask{
		GfspReplicatePieceApprovalTask: &gfsptask.GfSpReplicatePieceApprovalTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure13(t *testing.T) {
	t.Log("Failure case description: failed to sign create global virtual group")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CreateGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	coin := sdk.NewCoin("mock", sdkmath.NewInt(1))
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup{
		CreateGlobalVirtualGroup: &gfspserver.GfSpCreateGlobalVirtualGroup{
			VirtualGroupFamilyId: 1,
			Deposit:              &coin,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure14(t *testing.T) {
	t.Log("Failure case description: failed to sign recover piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignRecoveryPieceTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspRecoverPieceTask{
		GfspRecoverPieceTask: &gfsptask.GfSpRecoverPieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure15(t *testing.T) {
	t.Log("Failure case description: failed to sign migrate gvg task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignMigrateGVG(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_GfspMigrateGvgTask{}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure16(t *testing.T) {
	t.Log("Failure case description: failed to sign complete migrate bucket")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CompleteMigrateBucket(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CompleteMigrateBucket{
		CompleteMigrateBucket: &storagetypes.MsgCompleteMigrateBucket{
			BucketName: "mockBucketName",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure17(t *testing.T) {
	t.Log("Failure case description: failed to sign secondary sp bls migration bucket")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignSecondarySPMigrationBucket(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SignSecondarySpMigrationBucket{
		SignSecondarySpMigrationBucket: &storagetypes.SecondarySpMigrationBucketSignDoc{
			BucketId: sdk.NewUint(1),
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure18(t *testing.T) {
	t.Log("Failure case description: failed to sign swap out")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SwapOut(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SwapOut{
		SwapOut: &virtual_types.MsgSwapOut{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure19(t *testing.T) {
	t.Log("Failure case description: failed to sign swap out approval")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SignSwapOut(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SignSwapOut{
		SignSwapOut: &virtual_types.MsgSwapOut{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure20(t *testing.T) {
	t.Log("Failure case description: failed to sign complete swap out")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CompleteSwapOut(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CompleteSwapOut{
		CompleteSwapOut: &virtual_types.MsgCompleteSwapOut{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure21(t *testing.T) {
	t.Log("Failure case description: failed to sign sp exit")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().SPExit(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SpExit{
		SpExit: &virtual_types.MsgStorageProviderExit{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure22(t *testing.T) {
	t.Log("Failure case description: failed to sign complete sp exit")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().CompleteSPExit(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_CompleteSpExit{
		CompleteSpExit: &virtual_types.MsgCompleteStorageProviderExit{
			StorageProvider: "mockSP",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpSignFailure23(t *testing.T) {
	t.Log("Failure case description: failed to sign update sp price")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockSigner(ctrl)
	g.signer = m
	m.EXPECT().UpdateSPPrice(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
	req := &gfspserver.GfSpSignRequest{Request: &gfspserver.GfSpSignRequest_SpStoragePrice{
		SpStoragePrice: &sptypes.MsgUpdateSpStoragePrice{
			SpAddress: "mockSpAddress",
		}}}
	result, err := g.GfSpSign(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}
