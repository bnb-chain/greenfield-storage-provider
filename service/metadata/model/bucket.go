package model

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Bucket struct {
	gorm.Model
	UserID uuid.UUID `gorm:"uniqueIndex:user_id_key;type:varchar(255)"`
}
