package utils

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OpenDB only open the db
func OpenDB(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// KeyHash create hash from key and return string
func KeyHash(key string) string {
	data := []byte(key)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// RandStr return rand string with length is n
func RandStr(n int) string {
	rand.Seed(time.Now().UnixMicro())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// PathExist check path exist
func PathExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Error(err)
		return false
	}
	return true
}

// DatabaseDir get database directory
func DatabaseDir() string {
	homeDir, err := os.UserHomeDir() //get home dir
	if err != nil {
		path, err := os.Executable()
		if err != nil {
			log.Fatalf("Get DatabaseDir Failed:%s", err)
		}
		return path
	}
	return homeDir
}

// FDNDir get fdn data directory
func FDNDir() string {
	dbDir := DatabaseDir()
	fdnDir := filepath.Join(dbDir, ".fdn")
	_, err := os.Lstat(fdnDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(fdnDir, os.ModePerm); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	return fdnDir
}
