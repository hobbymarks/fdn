package db

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

func EnsureDefaultCFG(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	_db := utils.OpenDB(dbPath)
	if err := _db.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{}); err != nil {
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
	for _, w := range []string{
		"-",
		" ",
		"\u201c",
		"\u201d",
		"\uff0c",
		"\uff1f",
		"(",
		")",
		"~",
		"\u3001",
		"\uff1a",
		"\u3010",
		"\u3011",
		"\u4e28",
		"\uff5c",
		"\uff01",
		"\uff08",
		"\u300a",
		"\u300b",
	} {
		sw := ToSepWord{KeyHash: utils.KeyHash(w), Value: w}
		if err := tx.Create(&sw).Error; err != nil {
			return err
		}
	}
	return nil
}
