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
	PreviousName    string
	CurrentNameHash string
}

// ConnectRDDB connect config database
func ConnectRDDB(path ...string) *gorm.DB {
	var dbPath string

	// len(paths)!=0
	if len(path) != 0 && utils.PathExist(path[0]) {
		db := utils.OpenDB(path[0])
		return db
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
