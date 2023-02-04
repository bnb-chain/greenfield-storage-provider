package piecestore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeSegmentPieceKey(t *testing.T) {
	cases := []struct {
		name       string
		req1       uint64
		req2       uint32
		wantedResp string
	}{
		{
			name:       "encode segment piece key successfully",
			req1:       123456789,
			req2:       0,
			wantedResp: "123456789_s0",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := EncodeSegmentPieceKey(tt.req1, tt.req2)
			assert.Equal(t, tt.wantedResp, resp)
		})
	}
}

func TestDecodeSegmentPieceKey(t *testing.T) {
	cases := []struct {
		name        string
		req         string
		wantedResp1 uint64
		wantedResp2 uint32
		wantedErr   error
	}{
		{
			name:        "decode segment piece key successfully",
			req:         "123456789_s0",
			wantedResp1: 123456789,
			wantedResp2: 0,
			wantedErr:   nil,
		},
		{
			name:        "invalid piece key 1",
			req:         "testID_s2",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedErr:   fmt.Errorf("invalid segment piece key"),
		},
		{
			name:        "invalid piece key 2",
			req:         "123456789_p",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedErr:   fmt.Errorf("invalid segment piece key"),
		},
		{
			name:        "invalid piece key 3",
			req:         "123456789_s123r",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedErr:   fmt.Errorf("invalid segment piece key"),
		},
		{
			name:        "invalid piece key 4",
			req:         "123456789_ss.123",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedErr:   fmt.Errorf("invalid segment piece key"),
		},
		{
			name:        "invalid segment piece key 4",
			req:         "123456789_s123/111",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedErr:   fmt.Errorf("invalid segment piece key"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp1, resp2, err := DecodeSegmentPieceKey(tt.req)
			assert.Equal(t, tt.wantedResp1, resp1)
			assert.Equal(t, tt.wantedResp2, resp2)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestEncodeECPieceKey(t *testing.T) {
	cases := []struct {
		name       string
		req1       uint64
		req2       uint32
		req3       uint32
		wantedResp string
	}{
		{
			name:       "encode ec piece key successfully",
			req1:       123456789,
			req2:       1,
			req3:       3,
			wantedResp: "123456789_s1_p3",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := EncodeECPieceKey(tt.req1, tt.req2, tt.req3)
			assert.Equal(t, tt.wantedResp, resp)
		})
	}
}

func TestDecodeECPieceKey(t *testing.T) {
	cases := []struct {
		name        string
		req         string
		wantedResp1 uint64
		wantedResp2 uint32
		wantedResp3 uint32
		wantedErr   error
	}{
		{
			name:        "encode ec piece key successfully",
			req:         "123456789_s1_p3",
			wantedResp1: 123456789,
			wantedResp2: 1,
			wantedResp3: 3,
			wantedErr:   nil,
		},
		{
			name:        "invalid ec piece key 1",
			req:         "ABCD123_s1",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedResp3: 0,
			wantedErr:   fmt.Errorf("invalid EC piece key"),
		},
		{
			name:        "invalid ec piece key 2",
			req:         "123456789_s1_p",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedResp3: 0,
			wantedErr:   fmt.Errorf("invalid EC piece key"),
		},
		{
			name:        "invalid ec piece key 3",
			req:         "123456789_s1_ps2",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedResp3: 0,
			wantedErr:   fmt.Errorf("invalid EC piece key"),
		},
		{
			name:        "invalid ec piece key 4",
			req:         "123456789_s1_p2_p3",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedResp3: 0,
			wantedErr:   fmt.Errorf("invalid EC piece key"),
		},
		{
			name:        "invalid ec piece key 5",
			req:         "123456789_s1_p2n",
			wantedResp1: 0,
			wantedResp2: 0,
			wantedResp3: 0,
			wantedErr:   fmt.Errorf("invalid EC piece key"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp1, resp2, resp3, err := DecodeECPieceKey(tt.req)
			assert.Equal(t, tt.wantedResp1, resp1)
			assert.Equal(t, tt.wantedResp2, resp2)
			assert.Equal(t, tt.wantedResp3, resp3)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestCheckSegmentPieceKey(t *testing.T) {
	cases := []struct {
		name       string
		req        []string
		wantedResp bool
	}{
		{
			name:       "check segment piece key successfully",
			req:        []string{"123456789", "s3"},
			wantedResp: true,
		},
		{
			name:       "invalid segment piece key 1",
			req:        []string{"ABCD", "s3", "y7"},
			wantedResp: false,
		},
		{
			name:       "invalid segment piece key 2",
			req:        []string{"123456789", "s3.."},
			wantedResp: false,
		},
		{
			name:       "invalid segment piece key 3",
			req:        []string{"123456789", "s3m2"},
			wantedResp: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := CheckSegmentPieceKey(tt.req)
			assert.Equal(t, tt.wantedResp, resp)
		})
	}
}

func TestCheckECPieceKey(t *testing.T) {
	cases := []struct {
		name       string
		req        []string
		wantedResp bool
	}{
		{
			name:       "check ec piece key successfully",
			req:        []string{"123456789", "s3", "p5"},
			wantedResp: true,
		},
		{
			name:       "invalid ec piece key 1",
			req:        []string{"ABCD", "s3", "p7"},
			wantedResp: false,
		},
		{
			name:       "invalid ec piece key 2",
			req:        []string{"123456789", "s3..", "p5/"},
			wantedResp: false,
		},
		{
			name:       "invalid ec piece key 3",
			req:        []string{"123456789", "s3m2", "p5*/"},
			wantedResp: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := CheckECPieceKey(tt.req)
			assert.Equal(t, tt.wantedResp, resp)
		})
	}
}
