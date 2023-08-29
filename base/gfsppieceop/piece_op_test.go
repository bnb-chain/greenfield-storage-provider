package gfsppieceop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGfSpPieceOp_SegmentPieceKey(t *testing.T) {
	p := &GfSpPieceOp{}
	result := p.SegmentPieceKey(1, 1)
	assert.Equal(t, "s1_s1", result)
}

func TestGfSpPieceOp_ECPieceKey(t *testing.T) {
	p := &GfSpPieceOp{}
	result := p.ECPieceKey(1, 1, 2)
	assert.Equal(t, "e1_s1_p2", result)
}

func TestGfSpPieceOp_ChallengePieceKey(t *testing.T) {
	cases := []struct {
		name          string
		objectID      uint64
		segmentIdx    uint32
		redundancyIdx int32
		wantedResult  string
	}{
		{
			name:          "1",
			objectID:      3,
			segmentIdx:    1,
			redundancyIdx: -1,
			wantedResult:  "s3_s1",
		},
		{
			name:          "2",
			objectID:      1,
			segmentIdx:    3,
			redundancyIdx: 2,
			wantedResult:  "e1_s3_p2",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &GfSpPieceOp{}
			result := p.ChallengePieceKey(tt.objectID, tt.segmentIdx, tt.redundancyIdx)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpPieceOp_MaxSegmentPieceSize(t *testing.T) {
	cases := []struct {
		name           string
		payloadSize    uint64
		maxSegmentSize uint64
		wantedResult   int64
	}{
		{
			name:           "1",
			payloadSize:    3,
			maxSegmentSize: 1,
			wantedResult:   1,
		},
		{
			name:           "2",
			payloadSize:    1,
			maxSegmentSize: 3,
			wantedResult:   1,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &GfSpPieceOp{}
			result := p.MaxSegmentPieceSize(tt.payloadSize, tt.maxSegmentSize)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpPieceOp_SegmentPieceSize(t *testing.T) {
	cases := []struct {
		name           string
		payloadSize    uint64
		segmentIdx     uint32
		maxSegmentSize uint64
		wantedResult   int64
	}{
		{
			name:           "1",
			payloadSize:    1,
			segmentIdx:     3,
			maxSegmentSize: 2,
			wantedResult:   1,
		},
		{
			name:           "2",
			payloadSize:    3,
			segmentIdx:     2,
			maxSegmentSize: 1,
			wantedResult:   1,
		},
		{
			name:           "3",
			payloadSize:    3,
			segmentIdx:     3,
			maxSegmentSize: 1,
			wantedResult:   1,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &GfSpPieceOp{}
			result := p.SegmentPieceSize(tt.payloadSize, tt.segmentIdx, tt.maxSegmentSize)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpPieceOp_ECPieceSize(t *testing.T) {
	p := &GfSpPieceOp{}
	result := p.ECPieceSize(3, 3, 1, 4)
	assert.Equal(t, int64(1), result)
}

func TestGfSpPieceOp_SegmentPieceCount(t *testing.T) {
	p := &GfSpPieceOp{}
	result := p.SegmentPieceCount(3, 3)
	assert.Equal(t, uint32(1), result)
}

func TestGfSpPieceOp_ParseSegmentIdx(t *testing.T) {
	cases := []struct {
		name         string
		segmentKey   string
		wantedResult uint32
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:         "1",
			segmentKey:   "s1_s1",
			wantedResult: 1,
			wantedIsErr:  false,
		},
		{
			name:         "2",
			segmentKey:   "s",
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErrStr: "invalid segmentKey format",
		},
		{
			name:         "3",
			segmentKey:   "s1_s4294967395",
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErrStr: "value out of range",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &GfSpPieceOp{}
			result, err := p.ParseSegmentIdx(tt.segmentKey)
			assert.Equal(t, tt.wantedResult, result)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpPieceOp_ParseECPieceKeyIdx(t *testing.T) {
	cases := []struct {
		name          string
		ecPieceKey    string
		wantedResult1 uint32
		wantedResult2 int32
		wantedIsErr   bool
		wantedErrStr  string
	}{
		{
			name:          "1",
			ecPieceKey:    "e1_s1_p2",
			wantedResult1: 1,
			wantedResult2: 2,
			wantedIsErr:   false,
		},
		{
			name:          "2",
			ecPieceKey:    "s",
			wantedResult1: 0,
			wantedResult2: 0,
			wantedIsErr:   true,
			wantedErrStr:  "invalid EC piece key",
		},
		{
			name:          "3",
			ecPieceKey:    "s1_s4294967395_p2",
			wantedResult1: 0,
			wantedResult2: 0,
			wantedIsErr:   true,
			wantedErrStr:  "value out of range",
		},
		{
			name:          "4",
			ecPieceKey:    "s1_s1_p4294967395",
			wantedResult1: 0,
			wantedResult2: 0,
			wantedIsErr:   true,
			wantedErrStr:  "value out of range",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &GfSpPieceOp{}
			result1, result2, err := p.ParseECPieceKeyIdx(tt.ecPieceKey)
			assert.Equal(t, tt.wantedResult1, result1)
			assert.Equal(t, tt.wantedResult2, result2)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpPieceOp_ParseChallengeIdx(t *testing.T) {
	cases := []struct {
		name          string
		challengeKey  string
		wantedResult1 uint32
		wantedResult2 int32
		wantedIsErr   bool
		wantedErrStr  string
	}{
		{
			name:          "1",
			challengeKey:  "s1_s1",
			wantedResult1: 1,
			wantedResult2: -1,
			wantedIsErr:   false,
		},
		{
			name:          "2",
			challengeKey:  "s1_s4294967395",
			wantedResult1: 0,
			wantedResult2: 0,
			wantedIsErr:   true,
			wantedErrStr:  "value out of range",
		},
		{
			name:          "3",
			challengeKey:  "s1_s2_p2",
			wantedResult1: 2,
			wantedResult2: 2,
			wantedIsErr:   false,
		},
		{
			name:          "4",
			challengeKey:  "2",
			wantedResult1: 0,
			wantedResult2: 0,
			wantedIsErr:   true,
			wantedErrStr:  "invalid challenge key",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &GfSpPieceOp{}
			result1, result2, err := p.ParseChallengeIdx(tt.challengeKey)
			assert.Equal(t, tt.wantedResult1, result1)
			assert.Equal(t, tt.wantedResult2, result2)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
