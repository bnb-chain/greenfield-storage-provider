package gfspvgmgr

import (
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const (
	DefaultFreezingPeriodForSP = 1 * time.Second
	ReleaseSPJobInterval       = 1 * time.Minute
)

type FreezeSPPool struct {
	sync.Map
}

type SPStats struct {
	JoinedGVGs  []*virtual_types.GlobalVirtualGroup
	FreezeUntil int64
}

func NewFreezeSPPool() *FreezeSPPool {
	return &FreezeSPPool{sync.Map{}}
}

// FreezeSPAndGVGs puts a Secondary SP and its joined Global virtual groups into the Freeze Pool.
func (s *FreezeSPPool) FreezeSPAndGVGs(spID uint32, joinedGVGs []*virtual_types.GlobalVirtualGroup) {
	stats := &SPStats{
		JoinedGVGs:  joinedGVGs,
		FreezeUntil: time.Now().Add(DefaultFreezingPeriodForSP).Unix(),
	}
	s.Store(spID, stats)
}

func (s *FreezeSPPool) GetFreezeSPIDs() vgmgr.IDSet {
	idSet := make(map[uint32]struct{}, 0)
	s.Range(func(k, v interface{}) bool {
		idSet[k.(uint32)] = struct{}{}
		return true
	})
	return idSet
}

func (s *FreezeSPPool) GetFreezeGVGsInFamily(familyID uint32) vgmgr.IDSet {
	idSet := make(map[uint32]struct{}, 0)
	s.Range(func(k, v interface{}) bool {
		for _, gvg := range v.(*SPStats).JoinedGVGs {
			if familyID == gvg.GetFamilyId() {
				idSet[gvg.Id] = struct{}{}
			}
		}
		return true
	})
	return idSet
}

func (s *FreezeSPPool) ReleaseSP() {
	s.Range(func(k interface{}, v interface{}) bool {
		stats := v.(*SPStats)
		if time.Now().Unix() > stats.FreezeUntil {
			s.Delete(k)
		}
		return true
	})
}

func (s *FreezeSPPool) ReleaseAllSP() {
	s.Range(func(k interface{}, v interface{}) bool {
		s.Delete(k)
		return true
	})
}
