package piecestore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var numberRegex = regexp.MustCompile("[0-9]+")

// EncodeSegmentPieceKey encodes segment piece store key
func EncodeSegmentPieceKey(objectID uint64, segmentIndex int) string {
	return fmt.Sprintf("%d_s%d", objectID, segmentIndex)
}

// DecodeSegmentPieceKey decodes segment piece store key
// Valid segment piece key: objectID_s0
func DecodeSegmentPieceKey(pieceKey string) (uint64, int, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckSegmentPieceKey(keys); !valid {
		log.Errorw("Invalid segment piece key", "piece key", pieceKey)
		return 0, 0, fmt.Errorf("Invalid segment piece key")
	}

	objectID, _ := strconv.ParseUint(keys[0], 10, 64)
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.Atoi(s)

	return objectID, segmentIndex, nil
}

// EncodeECPieceKey encodes ec piece store key
func EncodeECPieceKey(objectID uint64, segmentIndex, pieceIndex int) string {
	return fmt.Sprintf("%d_s%d_p%d", objectID, segmentIndex, pieceIndex)
}

// DecodeECPieceKey decodes ec piece store key
// Valid EC piece key: objectID_s0_p0
func DecodeECPieceKey(pieceKey string) (uint64, int, int, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckECPieceKey(keys); !valid {
		log.Errorw("Invalid EC piece key", "piece key", pieceKey)
		return 0, 0, 0, fmt.Errorf("Invalid EC piece key")
	}

	objectID, _ := strconv.ParseUint(keys[0], 10, 64)
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.Atoi(s)
	e := numberRegex.FindString(keys[2])
	ecIndex, _ := strconv.Atoi(e)

	return objectID, segmentIndex, ecIndex, nil
}

var (
	objectRegex  = regexp.MustCompile("^[0-9]+$")
	segmentRegex = regexp.MustCompile("^[s][0-9]+$")
	ecRegex      = regexp.MustCompile("^[p][0-9]+$")
)

// CheckSegmentPieceKey checks ec piece key is correct
func CheckSegmentPieceKey(keys []string) bool {
	if len(keys) != 2 {
		log.Errorw("Invalid segment piece key", "piece key", keys)
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
		log.Errorw("Invalid ec piece key", "piece key", keys)
		return false
	}

	if objectRegex.MatchString(keys[0]) && segmentRegex.MatchString(keys[1]) && ecRegex.MatchString(keys[2]) {
		return true
	}
	return false
}
