package bsdb

type ListObjectsResult struct {
	object     Object
	ResultType string `gorm:"result_type"`
}
