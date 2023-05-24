package bsdb

import (
	"gorm.io/gorm"
)

func ContinuationTokenFilter(continuationToken string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("object_name >= ?", continuationToken)
	}
}

func PrefixFilter(prefix string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("object_name LIKE ?", prefix+"%")
	}
}
