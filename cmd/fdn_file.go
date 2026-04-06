package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

func FDNFile(currentPath string, toBePath string, reversed bool) error {
	_to := filepath.Base(toBePath)
	_cur := filepath.Base(currentPath)

	if err := os.Rename(currentPath, toBePath); err != nil {
		log.Error(err)
		return err
	}

	_db := db.ConnectRDDB()
	defer utils.DBClose(_db)
	var journalErr error
	if !reversed {
		_rd := db.Record{
			EncryptedPreviousName: utils.Encrypt(_to, _cur),
			HashedCurrentName:     utils.KeyHash(_to),
		}
		journalErr = AddRecord(_db, _rd)
	} else {
		_rd := db.Record{
			EncryptedPreviousName: utils.Encrypt(_cur, _to),
			HashedCurrentName:     utils.KeyHash(_cur),
		}
		journalErr = DeleteRecord(_db, _rd)
	}
	if journalErr != nil {
		if rb := os.Rename(toBePath, currentPath); rb != nil {
			return fmt.Errorf("%w; rename rollback failed: %v", journalErr, rb)
		}
		log.Error(journalErr)
		return journalErr
	}
	return nil
}

func AddRecord(_db *gorm.DB, _rd db.Record) error {
	var rd db.Record
	rlt := _db.First(
		&rd,
		"encrypted_previous_name = ? AND hashed_current_name = ?",
		_rd.EncryptedPreviousName,
		_rd.HashedCurrentName,
	)
	if rlt.Error != nil {
		if errors.Is(rlt.Error, gorm.ErrRecordNotFound) {
			return _db.Create(&_rd).Error
		}
		return rlt.Error
	}
	rd.Count++
	return _db.Save(&rd).Error
}

func DeleteRecord(_db *gorm.DB, _rd db.Record) error {
	var rd db.Record
	rlt := _db.First(
		&rd,
		"encrypted_previous_name = ? AND hashed_current_name = ?",
		_rd.EncryptedPreviousName,
		_rd.HashedCurrentName,
	)
	if rlt.Error != nil {
		return rlt.Error
	}
	rd.Count--
	if rd.Count != 0 {
		return _db.Save(&rd).Error
	}
	rd.Count++
	return _db.Unscoped().Delete(&rd).Error
}

func CheckDoFDN(
	currentPath string,
	toBePath string,
	reverse bool,
	overwrite bool,
) error {
	if utils.PathExist(toBePath) {
		if overwrite {
			err := FDNFile(currentPath, toBePath, reverse)
			if err != nil {
				log.Error(err)
				return err
			}
			OutputResult(currentPath, toBePath, true, fullpath)
		} else {
			same, err := utils.SameFiles(currentPath, toBePath)
			if err != nil {
				fmt.Println("[ERROR]Skip:", currentPath)
			} else {
				if same {
					err := FDNFile(currentPath, toBePath, reverse)
					if err != nil {
						log.Error(err)
						return err
					}
					OutputResult(currentPath, toBePath, true, fullpath)
				} else {
					fmt.Println("[EXIST]Skip:", currentPath)
				}
			}
		}
	} else {
		err := FDNFile(currentPath, toBePath, reverse)
		if err != nil {
			log.Error(err)
			return err
		}
		OutputResult(currentPath, toBePath, true, fullpath)
	}
	return nil
}
