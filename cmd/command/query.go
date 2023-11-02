package command

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

const (
	GfSpCliUserName = "gfsp-cli"
	queryCommands   = "QUERY COMMANDS"
)

var endpointFlag = &cli.StringFlag{
	Name:  "endpoint",
	Usage: "The address of machine that to query tasks",
	Value: "",
}

var keyFlag = &cli.StringFlag{
	Name:     "task.key",
	Usage:    "The sub key of task",
	Value:    "",
	Required: true,
}

var objectIDFlag = &cli.StringFlag{
	Name:     "object.id",
	Usage:    "The ID key of Object",
	Required: true,
}

var spIDFlag = &cli.StringFlag{
	Name:     "sp.id",
	Usage:    "The ID of a SP",
	Required: true,
}

var redundancyIdxFlag = &cli.Int64Flag{
	Name:     "redundancy.index",
	Usage:    "The object replicate index of SP",
	Required: true,
}

var segmentIdxFlag = &cli.Uint64Flag{
	Name:     "segment.index",
	Usage:    "The segment index",
	Required: true,
}

var ListModulesCmd = &cli.Command{
	Action:      listModulesAction,
	Name:        "list.modules",
	Usage:       "List the modules in greenfield storage provider",
	Category:    queryCommands,
	Description: `The list command output the services in greenfield storage provider.`,
}

var ListErrorsCmd = &cli.Command{
	Action:      listErrorsAction,
	Name:        "list.errors",
	Usage:       "List the predefine errors in greenfield storage provider",
	Category:    queryCommands,
	Description: `The list command output the services in greenfield storage provider.`,
}

var GetObjectCmd = &cli.Command{
	Action: CW.getObjectAction,
	Name:   "get.object",
	Usage:  "Get object payload data",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		objectIDFlag,
	},
	Category:    queryCommands,
	Description: `The get.object command send rpc request to downloader server to get object payload data.`,
}

var QueryTaskCmd = &cli.Command{
	Action:   CW.queryTasksAction,
	Name:     "query.task",
	Usage:    "Query running tasks in modules by task sub key",
	Category: queryCommands,
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		endpointFlag,
		keyFlag,
	},
	Description: `Query running tasks in modules by task sub key, show the tasks that task key contains the inout key detail info`,
}

var ChallengePieceCmd = &cli.Command{
	Action: CW.challengePieceAction,
	Name:   "challenge.piece",
	Usage:  "Challenge piece integrity hash",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		objectIDFlag,
		redundancyIdxFlag,
		segmentIdxFlag,
	},
	Category:    queryCommands,
	Description: `The challenge.piece command send rpc request to downloader, get integrity meta and check the piece checksums.`,
}

var GetSegmentIntegrityCmd = &cli.Command{
	Action: CW.getSegmentIntegrityAction,
	Name:   "get.piece.integrity",
	Usage:  "Get piece integrity hash and signature",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		objectIDFlag,
	},
	Category:    queryCommands,
	Description: `The get.segment.integrity command send rpc request to spdb, get integrity hash and signature.`,
}

var QueryBucketMigrateCmd = &cli.Command{
	Action: CW.getBucketMigrateAction,
	Name:   "query.bucket.migrate",
	Usage:  "Query bucket migrate plan and status",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		endpointFlag,
	},
	Category:    queryCommands,
	Description: `The query.bucket.migrate command send rpc request to manager get plan and status.`,
}

var QuerySPExitCmd = &cli.Command{
	Action: CW.getSPExitAction,
	Name:   "query.sp.exit",
	Usage:  "Query sp exit swap plan and migrate gvg task status",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		endpointFlag,
	},
	Category:    queryCommands,
	Description: `The query.sp.exit command send rpc request to manager, get sp exit swap plan and migrate gvg task status.`,
}

