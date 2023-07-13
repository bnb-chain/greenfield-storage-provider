package bsdb

import (
	"github.com/forbole/juno/v4/common"
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

func PathNameFilter(pathName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path_name = ?", pathName)
	}
}

func NameFilter(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name like ?", name+"%")
	}
}

func FullNameFilter(fullName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("full_name >= ?", fullName)
	}
}

func SourceTypeFilter(sourceType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source_type = ?", sourceType)
	}
}

func RemovedFilter(removed bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("removed = ?", removed)
	}
}

func BucketIDStartAfterFilter(bucketID common.Hash) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bucket_id > ?", bucketID)
	}
}

func ObjectIDStartAfterFilter(objectID common.Hash) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("object_id > ?", objectID)
	}
}

func WithLimit(limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}
