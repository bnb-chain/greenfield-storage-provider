package storage

import (
	"errors"
)

// piece store errors
var (
	// ErrNoSuchObject defines not existed object error
	ErrNoSuchObject = errors.New("the specified key does not exist")
	// ErrNoSuchBucket defines not existed bucket error
	ErrNoSuchBucket = errors.New("the specified bucket does not exist")
	// ErrUnsupportedDelimiter defines invalid key with delimiter error
	ErrUnsupportedDelimiter = errors.New("unsupported delimiter")
	// ErrInvalidObjectKey defines invalid object key error
	ErrInvalidObjectKey = errors.New("invalid object key")
	// ErrUnsupportedMethod defines unsupported method error
	ErrUnsupportedMethod = errors.New("unsupported method")
	// ErrNoPermissionAccessBucket defines deny access bucket error
	ErrNoPermissionAccessBucket = errors.New("deny access bucket")
)
