package uploader

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// stoneHubClientConfig is the stoneHubClient config.
type stoneHubClientConfig struct {
	Address    string
	TimeoutSec uint32
}

// stoneHubClient wrapper StoneHubService, respond to maintain job meta/status,
// also schedule background task execution.
type stoneHubClient struct {
	config          *stoneHubClientConfig
	storageProvider string
}

func newStoneHubClient(config *stoneHubClientConfig, sp string) *stoneHubClient {
	return &stoneHubClient{config: config, storageProvider: sp}
}

// registerJobMeta store job meta in stoneHub database, record Job execute status.
func (shc *stoneHubClient) registerJobMeta(ctx context.Context, req *pbService.UploaderServiceCreateObjectRequest, txHash []byte) error {
	ctx, _ = context.WithTimeout(ctx, time.Duration(shc.config.TimeoutSec)*time.Second)
	conn, err := grpc.DialContext(ctx, shc.config.Address, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dial", "err", err)
		return err
	}
	defer conn.Close()
	c := pbService.NewStoneHubServiceClient(conn)
	resp, err := c.CreateObject(ctx, &pbService.StoneHubServiceCreateObjectRequest{
		TxHash:     txHash,
		ObjectInfo: req.ObjectInfo,
	})
	if err != nil {
		log.Warnw("failed to send grpc", "err", err)
		return err
	}
	if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Warnw("failed to grpc", "err", resp.ErrMessage)
		return fmt.Errorf(resp.ErrMessage.ErrMsg)
	}
	log.Infow("succeed to send grpc", "response", resp)
	return nil
}

// updateJobMeta is used to set JobMeta height.
func (shc *stoneHubClient) updateJobMeta(txHash []byte, height uint64, objectID uint64) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(shc.config.TimeoutSec)*time.Second)
	conn, err := grpc.DialContext(ctx, shc.config.Address, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dial", "err", err)
		return err
	}
	defer conn.Close()
	c := pbService.NewStoneHubServiceClient(conn)
	resp, err := c.SetObjectCreateInfo(ctx, &pbService.StoneHubServiceSetObjectCreateInfoRequest{
		TxHash:   txHash,
		TxHeight: height,
		ObjectId: objectID,
	})
	if err != nil {
		log.Warnw("failed to send grpc", "err", err)
		return err
	}
	if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Warnw("failed to grpc", "err", resp.ErrMessage)
		return fmt.Errorf(resp.ErrMessage.ErrMsg)
	}
	log.Infow("succeed to send grpc", "response", resp)
	return nil
}

// JobMeta is Job Context, got from stone hub.
type JobMeta struct {
	objectID      uint64
	toUploadedIDs map[uint32]bool
	txHash        []byte
	pieceJob      *pbService.PieceJob
}

// fetchJobMeta fetch job meta from stone hub.
func (shc *stoneHubClient) fetchJobMeta(txHash []byte) (*JobMeta, error) {
	var (
		jm  JobMeta
		err error
	)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(shc.config.TimeoutSec)*time.Second)
	conn, err := grpc.DialContext(ctx, shc.config.Address, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dial", "err", err)
		return nil, err
	}
	defer conn.Close()
	c := pbService.NewStoneHubServiceClient(conn)
	resp, err := c.BeginUploadPayload(ctx, &pbService.StoneHubServiceBeginUploadPayloadRequest{
		TxHash: txHash,
	})
	if err != nil {
		log.Warnw("failed to send grpc", "err", err)
		return nil, err
	}
	if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Warnw("failed to grpc", "err", resp.ErrMessage)
		return nil, fmt.Errorf(resp.ErrMessage.ErrMsg)
	}
	// todo: fill jm
	jm.pieceJob = resp.PieceJob
	jm.txHash = txHash
	jm.objectID = resp.PieceJob.ObjectId
	jm.toUploadedIDs = make(map[uint32]bool)
	if idxs := resp.PieceJob.GetTargetIdx(); idxs != nil {
		for idx := range idxs {
			jm.toUploadedIDs[uint32(idx)] = true
		}
	}
	log.Infow("succeed to send grpc", "response", resp)
	return &jm, err

}

// reportJobProgress report done piece index to stone hub.
func (shc *stoneHubClient) reportJobProgress(jm *JobMeta, uploadID uint32, checkSum []byte) error {
	var (
		req        *pbService.StoneHubServiceDonePrimaryPieceJobRequest
		spSealInfo *pbService.StorageProviderSealInfo
	)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(shc.config.TimeoutSec)*time.Second)
	conn, err := grpc.DialContext(ctx, shc.config.Address, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dial", "err", err)
		return err
	}
	defer conn.Close()
	c := pbService.NewStoneHubServiceClient(conn)
	req = &pbService.StoneHubServiceDonePrimaryPieceJobRequest{
		TxHash:   jm.txHash,
		PieceJob: jm.pieceJob,
	}
	spSealInfo = &pbService.StorageProviderSealInfo{
		StorageProviderId: shc.storageProvider,
		PieceIdx:          uploadID,
		PieceChecksum:     [][]byte{checkSum},
	}
	req.PieceJob.StorageProviderSealInfo = spSealInfo
	resp, err := c.DonePrimaryPieceJob(ctx, req)
	if err != nil {
		log.Warnw("failed to send grpc", "err", err)
		return err
	}
	if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Warnw("failed to grpc", "err", resp.ErrMessage)
		return fmt.Errorf(resp.ErrMessage.ErrMsg)
	}
	log.Infow("succeed to send grpc", "response", resp)
	return err
}
