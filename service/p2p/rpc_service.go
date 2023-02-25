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
	req.GetCreateObjectMsg().BucketName = "asdasd"
	req.GetCreateObjectMsg().ObjectName = "asdasd"
	var (
		bucket    = req.GetCreateObjectMsg().GetBucketName()
		object    = req.GetCreateObjectMsg().GetObjectName()
		routerKey = hash.HexStringHash(bucket, object)
		ackNum    int32
	)
	resp = &stypes.P2PServiceAskSecondaryApprovalResponse{}
	//defer func() {
	//	if err != nil {
	//		resp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//		log.CtxErrorw(ctx, "fail to ask approval", "bucket", bucket, "object", object, "error", err)
	//		return
	//	}
	//	if atomic.LoadInt32(&ackNum) < req.GetMinApprovalCount() {
	//		resp.ErrMessage = merrors.MakeErrMsgResponse(errors.New("ack approval less expect"))
	//		log.CtxErrorw(ctx, "ack approval less expect", "bucket", bucket, "object", object,
	//			"expected_min", req.GetMinApprovalCount(), "expected_max", req.GetMaxApprovalCount(), "ask_count", atomic.LoadInt32(&ackNum))
	//		return
	//	}
	//	log.CtxInfow(ctx, "success to ask approval", "bucket", bucket, "object", object)
	//}()

	err = service.addRouter(routerKey)
	if err != nil {
		return
	}

	//go func() {
	//	defer func() {
	//		service.deleteRouter(routerKey)
	//		close(syncCh)
	//	}()
	//	resCh, err := service.getRouterCh(routerKey)
	//	if err != nil {
	//		return
	//	}
	//	ticker := time.NewTicker(time.Duration(AskSecondaryApprovalTimeout) * time.Second)
	//	for {
	//		select {
	//		case envelope := <-resCh:
	//			switch msg := envelope.Message.(type) {
	//			case *ptypes.AskApprovalRequest:
	//				log.CtxInfow(ctx, "receive the ack approval", "from", envelope.From)
	//				resp.CreateObjectMsg = append(resp.CreateObjectMsg, msg.CreateObjectMsg)
	//				atomic.AddInt32(&ackNum, 1)
	//				if atomic.LoadInt32(&ackNum) >= req.GetMaxApprovalCount() {
	//					return
	//				}
	//			default:
	//				continue
	//			}
	//		case <-ticker.C:
	//			err = errors.New("ask approval timeout")
	//			return
	//		}
	//	}
	//}()

	go func() {
		service.publishInfo(&libs.Envelope{
			Message: &ptypes.AskApprovalRequest{
				CreateObjectMsg: req.CreateObjectMsg,
			},
			Broadcast: true,
		})
	}()

	resCh, err := service.getRouterCh(routerKey)
	if err != nil {
		return
	}
	ticker := time.NewTicker(time.Duration(AskSecondaryApprovalTimeout) * time.Second)
	for {
		select {
		case envelope := <-resCh:
			switch msg := envelope.Message.(type) {
			case *ptypes.AckApproval:
				log.CtxInfow(ctx, "receive the ack approval", "from", envelope.From)
				resp.CreateObjectMsg = append(resp.CreateObjectMsg, msg.CreateObjectMsg)
				ackNum++
				if ackNum >= req.GetMaxApprovalCount() {
					goto finish
				}
			default:
				continue
			}
		case <-ticker.C:
			err = errors.New("ask approval timeout")
			goto finish
		}
	}
finish:
	if err != nil {
		resp.ErrMessage = merrors.MakeErrMsgResponse(err)
	} else if ackNum < req.GetMinApprovalCount() {
		resp.ErrMessage = merrors.MakeErrMsgResponse(errors.New("ack approval less expect"))
	}
	service.deleteRouter(routerKey)
	log.CtxInfow(ctx, "finish to ask approval", "bucket", bucket, "object", object, "ack", ackNum, "error", err)
	return
}
