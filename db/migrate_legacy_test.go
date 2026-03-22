package db

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/utils"
	"github.com/stretchr/testify/assert"
)

func TestMigrateLegacyFDNDatabases_mergeCfgAndRd(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir := utils.FDNDir()

	cfgPath := filepath.Join(fdnDir, "cfg.db")
	rdPath := filepath.Join(fdnDir, "rd.db")
	fdnPath := filepath.Join(fdnDir, FDNDBFileName)

	if err := EnsureDefaultCFG(cfgPath); err != nil {
		t.Fatal(err)
	}

	rdDB := utils.OpenDB(rdPath)
	if err := rdDB.AutoMigrate(&Record{}); err != nil {
		utils.DBClose(rdDB)
		t.Fatal(err)
	}
	if err := rdDB.Create(&Record{
		EncryptedPreviousName: "encdemo",
		HashedCurrentName:     "hashdemo",
		Count:                 1,
	}).Error; err != nil {
		utils.DBClose(rdDB)
		t.Fatal(err)
	}
	utils.DBClose(rdDB)

	if err := MigrateLegacyFDNDatabases(fdnDir); err != nil {
		t.Fatal(err)
	}

	assert.False(t, utils.PathExist(cfgPath))
	assert.False(t, utils.PathExist(rdPath))
	assert.True(t, utils.PathExist(fdnPath))

	_db := ConnectCFGDB()
	defer utils.DBClose(_db)
	var nTerm, nRec int64
	if err := _db.Model(&TermWord{}).Count(&nTerm).Error; err != nil {
		t.Fatal(err)
	}
	if err := _db.Model(&Record{}).Count(&nRec).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(7), nTerm)
	assert.Equal(t, int64(1), nRec)
}

func TestMigrateLegacyFDNDatabases_cfgOnlyRenames(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir := utils.FDNDir()
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
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir := utils.FDNDir()
	rdPath := filepath.Join(fdnDir, "rd.db")
	fdnPath := filepath.Join(fdnDir, FDNDBFileName)

	rdDB := utils.OpenDB(rdPath)
	if err := rdDB.AutoMigrate(&Record{}); err != nil {
		utils.DBClose(rdDB)
		t.Fatal(err)
	}
	utils.DBClose(rdDB)

	if err := MigrateLegacyFDNDatabases(fdnDir); err != nil {
		t.Fatal(err)
	}
	assert.False(t, utils.PathExist(rdPath))
	assert.True(t, utils.PathExist(fdnPath))
}
