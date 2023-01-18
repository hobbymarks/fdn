/*
Package db fdn cfg
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

// TermWord Term Word
type TermWord struct {
	gorm.Model
	KeyHash       string `gorm:"unique"`
	OriginalLower string
	TargetWord    string
}

// ToSepWord will be change to separator
type ToSepWord struct {
	gorm.Model
	KeyHash string `gorm:"unique"`
	Value   string
}

// Separator separator
type Separator struct {
	gorm.Model
	KeyHash string `gorm:"unique"`
	Value   string
}

// ConnectCFGDB connect config database
func ConnectCFGDB(path ...string) *gorm.DB {
	var dbPath string

	// len(paths)!=0
	if len(path) != 0 && utils.PathExist(path[0]) {
		dbPath = path[0]
	}
	// len(paths)==0
	if len(path) == 0 {
		fdnDir := utils.FDNDir()
		dbPath = filepath.Join(fdnDir, "cfg.db")
	} else {
		// path not exist
		if err := os.MkdirAll(filepath.Dir(path[0]), os.ModePerm); err != nil {
			log.Fatal(err)
		}
		dbPath = path[0]
	}
	// open _db and init
	_db := utils.OpenDB(dbPath)
	err := _db.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{})
	if err != nil {
		log.Fatal(err)
	}
	return _db
}
