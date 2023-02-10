package piecestore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var numberRegex = regexp.MustCompile("[0-9]+")

// EncodeSegmentPieceKey encodes segment piece store key
func EncodeSegmentPieceKey(objectID uint64, segmentIndex uint32) string {
	return fmt.Sprintf("%d_s%d", objectID, segmentIndex)
}

// DecodeSegmentPieceKey decodes segment piece store key
// Valid segment piece key: objectID_s0
func DecodeSegmentPieceKey(pieceKey string) (uint64, uint32, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckSegmentPieceKey(keys); !valid {
		log.Errorw("segment piece key is wrong", "segment piece key", pieceKey)
		return 0, 0, fmt.Errorf("invalid segment piece key")
	}

	objectID, _ := strconv.ParseUint(keys[0], 10, 64)
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.ParseUint(s, 10, 32)

	return objectID, uint32(segmentIndex), nil
}

// EncodePieceKey encodes piece store key
func EncodePieceKey(rType ptypes.RedundancyType, objectId uint64, segmentIndex, ecIndex uint32) (string, error) {
	var pieceKey string
	switch rType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		pieceKey = EncodeECPieceKey(objectId, segmentIndex, ecIndex)
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE:
		pieceKey = EncodeSegmentPieceKey(objectId, segmentIndex)
	case ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceKey = EncodeSegmentPieceKey(objectId, segmentIndex)
	default:
		return "", merrors.ErrRedundancyType
	}

	return pieceKey, nil
}

// EncodeECPieceKey encodes ec piece store key
func EncodeECPieceKey(objectID uint64, segmentIndex, ecIndex uint32) string {
	return fmt.Sprintf("%d_s%d_p%d", objectID, segmentIndex, ecIndex)
}

// EncodeECPieceKeyBySegmentKey encodes ec piece store key
func EncodeECPieceKeyBySegmentKey(segmentKey string, ecIndex uint32) string {
	return fmt.Sprintf("%s_p%d", segmentKey, ecIndex)
}

// DecodeECPieceKey decodes ec piece store key
// Valid EC piece key: objectID_s0_p0
func DecodeECPieceKey(pieceKey string) (uint64, uint32, uint32, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckECPieceKey(keys); !valid {
		log.Errorw("ec piece key is wrong", "ec piece key", pieceKey)
		return 0, 0, 0, fmt.Errorf("invalid EC piece key")
	}

	objectID, _ := strconv.ParseUint(keys[0], 10, 64)
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.ParseUint(s, 10, 32)
	e := numberRegex.FindString(keys[2])
	ecIndex, _ := strconv.ParseUint(e, 10, 32)

	return objectID, uint32(segmentIndex), uint32(ecIndex), nil
}

var (
	objectRegex  = regexp.MustCompile("^[0-9]+$")
	segmentRegex = regexp.MustCompile("^[s][0-9]+$")
	ecRegex      = regexp.MustCompile("^[p][0-9]+$")
)

// CheckSegmentPieceKey checks ec piece key is correct
func CheckSegmentPieceKey(keys []string) bool {
	if len(keys) != 2 {
		log.Errorw("invalid segment piece key", "segment piece key", keys)
		return false
	}

	if objectRegex.MatchString(keys[0]) && segmentRegex.MatchString(keys[1]) {
		return true
	}
	return false
}

// CheckECPieceKey checks EC piece key is correct
func CheckECPieceKey(keys []string) bool {
	if len(keys) != 3 {
		log.Errorw("invalid EC piece key", "ec piece key", keys)
		return false
	}

	if objectRegex.MatchString(keys[0]) && segmentRegex.MatchString(keys[1]) && ecRegex.MatchString(keys[2]) {
		return true
	}
	return false
}
