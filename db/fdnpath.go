package db

import (
	"path/filepath"
	"strings"

	"github.com/hobbymarks/fdn/utils"
)

const FDNDBFileName = "fdn.db"

func DefaultFDNDBPath() string {
	fdnDir, err := utils.FDNDir()
	if err != nil {
		panic("fdn data directory unavailable: " + err.Error())
	}
	return filepath.Join(fdnDir, FDNDBFileName)
}

func sqliteQuotePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	abs = filepath.ToSlash(abs)
	return strings.ReplaceAll(abs, "'", "''"), nil
}
