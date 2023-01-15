package storage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// customS3Retryer wraps aws-sdk-go's built-in DefaultRetryer and adding additional error codes
// such as retry for S3 InternalError code
type customS3Retryer struct {
	client.DefaultRetryer
}

func newCustomS3Retryer(maxRetries int, minRetryDelay time.Duration) *customS3Retryer {
	return &customS3Retryer{
		DefaultRetryer: client.DefaultRetryer{
			NumMaxRetries: maxRetries,
			MinRetryDelay: minRetryDelay,
		},
	}
}

// ShouldRetry overrides built-in DefaultRetryer of SDK, adding custom retry
// logics that are not included in the SDK.
func (c *customS3Retryer) ShouldRetry(r *request.Request) bool {
	shouldRetry := errHasCode(r.Error, "InternalError") || errHasCode(r.Error, "RequestTimeTooSkewed") || errHasCode(r.Error, "SlowDown") || strings.Contains(r.Error.Error(), "connection reset") || strings.Contains(r.Error.Error(), "connection timed out")
	if !shouldRetry {
		shouldRetry = c.DefaultRetryer.ShouldRetry(r)
	}

	// Errors related to tokens
	if errHasCode(r.Error, "ExpiredToken") || errHasCode(r.Error, "ExpiredTokenException") || errHasCode(r.Error, "InvalidToken") {
		return false
	}

	if shouldRetry && r.Error != nil {
		err := fmt.Errorf("retryable error: %v", r.Error)
		log.Error(err)
	}

	return shouldRetry
}

func errHasCode(err error, code string) bool {
	if err == nil || code == "" {
		return false
	}
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		if awsErr.Code() == code {
			return true
		}
	}
	return false
}
