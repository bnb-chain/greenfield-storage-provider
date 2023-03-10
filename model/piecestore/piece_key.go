package piecestore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var numberRegex = regexp.MustCompile("[0-9]+")

// EncodeSegmentPieceKey encodes segment piece store key
func EncodeSegmentPieceKey(objectID string, segmentIndex uint32) string {
	return fmt.Sprintf("%s_s%d", objectID, segmentIndex)
}

// DecodeSegmentPieceKey decodes segment piece store key
// Valid segment piece key: objectID_s0
func DecodeSegmentPieceKey(pieceKey string) (string, uint32, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckSegmentPieceKey(keys); !valid {
		log.Errorw("segment piece key is wrong", "segment piece key", pieceKey)
		return "0", 0, fmt.Errorf("invalid segment piece key")
	}

	objectID := keys[0]
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.ParseUint(s, 10, 32)

	return objectID, uint32(segmentIndex), nil
}

// EncodeECPieceKey encodes ec piece store key
func EncodeECPieceKey(objectID string, segmentIndex, ecIndex uint32) string {
	return fmt.Sprintf("%s_s%d_p%d", objectID, segmentIndex, ecIndex)
}

// DecodeECPieceKey decodes ec piece store key
// Valid EC piece key: objectID_s0_p0
func DecodeECPieceKey(pieceKey string) (string, uint32, uint32, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckECPieceKey(keys); !valid {
		log.Errorw("ec piece key is wrong", "ec piece key", pieceKey)
		return "0", 0, 0, fmt.Errorf("invalid EC piece key")
	}

	objectID := keys[0]
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

// ComputeSegmentCount return the segments count by payload size.
func ComputeSegmentCount(size uint64, spiltSize uint64) uint32 {
	segmentCount := uint32(size / spiltSize)
	if (size % spiltSize) > 0 {
		segmentCount++
	}
	return segmentCount
}
