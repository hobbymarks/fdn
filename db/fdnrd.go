/*
Package db fdn rd
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

// Record a filename changed record
type Record struct {
	gorm.Model
	EncryptedPreviousName string `gorm:"unique"`
	HashedCurrentName     string
}

// ConnectRDDB connect config database
func ConnectRDDB(path ...string) *gorm.DB {
	var dbPath string

	// len(paths)!=0
	if len(path) != 0 && utils.PathExist(path[0]) {
		dbPath = path[0]
	}
	// len(paths)==0
	if len(path) == 0 {
		fdnDir := utils.FDNDir()
		dbPath = filepath.Join(fdnDir, "rd.db")
	} else {
		// path not exist
		if err := os.MkdirAll(filepath.Dir(path[0]), os.ModePerm); err != nil {
			log.Fatal(err)
		}
		dbPath = path[0]
	}
	// open db and init
	db := utils.OpenDB(dbPath)
	err := db.AutoMigrate(&Record{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}
