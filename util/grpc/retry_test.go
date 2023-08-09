package grpc

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultGRPCRetryPolicy(t *testing.T) {
	option, err := GetDefaultGRPCRetryPolicy("mockService")
	assert.NotNil(t, option)
	assert.Equal(t, nil, err)
}

func TestSetGRPCRetryPolicy(t *testing.T) {
	cases := []struct {
		name        string
		retryConfig *RetryConfig
		wantedIsErr bool
	}{
		{
			name: "Set gRPC retry policy correctly",
			retryConfig: &RetryConfig{MethodConfig: []*MethodConfig{
				{
					Name: []Name{{
						Service: "mockService",
						Method:  "mockMethod",
					}},
					WaitForReady:            true,
					Timeout:                 "100",
					MaxRequestMessageBytes:  1024,
					MaxResponseMessageBytes: 1024,
				},
			}},
			wantedIsErr: false,
		},
		{
			name: "Invalid retry config struct",
			retryConfig: &RetryConfig{MethodConfig: []*MethodConfig{
				{
					Name: []Name{{
						Service: "mockService",
						Method:  "mockMethod",
					}},
					RetryPolicy: &RetryPolicy{
						BackoffMultiplier: math.NaN(),
					},
				},
			}},
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetGRPCRetryPolicy(tt.retryConfig)
			if tt.wantedIsErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetGRPCErrorCodes(t *testing.T) {
	codes := GetGRPCErrorCodes()
	assert.Equal(t, 16, len(codes))
}
