package p2p

import (
	"context"
	"errors"
	"time"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var AskSecondaryApprovalTimeout = 10
var _ stypes.P2PServiceServer = &P2PService{}

func (service *P2PService) AskSecondaryApproval(
	ctx context.Context,
	req *stypes.P2PServiceAskSecondaryApprovalRequest) (
	resp *stypes.P2PServiceAskSecondaryApprovalResponse,
	err error) {
	ctx = log.Context(ctx)
	var (
		bucket    = req.GetCreateObjectMsg().GetBucketName()
		object    = req.GetCreateObjectMsg().GetObjectName()
		routerKey = hash.HexStringHash(bucket, object)
		syncCh    = make(chan struct{})
		ackCount  int32
	)
	resp = &stypes.P2PServiceAskSecondaryApprovalResponse{}
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "fail to ask approval", "bucket", bucket, "object", object, "error", err)
			return
		}
		if ackCount < req.GetMinApprovalCount() {
			resp.ErrMessage = merrors.MakeErrMsgResponse(errors.New("ack approval less expect"))
			log.CtxErrorw(ctx, "ack approval less expect", "bucket", bucket, "object", object, "error", err)
			return
		}
		log.CtxInfow(ctx, "success to ask approval", "bucket", bucket, "object", object)
	}()

	err = service.addRouter(routerKey)
	if err != nil {
		return
	}

	resCh, err := service.getRouterCh(routerKey)
	if err != nil {
		return
	}

	go func() {
		defer func() {
			service.deleteRouter(routerKey)
			syncCh <- struct{}{}
		}()

		ticker := time.NewTicker(time.Duration(AskSecondaryApprovalTimeout) * time.Second)
		for {
			select {
			case envelope := <-resCh:
				switch msg := envelope.Message.(type) {
				case *ptypes.AskApprovalRequest:
					resp.CreateObjectMsg = append(resp.CreateObjectMsg, msg.CreateObjectMsg)
					ackCount++
					if ackCount >= req.GetMaxApprovalCount() {
						return
					}
				default:
					continue
				}
			case <-ticker.C:
				err = errors.New("ask approval timeout")
				return
			}
		}
	}()
	service.publishInfo(&libs.Envelope{
		Message: &ptypes.AskApprovalRequest{
			CreateObjectMsg: req.CreateObjectMsg,
		},
		Broadcast: true,
	})
	<-syncCh
	return
}
