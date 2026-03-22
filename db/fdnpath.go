package db

import (
	"path/filepath"
	"strings"

	"github.com/hobbymarks/fdn/utils"
)

const FDNDBFileName = "fdn.db"

func DefaultFDNDBPath() string {
	return filepath.Join(utils.FDNDir(), FDNDBFileName)
}

func sqliteQuotePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	abs = filepath.ToSlash(abs)
	return strings.ReplaceAll(abs, "'", "''"), nil
}
