package cmd

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

func testIsolatedHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return tmp
}

func seedEmbeddedCfgDB(t *testing.T) {
	t.Helper()
	testIsolatedHome(t)
	p := filepath.Join(utils.FDNDir(), db.FDNDBFileName)
	if err := db.EnsureDefaultCFG(p); err != nil {
		t.Fatal(err)
	}
}
