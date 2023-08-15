package group

import (
	"context"
	"errors"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/log"

	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/models"
)

var (
	EventCreateGroup       = proto.MessageName(&storagetypes.EventCreateGroup{})
	EventDeleteGroup       = proto.MessageName(&storagetypes.EventDeleteGroup{})
	EventLeaveGroup        = proto.MessageName(&storagetypes.EventLeaveGroup{})
	EventUpdateGroupMember = proto.MessageName(&storagetypes.EventUpdateGroupMember{})
	EventRenewGroupMember  = proto.MessageName(&storagetypes.EventRenewGroupMember{})
)

var GroupEvents = map[string]bool{
	EventCreateGroup:       true,
	EventDeleteGroup:       true,
	EventLeaveGroup:        true,
	EventUpdateGroupMember: true,
	EventRenewGroupMember:  true,
}

func (m *Module) ExtractEventStatements(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	if !GroupEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventCreateGroup:
		createGroup, ok := typedEvent.(*storagetypes.EventCreateGroup)
		if !ok {
			log.Errorw("type assert error", "type", "EventCreateGroup", "event", typedEvent)
			return nil, errors.New("create group event assert error")
		}
		return m.handleCreateGroup(ctx, block, createGroup), nil
	case EventUpdateGroupMember:
		updateGroupMember, ok := typedEvent.(*storagetypes.EventUpdateGroupMember)
		if !ok {
			log.Errorw("type assert error", "type", "EventUpdateGroupMember", "event", typedEvent)
			return nil, errors.New("update group member event assert error")
		}
		return m.handleUpdateGroupMember(ctx, block, updateGroupMember), nil

	case EventDeleteGroup:
		deleteGroup, ok := typedEvent.(*storagetypes.EventDeleteGroup)
		if !ok {
			log.Errorw("type assert error", "type", "EventDeleteGroup", "event", typedEvent)
			return nil, errors.New("delete group event assert error")
		}
		return m.handleDeleteGroup(ctx, block, deleteGroup), nil
	case EventLeaveGroup:
		leaveGroup, ok := typedEvent.(*storagetypes.EventLeaveGroup)
		if !ok {
			log.Errorw("type assert error", "type", "EventLeaveGroup", "event", typedEvent)
			return nil, errors.New("leave group event assert error")
		}
		return m.handleLeaveGroup(ctx, block, leaveGroup), nil
	case EventRenewGroupMember:
		renewGroupMember, ok := typedEvent.(*storagetypes.EventRenewGroupMember)
		if !ok {
			log.Errorw("type assert error", "type", "EventRenewGroupMember", "event", typedEvent)
			return nil, errors.New("renew group member event assert error")
		}
		return m.handleRenewGroupMember(ctx, block, renewGroupMember), nil
	}
	return nil, nil
}

func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, _ common.Hash, event sdk.Event) error {
	return nil
}

