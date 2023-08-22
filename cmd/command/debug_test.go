package command

import (
	"fmt"
	"os"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCreateBucketApproval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	o1 := mockGRPCAPI.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(true, &gfsptask.GfSpCreateBucketApprovalTask{}, nil)
	o2 := mockGRPCAPI.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(false, &gfsptask.GfSpCreateBucketApprovalTask{}, nil)
	o3 := mockGRPCAPI.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(true, &gfsptask.GfSpCreateBucketApprovalTask{}, fmt.Errorf("failed to get create bucket approval"))
	gomock.InOrder(o1, o2, o3)

	CW.grpcAPI = mockGRPCAPI
	// succeed
	err := DebugCreateBucketApprovalCmd.Action(&cli.Context{})
	assert.Nil(t, err)
	// not allow
	err = DebugCreateBucketApprovalCmd.Action(&cli.Context{})
	assert.NotNil(t, err)
	// failed
	err = DebugCreateBucketApprovalCmd.Action(&cli.Context{})
	assert.NotNil(t, err)
}

func TestCreateObjectApproval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	o1 := mockGRPCAPI.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(true, &gfsptask.GfSpCreateObjectApprovalTask{}, nil)
	o2 := mockGRPCAPI.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(false, &gfsptask.GfSpCreateObjectApprovalTask{}, nil)
	o3 := mockGRPCAPI.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(true, &gfsptask.GfSpCreateObjectApprovalTask{}, fmt.Errorf("failed to get create object approval"))
	gomock.InOrder(o1, o2, o3)

	CW.grpcAPI = mockGRPCAPI
	// succeed
	err := DebugCreateObjectApprovalCmd.Action(&cli.Context{})
	assert.Nil(t, err)
	// not allow
	err = DebugCreateObjectApprovalCmd.Action(&cli.Context{})
	assert.NotNil(t, err)
	// failed
	err = DebugCreateObjectApprovalCmd.Action(&cli.Context{})
	assert.NotNil(t, err)
}

func TestReplicatePieceApproval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	o1 := mockGRPCAPI.EXPECT().AskSecondaryReplicatePieceApproval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to get replicate piece approval"))
	o2 := mockGRPCAPI.EXPECT().AskSecondaryReplicatePieceApproval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*gfsptask.GfSpReplicatePieceApprovalTask{{}}, nil)
	o4 := mockGRPCAPI.EXPECT().AskSecondaryReplicatePieceApproval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*gfsptask.GfSpReplicatePieceApprovalTask{{}}, nil)
	mockDBAPI := spdb.NewMockSPDB(ctrl)
	o3 := mockDBAPI.EXPECT().GetSpByAddress(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil)
	o5 := mockDBAPI.EXPECT().GetSpByAddress(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to query sp db"))
	gomock.InOrder(o1, o2, o3, o4, o5)

	CW.grpcAPI = mockGRPCAPI
	CW.spDBAPI = mockDBAPI
	// failed to get replicate approval
	err := DebugReplicateApprovalCmd.Action(&cli.Context{})
	assert.NotNil(t, err)
	// succeed
	err = DebugReplicateApprovalCmd.Action(&cli.Context{})
	assert.Nil(t, err)
	// failed to query sp db
	err = DebugReplicateApprovalCmd.Action(&cli.Context{})
	assert.NotNil(t, err)
}

func TestPutObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	o1 := mockGRPCAPI.EXPECT().UploadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to upload"))
	o2 := mockGRPCAPI.EXPECT().UploadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	gomock.InOrder(o1, o2)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		DebugPutObjectCmd,
	}
	// failed, due to not exist file
	err := app.Run([]string{"./gnfd-sp", "debug.put.object", "--file", "./not_exist_file"})
	assert.NotNil(t, err)

	// failed, due to file too large
	fileContent := util.RandomString(DefaultMaxSegmentPieceSize + 1)
	err = os.WriteFile("./too_large_file", []byte(fileContent), os.ModePerm)
	assert.Nil(t, err)
	err = app.Run([]string{"./gnfd-sp", "debug.put.object", "--file", "./too_large_file"})
	assert.NotNil(t, err)
	os.Remove("./too_large_file")

	// failed due to uploader
	fileContent = util.RandomString(8 * 1024 * 1024)
	err = os.WriteFile("./normal_file", []byte(fileContent), os.ModePerm)
	assert.Nil(t, err)
	err = app.Run([]string{"./gnfd-sp", "debug.put.object", "--file", "./normal_file"})
	assert.NotNil(t, err)

	// succeed
	err = app.Run([]string{"./gnfd-sp", "debug.put.object", "--file", "./normal_file"})
	assert.Nil(t, err)
	os.Remove("./normal_file")
}
