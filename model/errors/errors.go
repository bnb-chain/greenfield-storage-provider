package errors

import "errors"

// errors
var (
	NotSupportedMethod    = errors.New("Not supported method")
	NotSupportedDelimiter = errors.New("Not supported delimiter")
	EmptyObjectKey        = errors.New("Object key cannot be empty")
	EmptyMemoryObject     = errors.New("Memory object is empty")
	BucketNotExisted      = errors.New("Bucket not existed")
)
