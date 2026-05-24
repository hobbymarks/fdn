package db

import (
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

	conn, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}

	var ts, sws, seps int64
	if err := conn.Model(&TermWord{}).Count(&ts).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Model(&ToSepWord{}).Count(&sws).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Model(&Separator{}).Count(&seps).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(7), ts)
	assert.Equal(t, int64(6323), sws)
	assert.Equal(t, int64(1), seps)
}

func TestEnsureDefaultCFG_marksBuiltinSource(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}

	var seps []Separator
	require.NoError(t, conn.Find(&seps).Error)
	require.Len(t, seps, 1)
	assert.Equal(t, BuiltinSource, seps[0].Source)

	var tws []TermWord
	require.NoError(t, conn.Find(&tws).Error)
	for _, tw := range tws {
		assert.Equal(t, BuiltinSource, tw.Source, "term %q should be builtin", tw.OriginalLower)
	}

	var sws []ToSepWord
	require.NoError(t, conn.Where("source = ?", BuiltinSource).Find(&sws).Error)
	assert.Equal(t, int64(6323), conn.Where("source = ?", BuiltinSource).Find(&[]ToSepWord{}).RowsAffected)
}

func TestEnsureDefaultCFG_userOverridesProtected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}

	lo := "wikipedia"
	keyHash := utils.KeyHash(lo)
	conn.Model(&TermWord{}).Where("key_hash = ?", keyHash).
		Updates(map[string]any{
			"target_word": "wiki",
			"source":      UserSource,
		})

	require.NoError(t, EnsureDefaultCFG(p))

	var tw TermWord
	require.NoError(t, conn.Where("key_hash = ?", keyHash).First(&tw).Error)
	assert.Equal(t, "wiki", tw.TargetWord, "user override should not be overwritten by builtin sync")
	assert.Equal(t, UserSource, tw.Source)
}

func TestEnsureDefaultCFG_builtinValueChangedOnUpgrade(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}
	var before TermWord
	require.NoError(t, conn.Where("key_hash = ?", utils.KeyHash("wikipedia")).First(&before).Error)
	assert.Equal(t, "wikipedia", before.TargetWord)
	assert.Equal(t, BuiltinSource, before.Source)

	saved := defaultTermWordDefs
	for i := range defaultTermWordDefs {
		if defaultTermWordDefs[i].orig == "wikipedia" {
			defaultTermWordDefs[i].target = "wp"
		}
	}
	defer func() { defaultTermWordDefs = saved }()

	require.NoError(t, EnsureDefaultCFG(p))

	conn2, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}
	var after TermWord
	require.NoError(t, conn2.Where("key_hash = ?", utils.KeyHash("wikipedia")).First(&after).Error)
	assert.Equal(t, "wp", after.TargetWord, "changed builtin default should be applied on upgrade")
	assert.Equal(t, BuiltinSource, after.Source)
}

func TestEnsureDefaultCFG_removesObsoleteBuiltin(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := utils.OpenDB(p)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, conn.AutoMigrate(&TermWord{}))

	require.NoError(t, conn.Create(&TermWord{
		KeyHash:       utils.KeyHash("obsolete_term"),
		OriginalLower: "obsolete_term",
		TargetWord:    "obsolete_val",
		Source:        BuiltinSource,
	}).Error)

	require.NoError(t, EnsureDefaultCFG(p))

	var cnt int64
	require.NoError(t, conn.Model(&TermWord{}).Where("key_hash = ?", utils.KeyHash("obsolete_term")).Count(&cnt).Error)
	assert.Equal(t, int64(0), cnt, "obsolete builtin should be removed")

	var total int64
	require.NoError(t, conn.Model(&TermWord{}).Count(&total).Error)
	assert.Equal(t, int64(7), total)
}

