package db

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/utils"
	"github.com/stretchr/testify/assert"
)

func TestMigrateLegacyFDNDatabases_mergeCfgAndRd(t *testing.T) {
	resetSharedDBOnCleanup(t)
	defer ResetSharedDB()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir, err := utils.FDNDir()
	if err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(fdnDir, "cfg.db")
	rdPath := filepath.Join(fdnDir, "rd.db")
	fdnPath := filepath.Join(fdnDir, FDNDBFileName)

	if err := EnsureDefaultCFG(cfgPath); err != nil {
		t.Fatal(err)
	}

	rdDB, err := utils.OpenDB(rdPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := rdDB.AutoMigrate(&Record{}); err != nil {
		t.Fatal(err)
	}
	if err := rdDB.Create(&Record{
		EncryptedPreviousName: "encdemo",
		HashedCurrentName:     "hashdemo",
		Count:                 1,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyFDNDatabases(fdnDir); err != nil {
		t.Fatal(err)
	}

	assert.False(t, utils.PathExist(cfgPath))
	assert.False(t, utils.PathExist(rdPath))
	assert.True(t, utils.PathExist(fdnPath))

	conn, err := ConnectCFGDB()
	if err != nil {
		t.Fatal(err)
	}
	var nTerm, nRec int64
	if err := conn.Model(&TermWord{}).Count(&nTerm).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Model(&Record{}).Count(&nRec).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(7), nTerm)
	assert.Equal(t, int64(1), nRec)
}

func TestMigrateLegacyFDNDatabases_cfgOnlyRenames(t *testing.T) {
	resetSharedDBOnCleanup(t)
	defer ResetSharedDB()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir, err := utils.FDNDir()
	if err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(fdnDir, "cfg.db")
	fdnPath := filepath.Join(fdnDir, FDNDBFileName)

	if err := EnsureDefaultCFG(cfgPath); err != nil {
		t.Fatal(err)
	}
	if err := MigrateLegacyFDNDatabases(fdnDir); err != nil {
		t.Fatal(err)
	}
	assert.False(t, utils.PathExist(cfgPath))
	assert.True(t, utils.PathExist(fdnPath))
}

func TestMigrateLegacyFDNDatabases_rdOnlyRenames(t *testing.T) {
	resetSharedDBOnCleanup(t)
	defer ResetSharedDB()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir, err := utils.FDNDir()
	if err != nil {
		t.Fatal(err)
	}
	rdPath := filepath.Join(fdnDir, "rd.db")
	fdnPath := filepath.Join(fdnDir, FDNDBFileName)

	rdDB, err := utils.OpenDB(rdPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := rdDB.AutoMigrate(&Record{}); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyFDNDatabases(fdnDir); err != nil {
		t.Fatal(err)
	}
	assert.False(t, utils.PathExist(rdPath))
	assert.True(t, utils.PathExist(fdnPath))
}
