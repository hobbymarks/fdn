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
	"\n",
	"\r",
	"\t",
	"-",
	" ",
	"\u201C",
	"\u201D",
	"\uFF0C",
	"\uFF1F",
	"(",
	")",
	"~",
	"\u3001",
	"\u3002",
	"\u30FB",
	"\uFF02",
	"\uFF1A",
	"\u3010",
	"\u3011",
	"\u4E28",
	"\uFF5C",
	"\uFF01",
	"\uFF08",
	"\uFF09",
	"\u300A",
	"\u300B",
	"<",
	">",
	":",
	"\"",
	"/",
	"\u29F8",
	"\\",
	"|",
	"?",
	"*",
	"!",
	"@",
	"#",
	"$",
	"%",
	"^",
	"&",
	"`",
	";",
	",",
	"[",
	"]",
	"{",
	"}",
	"'",
	"+",
	"=",
}

func getOrConnectDB(dbPath string) (*gorm.DB, error) {
	conn := GetDB()
	if conn != nil {
		return conn, nil
	}
	return ConnectDB(dbPath)
}

func EnsureDefaultCFG(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	conn, err := getOrConnectDB(dbPath)
	if err != nil {
		return err
	}
	return conn.Transaction(func(tx *gorm.DB) error {
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
	for line := range strings.SplitSeq(string(defaultEmojiSepwordsData), "\n") {
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
	for line := range strings.SplitSeq(string(defaultEmojiSepwordsData), "\n") {
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
		DoUpdates: clause.Assignments(map[string]any{
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
		DoUpdates: clause.Assignments(map[string]any{
			"value": sw.Value,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "source", Value: BuiltinSource},
		}},
	}).Create(&sw)
}

func ResetCFG(dbPath string) error {
	conn, err := getOrConnectDB(dbPath)
	if err != nil {
		return err
	}
	return conn.Transaction(func(tx *gorm.DB) error {
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
	if before, _, ok := strings.Cut(line, "\t"); ok {
		return strings.TrimSpace(before), true
	}
	return line, true
}
