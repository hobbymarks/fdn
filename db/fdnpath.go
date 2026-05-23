package db

import (
	"path/filepath"
	"strings"

	"github.com/hobbymarks/fdn/utils"
)

const FDNDBFileName = "fdn.db"

func DefaultFDNDBPath() (string, error) {
	fdnDir, err := utils.FDNDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(fdnDir, FDNDBFileName), nil
}

func sqliteQuotePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	abs = filepath.ToSlash(abs)
	return strings.ReplaceAll(abs, "'", "''"), nil
}
