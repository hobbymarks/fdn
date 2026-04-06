package db

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/utils"
	"github.com/stretchr/testify/assert"
)

func TestEnsureDefaultCFG_seedsOnce(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "fdn.db")
	if err := EnsureDefaultCFG(p); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDefaultCFG(p); err != nil {
		t.Fatal(err)
	}

	_db := ConnectCFGDB(p)
	defer utils.DBClose(_db)

	var ts, sws, seps int64
	if err := _db.Model(&TermWord{}).Count(&ts).Error; err != nil {
		t.Fatal(err)
	}
	if err := _db.Model(&ToSepWord{}).Count(&sws).Error; err != nil {
		t.Fatal(err)
	}
	if err := _db.Model(&Separator{}).Count(&seps).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(7), ts)
	assert.Equal(t, int64(6320), sws)
	assert.Equal(t, int64(1), seps)
}
