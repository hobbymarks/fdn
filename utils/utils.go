package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
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

// HashTo32B hash input to 32 bytes and return
func HashTo32B(key string) []byte {
	h := sha256.New()
	h.Write([]byte(key))
	bs := h.Sum(nil)
	return bs
}

// RandEnAlphDigitShiftDigit return rand string with length is n
func RandEnAlphDigitShiftDigit(n int) string {
	rand.Seed(time.Now().UnixMicro())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// RandEnAlphDigit return rand string with length is n
func RandEnAlphDigit(n int) string {
	rand.Seed(time.Now().UnixMicro())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// RandEnAlph return rand string with length is n
func RandEnAlph(n int) string {
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

// PathMaker make path with random string
func PathMaker(_type string) string {
	var path string

	// defer os.RemoveAll(filepath.Dir(_f.Name()))

	if _type == "f" {
		_f, err := os.CreateTemp("", "fdn"+RandEnAlphDigit(18)+".*")
		if err != nil {
			log.Fatal(err)
		}
		path = _f.Name()
	}
	if _type == "d" {
		_f, err := os.MkdirTemp("", "fdn"+RandEnAlphDigit(32))
		if err != nil {
			log.Fatal(err)
		}
		path = _f
	}
	return path
}

// Ext retrieve path extension if directory or error return empty string
func Ext(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Error(err)
		return ""
	}
	if fileInfo.IsDir() {
		return ""
	} else if fileInfo.Mode().IsRegular() {
		return filepath.Ext(path)
	} else {
		log.Trace("skipped:", path)
		return ""
	}
}

var iv = []byte{
	97,
	70,
	68,
	78,
	105,
	110,
	116,
	101,
	114,
	110,
	97,
	108,
	117,
	115,
	101,
	100,
} /*aFDNinternalused*/

// EncodeBase64 encode bytes with base64 to string
func EncodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeBase64 decode string with base64 to bytes
func DecodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// Encrypt encrypt text with key return string
func Encrypt(key, text string) string {
	block, err := aes.NewCipher(HashTo32B(key))
	if err != nil {
		panic(err)
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return EncodeBase64(ciphertext)
}

// Decrypt decrypt text with key return string
func Decrypt(key, text string) string {
	block, err := aes.NewCipher(HashTo32B(key))
	if err != nil {
		panic(err)
	}
	ciphertext := DecodeBase64(text)
	cfb := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext)
}

// DBClose close the gorm.DB via sqlDB
func DBClose(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	err = sqlDB.Close()
	if err != nil {
		log.Fatal(err)
	}
}
