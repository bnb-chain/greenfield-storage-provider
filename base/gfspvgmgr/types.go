package gfspvgmgr

import "github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"

func NewIDSetFromList(list []uint32) vgmgr.IDSet {
	set := make(map[uint32]struct{}, 0)
	for _, v := range list {
		set[v] = struct{}{}
	}
	return set
}
