//go:build darwin || linux

/*
Package cmd notwin
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import "path/filepath"

// IsHidden check file is hidden
func IsHidden(abspath string) (bool, error) {
	abspath = filepath.Clean(abspath)
	bn := filepath.Base(abspath)

	if bn[0:1] == "." {
		return true, nil
	}
	return false, nil
}