var QueryPrimarySPIncomeCmd = &cli.Command{
	Action: CW.getPrimarySPIncomeAction,
	Name:   "query.primary.sp.income",
	Usage:  "Query primary sp income details",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		spIDFlag,
	},
	Category:    queryCommands,
	Description: `The query.primary.sp.income command send rpc request to metadata, get the primary sp incomes details for the current timestamp`,
}

var QuerySecondarySPIncomeCmd = &cli.Command{
	Action: CW.getSecondarySPIncomeAction,
	Name:   "query.secondary.sp.income",
	Usage:  "Query secondary sp income details",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		spIDFlag,
	},
	Category:    queryCommands,
	Description: `The query.secondary.sp.income command send rpc request to metadata, get the secondary sp incomes details for the current timestamp`,
}

func listModulesAction(ctx *cli.Context) error {
	fmt.Println(gfspapp.GetRegisterModuleDescription())
	return nil
}

func listErrorsAction(ctx *cli.Context) error {
	gfspErrors := gfsperrors.GfSpErrorList()
	for _, gfspError := range gfspErrors {
		fmt.Println(gfspError.String())
	}
	return nil
}

func (w *CMDWrapper) queryTasksAction(ctx *cli.Context) error {
	w.initEmptyGRPCAPI()
	endpoint := gfspapp.DefaultGRPCAddress
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg := &gfspconfig.GfSpConfig{}
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return err
		}
		endpoint = cfg.GRPCAddress
	}
	if ctx.IsSet(endpointFlag.Name) {
		endpoint = ctx.String(endpointFlag.Name)
	}
	key := ctx.String(keyFlag.Name)
	infos, err := w.grpcAPI.QueryTasks(context.Background(), endpoint, key)
	if err != nil {
		fmt.Printf("failed to query task, endpoint:%v, key:%v, error:%v\n", endpoint, key, err)
		return err
	}
	if len(infos) == 0 {
		fmt.Printf("failed to query task due to no task, endpoint:%v, key:%v\n", endpoint, key)
	}
	for _, info := range infos {
		fmt.Printf(info + "\n")
	}
	return nil
}

func (w *CMDWrapper) getObjectAction(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	if err = w.initChainAPI(ctx); err != nil {
		return err
	}

	objectID := ctx.String(objectIDFlag.Name)
	objectInfo, err := w.chainAPI.QueryObjectInfoByID(context.Background(), objectID)
	if err != nil {
		return fmt.Errorf("failed to query object info, error: %v", err)
	}
	bucketInfo, err := w.chainAPI.QueryBucketInfo(context.Background(), objectInfo.GetBucketName())
	if err != nil {
		return fmt.Errorf("failed to query bucket info, error: %v", err)
	}
	params, err := w.chainAPI.QueryStorageParamsByTimestamp(context.Background(), objectInfo.GetCreateAt())
	if err != nil {
		return fmt.Errorf("failed to query storage params, error: %v", err)
	}
	task := &gfsptask.GfSpDownloadObjectTask{}
	task.InitDownloadObjectTask(objectInfo, bucketInfo, params, coretask.UnSchedulingPriority,
		GfSpCliUserName, 0, int64(objectInfo.GetPayloadSize()-1), 0, 0)
	data, err := w.grpcAPI.GetObject(context.Background(), task)
	if err != nil {
		return fmt.Errorf("failed to get object, error: %v", err)
	}
	if err = os.WriteFile("./"+objectInfo.GetObjectName(), data, os.ModePerm); err != nil {
		fmt.Printf("failed to create file to wirte object payload data, error: %v", err)
	}
	fmt.Printf("succeed to get object\n\n"+
		"BucketInfo: %s\n\n "+
		"ObjectInfo: %s \n\n"+
		"StorageParam: %s\n\n",
		bucketInfo.String(), objectInfo.String(), params.String())
	return nil
}

