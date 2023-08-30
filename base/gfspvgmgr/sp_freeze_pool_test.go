package gfspvgmgr

import (
	"testing"

	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/stretchr/testify/assert"
)

func TestFreezeSPPool_FreezeSPAndGVGs(t *testing.T) {
	pool := NewFreezeSPPool()
	joinedGVGs := []*virtual_types.GlobalVirtualGroup{{Id: 1}}
	pool.FreezeSPAndGVGs(1, joinedGVGs)
}

func TestFreezeSPPool_GetFreezeSPIDs(t *testing.T) {
	pool := NewFreezeSPPool()
	joinedGVGs := []*virtual_types.GlobalVirtualGroup{{Id: 1, FamilyId: 1}}
	pool.FreezeSPAndGVGs(1, joinedGVGs)
	result := pool.GetFreezeSPIDs()
	assert.Equal(t, map[uint32]struct{}{0x1: {}}, result)
}

func TestFreezeSPPool_GetFreezeGVGsInFamily(t *testing.T) {
	pool := NewFreezeSPPool()
	joinedGVGs := []*virtual_types.GlobalVirtualGroup{{Id: 1, FamilyId: 1}}
	pool.FreezeSPAndGVGs(1, joinedGVGs)
	result := pool.GetFreezeGVGsInFamily(1)
	assert.Equal(t, map[uint32]struct{}{0x1: {}}, result)
}

func TestFreezeSPPool_ReleaseSP(t *testing.T) {
	pool := NewFreezeSPPool()
	joinedGVGs := []*virtual_types.GlobalVirtualGroup{{Id: 1, FamilyId: 1}}
	pool.FreezeSPAndGVGs(1, joinedGVGs)
	pool.ReleaseSP()
}
