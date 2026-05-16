package cmd

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

func ConfigTermWords(keyValueMap map[string]string) error {
	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}
	for key, value := range keyValueMap {
		keyHash := utils.KeyHash(strings.ToLower(key))
		termWord := db.TermWord{
			KeyHash:       keyHash,
			OriginalLower: strings.ToLower(key),
			TargetWord:    value,
			Source:        db.UserSource,
		}
		var count int64 = 0
		conn.Model(&db.TermWord{}).
			Where("key_hash = ?", termWord.KeyHash).
			Count(&count)
		if count == 0 {
			result := conn.Create(&termWord)
			if result.Error != nil {
				log.Error(result.Error)
			}
		} else {
			result := conn.Model(&db.TermWord{}).
				Where("key_hash = ?", termWord.KeyHash).
				Updates(map[string]interface{}{
					"target_word": value,
					"source":      db.UserSource,
				})
			if result.Error != nil {
				log.Error(result.Error)
			}
			log.Debugf("updated:%s -> %s", termWord.OriginalLower, value)
		}
	}
	invalidateTermWordRegexCache()
	return nil
}

func DeleteTermWords(keys []string) error {
	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}
	for _, key := range keys {
		keyHash := utils.KeyHash(key)
		var tw db.TermWord
		if result := conn.Where("key_hash = ?", keyHash).First(&tw); result.Error == nil {
			if tw.Source == db.BuiltinSource {
				log.Infof("removing built-in term: %s (will be restored on next sync)", tw.OriginalLower)
			}
		}
		result := conn.Unscoped().Delete(&db.TermWord{}, keyHash)
		if result.Error != nil {
			log.Error(result.Error)
		}
	}
	invalidateTermWordRegexCache()
	return nil
}

func ConfigToSepWords(words []string) error {
	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}
	for _, word := range words {
		keyHash := utils.KeyHash(word)
		toSepWord := db.ToSepWord{KeyHash: keyHash, Value: word, Source: db.UserSource}
		var count int64 = 0
		conn.Model(&db.ToSepWord{}).
			Where("key_hash = ?", toSepWord.KeyHash).
			Count(&count)
		if count == 0 {
			result := conn.Create(&toSepWord)
			if result.Error != nil {
				log.Error(result.Error)
			}
		} else {
			result := conn.Model(&db.ToSepWord{}).
				Where("key_hash = ?", toSepWord.KeyHash).
				Updates(map[string]interface{}{
					"value":  word,
					"source": db.UserSource,
				})
			if result.Error != nil {
				log.Error(result.Error)
			}
			log.Debugf("updated sepword:%s", word)
		}
	}
	return nil
}

func DeleteToSepWords(keys []string) error {
	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}
	for _, key := range keys {
		keyHash := utils.KeyHash(key)
		var sw db.ToSepWord
		if result := conn.Where("key_hash = ?", keyHash).First(&sw); result.Error == nil {
			if sw.Source == db.BuiltinSource {
				log.Infof("removing built-in sepword: %s (will be restored on next sync)", sw.Value)
			}
		}
		result := conn.Unscoped().Delete(&db.ToSepWord{}, keyHash)
		if result.Error != nil {
			log.Error(result.Error)
		}
	}
	return nil
}

func ConfigSeparator(separator string) error {
	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}
	var existing db.Separator
	result := conn.First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			sep := db.Separator{
				KeyHash: utils.KeyHash(separator),
				Value:   separator,
				Source:  db.UserSource,
			}
			createResult := conn.Create(&sep)
			if createResult.Error != nil {
				log.Error(createResult.Error)
			}
		} else {
			return result.Error
		}
	} else {
		updateResult := conn.Model(&existing).Updates(map[string]interface{}{
			"key_hash": utils.KeyHash(separator),
			"value":    separator,
			"source":   db.UserSource,
		})
		if updateResult.Error != nil {
			log.Error(updateResult.Error)
		}
		log.Debugf("updated separator:%s", separator)
	}
	return nil
}
