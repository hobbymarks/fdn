/*
Package utils provides database, crypto, and path helpers for fdn.
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenDB(path string) (*gorm.DB, error) {
	conn, err := gorm.Open(
		sqlite.Open(path),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func KeyHash(key string) string {
	data := []byte(key)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func HashTo32B(key string) []byte {
	h := sha256.New()
	h.Write([]byte(key))
	bs := h.Sum(nil)
	return bs
}

func RandEnAlphDigitShiftDigit(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandEnAlphDigit(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandEnAlph(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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

func PathIsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func DBBaseDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		path, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("get database dir failed: %w", err)
		}
		return path, nil
	}
	return homeDir, nil
}

func FDNDir() (string, error) {
	dbBaseDir, err := DBBaseDir()
	if err != nil {
		return "", err
	}
	fdnDir := filepath.Join(dbBaseDir, ".fdn")
	if _, err := os.Lstat(fdnDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(fdnDir, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return fdnDir, nil
}

func PathMaker(typ string) (string, error) {
	var path string

	if typ == "f" {
		tmp, err := os.CreateTemp("", "fdn"+RandEnAlphDigit(18)+".*")
		if err != nil {
			return "", err
		}
		path = tmp.Name()
	}
	if typ == "d" {
		tmp, err := os.MkdirTemp("", "fdn"+RandEnAlphDigit(32))
		if err != nil {
			return "", err
		}
		path = tmp
	}
	return path, nil
}

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

func EncodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func DecodeBase64(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Encrypt(key, text string) (string, error) {
	block, err := aes.NewCipher(HashTo32B(key))
	if err != nil {
		return "", err
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return EncodeBase64(ciphertext), nil
}

func Decrypt(key, text string) (string, error) {
	block, err := aes.NewCipher(HashTo32B(key))
	if err != nil {
		return "", err
	}
	ciphertext, err := DecodeBase64(text)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}

func DBClose(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func FileMD5(filePath string) (string, error) {
	var md5s string

	file, err := os.Open(filePath)
	if err != nil {
		return md5s, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return md5s, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	md5s = hex.EncodeToString(hashInBytes)

	return md5s, nil
}

func SameFiles(
	firstPath string,
	secondPath string,
	morePaths ...string,
) (bool, error) {
	fHash, err := FileMD5(firstPath)
	if err != nil {
		return false, err
	}
	sHash, err := FileMD5(secondPath)
	if err != nil {
		return false, err
	}
	if len(morePaths) == 0 {
		if strings.Compare(fHash, sHash) == 0 {
			return true, nil
		}
		return false, nil
	} else {
		if strings.Compare(fHash, sHash) != 0 {
			return false, nil
		}
		for _, path := range morePaths {
			aHash, err := FileMD5(path)
			if err != nil {
				return false, err
			}
			if strings.Compare(sHash, aHash) != 0 {
				return false, nil
			}
		}
		return true, nil
	}
}
