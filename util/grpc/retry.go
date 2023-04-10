package grpc

import (
	"encoding/json"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// RetryConfig defines the retry config of gRPC
//
// An example of gRPC retry config in json format:
//
//	`{
//		"methodConfig": [{
//		"name": [{"service": "service.uploader.types.UploaderService"}],
//		"waitForReady": true,
//		"retryPolicy": {
//			"maxAttempts": 3,
//			"initialBackoff": ".2s",
//			"maxBackoff": "20s",
//			"backoffMultiplier": 2,
//			"retryableStatusCodes": ["CANCELLED"]
//		}}]
//	}`
type RetryConfig struct {
	MethodConfig []*MethodConfig `json:"methodConfig"`
}

type MethodConfig struct {
	Name []Name `json:"name"`
	// WaitForReady indicates whether RPCs sent to this method should wait until
	// the connection is ready by default (!failfast). The value specified via the
	// gRPC client API will override the value set here.
	WaitForReady bool `json:"waitForReady,omitempty"`
	// Timeout is the default timeout for RPCs sent to this method. The actual
	// deadline used will be the minimum of the value specified here and the value
	// set by the application via the gRPC client API.  If either one is not set,
	// then the other will be used.  If neither is set, then the RPC has no deadline.
	Timeout string `json:"timeout,omitempty"`
	// MaxRequestMessageBytes is the maximum allowed payload size for an individual request in a
	// stream (client->server) in bytes. The size which is measured is the serialized
	// payload after per-message compression (but before stream compression) in bytes.
	// The actual value used is the minimum of the value specified here and the value set
	// by the application via the gRPC client API. If either one is not set, then the other
	// will be used.  If neither is set, then the built-in default is used.
	MaxRequestMessageBytes int `json:"maxRequestMessageBytes,omitempty"`
	// MaxResponseMessageBytes is the maximum allowed payload size for an individual response in a
	// stream (server->client) in bytes.
	MaxResponseMessageBytes int `json:"maxResponseMessageBytes,omitempty"`
	// RetryPolicy configures retry options for the method.
	RetryPolicy *RetryPolicy `json:"retryPolicy"`
}

// Name describes a gRPC service
type Name struct {
	Service string `json:"service"`          // Service defines gRPC service name
	Method  string `json:"method,omitempty"` // Method defines gRPC method name
}

// MaxAttempts is the maximum number of attempts, including the original RPC.
//
// This field is required and must be two or greater.
type RetryPolicy struct {
	// MaxAttempts is the maximum number of attempts, including the original RPC.
	//
	// This field is required and must be two or greater.
	MaxAttempts int `json:"maxAttempts"`
	// Exponential backoff parameters. The initial retry attempt will occur at
	// random(0, initialBackoff). In general, the nth attempt will occur at
	// random(0,
	//   min(initialBackoff*backoffMultiplier**(n-1), maxBackoff)).
	//
	// These fields are required and must be greater than zero.
	InitialBackoff string `json:"initialBackoff"`
	// MaxBackoff sets max backoff
	MaxBackoff string `json:"maxBackoff"`
	// BackoffMultiplier sets backoff multiplier
	BackoffMultiplier float64 `json:"backoffMultiplier"`
	// The set of status codes which may be retried.
	//
	// Status codes are specified as strings, e.g., "UNAVAILABLE".
	//
	// This field is required and must be non-empty.
	// Note: a set is used to store this for easy lookup.
	RetryableStatusCodes []string `json:"retryableStatusCodes"`
}

// If a RetryThrottling is provided, gRPC will automatically throttle
// retry attempts and hedged RPCs when the client’s ratio of failures to
// successes exceeds a threshold.
type RetryThrottling struct {
	// The number of tokens starts at maxTokens. The token_count will always be
	// between 0 and maxTokens.
	//
	// This field is required and must be in the range (0, 1000].  Up to 3
	// decimal places are supported
	MaxTokens int `json:"maxTokens"`
	// The amount of tokens to add on each successful RPC. Typically, this will
	// be some number between 0 and 1, e.g., 0.1.
	//
	// This field is required and must be greater than zero. Up to 3 decimal
	// places are supported.
	TokenRatio int `json:"tokenRatio"`
}

// GetDefaultGRPCRetryPolicy returns default gRPC retry policy
func GetDefaultGRPCRetryPolicy(service string) (grpc.DialOption, error) {
	return SetGRPCRetryPolicy(&RetryConfig{
		MethodConfig: []*MethodConfig{
			{
				Name:         []Name{{Service: service}},
				WaitForReady: true,
				Timeout:      "30s",
				RetryPolicy: &RetryPolicy{
					MaxAttempts:          3,
					InitialBackoff:       ".2s",
					MaxBackoff:           "20s",
					BackoffMultiplier:    2,
					RetryableStatusCodes: GetGRPCErrorCodes(),
				},
			},
		}})
}

// SetGRPCRetryPolicy sets gRPC retry policy
func SetGRPCRetryPolicy(retryConfig *RetryConfig) (grpc.DialOption, error) {
	data, err := json.Marshal(retryConfig)
	if err != nil {
		log.Errorw("failed to use json marshal method config", "error", err)
		return nil, err
	}
	return grpc.WithDefaultServiceConfig(string(data)), nil
}

// GetGRPCErrorCodes returns all gRPC error codes
func GetGRPCErrorCodes() []string {
	return []string{"CANCELLED", "UNKNOWN", "INVALID_ARGUMENT", "DEADLINE_EXCEEDED", "NOT_FOUND", "ALREADY_EXISTS",
		"PERMISSION_DENIED", "RESOURCE_EXHAUSTED", "FAILED_PRECONDITION", "ABORTED", "OUT_OF_RANGE", "UNIMPLEMENTED",
		"INTERNAL", "UNAVAILABLE", "DATA_LOSS", "UNAUTHENTICATED"}
}
