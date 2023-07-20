package object

import (
	"context"
	"errors"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"
	"github.com/forbole/juno/v4/models"
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
	return nil
}

func (m *Module) ExtractEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
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
		return m.handleCreateObject(ctx, block, txHash, createObject), nil
	case EventCancelCreateObject:
		cancelCreateObject, ok := typedEvent.(*storagetypes.EventCancelCreateObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCancelCreateObject", "event", typedEvent)
			return nil, errors.New("cancel create object event assert error")
		}
		return m.handleCancelCreateObject(ctx, block, txHash, cancelCreateObject), nil

	case EventSealObject:
		sealObject, ok := typedEvent.(*storagetypes.EventSealObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventSealObject", "event", typedEvent)
			return nil, errors.New("seal object event assert error")
		}
		return m.handleSealObject(ctx, block, txHash, sealObject), nil
	case EventCopyObject:
		copyObject, ok := typedEvent.(*storagetypes.EventCopyObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCopyObject", "event", typedEvent)
			return nil, errors.New("copy object event assert error")
		}
		return m.handleCopyObject(ctx, block, txHash, copyObject)
	case EventDeleteObject:
		deleteObject, ok := typedEvent.(*storagetypes.EventDeleteObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventDeleteObject", "event", typedEvent)
			return nil, errors.New("delete object event assert error")
		}
		return m.handleDeleteObject(ctx, block, txHash, deleteObject), nil
	case EventRejectSealObject:
		rejectSealObject, ok := typedEvent.(*storagetypes.EventRejectSealObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventRejectSealObject", "event", typedEvent)
			return nil, errors.New("reject seal object event assert error")
		}
		return m.handleRejectSealObject(ctx, block, txHash, rejectSealObject), nil
	case EventDiscontinueObject:
		discontinueObject, ok := typedEvent.(*storagetypes.EventDiscontinueObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventDiscontinueObject", "event", typedEvent)
			return nil, errors.New("discontinue object event assert error")
		}
		return m.handleEventDiscontinueObject(ctx, block, txHash, discontinueObject), nil
	case EventUpdateObjectInfo:
		updateObjectInfo, ok := typedEvent.(*storagetypes.EventUpdateObjectInfo)
		if !ok {
			log.Errorw("type assert error", "type", "EventUpdateObjectInfo", "event", typedEvent)
			return nil, errors.New("update object event assert error")
		}
		return m.handleUpdateObjectInfo(ctx, block, txHash, updateObjectInfo), nil
	}

	return nil, nil
}

func (m *Module) handleCreateObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, createObject *storagetypes.EventCreateObject) map[string][]interface{} {
	object := &models.Object{
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

	k, v := m.db.SaveObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleSealObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, sealObject *storagetypes.EventSealObject) map[string][]interface{} {
	object := &models.Object{
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

	k, v := m.db.UpdateObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleCancelCreateObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, cancelCreateObject *storagetypes.EventCancelCreateObject) map[string][]interface{} {
	object := &models.Object{
		BucketName:   cancelCreateObject.BucketName,
		ObjectName:   cancelCreateObject.ObjectName,
		ObjectID:     common.BigToHash(cancelCreateObject.ObjectId.BigInt()),
		Operator:     common.HexToAddress(cancelCreateObject.Operator),
		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      true,
	}

	k, v := m.db.UpdateObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleCopyObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, copyObject *storagetypes.EventCopyObject) (map[string][]interface{}, error) {
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

	k, v := m.db.UpdateObjectToSQL(ctx, destObject)
	return map[string][]interface{}{
		k: v,
	}, nil
}

func (m *Module) handleDeleteObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, deleteObject *storagetypes.EventDeleteObject) map[string][]interface{} {
	object := &models.Object{
		BucketName:          deleteObject.BucketName,
		ObjectName:          deleteObject.ObjectName,
		ObjectID:            common.BigToHash(deleteObject.ObjectId.BigInt()),
		LocalVirtualGroupId: deleteObject.LocalVirtualGroupId,

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      true,
	}

	k, v := m.db.UpdateObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}

// RejectSeal event won't emit a delete event, need to be deleted manually here in metadata service
// handle logic is set as removed, no need to set status
func (m *Module) handleRejectSealObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, rejectSealObject *storagetypes.EventRejectSealObject) map[string][]interface{} {
	object := &models.Object{
		BucketName: rejectSealObject.BucketName,
		ObjectName: rejectSealObject.ObjectName,
		ObjectID:   common.BigToHash(rejectSealObject.ObjectId.BigInt()),
		Operator:   common.HexToAddress(rejectSealObject.Operator),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
		Removed:      true,
	}

	k, v := m.db.UpdateObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleEventDiscontinueObject(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, discontinueObject *storagetypes.EventDiscontinueObject) map[string][]interface{} {
	object := &models.Object{
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

	k, v := m.db.UpdateObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleUpdateObjectInfo(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, updateObject *storagetypes.EventUpdateObjectInfo) map[string][]interface{} {
	object := &models.Object{
		BucketName: updateObject.BucketName,
		ObjectID:   common.BigToHash(updateObject.ObjectId.BigInt()),
		ObjectName: updateObject.ObjectName,
		Operator:   common.HexToAddress(updateObject.Operator),
		Visibility: updateObject.Visibility.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateObjectToSQL(ctx, object)
	return map[string][]interface{}{
		k: v,
	}
}
