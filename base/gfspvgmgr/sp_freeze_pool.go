package gfspvgmgr

import (
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"sync"
	"time"
)

const (
	DefaultFreezingPeriodForSPInMin = 1 * time.Minute
	ReleaseSPCronJobInterval        = 20 * time.Second
)

type FreezeSPPool struct {
	sync.Map
}

func NewSpFreezingPool() *FreezeSPPool {
	return &FreezeSPPool{sync.Map{}}
}

// FreezeSPAndGVGs put a Secondary SP and its joined Global virtual group into the Freeze Pool.
func (s *FreezeSPPool) FreezeSPAndGVGs(spID uint32, joinedGVGs []*virtual_types.GlobalVirtualGroup) {
	stats := &SPStats{
		JoinedGVGs:  joinedGVGs,
		FreezeUntil: time.Now().Add(DefaultFreezingPeriodForSPInMin).Unix(),
	}
	s.Store(spID, stats)
}

func (s *FreezeSPPool) Remove(SpID uint32) {
	s.Delete(SpID)
}

func (s *FreezeSPPool) GetFreezeSPIDs() vgmgr.IDSet {
	idSet := make(map[uint32]struct{}, 0)
	s.Range(func(k, v interface{}) bool {
		fmt.Println("printing sync map")
		fmt.Println(k, v)
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

func (s *FreezeSPPool) ReleaseSP() []uint32 {
	sps := make([]uint32, 0)
	s.Range(func(k interface{}, v interface{}) bool {
		stats := v.(*SPStats)
		fmt.Printf("ReleaseSP loop k-v, %v, %v \n", k, v)
		if time.Now().Unix() > stats.FreezeUntil {
			s.Delete(k)
			sps = append(sps, k.(uint32))
		}
		return true
	})
	fmt.Printf("ReleaseSP sps %v is released  \n", sps)
	return sps
}

type SPStats struct {
	JoinedGVGs  []*virtual_types.GlobalVirtualGroup
	FreezeUntil int64
}
