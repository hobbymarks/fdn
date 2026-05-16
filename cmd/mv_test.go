package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/db"
	"github.com/stretchr/testify/assert"
)

func testMvHome(t *testing.T) string {
	t.Helper()
	db.ResetSharedDB()
	t.Cleanup(func() { db.ResetSharedDB() })
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return tmp
}

func TestMvCmd_exactArgs(t *testing.T) {
	assert.Error(t, mvCmd.Args(mvCmd, []string{}))
	assert.Error(t, mvCmd.Args(mvCmd, []string{"only"}))
	assert.NoError(t, mvCmd.Args(mvCmd, []string{"a", "b"}))
}

func TestMvCmd_helpText(t *testing.T) {
	var buf bytes.Buffer
	mvCmd.SetOut(&buf)
	mvCmd.SetErr(&buf)
	assert.NoError(t, mvCmd.Help())
	out := buf.String()
	assert.Contains(t, out, "SOURCE")
	assert.Contains(t, out, "DEST")
	assert.Contains(t, out, "Examples:")
	mvCmd.SetOut(os.Stdout)
	mvCmd.SetErr(os.Stderr)
}

func TestMvCmd_originMissing(t *testing.T) {
	testMvHome(t)
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")
	dest := filepath.Join(dir, "out.txt")

	mvCmd.Run(mvCmd, []string{missing, dest})
	assert.NoFileExists(t, dest)
}

func TestMvCmd_targetNewPath(t *testing.T) {
	testMvHome(t)
	dir := t.TempDir()
	origin := filepath.Join(dir, "a.txt")
	dest := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(origin, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	mvCmd.Run(mvCmd, []string{origin, dest})

	assert.NoFileExists(t, origin)
	assert.FileExists(t, dest)
}

func TestMvCmd_targetDirectory(t *testing.T) {
	testMvHome(t)
	dir := t.TempDir()
	sub := filepath.Join(dir, "d")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	origin := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(origin, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	mvCmd.Run(mvCmd, []string{origin, sub})

	want := filepath.Join(sub, "a.txt")
	assert.NoFileExists(t, origin)
	assert.FileExists(t, want)
}

func TestMvCmd_targetExistingFile(t *testing.T) {
	testMvHome(t)
	dir := t.TempDir()
	origin := filepath.Join(dir, "a.txt")
	target := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(origin, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	mvCmd.Run(mvCmd, []string{origin, target})

	assert.FileExists(t, origin, "origin must remain when target is a non-directory")
	assert.FileExists(t, target)
}