func (w *CMDWrapper) challengePieceAction(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	if err = w.initChainAPI(ctx); err != nil {
		return err
	}

	objectID := ctx.String(objectIDFlag.Name)
	objectInfo, err := w.chainAPI.QueryObjectInfoByID(context.Background(), objectID)
	if err != nil {
		return fmt.Errorf("failed to query object info, error: %v", err)
	}
	bucketInfo, err := w.chainAPI.QueryBucketInfo(context.Background(), objectInfo.GetBucketName())
	if err != nil {
		return fmt.Errorf("failed to query bucket info, error: %v", err)
	}
	params, err := w.chainAPI.QueryStorageParamsByTimestamp(context.Background(), objectInfo.GetCreateAt())
	if err != nil {
		return fmt.Errorf("failed to query storage params, error: %v", err)
	}

	redundancyIdx := ctx.Int64(redundancyIdxFlag.Name)
	segmentIdx := ctx.Uint64(segmentIdxFlag.Name)

	task := &gfsptask.GfSpChallengePieceTask{}
	task.InitChallengePieceTask(objectInfo, bucketInfo, params, coretask.UnSchedulingPriority,
		GfSpCliUserName, int32(redundancyIdx), uint32(segmentIdx), 0, 0)
	integrityHash, checksums, data, err := w.grpcAPI.GetChallengeInfo(context.Background(), task)
	if err != nil {
		return fmt.Errorf("failed to get challenge info, error: %v", err)
	}
	fmt.Printf("integrity meta info: \n\n")
	fmt.Printf("integrity hash[%s]\n\n", hex.EncodeToString(integrityHash))
	for i, checksum := range checksums {
		fmt.Printf("piece[%d] checksum[%s]\n", i, hex.EncodeToString(checksum))
	}
	challengePieceChecksum := hash.GenerateChecksum(data)
	fmt.Printf("\nchallenge piece info\n: redundancy_idx[%d], segment_idx[%d], piece_checksum[%s]\n\n",
		redundancyIdx, segmentIdx, hex.EncodeToString(challengePieceChecksum))

	if !bytes.Equal(challengePieceChecksum, checksums[segmentIdx]) {
		return fmt.Errorf("piece data hash[%s] not equal to checksum list value[%s]",
			hex.EncodeToString(challengePieceChecksum), hex.EncodeToString(checksums[segmentIdx]))
	}

	if !bytes.Equal(integrityHash, hash.GenerateIntegrityHash(checksums)) {
		return fmt.Errorf("integrity hash[%s] mismatch checksum list result[%s]",
			hex.EncodeToString(integrityHash),
			hex.EncodeToString(hash.GenerateIntegrityHash(checksums)))
	}
	fmt.Printf("succeed to check integrity hash!!!\n")
	return nil
}

