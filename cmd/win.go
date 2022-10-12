//go:build windows

/*
Package cmd win only for windows
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"path/filepath"
	"syscall"
)

// IsHidden check file is hidden
func IsHidden(abspath string) (bool, error) {
	abspath = filepath.Clean(abspath)

	ptr, err := syscall.UTF16PtrFromString(`\\?\` + abspath)
	if err != nil {
		return false, err
	}
	attr, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		return false, err
	}
	return attr&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
