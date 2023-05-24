package bsdb

// ListObjectsResult represents the result of a List Objects operation.
type ListObjectsResult struct {
	PathName   string `gorm:"path_name"`
	ResultType string `gorm:"result_type"`
	Object
}
