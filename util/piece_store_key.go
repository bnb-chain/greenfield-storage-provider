package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var numberRegex = regexp.MustCompile("[0-9]+")

// EncodeSegmentPieceKey encodes segment piece store key
func EncodeSegmentPieceKey(objectID string, segmentIndex int) string {
	return fmt.Sprintf("%s_s%d", objectID, segmentIndex)
}

// DecodeSegmentPieceKey decodes segment piece store key
// Valid segment piece key: objectID_s0
func DecodeSegmentPieceKey(pieceKey string) (string, int, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckSegmentPieceKey(keys); !valid {
		log.Errorw("Invalid segment piece key", "piece key", pieceKey)
		return "", 0, fmt.Errorf("Invalid segment piece key")
	}

	objectID := keys[0]
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.Atoi(s)

	return objectID, segmentIndex, nil
}

// EncodeECPieceKey encodes ec piece store key
func EncodeECPieceKey(objectID string, segmentIndex, pieceIndex int) string {
	return fmt.Sprintf("%s_s%d_p%d", objectID, segmentIndex, pieceIndex)
}

// DecodeECPieceKey decodes ec piece store key
// Valid EC piece key: objectID_s0_p0
func DecodeECPieceKey(pieceKey string) (string, int, int, error) {
	keys := strings.Split(pieceKey, "_")
	if valid := CheckECPieceKey(keys); !valid {
		log.Errorw("Invalid EC piece key", "piece key", pieceKey)
		return "", 0, 0, fmt.Errorf("Invalid EC piece key")
	}

	objectID := keys[0]
	s := numberRegex.FindString(keys[1])
	segmentIndex, _ := strconv.Atoi(s)
	e := numberRegex.FindString(keys[2])
	ecIndex, _ := strconv.Atoi(e)

	return objectID, segmentIndex, ecIndex, nil
}

var (
	segmentRegex = regexp.MustCompile("^[s][0-9]+$")
	ecRegex      = regexp.MustCompile("^[p][0-9]+$")
)

// CheckSegmentPieceKey checks ec piece key is correct
func CheckSegmentPieceKey(keys []string) bool {
	if len(keys) != 2 {
		log.Errorw("Invalid segment piece key", "piece key", keys)
		return false
	}

	if segmentRegex.MatchString(keys[1]) {
		return true
	}
	return false
}

// CheckECPieceKey checks EC piece key is correct
func CheckECPieceKey(keys []string) bool {
	if len(keys) != 3 {
		log.Errorw("Invalid segment piece key", "piece key", keys)
		return false
	}

	if segmentRegex.MatchString(keys[1]) && ecRegex.MatchString(keys[2]) {
		return true
	}
	return false
}
