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

func RemovedFilter() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("removed = ?", false)
	}
}
