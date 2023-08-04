package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/stretchr/testify/assert"
)

func TestShouldRetry(t *testing.T) {
	cases := []struct {
		name         string
		r            *request.Request
		wantedResult bool
	}{
		{
			name: "1",
			r: &request.Request{
				Error: awserr.New("InternalError", "mock error", ErrNoSuchObject),
			},
			wantedResult: true,
		},
		{
			name: "2",
			r: &request.Request{
				Error: awserr.New("UnknownError", "mock error", ErrUnsupportedMethod),
			},
			wantedResult: true,
		},
		{
			name: "3",
			r: &request.Request{
				Error: awserr.New("ExpiredToken", "mock error", ErrInvalidObjectKey),
			},
			wantedResult: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c := newCustomS3Retryer(1, 2)
			result := c.ShouldRetry(tt.r)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func Test_errHasCode(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		code         string
		wantedResult bool
	}{
		{
			name:         "1",
			err:          nil,
			code:         "",
			wantedResult: false,
		},
		{
			name:         "2",
			err:          awserr.New("test", "mock error", ErrNoSuchObject),
			code:         "test",
			wantedResult: true,
		},
		{
			name:         "3",
			err:          ErrNoSuchBucket,
			code:         "",
			wantedResult: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := errHasCode(tt.err, tt.code)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}
