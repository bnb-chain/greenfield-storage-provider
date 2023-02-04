package stonenode

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeSegmentPieceKey(t *testing.T) {
	cases := []struct {
		name       string
		req1       int
		req2       string
		wantedResp string
	}{
		{
			name:       "encode segment piece key successfully",
			req1:       7,
			req2:       "AB875UT96",
			wantedResp: "7_AB875UT96",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := encodeSPKey(tt.req1, tt.req2)
			assert.Equal(t, tt.wantedResp, resp)
		})
	}
}

func TestDecodeSegmentPieceKey(t *testing.T) {
	cases := []struct {
		name        string
		req         string
		wantedResp1 uint32
		wantedErr   error
	}{
		{
			name:        "decode segment piece key successfully",
			req:         "12_S065BMTE",
			wantedResp1: 12,
			wantedErr:   nil,
		},
		{
			name:        "invalid piece key 1",
			req:         "testID_s2",
			wantedResp1: 0,
			wantedErr:   fmt.Errorf("invalid sp key"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp1, err := decodeSPKey(tt.req)
			assert.Equal(t, tt.wantedResp1, resp1)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}