func (m *Module) handleCreateGroup(ctx context.Context, block *tmctypes.ResultBlock, createGroup *storagetypes.EventCreateGroup) map[string][]interface{} {

	var membersToAddList []*models.Group

	//create group first
	groupItem := &models.Group{
		Owner:      common.HexToAddress(createGroup.Owner),
		GroupID:    common.BigToHash(createGroup.GroupId.BigInt()),
		GroupName:  createGroup.GroupName,
		SourceType: createGroup.SourceType.String(),
		AccountID:  common.HexToAddress("0"),
		Extra:      createGroup.Extra,

		CreateAt:   block.Block.Height,
		CreateTime: block.Block.Time.UTC().Unix(),
		UpdateAt:   block.Block.Height,
		UpdateTime: block.Block.Time.UTC().Unix(),
		Removed:    false,
	}
	membersToAddList = append(membersToAddList, groupItem)

	k, v := m.db.CreateGroupToSQL(ctx, membersToAddList)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleDeleteGroup(ctx context.Context, block *tmctypes.ResultBlock, deleteGroup *storagetypes.EventDeleteGroup) map[string][]interface{} {
	group := &models.Group{
		Owner:     common.HexToAddress(deleteGroup.Owner),
		GroupID:   common.BigToHash(deleteGroup.GroupId.BigInt()),
		GroupName: deleteGroup.GroupName,

		UpdateAt:   block.Block.Height,
		UpdateTime: block.Block.Time.UTC().Unix(),
		Removed:    true,
	}

	res := make(map[string][]interface{})

	k, v := m.db.DeleteGroupToSQL(ctx, group)
	res[k] = v
	return res
}

func (m *Module) handleLeaveGroup(ctx context.Context, block *tmctypes.ResultBlock, leaveGroup *storagetypes.EventLeaveGroup) map[string][]interface{} {
	group := &models.Group{
		Owner:     common.HexToAddress(leaveGroup.Owner),
		GroupID:   common.BigToHash(leaveGroup.GroupId.BigInt()),
		GroupName: leaveGroup.GroupName,
		AccountID: common.HexToAddress(leaveGroup.MemberAddress),

		UpdateAt:   block.Block.Height,
		UpdateTime: block.Block.Time.UTC().Unix(),
		Removed:    true,
	}

	//update group item
	groupItem := &models.Group{
		GroupID:   common.BigToHash(leaveGroup.GroupId.BigInt()),
		AccountID: common.HexToAddress("0"),

		UpdateAt:   block.Block.Height,
		UpdateTime: block.Block.Time.UTC().Unix(),
		Removed:    false,
	}
	res := make(map[string][]interface{})
	k, v := m.db.UpdateGroupToSQL(ctx, groupItem)
	res[k] = v

	k, v = m.db.UpdateGroupToSQL(ctx, group)
	res[k] = v
	return res
}

func (m *Module) handleUpdateGroupMember(ctx context.Context, block *tmctypes.ResultBlock, updateGroupMember *storagetypes.EventUpdateGroupMember) map[string][]interface{} {

	membersToAdd := updateGroupMember.MembersToAdd
	membersToDelete := updateGroupMember.MembersToDelete

	var membersToAddList []*models.Group
	res := make(map[string][]interface{})

	if len(membersToAdd) > 0 {
		for _, memberToAdd := range membersToAdd {
			groupItem := &models.Group{
				Owner:          common.HexToAddress(updateGroupMember.Owner),
				GroupID:        common.BigToHash(updateGroupMember.GroupId.BigInt()),
				GroupName:      updateGroupMember.GroupName,
				AccountID:      common.HexToAddress(memberToAdd.Member),
				Operator:       common.HexToAddress(updateGroupMember.Operator),
				ExpirationTime: memberToAdd.ExpirationTime.Unix(),

				CreateAt:   block.Block.Height,
				CreateTime: block.Block.Time.UTC().Unix(),
				UpdateAt:   block.Block.Height,
				UpdateTime: block.Block.Time.UTC().Unix(),
				Removed:    false,
			}
			membersToAddList = append(membersToAddList, groupItem)
		}
		k, v := m.db.CreateGroupToSQL(ctx, membersToAddList)
		res[k] = v
	}

	for _, memberToDelete := range membersToDelete {
		groupItem := &models.Group{
			Owner:     common.HexToAddress(updateGroupMember.Owner),
			GroupID:   common.BigToHash(updateGroupMember.GroupId.BigInt()),
			GroupName: updateGroupMember.GroupName,
			AccountID: common.HexToAddress(memberToDelete),
			Operator:  common.HexToAddress(updateGroupMember.Operator),

			UpdateAt:   block.Block.Height,
			UpdateTime: block.Block.Time.UTC().Unix(),
			Removed:    true,
		}
		k, v := m.db.UpdateGroupToSQL(ctx, groupItem)
		res[k] = v
	}

	//update group item
	groupItem := &models.Group{
		GroupID:   common.BigToHash(updateGroupMember.GroupId.BigInt()),
		AccountID: common.HexToAddress("0"),

		UpdateAt:   block.Block.Height,
		UpdateTime: block.Block.Time.UTC().Unix(),
		Removed:    false,
	}
	k, v := m.db.UpdateGroupToSQL(ctx, groupItem)
	res[k] = v

	return res
}

func (m *Module) handleRenewGroupMember(ctx context.Context, block *tmctypes.ResultBlock, renewGroupMember *storagetypes.EventRenewGroupMember) map[string][]interface{} {
	res := map[string][]interface{}{}
	for _, e := range renewGroupMember.Members {
		k, v := m.db.UpdateGroupToSQL(ctx, &models.Group{
			GroupID:        common.BigToHash(renewGroupMember.GroupId.BigInt()),
			AccountID:      common.HexToAddress(e.Member),
			ExpirationTime: e.ExpirationTime.Unix(),

			UpdateAt:   block.Block.Height,
			UpdateTime: block.Block.Time.UTC().Unix(),
		})
		res[k] = v
	}
	return res
}
