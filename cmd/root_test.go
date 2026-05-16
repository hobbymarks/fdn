/*
Package cmd root test
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/db"
	"github.com/stretchr/testify/assert"
)

func TestRetrievedAbsPaths_returnsErrorWhenAllMissing(t *testing.T) {
	_, err := RetrievedAbsPaths(
		[]string{filepath.Join(t.TempDir(), "nonexistent", "path")},
		1,
		true,
	)
	assert.Error(t, err, "should return error when all input paths are missing")
}

func TestFDNedFrom(t *testing.T) {
	seedEmbeddedCfgDB(t)
	out, err := FDNedFrom("123 456 789")
	assert.NoError(t, err)
	assert.Equal(t, "123_456_789", out)
}

func TestPrepareFDNDataDir(t *testing.T) {
	db.ResetSharedDB()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	err := prepareFDNDataDir()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(home, ".fdn", db.FDNDBFileName))
}
