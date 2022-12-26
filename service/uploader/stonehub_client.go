package uploader

import (
	"context"
	"time"

	"google.golang.org/grpc"

	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// stoneHubClient wrapper StoneHubService, respond to maintain job meta/status,
// also schedule background task execution.
// todo: cache grpc client
type stoneHubClient struct {
	grpcAddr    string
	grpcTimeout time.Duration
}

func newStoneHubClient() *stoneHubClient {
	// mock return
	// todo: stoneHub config
	return &stoneHubClient{"127.0.0.1:9092", 5 * time.Second}
}

// registerJobMeta store job meta in stoneHub database, record Job execute status.
func (shc *stoneHubClient) registerJobMeta(ctx context.Context, req *pbService.UploaderServiceCreateObjectRequest, txHash []byte) error {
	// mock return
	return nil

	// todo: debug and polish(CreateObject)
	ctx, _ = context.WithTimeout(ctx, shc.grpcTimeout)
	conn, err := grpc.DialContext(ctx, shc.grpcAddr, grpc.WithInsecure())
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
	log.Infow("succeed to send grpc", "response", resp)
	return nil
}

// updateJobMeta is used to set JobMeta height.
func (shc *stoneHubClient) updateJobMeta(txHash []byte, height uint64) error {
	// mock return
	return nil

	// todo: debug and polish(SetObjectCreateHeight)
	ctx, _ := context.WithTimeout(context.Background(), shc.grpcTimeout)
	conn, err := grpc.DialContext(ctx, shc.grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dial", "err", err)
		return err
	}
	defer conn.Close()
	c := pbService.NewStoneHubServiceClient(conn)
	resp, err := c.SetObjectCreateHeight(ctx, &pbService.StoneHubServiceSetObjectCreateHeightRequest{
		TxHash:   txHash,
		TxHeight: height,
	})
	if err != nil {
		log.Warnw("failed to send grpc", "err", err)
		return err
	}
	log.Infow("succeed to send grpc", "response", resp)
	return nil
}

type JobMeta struct {
	bucket      string
	object      string
	payloadSize uint32
	uploadedIDs map[uint32]bool
	txHash      []byte
	pieceJob    *pbService.PieceJob
}

func (shc *stoneHubClient) fetchJobMeta(txHash []byte) (*JobMeta, error) {
	// mock return
	var (
		jm  JobMeta
		err error
	)
	jm.bucket = "test_bucket"
	jm.object = "test_object"
	jm.uploadedIDs = make(map[uint32]bool)
	return &jm, err

	// todo: debug and polish(BeginUploadPayload)
	ctx, _ := context.WithTimeout(context.Background(), shc.grpcTimeout)
	conn, err := grpc.DialContext(ctx, shc.grpcAddr, grpc.WithInsecure())
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
	// todo: fill jm
	jm.pieceJob = resp.PieceJob
	jm.txHash = txHash
	log.Infow("succeed to send grpc", "response", resp)
	return &jm, err

}

func (shc *stoneHubClient) reportJobProgress(jm *JobMeta, uploadID uint32) error {
	// mock return
	return nil

	// todo: debug and polish(DonePrimaryPieceJob)
	ctx, _ := context.WithTimeout(context.Background(), shc.grpcTimeout)
	conn, err := grpc.DialContext(ctx, shc.grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dial", "err", err)
		return err
	}
	defer conn.Close()
	c := pbService.NewStoneHubServiceClient(conn)
	// todo: fill piece job
	resp, err := c.DonePrimaryPieceJob(ctx, &pbService.StoneHubServiceDonePrimaryPieceJobRequest{
		TxHash:   jm.txHash,
		PieceJob: jm.pieceJob,
	})
	if err != nil {
		log.Warnw("failed to send grpc", "err", err)
		return err
	}
	log.Infow("succeed to send grpc", "response", resp)
	return err
}
