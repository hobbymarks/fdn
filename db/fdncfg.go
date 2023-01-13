package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type TermWord struct {
	KeyHash       string
	OriginalLower string
	Value         string
}

type ToSepWord struct {
	KeyHash string
	Value   string
}

type Separator struct {
	KeyHash string
	Value   string
}

// OpenDB ...
func OpenDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("fdn.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db, nil
}
