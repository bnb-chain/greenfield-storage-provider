package general

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

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/util"
	"github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/resource"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	EventSetTag = proto.MessageName(&storagetypes.EventSetTag{})
)

var GeneralEvents = map[string]bool{
	EventSetTag: true,
}

func (m *Module) ExtractEventStatements(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	if !GeneralEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventSetTag:
		setTag, ok := typedEvent.(*storagetypes.EventSetTag)
		if !ok {
			log.Errorw("type assert error", "type", "EventSetTag", "event", typedEvent)
			return nil, errors.New("set tag event assert error")
		}
		var grn types.GRN
		parsingGrnErr := grn.ParseFromString(setTag.Resource, true)
		if parsingGrnErr != nil {
			log.Errorw("type parsing error", "type", "GRN", "value", setTag.Resource)
			return nil, errors.New("set tag event assert error")
		}

		return m.handleSetTag(ctx, block, grn, setTag.Tags), nil
	}
	return nil, nil
}

func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, _ common.Hash, event sdk.Event) error {
	return nil
}

func (m *Module) handleSetTag(ctx context.Context, block *tmctypes.ResultBlock, grn types.GRN, tags *storagetypes.ResourceTags) map[string][]interface{} {
	res := make(map[string][]interface{})
	switch grn.ResourceType() {
	case resource.RESOURCE_TYPE_BUCKET:
		bucketName, _ := grn.GetBucketName() // err will never happen here
		bucket := &models.Bucket{
			BucketName: bucketName,
			Tags:       util.GetTagJson(tags),
			UpdateAt:   block.Block.Height,
			UpdateTime: block.Block.Time.UTC().Unix(),
		}

		k, v := m.db.UpdateBucketByNameToSQL(ctx, bucket)
		res[k] = v
		return res
	case resource.RESOURCE_TYPE_OBJECT:
		bucketName, objectName, _ := grn.GetBucketAndObjectName() // err will never happen here
		object := &models.Object{
			BucketName: bucketName,
			ObjectName: objectName,
			UpdateAt:   block.Block.Height,
			UpdateTime: block.Block.Time.UTC().Unix(),
			Tags:       util.GetTagJson(tags),
		}

		k, v := m.db.UpdateObjectByBucketNameAndObjectNameToSQL(ctx, object)
		res[k] = v
		return res
	case resource.RESOURCE_TYPE_GROUP:
		groupOwner, groupName, _ := grn.GetGroupOwnerAndAccount() // err will never happen here
		//update group item
		groupItem := &models.Group{
			Owner:      common.Address(groupOwner),
			GroupName:  groupName,
			AccountID:  common.HexToAddress("0"),
			Tags:       util.GetTagJson(tags),
			UpdateAt:   block.Block.Height,
			UpdateTime: block.Block.Time.UTC().Unix(),
		}
		k, v := m.db.UpdateGroupByOwnerAndNameToSQL(ctx, groupItem)
		res[k] = v
		return res
	case resource.RESOURCE_TYPE_UNSPECIFIED:
		return res
	}

	return res
}
