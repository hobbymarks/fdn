package db

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

var (
	sharedDB   *gorm.DB
	sharedMu   sync.Mutex
)

func ConnectDB(path ...string) (*gorm.DB, error) {
	var dbPath string
	if len(path) == 0 {
		dbPath = DefaultFDNDBPath()
	} else {
		if err := os.MkdirAll(filepath.Dir(path[0]), os.ModePerm); err != nil {
			return nil, err
		}
		dbPath = path[0]
	}

	conn, err := utils.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}
	if err := conn.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{}, &Record{}); err != nil {
		return nil, err
	}
	return conn, nil
}

func GetDB() *gorm.DB {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	return sharedDB
}

func SetDB(conn *gorm.DB) {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	sharedDB = conn
}

func CloseDB() error {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	if sharedDB != nil {
		if err := utils.DBClose(sharedDB); err != nil {
			return err
		}
		sharedDB = nil
	}
	return nil
}

func ResetSharedDB() {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	sharedDB = nil
}

func InitDB(path string) error {
	conn, err := ConnectDB(path)
	if err != nil {
		return err
	}
	SetDB(conn)
	return nil
}
