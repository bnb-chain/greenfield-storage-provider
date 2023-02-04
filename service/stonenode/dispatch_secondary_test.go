package stonenode

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_encodeSPKey(t *testing.T) {
	cases := []struct {
		name       string
		req1       int
		req2       string
		wantedResp string
	}{
		{
			name:       "encode sp key successfully",
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

func Test_decodeSPKey(t *testing.T) {
	cases := []struct {
		name        string
		req         string
		wantedResp1 uint32
		wantedResp2 string
		wantedErr   error
	}{
		{
			name:        "decode sp key successfully",
			req:         "12_S065BMTE",
			wantedResp1: 12,
			wantedResp2: "S065BMTE",
			wantedErr:   nil,
		},
		{
			name:        "invalid sp key 1",
			req:         "testID_s2",
			wantedResp1: 0,
			wantedResp2: "",
			wantedErr:   fmt.Errorf("invalid sp key"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp1, resp2, err := decodeSPKey(tt.req)
			assert.Equal(t, tt.wantedResp1, resp1)
			assert.Equal(t, tt.wantedResp2, resp2)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}
