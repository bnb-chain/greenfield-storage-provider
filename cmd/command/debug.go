package command

import (
	"bytes"
	"context"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/types/common"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	DebugCommandPrefix = "gfsp-cli-debug-"
)

var fileFlag = &cli.StringFlag{
	Name:     "f",
	Usage:    "The file of uploading",
	Required: true,
}

var DebugCreateBucketApprovalCmd = &cli.Command{
	Action: createBucketApproval,
	Name:   "debug.create.bucket.approval",
	Usage:  "Create random CreateBucketApproval and send to approver for debugging and testing",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
	},
	Category: "DEBUG COMMANDS",
	Description: `The debug.create.bucket.approval command create a random 
CreateBucketApproval request and send it to approver for debugging and testing
the approver on Dev Env.`,
}

var DebugCreateObjectApprovalCmd = &cli.Command{
	Action: createObjectApproval,
	Name:   "debug.create.object.approval",
	Usage:  "Create random CreateObjectApproval and send to approver for debugging and testing",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
	},
	Category: "DEBUG COMMANDS",
	Description: `The debug.create.object.approval command create a random 
CreateObjectApproval request and send it to approver for debugging and testing
the approver on Dev Env.`,
}

var DebugReplicateApprovalCmd = &cli.Command{
	Action: replicatePieceApprovalAction,
	Name:   "debug.replicate.approval",
	Usage:  "Create random ObjectInfo and send to p2p for debugging and testing p2p protocol network",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		numberFlag,
	},
	Category: "DEBUG COMMANDS",
	Description: `The debug.ask.replicate.approval command create a random 
ObjectInfo and send it to p2p node for debugging and testing p2p protocol 
network on Dev Env.`,
}

var DebugPutObjectCmd = &cli.Command{
	Action: putObjectAction,
	Name:   "debug.put.object",
	Usage:  "Create random ObjectInfo and send to uploader for debugging and testing uploading primary sp",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		fileFlag,
	},
	Category: "DEBUG COMMANDS",
	Description: `The debug.put.object command create a random ObjectInfo 
and send it to uploader for debugging and testing upload primary sp on Dev Env.`,
}

// createBucketApproval is the debug.create.bucket.approval command action.
func createBucketApproval(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)

	msg := &storagetypes.MsgCreateBucket{
		BucketName:        DebugCommandPrefix + util.GetRandomBucketName(),
		PrimarySpApproval: &common.Approval{},
	}
	task := &gfsptask.GfSpCreateBucketApprovalTask{}
	task.InitApprovalCreateBucketTask("cmd_debug", msg, coretask.UnSchedulingPriority)
	allow, res, err := client.AskCreateBucketApproval(context.Background(), task)
	if err != nil {
		return err
	}
	if !allow {
		return fmt.Errorf("refuse create bucket approval")
	}
	fmt.Printf("succeed to ask create bucket approval, BucketName: %s, ExpiredHeight: %d \n",
		res.GetCreateBucketInfo().GetBucketName(), res.GetCreateBucketInfo().GetPrimarySpApproval().GetExpiredHeight())
	return nil
}

// createBucketApproval is the debug.create.bucket.approval command action.
func createObjectApproval(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)

	msg := &storagetypes.MsgCreateObject{
		BucketName:        DebugCommandPrefix + util.GetRandomBucketName(),
		ObjectName:        DebugCommandPrefix + util.GetRandomObjectName(),
		PrimarySpApproval: &common.Approval{},
	}
	task := &gfsptask.GfSpCreateObjectApprovalTask{}
	task.InitApprovalCreateObjectTask("cmd_debug", msg, coretask.UnSchedulingPriority)
	allow, res, err := client.AskCreateObjectApproval(context.Background(), task)
	if err != nil {
		return err
	}
	if !allow {
		return fmt.Errorf("refuse create object approval")
	}
	fmt.Printf("succeed to ask create object approval, BucketName: %s, ObjectName: %s, ExpiredHeight: %d \n",
		res.GetCreateObjectInfo().GetBucketName(), res.GetCreateObjectInfo().GetBucketName(),
		res.GetCreateObjectInfo().GetPrimarySpApproval().GetExpiredHeight())
	return nil
}

// createBucketApproval is the debug.create.bucket.approval command action.
func replicatePieceApprovalAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)

	objectInfo := &storagetypes.ObjectInfo{
		Id:         sdk.NewUint(uint64(util.RandInt64(0, 100000))),
		BucketName: DebugCommandPrefix + util.GetRandomBucketName(),
		ObjectName: DebugCommandPrefix + util.GetRandomObjectName(),
	}
	task := &gfsptask.GfSpReplicatePieceApprovalTask{}
	task.InitApprovalReplicatePieceTask(objectInfo, &storagetypes.Params{},
		coretask.UnSchedulingPriority, GfSpCliUserName)

	expectNumber := ctx.Int(numberFlag.Name)
	approvals, err := client.AskSecondaryReplicatePieceApproval(
		context.Background(), task, expectNumber, expectNumber, 10)
	if err != nil {
		return err
	}
	fmt.Printf("receive %d accepted approvals\n", len(approvals))

	db, err := utils.MakeSPDB(cfg)
	if err != nil {
		return err
	}
	for _, approval := range approvals {
		spInfo, err := db.GetSpByAddress(approval.GetApprovedSpOperatorAddress(), spdb.OperatorAddressType)
		if err != nil {
			return err
		}
		fmt.Printf("%s[%s] accepted\n", approval.GetApprovedSpOperatorAddress(), spInfo.GetEndpoint())
	}
	return nil
}

func putObjectAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)
	filePath := ctx.String(fileFlag.Name)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if len(data) > 16*1024*1024 {
		return fmt.Errorf("debug upload data too big size [%d], limit[%d]", len(data), 16*1024*1024)
	}
	checksum := hash.GenerateChecksum(data)
	integrity := hash.GenerateIntegrityHash([][]byte{checksum})
	objectInfo := &storagetypes.ObjectInfo{
		Id:          sdk.NewUint(uint64(util.RandInt64(0, 100000))),
		BucketName:  DebugCommandPrefix + util.GetRandomBucketName(),
		ObjectName:  DebugCommandPrefix + filePath,
		PayloadSize: uint64(len(data)),
		Checksums:   [][]byte{integrity},
	}
	params := &storagetypes.Params{
		VersionedParams: storagetypes.VersionedParams{
			MaxSegmentSize:          16 * 1024 * 1024,
			RedundantDataChunkNum:   4,
			RedundantParityChunkNum: 2,
		},
	}
	stream := bytes.NewReader(data)
	task := &gfsptask.GfSpUploadObjectTask{}
	// TODO: refine it
	task.InitUploadObjectTask(0, objectInfo, params, 0)
	err = client.UploadObject(context.Background(), task, stream)
	if err != nil {
		return fmt.Errorf("failed to upload %s to uploader, error: %v", filePath, err)
	}
	fmt.Printf("succeed to upload %s, len %d to uploader\n", filePath, len(data))
	return nil
}
