package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	maxRecoveryRetry    = 3
	MaxRecoveryTime     = 50
	recoveryCommands    = "RECOVERY COMMANDS"
	MaxRecoveryJob      = 1000
	FullQueueWaitTime   = 5 * time.Second
	RecoveryJobSyncTime = 30 * time.Second
)

var bucketFlag = &cli.StringFlag{
	Name:     "bucket",
	Usage:    "The bucket name",
	Aliases:  []string{"b"},
	Required: true,
}

var objectFlag = &cli.StringFlag{
	Name:     "object",
	Usage:    "The object name. If there is space inside name, use double quote around the name (except when used in k8s job config). If you use --objectList flag, --object flag will be ignored. ",
	Aliases:  []string{"o"},
	Required: false,
}

var objectListFlag = &cli.StringFlag{
	Name:     "objectList",
	Usage:    "The object name list, shall be in the format of object1//object2//object3. If there is space inside name, use double quote around the name (except when used in k8s job config).",
	Aliases:  []string{"l"},
	Required: false,
}

var RecoverObjectCmd = &cli.Command{
	Action: recoverObjectAction,
	Name:   "recover.object",
	Usage:  "Generate recover piece data tasks to recover the object data",

	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		bucketFlag,
		objectFlag,
		objectListFlag,
	},
	Category:    recoveryCommands,
	Description: `The recover.object command is used to recover the object data on the primarySP or the secondary SP.`,
}

var RecoverPieceCmd = &cli.Command{
	Action: recoverPieceAction,
	Name:   "recover.piece",
	Usage:  "Generate recover piece data task to recover the object piece",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		bucketFlag,
		objectFlag,
		segmentIdxFlag,
	},
	Category:    recoveryCommands,
	Description: `The recover.piece command is used to recover the object piece data on the primarySP or the secondary SP.`,
}

func recoverObjectAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)
	bucketName := ctx.String(bucketFlag.Name)

	if !ctx.IsSet(objectFlag.Name) && !ctx.IsSet(objectListFlag.Name) {
		return fmt.Errorf("either object flag or objectList flag has to be set for object(s) recovery cmd")
	}

	if ctx.IsSet(objectListFlag.Name) {
		objectNames := ctx.String(objectListFlag.Name)
		objectNameList := strings.Split(objectNames, "//")
		for _, objectName := range objectNameList {
			recoverObject(bucketName, objectName, cfg, client)
		}
	} else {
		objectName := ctx.String(objectFlag.Name)
		recoverObject(bucketName, objectName, cfg, client)
	}

	taskCheckTicker := time.NewTicker(RecoveryJobSyncTime)

	for range taskCheckTicker.C {
		stats, _ := client.GetTasksStats(context.Background())
		if stats.GetRecoveryProcessCount() > 0 {
			fmt.Printf("Processing recovery tasks, update every 30 seconds. \n Waiting Tasks Number: %d \n Failed: %s\n\n", stats.GetRecoveryProcessCount(), stats.GetRecoveryFailedList())
		} else {
			fmt.Printf("Finished process all recovery objects task on background. \n Failed: %s \n", stats.GetRecoveryFailedList())
			break
		}
	}

	return nil
}

func recoverObject(bucketName string, objectName string, cfg *gfspconfig.GfSpConfig, client *gfspclient.GfSpClient) error {
	var (
		replicateIdx    int
		processingCount uint32
	)

	stats, _ := client.GetTasksStats(context.Background())
	processingCount = stats.GetRecoveryProcessCount()
	for processingCount >= MaxRecoveryJob {
		log.Infof("Recovery object job waiting as there are already %d recovery job in queue.", MaxRecoveryJob)
		time.Sleep(FullQueueWaitTime)
		stats, _ := client.GetTasksStats(context.Background())
		processingCount = stats.GetRecoveryProcessCount()
	}

	bucketInfo, objectInfo, storageParams, err := getChainInfo(bucketName, objectName, cfg)
	if err != nil {
		return err
	}

	replicateIdx, err = getReplicateIdxBySP(bucketInfo, objectInfo, cfg)
	if err != nil {
		return err
	}

	maxSegmentSize := storageParams.GetMaxSegmentSize()
	segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)

	// recovery primary SP
	if replicateIdx == -1 {
		fmt.Printf("begin to recover the object of the primary SP: %s \n", objectName)
		for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
			task := &gfsptask.GfSpRecoverPieceTask{}
			task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, int32(-1), maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
			err = client.ReportTask(context.Background(), task)
			time.Sleep(time.Second)
		}
	} else {
		// recovery secondary SP
		fmt.Printf("begin to recovery the object of the secondary SP: %s \n", objectName)
		for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
			task := &gfsptask.GfSpRecoverPieceTask{}
			task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, int32(replicateIdx), maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
			err = client.ReportTask(context.Background(), task)
			time.Sleep(time.Second)
		}
	}

	if err != nil {
		fmt.Printf("failed to generate recovery object %s task on background \n, error: %s", objectName, err.Error())
	} else {
		fmt.Printf("succeed to generate recovery object %s task on background \n", objectName)
	}
	return err
}

func recoverPieceAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)

	objectName := ctx.String(objectFlag.Name)
	bucketName := ctx.String(bucketFlag.Name)
	segmentIdx := ctx.Uint64(segmentIdxFlag.Name)
	var replicateIdx int

	bucketInfo, objectInfo, storageParams, err := getChainInfo(bucketName, objectName, cfg)
	if err != nil {
		return err
	}

	replicateIdx, err = getReplicateIdxBySP(bucketInfo, objectInfo, cfg)
	if err != nil {
		return err
	}

	maxSegmentSize := storageParams.GetMaxSegmentSize()
	segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)
	if uint32(segmentIdx) > segmentCount {
		return fmt.Errorf("invaild segment id %d, segment should be no more than %d \n", segmentIdx, segmentCount)
	}

	if replicateIdx == -1 {
		fmt.Printf("begin to recover piece data of the primary SP, object name %s, segmentIdx %d , \n", objectName, segmentIdx)
		task := &gfsptask.GfSpRecoverPieceTask{}
		task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, uint32(segmentIdx), int32(-1), maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
		client.ReportTask(context.Background(), task)
	} else {
		// recovery secondary SP
		fmt.Printf("begin to recover piece data of the secondary SP, object name %s, segmentIdx %d \n", objectName, segmentIdx)
		task := &gfsptask.GfSpRecoverPieceTask{}
		task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, uint32(segmentIdx), int32(replicateIdx), maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
		client.ReportTask(context.Background(), task)
	}

	fmt.Printf("succeed to gerate recover piece of object  %s task on background \n", objectName)
	return nil
}

func getReplicateIdxBySP(bucketInfo *types.BucketInfo, objectInfo *types.ObjectInfo, cfg *gfspconfig.GfSpConfig) (int, error) {
	replicateIdx := 0
	var isSecondarySp bool
	chain, err := utils.MakeGnfd(cfg)
	if err != nil {
		return 0, err
	}
	spClient := utils.MakeGfSpClient(cfg)
	sp, err := chain.QuerySP(context.Background(), cfg.SpAccount.SpOperatorAddress)
	if err != nil {
		return 0, err
	}

	bucketSPID, err := util.GetBucketPrimarySPID(context.Background(), chain, bucketInfo)
	if err != nil {
		return 0, err
	}

	// it is a primary sp
	if bucketSPID == sp.Id {
		replicateIdx = -1
		return replicateIdx, nil
	}

	replicateIdx, isSecondarySp, err = util.ValidateAndGetSPIndexWithinGVGSecondarySPs(context.Background(), spClient, sp.Id, bucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
	if err != nil {
		return 0, err
	}
	if !isSecondarySp {
		return 0, fmt.Errorf(" it is not primary SP nor secondarySP of the object, pls choose the right SP")
	}
	return replicateIdx, nil
}

func getChainInfo(bucketName, objectName string, cfg *gfspconfig.GfSpConfig) (*types.BucketInfo, *types.ObjectInfo, *types.Params, error) {
	chain, err := utils.MakeGnfd(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	objectInfo, err := chain.QueryObjectInfo(context.Background(), bucketName, objectName)
	if err != nil {
		return nil, nil, nil, err
	}

	bucketInfo, err := chain.QueryBucketInfo(context.Background(), bucketName)
	if err != nil {
		return nil, nil, nil, err
	}

	storageParams, err := chain.QueryStorageParams(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}

	return bucketInfo, objectInfo, storageParams, nil
}

func segmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}
