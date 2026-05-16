package cmd

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

func testIsolatedHome(t *testing.T) string {
	t.Helper()
	db.ResetSharedDB()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Cleanup(func() { db.ResetSharedDB() })
	return tmp
}

func seedEmbeddedCfgDB(t *testing.T) {
	t.Helper()
	testIsolatedHome(t)
	fdnDir, err := utils.FDNDir()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(fdnDir, db.FDNDBFileName)
	if err := db.InitDB(p); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureDefaultCFG(p); err != nil {
		t.Fatal(err)
	}
}
