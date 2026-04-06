package cmd

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
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
			log.Debugf("skipped:%s", _termWord.OriginalLower)
		}
	}
	return nil
}

func DeleteTermWords(keys []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, key := range keys {
		_key := utils.KeyHash(key)
		_rlt := _db.Unscoped().Delete(&db.TermWord{}, _key)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	}
	return nil
}

func ConfigToSepWords(words []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, word := range words {
		_key := utils.KeyHash(word)
		_toSepWord := db.ToSepWord{KeyHash: _key, Value: word}
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
			log.Debugf("skipped:%s", _toSepWord.Value)
		}
	}
	return nil
}

func DeleteToSepWords(keys []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, key := range keys {
		_key := utils.KeyHash(key)
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
	_sep := db.Separator{KeyHash: utils.KeyHash(separator), Value: separator}
	var _cnt int64 = 0
	_db.Model(&db.Separator{}).
		Where("key_hash = ?", _sep.KeyHash).
		Count(&_cnt)
	if _cnt == 0 {
		_rlt := _db.Create(&_sep)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	} else {
		log.Debugf("skipped:%s", _sep.Value)
	}
	return nil
}
