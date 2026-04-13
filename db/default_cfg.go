package db

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed default_emoji_sepwords.txt
var defaultEmojiSepwordsData []byte

var defaultTermWordDefs = []struct{ orig, target string }{
	{"凤凰网", "ifeng"},
	{"wikipedia", "wikipedia"},
	{"新浪网", "sina"},
	{"自由的百科全书", "_"},
	{"维基百科", "wikipedia"},
	{"_·_", "_"},
	{"搜狐网", "sohu"},
}

var defaultPunctWords = []string{
	"-",      // U+002D
	" ",      // U+0020
	"\u201C", // U+201C "
	"\u201D", // U+201D "
	"\uFF0C", // U+FF0C ，
	"\uFF1F", // U+FF1F ？
	"(",      // U+0028
	")",      // U+0029
	"~",      // U+007E
	"\u3001", // U+3001 、
	"\u3002", // U+3002 。
	"\u30FB", // U+30FB ・
	"\uFF02", // U+FF02 ＂
	"\uFF1A", // U+FF1A ：
	"\u3010", // U+3010 【
	"\u3011", // U+3011 】
	"\u4E28", // U+4E28 丨
	"\uFF5C", // U+FF5C ｜
	"\uFF01", // U+FF01 ！
	"\uFF08", // U+FF08 （
	"\uFF09", // U+FF09 ）
	"\u300A", // U+300A 《
	"\u300B", // U+300B 》
	"<",      // U+003C
	">",      // U+003E
	":",      // U+003A
	"\"",     // U+0022
	"/",      // U+002F
	"\u29F8", // U+29F8 ⧸
	"\\",     // U+005C
	"|",      // U+007C
	"?",      // U+003F
	"*",      // U+002A
	"!",      // U+0021
	"@",      // U+0040
	"#",      // U+0023
	"$",      // U+0024
	"%",      // U+0025
	"^",      // U+005E
	"&",      // U+0026
	"`",      // U+0060
	";",      // U+003B
	",",      // U+002C
	"[",      // U+005B
	"]",      // U+005D
	"{",      // U+007B
	"}",      // U+007D
	"'",      // U+0027
	"+",      // U+002B
	"=",      // U+003D
}

func EnsureDefaultCFG(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	_db := utils.OpenDB(dbPath)
	defer utils.DBClose(_db)
	if err := _db.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{}, &Record{}); err != nil {
		return err
	}
	return _db.Transaction(func(tx *gorm.DB) error {
		if err := migrateExistingDefaults(tx); err != nil {
			return err
		}
		return syncDefaultCFG(tx)
	})
}

func migrateExistingDefaults(tx *gorm.DB) error {
	var twKeys []string
	for _, p := range defaultTermWordDefs {
		twKeys = append(twKeys, utils.KeyHash(strings.ToLower(p.orig)))
	}
	if len(twKeys) > 0 {
		tx.Model(&TermWord{}).
			Where("key_hash IN ? AND (source = '' OR source IS NULL)", twKeys).
			Update("source", BuiltinSource)
	}

	var swKeys []string
	for _, w := range defaultPunctWords {
		swKeys = append(swKeys, utils.KeyHash(w))
	}
	for _, line := range strings.Split(string(defaultEmojiSepwordsData), "\n") {
		w, ok := parseDefaultEmojiSepwordLine(line)
		if !ok {
			continue
		}
		swKeys = append(swKeys, utils.KeyHash(w))
	}
	if len(swKeys) > 0 {
		tx.Model(&ToSepWord{}).
			Where("key_hash IN ? AND (source = '' OR source IS NULL)", swKeys).
			Update("source", BuiltinSource)
	}

	tx.Model(&Separator{}).
		Where("key_hash = ? AND (source = '' OR source IS NULL)", utils.KeyHash("_")).
		Update("source", BuiltinSource)
	return nil
}

func syncDefaultCFG(tx *gorm.DB) error {
	var sep Separator
	if err := tx.Where("key_hash = ?", utils.KeyHash("_")).First(&sep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&Separator{
				KeyHash: utils.KeyHash("_"),
				Value:   "_",
				Source:  BuiltinSource,
			}).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	for _, p := range defaultTermWordDefs {
		lo := strings.ToLower(p.orig)
		tw := TermWord{
			KeyHash:       utils.KeyHash(lo),
			OriginalLower: lo,
			TargetWord:    p.target,
			Source:        BuiltinSource,
		}
		if err := upsertBuiltinTermWord(tx, tw).Error; err != nil {
			return err
		}
	}

	seenSepKey := make(map[string]struct{})
	for _, w := range defaultPunctWords {
		h := utils.KeyHash(w)
		if _, dup := seenSepKey[h]; dup {
			continue
		}
		seenSepKey[h] = struct{}{}
		sw := ToSepWord{KeyHash: h, Value: w, Source: BuiltinSource}
		if err := upsertBuiltinToSepWord(tx, sw).Error; err != nil {
			return err
		}
	}
	for _, line := range strings.Split(string(defaultEmojiSepwordsData), "\n") {
		w, ok := parseDefaultEmojiSepwordLine(line)
		if !ok {
			continue
		}
		h := utils.KeyHash(w)
		if _, dup := seenSepKey[h]; dup {
			continue
		}
		seenSepKey[h] = struct{}{}
		sw := ToSepWord{KeyHash: h, Value: w, Source: BuiltinSource}
		if err := upsertBuiltinToSepWord(tx, sw).Error; err != nil {
			return err
		}
	}

	var twKeys []string
	for _, p := range defaultTermWordDefs {
		twKeys = append(twKeys, utils.KeyHash(strings.ToLower(p.orig)))
	}
	if len(twKeys) > 0 {
		if err := tx.Where("source = ? AND key_hash NOT IN ?", BuiltinSource, twKeys).
			Delete(&TermWord{}).Error; err != nil {
			return err
		}
	}

	var swKeys []string
	for k := range seenSepKey {
		swKeys = append(swKeys, k)
	}
	if len(swKeys) > 0 {
		if err := tx.Where("source = ? AND key_hash NOT IN ?", BuiltinSource, swKeys).
			Delete(&ToSepWord{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func upsertBuiltinTermWord(tx *gorm.DB, tw TermWord) *gorm.DB {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key_hash"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"original_lower": tw.OriginalLower,
			"target_word":    tw.TargetWord,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "source", Value: BuiltinSource},
		}},
	}).Create(&tw)
}

func upsertBuiltinToSepWord(tx *gorm.DB, sw ToSepWord) *gorm.DB {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key_hash"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"value": sw.Value,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "source", Value: BuiltinSource},
		}},
	}).Create(&sw)
}

func ResetCFG(dbPath string) error {
	_db := utils.OpenDB(dbPath)
	defer utils.DBClose(_db)
	return _db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&TermWord{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&ToSepWord{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&Separator{}).Error; err != nil {
			return err
		}
		return syncDefaultCFG(tx)
	})
}

func parseDefaultEmojiSepwordLine(line string) (literal string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	if i := strings.IndexByte(line, '\t'); i >= 0 {
		return strings.TrimSpace(line[:i]), true
	}
	return line, true
}
