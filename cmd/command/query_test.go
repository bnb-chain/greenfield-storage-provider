package command

import (
	"fmt"
	"os"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestListModules(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListModulesCmd,
	}
	err := app.Run([]string{"./gnfd-sp", "list.modules"})
	assert.Nil(t, err)
}

func TestListErrors(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListErrorsCmd,
	}
	err := app.Run([]string{"./gnfd-sp", "list.errors"})
	assert.Nil(t, err)
}

func TestQueryTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	o1 := mockGRPCAPI.EXPECT().QueryTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to query task"))
	o2 := mockGRPCAPI.EXPECT().QueryTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	gomock.InOrder(o1, o2)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		QueryTaskCmd,
	}

	// failed due to config file is not found
	err := app.Run([]string{"./gnfd-sp", "query.task", "--task.key", "abc", "--config", "not_exist_config"})
	assert.NotNil(t, err)

	// failed due to query task error
	err = ConfigDumpCmd.Action(&cli.Context{})
	assert.Equal(t, nil, err)
	_, err = os.Stat(DefaultConfigFile)
	assert.Equal(t, nil, err)

	err = app.Run([]string{"./gnfd-sp", "query.task", "--task.key", "abc", "--config", DefaultConfigFile})
	assert.NotNil(t, err)

	// succeed
	err = app.Run([]string{"./gnfd-sp", "query.task", "--task.key", "abc", "--config", DefaultConfigFile})
	assert.Nil(t, err)

	// clear temp config file
	os.Remove(DefaultConfigFile)
}

func TestGetObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	mockConsensusAPI := consensus.NewMockConsensus(ctrl)
	CW.chainAPI = mockConsensusAPI

	o1 := mockConsensusAPI.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{}, nil)
	o2 := mockConsensusAPI.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil)
	o3 := mockConsensusAPI.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{}, nil)
	o4 := mockGRPCAPI.EXPECT().GetObject(gomock.Any(), gomock.Any()).Return([]byte{1}, nil)
	gomock.InOrder(o1, o2, o3, o4)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		GetObjectCmd,
	}
	err := app.Run([]string{"./gnfd-sp", "get.object", "--object.id", "100"})
	assert.Nil(t, err)
}

func TestGetChallengePieceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	mockConsensusAPI := consensus.NewMockConsensus(ctrl)
	CW.chainAPI = mockConsensusAPI

	o1 := mockConsensusAPI.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{}, nil)
	o2 := mockConsensusAPI.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil)
	o3 := mockConsensusAPI.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{}, nil)
	o4 := mockGRPCAPI.EXPECT().GetChallengeInfo(gomock.Any(), gomock.Any()).Return([]byte(""), [][]byte{[]byte("mock")}, []byte(""), nil)

	gomock.InOrder(o1, o2, o3, o4)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ChallengePieceCmd,
	}
	err := app.Run([]string{"./gnfd-sp", "challenge.piece", "--object.id", "100", "--redundancy.index", "0", "--segment.index", "0"})
	assert.NotNil(t, err)
}

func TestGetSegmentIntegrity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	mockDBAPI := spdb.NewMockSPDB(ctrl)
	CW.spDBAPI = mockDBAPI
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	mockConsensusAPI := consensus.NewMockConsensus(ctrl)
	CW.chainAPI = mockConsensusAPI

	o1 := mockGRPCAPI.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{}, nil)
	o2 := mockGRPCAPI.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&types.Bucket{BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(10)}}, nil)
	o3 := mockGRPCAPI.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{}, nil)
	o4 := mockConsensusAPI.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil)
	o5 := mockDBAPI.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&spdb.IntegrityMeta{}, nil)
	gomock.InOrder(o1, o2, o3, o4, o5)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		GetSegmentIntegrityCmd,
	}
	err := app.Run([]string{"./gnfd-sp", "get.piece.integrity", "--object.id", "100"})
	assert.Nil(t, err)
}

func TestQueryBucketMigrate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI

	o1 := mockGRPCAPI.EXPECT().QueryBucketMigrate(gomock.Any(), gomock.Any()).Return("bucket_migrate", nil)
	gomock.InOrder(o1)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		QueryBucketMigrateCmd,
	}
	err := ConfigDumpCmd.Action(&cli.Context{})
	assert.Equal(t, nil, err)
	_, err = os.Stat(DefaultConfigFile)
	assert.Equal(t, nil, err)

	err = app.Run([]string{"./gnfd-sp", "query.bucket.migrate", "--config", DefaultConfigFile, "--endpoint", "localhost:2012"})
	assert.Nil(t, err)
	// clear temp config file
	os.Remove(DefaultConfigFile)
}

func TestQuerySPExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI

	o1 := mockGRPCAPI.EXPECT().QuerySPExit(gomock.Any(), gomock.Any()).Return("sp_exit", nil)
	gomock.InOrder(o1)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		QuerySPExitCmd,
	}
	err := ConfigDumpCmd.Action(&cli.Context{})
	assert.Equal(t, nil, err)
	_, err = os.Stat(DefaultConfigFile)
	assert.Equal(t, nil, err)

	err = app.Run([]string{"./gnfd-sp", "query.sp.exit", "--config", DefaultConfigFile, "--endpoint", "localhost:2012"})
	assert.Nil(t, err)
	// clear temp config file
	os.Remove(DefaultConfigFile)
}
