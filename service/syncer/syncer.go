package syncer

import (
	"context"
	"io"
	"net"

	"google.golang.org/grpc"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPieceAPI used to mock
type UploadECPieceAPI interface {
	UploadECPiece(stream service.SyncerService_UploadECPieceServer) error
}

type syncerConfig struct {
	Port                 string
	PieceStoreConfigFile string
}

// SyncerService synchronizes ec data to piece store
type SyncerService struct {
	cfg       syncerConfig
	grpcSever *grpc.Server
}

type Syncer struct {
}

// Name describes the name of SyncerService
func (s *SyncerService) Name() string {
	return "SyncerService"
}

// Start running SyncerService
func (s *SyncerService) Start(ctx context.Context) error {
	l, err := net.Listen("tcp", ":"+s.cfg.Port)
	if err != nil {
		log.Errorw("Start SyncerService net.Listen error", "error", err)
		return err
	}
	s.grpcSever = grpc.NewServer()
	service.RegisterSyncerServiceServer(s.grpcSever, &Syncer{})
	go func() {
		if err := s.grpcSever.Serve(l); err != nil {
			log.Errorw("gRPC server Serve error", "error", err)
		}
	}()
	log.Info("Start SyncerService successfully")
	return nil
}

// Stop SyncerService
func (s *SyncerService) Stop(ctx context.Context) error {
	return nil
}

// UploadECPiece uploads piece data encoded using the ec algorithm to secondary storage provider
func (s *Syncer) UploadECPiece(stream service.SyncerService_UploadECPieceServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Errorw("UploadECPiece receive data error", "traceID", req.GetTraceId(), "err", err)
			return err
		}
		if err == io.EOF {
			log.Info("UploadECPiece client closed")
			if err := stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				TraceId:         req.GetTraceId(),
				SecondarySpInfo: nil,
				ErrMessage: &service.ErrMessage{
					ErrCode: service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED,
					ErrMsg:  "Successful",
				},
			}); err != nil {
				log.Errorw("UploadECPiece SendAndClose error", "traceID", req.GetTraceId(), "err", err)
				return err
			}
		}
	}
}

func handleRequest(req *service.SyncerServiceUploadECPieceRequest) error {
	req.ProtoReflect()
	return nil
}
