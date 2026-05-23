//go:build darwin || linux

package cmd

import (
	"path/filepath"
	"strings"
)

// IsHidden check file is hidden
func IsHidden(abspath string) (bool, error) {
	abspath = filepath.Clean(abspath)
	bn := filepath.Base(abspath)

	if strings.HasPrefix(bn, ".") {
		return true, nil
	}
	return false, nil
}