func TestEnsureDefaultCFG_migrateExistingDefaults(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")

	conn, err := utils.OpenDB(p)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, conn.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{}, &Record{}))

	conn.Exec("INSERT INTO term_words (key_hash, original_lower, target_word, source, created_at, updated_at) VALUES (?, ?, ?, '', datetime('now'), datetime('now'))",
		utils.KeyHash("wikipedia"), "wikipedia", "wikipedia")
	conn.Exec("INSERT INTO term_words (key_hash, original_lower, target_word, source, created_at, updated_at) VALUES (?, ?, ?, '', datetime('now'), datetime('now'))",
		utils.KeyHash("mycustom"), "mycustom", "custom_val")
	conn.Exec("INSERT INTO separators (key_hash, value, source, created_at, updated_at) VALUES (?, ?, '', datetime('now'), datetime('now'))",
		utils.KeyHash("_"), "_")

	require.NoError(t, EnsureDefaultCFG(p))

	var tw TermWord
	require.NoError(t, conn.Where("key_hash = ?", utils.KeyHash("wikipedia")).First(&tw).Error)
	assert.Equal(t, BuiltinSource, tw.Source, "matching default term should be marked builtin")

	var tw2 TermWord
	require.NoError(t, conn.Where("key_hash = ?", utils.KeyHash("mycustom")).First(&tw2).Error)
	assert.Equal(t, "", tw2.Source, "user-added term with empty source should stay empty (not builtin)")

	var sep Separator
	require.NoError(t, conn.First(&sep).Error)
	assert.Equal(t, BuiltinSource, sep.Source, "matching default separator should be marked builtin")
}

func TestResetCFG(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}
	conn.Create(&TermWord{
		KeyHash:       utils.KeyHash("user_term"),
		OriginalLower: "user_term",
		TargetWord:    "user_val",
		Source:        UserSource,
	})

	require.NoError(t, ResetCFG(p))

	conn2, err := ConnectCFGDB(p)
	if err != nil {
		t.Fatal(err)
	}

	var cnt int64
	conn2.Model(&TermWord{}).Where("key_hash = ?", utils.KeyHash("user_term")).Count(&cnt)
	assert.Equal(t, int64(0), cnt, "user entries should be removed by reset")

	var total int64
	conn2.Model(&TermWord{}).Count(&total)
	assert.Equal(t, int64(7), total, "builtin terms should be restored")

	var seps int64
	conn2.Model(&Separator{}).Count(&seps)
	assert.Equal(t, int64(1), seps)

	var sep Separator
	conn2.First(&sep)
	assert.Equal(t, BuiltinSource, sep.Source)
}

func TestSyncDefaultCFG_addsNewBuiltin(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := utils.OpenDB(p)
	if err != nil {
		t.Fatal(err)
	}

	saved := defaultTermWordDefs
	defaultTermWordDefs = append(defaultTermWordDefs, struct{ orig, target string }{"newterm", "newval"})
	defer func() { defaultTermWordDefs = saved }()

	require.NoError(t, conn.Transaction(func(tx *gorm.DB) error {
		return syncDefaultCFG(tx)
	}))

	var tw TermWord
	err = conn.Where("key_hash = ?", utils.KeyHash("newterm")).First(&tw).Error
	assert.NoError(t, err, "new builtin term should be added")
	assert.Equal(t, BuiltinSource, tw.Source)
	assert.Equal(t, "newval", tw.TargetWord)
}

func TestSyncDefaultCFG_updatesBuiltinValue(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fdn.db")
	require.NoError(t, EnsureDefaultCFG(p))

	conn, err := utils.OpenDB(p)
	if err != nil {
		t.Fatal(err)
	}

	saved := defaultTermWordDefs
	for i := range defaultTermWordDefs {
		if defaultTermWordDefs[i].orig == "wikipedia" {
			defaultTermWordDefs[i].target = "wp"
		}
	}
	defer func() { defaultTermWordDefs = saved }()

	require.NoError(t, conn.Transaction(func(tx *gorm.DB) error {
		return syncDefaultCFG(tx)
	}))

	var tw TermWord
	require.NoError(t, conn.Where("key_hash = ?", utils.KeyHash("wikipedia")).First(&tw).Error)
	assert.Equal(t, "wp", tw.TargetWord, "builtin value should be updated")
	assert.Equal(t, BuiltinSource, tw.Source)
}
