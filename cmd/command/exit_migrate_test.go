package command

import (
	"fmt"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestSPExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	o1 := mockGRPCAPI.EXPECT().SPExit(gomock.Any(), gomock.Any()).Return("txHash", nil)
	o2 := mockGRPCAPI.EXPECT().SPExit(gomock.Any(), gomock.Any()).Return("", fmt.Errorf("failed to send signer"))
	gomock.InOrder(o1, o2)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		SPExitCmd,
	}
	// failed due to check operator address
	err := app.Run([]string{"./gnfd-sp", "sp.exit", "--operatorAddress", "abc"})
	assert.NotNil(t, err)

	// succeed
	CW.config.SpAccount.SpOperatorAddress = "abc"
	err = app.Run([]string{"./gnfd-sp", "sp.exit", "--operatorAddress", "abc"})
	assert.Nil(t, err)

	// failed due to send signer error
	err = app.Run([]string{"./gnfd-sp", "sp.exit", "--operatorAddress", "abc"})
	assert.NotNil(t, err)
}

func TestCompleteSPExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	o1 := mockGRPCAPI.EXPECT().CompleteSPExit(gomock.Any(), gomock.Any()).Return("txHash", nil)
	o2 := mockGRPCAPI.EXPECT().CompleteSPExit(gomock.Any(), gomock.Any()).Return("", fmt.Errorf("failed to send signer"))
	gomock.InOrder(o1, o2)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		CompleteSPExitCmd,
	}
	// failed due to check operator address
	err := app.Run([]string{"./gnfd-sp", "sp.complete.exit", "--operatorAddress", "abc"})
	assert.NotNil(t, err)

	// succeed
	CW.config.SpAccount.SpOperatorAddress = "abc"
	err = app.Run([]string{"./gnfd-sp", "sp.complete.exit", "--operatorAddress", "abc"})
	assert.Nil(t, err)

	// failed due to send signer error
	err = app.Run([]string{"./gnfd-sp", "sp.complete.exit", "--operatorAddress", "abc"})
	assert.NotNil(t, err)
}

func TestCompleteSwapOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	CW.config = &gfspconfig.GfSpConfig{}
	CW.spDBAPI = spdb.NewMockSPDB(ctrl)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	CW.grpcAPI = mockGRPCAPI
	o1 := mockGRPCAPI.EXPECT().CompleteSwapOut(gomock.Any(), gomock.Any()).Return("txHash", nil)
	o2 := mockGRPCAPI.EXPECT().CompleteSwapOut(gomock.Any(), gomock.Any()).Return("", fmt.Errorf("failed to send signer"))
	gomock.InOrder(o1, o2)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		CompleteSwapOutCmd,
	}
	// failed due to check operator address
	err := app.Run([]string{"./gnfd-sp", "sp.complete.swapout", "--operatorAddress", "abc", "-gvgIDList", ""})
	assert.NotNil(t, err)

	// failed due to invalid gvg list
	err = app.Run([]string{"./gnfd-sp", "sp.complete.swapout", "--operatorAddress", "abc", "-gvgIDList", "abc"})
	assert.NotNil(t, err)

	// succeed
	CW.config.SpAccount.SpOperatorAddress = "abc"
	err = app.Run([]string{"./gnfd-sp", "sp.complete.swapout", "--operatorAddress", "abc", "--gvgIDList", "1,2,3", "--familyID", "1"})
	assert.Nil(t, err)

	// failed due to send signer error
	err = app.Run([]string{"./gnfd-sp", "sp.complete.swapout", "--operatorAddress", "abc", "--gvgIDList", "1,2,3"})
	assert.NotNil(t, err)
}
