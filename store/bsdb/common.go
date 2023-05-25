package bsdb

// ListObjectsResult represents the result of a List Objects operation.
type ListObjectsResult struct {
	PathName   string
	ResultType string
	*Object
}
