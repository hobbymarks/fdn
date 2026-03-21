package cmd

import (
	"os"
	"path/filepath"
	"testing"

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
	fdn := utils.FDNDir()
	p := filepath.Join(fdn, "cfg.db")
	data, err := defaultCFG.ReadFile("cfg.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
