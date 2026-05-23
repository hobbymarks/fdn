package db

import (
	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	EncryptedPreviousName string `gorm:"unique"`
	HashedCurrentName     string
	Count                 int64 `gorm:"default:1"`
}
