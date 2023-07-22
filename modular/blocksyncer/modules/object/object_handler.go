package object

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"
	"github.com/forbole/juno/v4/models"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	EventCreateObject       = proto.MessageName(&storagetypes.EventCreateObject{})
	EventCancelCreateObject = proto.MessageName(&storagetypes.EventCancelCreateObject{})
	EventSealObject         = proto.MessageName(&storagetypes.EventSealObject{})
	EventCopyObject         = proto.MessageName(&storagetypes.EventCopyObject{})
	EventDeleteObject       = proto.MessageName(&storagetypes.EventDeleteObject{})
	EventRejectSealObject   = proto.MessageName(&storagetypes.EventRejectSealObject{})
	EventDiscontinueObject  = proto.MessageName(&storagetypes.EventDiscontinueObject{})
	EventUpdateObjectInfo   = proto.MessageName(&storagetypes.EventUpdateObjectInfo{})
)

var ObjectEvents = map[string]bool{
	EventCreateObject:       true,
	EventCancelCreateObject: true,
	EventSealObject:         true,
	EventCopyObject:         true,
	EventDeleteObject:       true,
	EventRejectSealObject:   true,
	EventDiscontinueObject:  true,
	EventUpdateObjectInfo:   true,
}

func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	_, err := m.Handle(ctx, block, txHash, event, true)
	return err
}

func (m *Module) ExtractEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (interface{}, error) {
	return m.Handle(ctx, block, txHash, event, false)
}

func (m *Module) Handle(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event, OperationDB bool) (interface{}, error) {
	if !ObjectEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventCreateObject:
		createObject, ok := typedEvent.(*storagetypes.EventCreateObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCreateObject", "event", typedEvent)
			return nil, errors.New("create object event assert error")
		}
		data := m.handleCreateObject(ctx, block, txHash, createObject)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.SaveObject(ctx, data)
	case EventCancelCreateObject:
		cancelCreateObject, ok := typedEvent.(*storagetypes.EventCancelCreateObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCancelCreateObject", "event", typedEvent)
			return nil, errors.New("cancel create object event assert error")
		}
		data := m.handleCancelCreateObject(ctx, block, txHash, cancelCreateObject)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	case EventSealObject:
		sealObject, ok := typedEvent.(*storagetypes.EventSealObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventSealObject", "event", typedEvent)
			return nil, errors.New("seal object event assert error")
		}
		data := m.handleSealObject(ctx, block, txHash, sealObject)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	case EventCopyObject:
		copyObject, ok := typedEvent.(*storagetypes.EventCopyObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCopyObject", "event", typedEvent)
			return nil, errors.New("copy object event assert error")
		}
		data, copyErr := m.handleCopyObject(ctx, block, txHash, copyObject)
		if copyErr != nil {
			return nil, copyErr
		}
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	case EventDeleteObject:
		deleteObject, ok := typedEvent.(*storagetypes.EventDeleteObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventDeleteObject", "event", typedEvent)
			return nil, errors.New("delete object event assert error")
		}
		data := m.handleDeleteObject(ctx, block, txHash, deleteObject)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	case EventRejectSealObject:
		rejectSealObject, ok := typedEvent.(*storagetypes.EventRejectSealObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventRejectSealObject", "event", typedEvent)
			return nil, errors.New("reject seal object event assert error")
		}
		data := m.handleRejectSealObject(ctx, block, txHash, rejectSealObject)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	case EventDiscontinueObject:
		discontinueObject, ok := typedEvent.(*storagetypes.EventDiscontinueObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventDiscontinueObject", "event", typedEvent)
			return nil, errors.New("discontinue object event assert error")
		}
		data := m.handleEventDiscontinueObject(ctx, block, txHash, discontinueObject)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	case EventUpdateObjectInfo:
		updateObjectInfo, ok := typedEvent.(*storagetypes.EventUpdateObjectInfo)
		if !ok {
			log.Errorw("type assert error", "type", "EventUpdateObjectInfo", "event", typedEvent)
			return nil, errors.New("update object event assert error")
		}
		data := m.handleUpdateObjectInfo(ctx, block, txHash, updateObjectInfo)
		if !OperationDB {
			return data, nil
		}
		return nil, m.db.UpdateObject(ctx, data)
	}

	return nil, nil
}

