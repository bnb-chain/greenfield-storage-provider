package command

import (
	"context"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/urfave/cli/v2"
)

var bucketFlag = &cli.StringFlag{
	Name:     "b",
	Usage:    "The bucket name",
	Required: true,
}

var objectFlag = &cli.StringFlag{
	Name:     "o",
	Usage:    "The object name",
	Required: true,
}

var RecoverObjectCmd = &cli.Command{
	Action: recoverObjectAction,
	Name:   "recover.object",
	Usage:  "Generate recover piece data task to recover the object data",

	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		bucketFlag,
		objectFlag,
	},

	Category: "RECOVERY COMMANDS",
	Description: `The recover.object command is used to recover the object piece data on the primarySP or the secondary SP", 
  `,
}

func recoverObjectAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)

	objectName := ctx.String(objectFlag.Name)
	bucketName := ctx.String(bucketFlag.Name)

	chain, err := utils.MakeGnfd(cfg)
	if err != nil {
		return err
	}

	objectInfo, err := chain.QueryObjectInfo(context.Background(), bucketName, objectName)
	if err != nil {
		return err
	}

	// TODO: refine it
	//bucketInfo, err := chain.QueryBucketInfo(context.Background(), bucketName)
	//if err != nil {
	//	return err
	//}

	storageParams, err := chain.QueryStorageParams(context.Background())
	if err != nil {
		return err
	}

	replicateIdx := 0

	// TODO: refine it
	isSecondarySP := true
	//if strings.EqualFold(bucketInfo.PrimarySpAddress, cfg.SpAccount.SpOperatorAddress) {
	//	replicateIdx = -1
	//}
	//
	//var isSecondarySP bool
	//for i, addr := range objectInfo.GetSecondarySpAddresses() {
	//	if strings.EqualFold(addr, cfg.SpAccount.SpOperatorAddress) {
	//		replicateIdx = i
	//		isSecondarySP = true
	//		break
	//	}
	//}
	if replicateIdx != -1 && !isSecondarySP {
		return fmt.Errorf(" it is not primary SP nor secondarySP of the object, pls choose the right SP")
	}

	maxSegmentSize := storageParams.GetMaxSegmentSize()
	segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)
	// recovery primary SP
	if replicateIdx == -1 {
		fmt.Printf("begin to recovery the primary SP object: %s \n", objectName)
		for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
			task := &gfsptask.GfSpRecoverPieceTask{}
			task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, int32(-1), maxSegmentSize, 0, 2)
			client.ReportTask(context.Background(), task)
			time.Sleep(time.Second)
		}
	} else {
		// recovery secondary SP
		fmt.Printf("begin to recovery the secondary SP object: %s \n", objectName)
		for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
			task := &gfsptask.GfSpRecoverPieceTask{}
			task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, int32(replicateIdx), maxSegmentSize, 0, 2)
			client.ReportTask(context.Background(), task)
			time.Sleep(time.Second)
		}
	}
	// TODO support query recovery task status command
	fmt.Printf("succeed to gerate recovery object %s task on background \n", objectName)
	return nil
}

func segmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}
