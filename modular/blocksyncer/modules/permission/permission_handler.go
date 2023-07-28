package permission

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"

	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"
	"github.com/forbole/juno/v4/models"
)

var (
	EventPutPolicy    = proto.MessageName(&permissiontypes.EventPutPolicy{})
	EventDeletePolicy = proto.MessageName(&permissiontypes.EventDeletePolicy{})
)

var PolicyEvents = map[string]bool{
	EventPutPolicy:    true,
	EventDeletePolicy: true,
}

var actionTypeMap = map[permissiontypes.ActionType]int{
	permissiontypes.ACTION_TYPE_ALL:            0,
	permissiontypes.ACTION_UPDATE_BUCKET_INFO:  1,
	permissiontypes.ACTION_DELETE_BUCKET:       2,
	permissiontypes.ACTION_CREATE_OBJECT:       3,
	permissiontypes.ACTION_DELETE_OBJECT:       4,
	permissiontypes.ACTION_COPY_OBJECT:         5,
	permissiontypes.ACTION_GET_OBJECT:          6,
	permissiontypes.ACTION_EXECUTE_OBJECT:      7,
	permissiontypes.ACTION_LIST_OBJECT:         8,
	permissiontypes.ACTION_UPDATE_GROUP_MEMBER: 9,
	permissiontypes.ACTION_DELETE_GROUP:        10,
	permissiontypes.ACTION_UPDATE_OBJECT_INFO:  11,
}

func (m *Module) ExtractEventStatements(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	if !PolicyEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}
	switch event.Type {
	case EventPutPolicy:
		putPolicy, ok := typedEvent.(*permissiontypes.EventPutPolicy)
		if !ok {
			log.Errorw("type assert error", "type", "EventCreateObject", "event", typedEvent)
			return nil, errors.New("put policy event assert error")
		}
		return m.handlePutPolicy(ctx, block, putPolicy)
	case EventDeletePolicy:
		deletePolicy, ok := typedEvent.(*permissiontypes.EventDeletePolicy)
		if !ok {
			log.Errorw("type assert error", "type", "EventCancelCreateObject", "event", typedEvent)
			return nil, errors.New("cancel delete policy event assert error")
		}
		return m.handleDeletePolicy(ctx, block, deletePolicy), nil
	}
	return nil, nil
}

func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, _ common.Hash, event sdk.Event) error {
	return nil
}

func (m *Module) handlePutPolicy(ctx context.Context, block *tmctypes.ResultBlock, policy *permissiontypes.EventPutPolicy) (map[string][]interface{}, error) {
	var expireTime int64
	if policy.ExpirationTime == nil {
		expireTime = 0
	} else {
		expireTime = policy.ExpirationTime.Unix()
	}
	p := &models.Permission{
		PrincipalType:   int32(policy.Principal.Type),
		PrincipalValue:  policy.Principal.Value,
		ResourceType:    policy.ResourceType.String(),
		ResourceID:      common.BigToHash(policy.ResourceId.BigInt()),
		PolicyID:        common.BigToHash(policy.PolicyId.BigInt()),
		CreateTimestamp: block.Block.Time.Unix(),
		ExpirationTime:  expireTime,
	}

	statements := make([]*models.Statements, 0)
	for _, statement := range policy.Statements {
		actionValue := 0
		for _, action := range statement.Actions {
			value, ok := actionTypeMap[action]
			if !ok {
				return nil, errors.New("unknown action type action")
			}
			actionValue |= 1 << value
		}
		s := &models.Statements{
			PolicyID:    common.BigToHash(policy.PolicyId.BigInt()),
			Effect:      statement.Effect.String(),
			ActionValue: actionValue,
		}
		if statement.ExpirationTime != nil {
			s.ExpirationTime = statement.ExpirationTime.UTC().Unix()
		}
		if statement.LimitSize != nil {
			s.LimitSize = statement.LimitSize.Value
		}
		if len(statement.Resources) != 0 {
			s.Resources = statement.Resources
		}
		statements = append(statements, s)
	}

	res := make(map[string][]interface{})
	k, v := m.db.SavePermissionToSQL(ctx, p)
	res[k] = v
	k, v = m.db.MultiSaveStatementToSQL(ctx, statements)
	res[k] = v

	return res, nil
}

func (m *Module) handleDeletePolicy(ctx context.Context, block *tmctypes.ResultBlock, event *permissiontypes.EventDeletePolicy) map[string][]interface{} {
	policyIDHash := common.BigToHash(event.PolicyId.BigInt())
	res := make(map[string][]interface{})
	k, v := m.db.UpdatePermissionToSQL(ctx, &models.Permission{
		PolicyID:        policyIDHash,
		Removed:         true,
		UpdateTimestamp: block.Block.Time.Unix(),
	})
	res[k] = v
	k, v = m.db.RemoveStatementsToSQL(ctx, policyIDHash)
	res[k] = v
	return res
}
