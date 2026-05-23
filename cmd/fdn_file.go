package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

func FDNFile(currentPath string, toBePath string, reversed bool) error {
	toBase := filepath.Base(toBePath)
	curBase := filepath.Base(currentPath)

	if err := os.Rename(currentPath, toBePath); err != nil {
		slog.Error(err.Error())
		return err
	}

	conn, err := db.ConnectCFGDB()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	var journalErr error
	if !reversed {
		encPrev, err := utils.Encrypt(toBase, curBase)
		if err != nil {
			journalErr = err
		} else {
			rec := db.Record{
				EncryptedPreviousName: encPrev,
				HashedCurrentName:     utils.KeyHash(toBase),
			}
			journalErr = AddRecord(conn, rec)
		}
	} else {
		encPrev, err := utils.Encrypt(curBase, toBase)
		if err != nil {
			journalErr = err
		} else {
			rec := db.Record{
				EncryptedPreviousName: encPrev,
				HashedCurrentName:     utils.KeyHash(curBase),
			}
			journalErr = DeleteRecord(conn, rec)
		}
	}
	if journalErr != nil {
		if rb := os.Rename(toBePath, currentPath); rb != nil {
			return fmt.Errorf("%w; rename rollback failed: %v", journalErr, rb)
		}
		slog.Error(journalErr.Error())
		return journalErr
	}
	return nil
}

func AddRecord(conn *gorm.DB, rec db.Record) error {
	var existing db.Record
	result := conn.First(
		&existing,
		"encrypted_previous_name = ? AND hashed_current_name = ?",
		rec.EncryptedPreviousName,
		rec.HashedCurrentName,
	)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return conn.Create(&rec).Error
		}
		return result.Error
	}
	existing.Count++
	return conn.Save(&existing).Error
}

func DeleteRecord(conn *gorm.DB, rec db.Record) error {
	var existing db.Record
	result := conn.First(
		&existing,
		"encrypted_previous_name = ? AND hashed_current_name = ?",
		rec.EncryptedPreviousName,
		rec.HashedCurrentName,
	)
	if result.Error != nil {
		return result.Error
	}
	existing.Count--
	if existing.Count != 0 {
		return conn.Save(&existing).Error
	}
	return conn.Unscoped().Delete(&existing).Error
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
				slog.Error(err.Error())
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
						slog.Error(err.Error())
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
			slog.Error(err.Error())
			return err
		}
		OutputResult(currentPath, toBePath, true, fullpath)
	}
	return nil
}
