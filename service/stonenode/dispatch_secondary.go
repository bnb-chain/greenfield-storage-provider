package stonenode

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// dispatchSecondarySP dispatch piece data to secondary storage provider.
// returned map key is spID, value map key is ec piece key or segment key, value map's value is piece data
func (node *StoneNodeService) dispatchSecondarySP(pieceDataBySegment map[string][][]byte, redundancyType ptypesv1pb.RedundancyType,
	secondarySPs []string, targetIdx []uint32) (map[string]map[string][]byte, error) {
	pieceDataBySecondary := make(map[string]map[string][]byte)

	// pieceDataBySegment key is segment key; if redundancyType is EC, value is [][]byte type,
	// a two-dimensional array which contains ec data from ec1 []byte data to ec6 []byte data
	// if redundancyType is replica or inline, value is [][]byte type, a two-dimensional array
	// which only contains one []byte data
	var err error
	switch redundancyType {
	case ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceDataBySecondary, err = dispatchReplicaOrInlineData(pieceDataBySegment, secondarySPs, targetIdx)
	default: // ec type
		pieceDataBySecondary, err = dispatchECData(pieceDataBySegment, secondarySPs, targetIdx)
	}
	if err != nil {
		log.Errorw("fill piece data by secondary error", "error", err)
		return nil, err
	}
	return pieceDataBySecondary, nil
}

// dispatchReplicaOrInlineData dispatches replica or inline data into different sp, each sp should store all segments data of an object
// if an object uses replica type, it's split into 10 segments and there are 6 sp, each sp should store 10 segments data
// if an object uses inline type, there is only one segment and there are 6 sp, each sp should store 1 segment data
func dispatchReplicaOrInlineData(pieceDataBySegment map[string][][]byte, secondarySPs []string, targetIdx []uint32) (
	map[string]map[string][]byte, error) {
	replicaOrInlineDataMap := make(map[string]map[string][]byte)
	spNumber := len(secondarySPs)
	if spNumber < 1 && spNumber > 6 {
		return replicaOrInlineDataMap, merrors.ErrSecondarySPNumber
	}

	keys := util.GenericSortedKeys(pieceDataBySegment)
	for i := 0; i < len(secondarySPs); i++ {
		sp := secondarySPs[i]
		spKey := encodeSPKey(i, sp)
		for j := 0; j < len(keys); j++ {
			pieceKey := keys[j]
			pieceData := pieceDataBySegment[pieceKey]
			if len(pieceData) != 1 {
				return nil, merrors.ErrInvalidSegmentData
			}

			for _, index := range targetIdx {
				if int(index) == i {
					if _, ok := replicaOrInlineDataMap[spKey]; !ok {
						replicaOrInlineDataMap[spKey] = make(map[string][]byte)
					}
					replicaOrInlineDataMap[spKey][pieceKey] = pieceData[0]
				}
			}
		}
	}
	return replicaOrInlineDataMap, nil
}

func encodeSPKey(spIndex int, sp string) string {
	return fmt.Sprintf("%d_%s", spIndex, sp)
}

func decodeSPKey(spKey string) (uint32, string, error) {
	keys := strings.Split(spKey, "_")
	if valid := checkSPKey(keys); !valid {
		log.Errorw("sp key is wrong", "sp key", spKey)
		return 0, "", fmt.Errorf("invalid sp key")
	}
	spIndex, _ := strconv.ParseUint(keys[0], 10, 32)
	return uint32(spIndex), keys[1], nil
}

var spRegex = regexp.MustCompile("^[0-9]+$")

func checkSPKey(keys []string) bool {
	if len(keys) != 2 {
		log.Errorw("invalid sp piece key", "sp key", keys)
		return false
	}
	if spRegex.MatchString(keys[0]) {
		return true
	}
	return false
}

// dispatchECData dispatched ec data into different sp
// one sp stores same ec column data: sp1 stores all ec1 data, sp2 stores all ec2 data, etc
func dispatchECData(pieceDataBySegment map[string][][]byte, secondarySPs []string, targetIdx []uint32) (map[string]map[string][]byte, error) {
	ecPieceDataMap := make(map[string]map[string][]byte)
	for pieceKey, pieceData := range pieceDataBySegment {
		if len(pieceData) != 6 {
			return map[string]map[string][]byte{}, merrors.ErrInvalidECData
		}

		for idx, data := range pieceData {
			if idx >= len(secondarySPs) {
				return map[string]map[string][]byte{}, merrors.ErrSecondarySPNumber
			}

			sp := secondarySPs[idx]
			for _, index := range targetIdx {
				if int(index) == idx {
					if _, ok := ecPieceDataMap[sp]; !ok {
						ecPieceDataMap[sp] = make(map[string][]byte)
					}
					key := piecestore.EncodeECPieceKeyBySegmentKey(pieceKey, uint32(idx))
					ecPieceDataMap[sp][key] = data
				}
			}
		}
	}
	return ecPieceDataMap, nil
}