func (m *Module) handleCreateObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, createObject *storagetypes.EventCreateObject) *models.Object {
	return &models.Object{
		BucketID:       common.BigToHash(createObject.BucketId.BigInt()),
		BucketName:     createObject.BucketName,
		ObjectID:       common.BigToHash(createObject.ObjectId.BigInt()),
		ObjectName:     createObject.ObjectName,
		Creator:        common.HexToAddress(createObject.Creator),
		Owner:          common.HexToAddress(createObject.Owner),
		PayloadSize:    createObject.PayloadSize,
		Visibility:     createObject.Visibility.String(),
		ContentType:    createObject.ContentType,
		Status:         createObject.Status.String(),
		RedundancyType: createObject.RedundancyType.String(),
		SourceType:     createObject.SourceType.String(),
		CheckSums:      createObject.Checksums,

		CreateTxHash: txHash,
		CreateAt:     block.Block.Height,
		CreateTime:   createObject.CreateAt,
		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   createObject.CreateAt,
		Removed:      false,
	}
}

func (m *Module) handleSealObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, sealObject *storagetypes.EventSealObject) *models.Object {
	return &models.Object{
		BucketName:          sealObject.BucketName,
		ObjectName:          sealObject.ObjectName,
		ObjectID:            common.BigToHash(sealObject.ObjectId.BigInt()),
		Operator:            common.HexToAddress(sealObject.Operator),
		LocalVirtualGroupId: sealObject.LocalVirtualGroupId,
		Status:              sealObject.Status.String(),
		SealedTxHash:        txHash,

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      false,
	}
}

func (m *Module) handleCancelCreateObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, cancelCreateObject *storagetypes.EventCancelCreateObject) *models.Object {
	return &models.Object{
		BucketName:   cancelCreateObject.BucketName,
		ObjectName:   cancelCreateObject.ObjectName,
		ObjectID:     common.BigToHash(cancelCreateObject.ObjectId.BigInt()),
		Operator:     common.HexToAddress(cancelCreateObject.Operator),
		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      true,
	}
}

func (m *Module) handleCopyObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, copyObject *storagetypes.EventCopyObject) (*models.Object, error) {
	destObject, err := m.db.GetObject(ctx, common.BigToHash(copyObject.SrcObjectId.BigInt()))
	if err != nil {
		return nil, err
	}

	destObject.ObjectID = common.BigToHash(copyObject.DstObjectId.BigInt())
	destObject.ObjectName = copyObject.DstObjectName
	destObject.BucketName = copyObject.DstBucketName
	destObject.Operator = common.HexToAddress(copyObject.Operator)
	destObject.CreateAt = block.Block.Height
	destObject.CreateTxHash = txHash
	destObject.CreateTime = block.Block.Time.UTC().Unix()
	destObject.UpdateAt = block.Block.Height
	destObject.UpdateTxHash = txHash
	destObject.UpdateTime = block.Block.Time.UTC().Unix()
	destObject.Removed = false

	return destObject, nil
}

func (m *Module) handleDeleteObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, deleteObject *storagetypes.EventDeleteObject) *models.Object {
	return &models.Object{
		BucketName:          deleteObject.BucketName,
		ObjectName:          deleteObject.ObjectName,
		ObjectID:            common.BigToHash(deleteObject.ObjectId.BigInt()),
		LocalVirtualGroupId: deleteObject.LocalVirtualGroupId,

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      true,
	}
}

// RejectSeal event won't emit a delete event, need to be deleted manually here in metadata service
// handle logic is set as removed, no need to set status
func (m *Module) handleRejectSealObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, rejectSealObject *storagetypes.EventRejectSealObject) *models.Object {
	return &models.Object{
		BucketName: rejectSealObject.BucketName,
		ObjectName: rejectSealObject.ObjectName,
		ObjectID:   common.BigToHash(rejectSealObject.ObjectId.BigInt()),
		Operator:   common.HexToAddress(rejectSealObject.Operator),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      true,
	}
}

func (m *Module) handleEventDiscontinueObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, discontinueObject *storagetypes.EventDiscontinueObject) *models.Object {
	return &models.Object{
		BucketName:   discontinueObject.BucketName,
		ObjectID:     common.BigToHash(discontinueObject.ObjectId.BigInt()),
		DeleteReason: discontinueObject.Reason,
		DeleteAt:     discontinueObject.DeleteAt,
		Status:       storagetypes.OBJECT_STATUS_DISCONTINUED.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      false,
	}
}

func (m *Module) handleUpdateObjectInfo(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, updateObject *storagetypes.EventUpdateObjectInfo) *models.Object {
	return &models.Object{
		BucketName: updateObject.BucketName,
		ObjectID:   common.BigToHash(updateObject.ObjectId.BigInt()),
		ObjectName: updateObject.ObjectName,
		Operator:   common.HexToAddress(updateObject.Operator),
		Visibility: updateObject.Visibility.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}
}
