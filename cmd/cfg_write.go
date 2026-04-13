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
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for key, value := range keyValueMap {
		_key := strings.ToLower(key)
		_termWord := db.TermWord{
			KeyHash:       utils.KeyHash(_key),
			OriginalLower: _key,
			TargetWord:    value,
			Source:        db.UserSource,
		}
		var _cnt int64 = 0
		_db.Model(&db.TermWord{}).
			Where("key_hash = ?", _termWord.KeyHash).
			Count(&_cnt)
		if _cnt == 0 {
			_rlt := _db.Create(&_termWord)
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
		} else {
			_rlt := _db.Model(&db.TermWord{}).
				Where("key_hash = ?", _termWord.KeyHash).
				Updates(map[string]interface{}{
					"target_word": value,
					"source":      db.UserSource,
				})
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
			log.Debugf("updated:%s -> %s", _termWord.OriginalLower, value)
		}
	}
	invalidateTermWordRegexCache()
	return nil
}

func DeleteTermWords(keys []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, key := range keys {
		_key := utils.KeyHash(key)
		var tw db.TermWord
		if rlt := _db.Where("key_hash = ?", _key).First(&tw); rlt.Error == nil {
			if tw.Source == db.BuiltinSource {
				log.Infof("removing built-in term: %s (will be restored on next sync)", tw.OriginalLower)
			}
		}
		_rlt := _db.Unscoped().Delete(&db.TermWord{}, _key)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	}
	invalidateTermWordRegexCache()
	return nil
}

func ConfigToSepWords(words []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, word := range words {
		_key := utils.KeyHash(word)
		_toSepWord := db.ToSepWord{KeyHash: _key, Value: word, Source: db.UserSource}
		var _cnt int64 = 0
		_db.Model(&db.ToSepWord{}).
			Where("key_hash = ?", _toSepWord.KeyHash).
			Count(&_cnt)
		if _cnt == 0 {
			_rlt := _db.Create(&_toSepWord)
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
		} else {
			_rlt := _db.Model(&db.ToSepWord{}).
				Where("key_hash = ?", _toSepWord.KeyHash).
				Updates(map[string]interface{}{
					"value":  word,
					"source": db.UserSource,
				})
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
			log.Debugf("updated sepword:%s", word)
		}
	}
	return nil
}

func DeleteToSepWords(keys []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, key := range keys {
		_key := utils.KeyHash(key)
		var sw db.ToSepWord
		if rlt := _db.Where("key_hash = ?", _key).First(&sw); rlt.Error == nil {
			if sw.Source == db.BuiltinSource {
				log.Infof("removing built-in sepword: %s (will be restored on next sync)", sw.Value)
			}
		}
		_rlt := _db.Unscoped().Delete(&db.ToSepWord{}, _key)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	}
	return nil
}

func ConfigSeparator(separator string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	var existing db.Separator
	rlt := _db.First(&existing)
	if rlt.Error != nil {
		if errors.Is(rlt.Error, gorm.ErrRecordNotFound) {
			_sep := db.Separator{
				KeyHash: utils.KeyHash(separator),
				Value:   separator,
				Source:  db.UserSource,
			}
			_rlt := _db.Create(&_sep)
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
		} else {
			return rlt.Error
		}
	} else {
		_rlt := _db.Model(&existing).Updates(map[string]interface{}{
			"key_hash": utils.KeyHash(separator),
			"value":    separator,
			"source":   db.UserSource,
		})
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
		log.Debugf("updated separator:%s", separator)
	}
	return nil
}