func (w *CMDWrapper) getSegmentIntegrityAction(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	if err = w.initChainAPI(ctx); err != nil {
		return err
	}

	objectID, err := util.StringToUint64(ctx.String(objectIDFlag.Name))
	if err != nil {
		return fmt.Errorf("invalid object id, it should be an unsigned integer")
	}
	objectInfo, err := w.grpcAPI.GetObjectByID(ctx.Context, objectID)
	if err != nil {
		return fmt.Errorf("failed to query object info, error: %v", err)
	}
	bucket, err := w.grpcAPI.GetBucketByBucketName(ctx.Context, objectInfo.GetBucketName(), true)
	if err != nil {
		return fmt.Errorf("failed to query bucket info by bucket name, error: %v", err)
	}
	gvg, err := w.grpcAPI.GetGlobalVirtualGroup(ctx.Context, bucket.BucketInfo.Id.Uint64(), objectInfo.GetLocalVirtualGroupId())
	if err != nil {
		return fmt.Errorf("failed to query global virtual group, error: %v\n", err)
	}

	spInfo, err := w.chainAPI.QuerySP(ctx.Context, w.config.SpAccount.SpOperatorAddress)
	if err != nil {
		return fmt.Errorf("failed to query sp info, error: %v\n", err)
	}
	selfSpID := spInfo.GetId()
	var redundancyIdx int
	// -1 represents this is a primarySP; secondarySP should firstly query its order in secondary sp list
	if selfSpID == gvg.GetPrimarySpId() {
		fmt.Printf("%d belongs to primary sp\n", objectID)
		redundancyIdx = piecestore.PrimarySPRedundancyIndex
	} else {
		for index, spID := range gvg.GetSecondarySpIds() {
			if selfSpID == spID {
				redundancyIdx = index
				break
			}
			if redundancyIdx == 0 {
				fmt.Printf("%d doesn't belong to this secondary sp list\n", objectID)
			}
		}
	}

	integrity, err := w.spDBAPI.GetObjectIntegrity(objectID, int32(redundancyIdx))
	if err != nil {
		return fmt.Errorf("failed to get object integrity by object id, error: %v", err)
	}
	fmt.Printf("succeed to get segment integrity:\n\nredundancyIdx[%d], integrity_hash[%s]\n\n",
		redundancyIdx, hex.EncodeToString(integrity.IntegrityChecksum))
	for i, checksum := range integrity.PieceChecksumList {
		fmt.Printf("piece[%d], checksum[%s]\n", i, hex.EncodeToString(checksum))
	}
	return nil
}

func (w *CMDWrapper) getBucketMigrateAction(ctx *cli.Context) error {
	w.initEmptyGRPCAPI()
	endpoint := gfspapp.DefaultGRPCAddress
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg := &gfspconfig.GfSpConfig{}
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return err
		}
		endpoint = cfg.GRPCAddress
	}
	if ctx.IsSet(endpointFlag.Name) {
		endpoint = ctx.String(endpointFlag.Name)
	}
	info, err := w.grpcAPI.QueryBucketMigrate(context.Background(), endpoint)
	if err != nil {
		return err
	}
	fmt.Println(info)
	return nil
}

func (w *CMDWrapper) getSPExitAction(ctx *cli.Context) error {
	w.initEmptyGRPCAPI()
	endpoint := gfspapp.DefaultGRPCAddress
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg := &gfspconfig.GfSpConfig{}
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return err
		}
		endpoint = cfg.GRPCAddress
	}
	if ctx.IsSet(endpointFlag.Name) {
		endpoint = ctx.String(endpointFlag.Name)
	}
	info, err := w.grpcAPI.QuerySPExit(context.Background(), endpoint)
	if err != nil {
		return err
	}
	fmt.Println(info)
	return nil
}

func (w *CMDWrapper) getPrimarySPIncomeAction(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg := &gfspconfig.GfSpConfig{}
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return err
		}
	}
	spID := ctx.Int(spIDFlag.Name)

	currentTimestamp, info, err := w.grpcAPI.PrimarySpIncomeDetails(context.Background(), uint32(spID))
	if err != nil {
		return err
	}
	details, _ := json.Marshal(info)

	fmt.Println("querying primary sp income details for sp ", spID)
	fmt.Println("query timestamp", currentTimestamp, time.Unix(currentTimestamp, 0))
	fmt.Println("query results:", string(details[:]))
	return nil
}

func (w *CMDWrapper) getSecondarySPIncomeAction(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg := &gfspconfig.GfSpConfig{}
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return err
		}
	}
	spID := ctx.Int(spIDFlag.Name)

	currentTimestamp, info, err := w.grpcAPI.SecondarySpIncomeDetails(context.Background(), uint32(spID))
	if err != nil {
		return err
	}
	details, _ := json.Marshal(info)

	if err != nil {
		return err
	}
	fmt.Println("querying secondary sp income details for sp ", spID)
	fmt.Println("query timestamp", currentTimestamp, time.Unix(currentTimestamp, 0))
	fmt.Println("query results:", string(details[:]))
	return nil
}
