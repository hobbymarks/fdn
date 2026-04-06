package db

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

// At runtime only default_emoji_sepwords.txt (below) is embedded and seeded. The Unicode
// source files under db/ and db/emojigen exist solely to regenerate that file; fdn does not
// read them when serving the CLI or opening fdn.db.
//
//go:embed default_emoji_sepwords.txt
var defaultEmojiSepwordsData []byte

func EnsureDefaultCFG(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	_db := utils.OpenDB(dbPath)
	if err := _db.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{}, &Record{}); err != nil {
		return err
	}
	var n int64
	if err := _db.Model(&Separator{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return _db.Transaction(func(tx *gorm.DB) error {
		return seedDefaultCFG(tx)
	})
}

func seedDefaultCFG(tx *gorm.DB) error {
	if err := tx.Create(&Separator{KeyHash: utils.KeyHash("_"), Value: "_"}).Error; err != nil {
		return err
	}
	for _, p := range []struct{ orig, target string }{
		{"凤凰网", "ifeng"},
		{"wikipedia", "wikipedia"},
		{"新浪网", "sina"},
		{"自由的百科全书", "_"},
		{"维基百科", "wikipedia"},
		{"_·_", "_"},
		{"搜狐网", "sohu"},
	} {
		lo := strings.ToLower(p.orig)
		tw := TermWord{
			KeyHash:       utils.KeyHash(lo),
			OriginalLower: lo,
			TargetWord:    p.target,
		}
		if err := tx.Create(&tw).Error; err != nil {
			return err
		}
	}
	seenSepKey := make(map[string]struct{})
	punct := []string{
		"-",  // U+002D
		" ",  // U+0020
		"“",  // U+201C
		"”",  // U+201D
		"，",  // U+FF0C
		"？",  // U+FF1F
		"(",  // U+0028
		")",  // U+0029
		"~",  // U+007E
		"、",  // U+3001
		"。",  // U+3002
		"・",  // U+30FB
		"＂",  // U+FF02
		"：",  // U+FF1A
		"【",  // U+3010
		"】",  // U+3011
		"丨",  // U+4E28
		"｜",  // U+FF5C
		"！",  // U+FF01
		"（",  // U+FF08
		"）",  // U+FF09
		"《",  // U+300A
		"》",  // U+300B
		"<",  // U+003C
		">",  // U+003E
		":",  // U+003A
		"\"", // U+0022
		"/",  // U+002F
		"⧸",  // U+29F8
		"\\", // U+005C
		"|",  // U+007C
		"?",  // U+003F
		"*",  // U+002A
		"!",  // U+0021
		"@",  // U+0040
		"#",  // U+0023
		"$",  // U+0024
		"%",  // U+0025
		"^",  // U+005E
		"&",  // U+0026
		"`",  // U+0060
		";",  // U+003B
		",",  // U+002C
		"[",  // U+005B
		"]",  // U+005D
		"{",  // U+007B
		"}",  // U+007D
		"'",  // U+0027
		"+",  // U+002B
		"=",  // U+003D
	}
	for _, w := range punct {
		h := utils.KeyHash(w)
		seenSepKey[h] = struct{}{}
		sw := ToSepWord{KeyHash: h, Value: w}
		if err := tx.Create(&sw).Error; err != nil {
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
		sw := ToSepWord{KeyHash: h, Value: w}
		if err := tx.Create(&sw).Error; err != nil {
			return err
		}
	}
	return nil
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
