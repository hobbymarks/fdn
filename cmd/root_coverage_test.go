package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func withDisplayFlags(t *testing.T, plain, prettyV, full bool) func() {
	t.Helper()
	op, opr, of := plainStyle, pretty, fullpath
	plainStyle = plain
	pretty = prettyV
	fullpath = full
	return func() {
		plainStyle, pretty, fullpath = op, opr, of
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	old := os.Stdout
	os.Stdout = w
	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		outCh <- buf.String()
	}()
	fn()
	assert.NoError(t, w.Close())
	os.Stdout = old
	return <-outCh
}

func TestDepthFiles_shallow(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := DepthFiles(dir, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.GreaterOrEqual(t, len(got), 1)
}

func TestDepthFiles_readDirError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent_dir_xyz")
	_, err := DepthFiles(missing, 1, false)
	assert.Error(t, err)
}

func TestFilteredSubPaths_missingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing_nested_xyz")
	_, err := FilteredSubPaths(missing, 1, false)
	assert.Error(t, err)
}

func TestFilteredSubPaths_walkMissingRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "walk_missing_xyz")
	_, err := FilteredSubPaths(missing, -1, false)
	assert.Error(t, err)
}

func TestDepthFiles_recursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "inner.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := DepthFiles(dir, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, p := range got {
		if filepath.Base(p) == "inner.txt" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestFilteredSubPaths_boundedDepth(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FilteredSubPaths(dir, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, got)
}

func TestFilteredSubPaths_unlimitedDepth(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "deep")
	if err := os.MkdirAll(filepath.Join(sub, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "f.txt"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FilteredSubPaths(dir, -1, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, got)
}

func TestFilteredSubPaths_onlyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "d"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FilteredSubPaths(dir, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, got)
}

func TestRetrievedAbsPaths_onlyDirSkipsPlainFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "onlyfile.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := RetrievedAbsPaths([]string{f}, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, out)
}

func TestRetrievedAbsPaths_fileAndDir(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "one.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := RetrievedAbsPaths([]string{f}, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 1)

	dirs, err := RetrievedAbsPaths([]string{dir}, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, dirs)
}

func TestRetrievedAbsPaths_missingPath(t *testing.T) {
	_, err := RetrievedAbsPaths([]string{"/nonexistent/path/that/does/not/exist/ever"}, 1, false)
	assert.Error(t, err)
}

func TestRemoveHidden_filtersDotFile(t *testing.T) {
	dir := t.TempDir()
	pub := filepath.Join(dir, "visible.txt")
	dot := filepath.Join(dir, ".hidden.txt")
	if err := os.WriteFile(pub, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dot, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	absPub, err := filepath.Abs(pub)
	if err != nil {
		t.Fatal(err)
	}
	absDot, err := filepath.Abs(dot)
	if err != nil {
		t.Fatal(err)
	}

	out := RemoveHidden([]string{absPub, absDot})
	assert.Contains(t, out, absPub)
}

func TestASCHead_variants(t *testing.T) {
	assert.Contains(t, ASCHead("hello"), "hello")
	assert.NotEmpty(t, ASCHead("中文"))
	d := ASCHead("123abc")
	assert.True(t, len(d) >= len("123abc"))
}

func TestFDNFile_forwardRename(t *testing.T) {
	testIsolatedHome(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "from.txt")
	dst := filepath.Join(dir, "to.txt")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := FDNFile(src, dst, false); err != nil {
		t.Fatal(err)
	}
	assert.FileExists(t, dst)
	assert.NoFileExists(t, src)
}

func TestAddRecord_incrementCount(t *testing.T) {
	testIsolatedHome(t)
	conn, err := db.ConnectRDDB()
	if err != nil {
		t.Fatal(err)
	}
	to := "bname"
	cur := "aname"
	encPrev, err := utils.Encrypt(to, cur)
	if err != nil {
		t.Fatal(err)
	}
	rec := db.Record{
		EncryptedPreviousName: encPrev,
		HashedCurrentName:     utils.KeyHash(to),
		Count:                 1,
	}
	if err := AddRecord(conn, rec); err != nil {
		t.Fatal(err)
	}
	if err := AddRecord(conn, rec); err != nil {
		t.Fatal(err)
	}
	var got db.Record
	if err := conn.Where("hashed_current_name = ?", utils.KeyHash(to)).First(&got).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(2), got.Count)
}

func TestDeleteRecord_decrementAndRemove(t *testing.T) {
	testIsolatedHome(t)
	conn, err := db.ConnectRDDB()
	if err != nil {
		t.Fatal(err)
	}
	to := "tb"
	cur := "ta"
	encPrev, err := utils.Encrypt(to, cur)
	if err != nil {
		t.Fatal(err)
	}
	rec := db.Record{
		EncryptedPreviousName: encPrev,
		HashedCurrentName:     utils.KeyHash(to),
		Count:                 2,
	}
	if err := conn.Create(&rec).Error; err != nil {
		t.Fatal(err)
	}

	del := db.Record{
		EncryptedPreviousName: encPrev,
		HashedCurrentName:     utils.KeyHash(to),
	}
	if err := DeleteRecord(conn, del); err != nil {
		t.Fatal(err)
	}
	var mid db.Record
	if err := conn.Where("id = ?", rec.ID).First(&mid).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), mid.Count)

	if err := DeleteRecord(conn, del); err != nil {
		t.Fatal(err)
	}
	err = conn.Where("id = ?", rec.ID).First(&db.Record{}).Error
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCheckDoFDN_targetMissing(t *testing.T) {
	testIsolatedHome(t)
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckDoFDN(a, b, false, false); err != nil {
		t.Fatal(err)
	}
	assert.NoFileExists(t, a)
	assert.FileExists(t, b)
}

func TestCheckDoFDN_sameFile(t *testing.T) {
	testIsolatedHome(t)
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	p := filepath.Join(dir, "same.txt")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckDoFDN(p, p, false, false); err != nil {
		t.Fatal(err)
	}
	assert.FileExists(t, p)
}

func TestCheckDoFDN_existDifferentFiles(t *testing.T) {
	testIsolatedHome(t)
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		_ = CheckDoFDN(a, b, false, false)
	})
	assert.Contains(t, out, "[EXIST]Skip:")
	assert.FileExists(t, a)
	assert.FileExists(t, b)
}

func TestCheckDoFDN_sameContentReplacesDestination(t *testing.T) {
	testIsolatedHome(t)
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	body := []byte("identical")
	if err := os.WriteFile(a, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, body, 0o644); err != nil {
		t.Fatal(err)
	}

	err := CheckDoFDN(a, b, false, false)
	assert.NoError(t, err)
	assert.NoFileExists(t, a)
	assert.FileExists(t, b)
}

func TestCheckDoFDN_sameFilesCompareError(t *testing.T) {
	testIsolatedHome(t)
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	d := filepath.Join(dir, "d")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		_ = CheckDoFDN(f, d, false, false)
	})
	assert.Contains(t, out, "[ERROR]Skip:")
}

func TestCheckDoFDN_overwrite(t *testing.T) {
	testIsolatedHome(t)
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("src"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckDoFDN(a, b, false, true); err != nil {
		t.Fatal(err)
	}
	assert.NoFileExists(t, a)
	assert.FileExists(t, b)
}

func TestOutputResult_plain(t *testing.T) {
	defer withDisplayFlags(t, true, false, false)()
	out := captureStdout(t, func() {
		OutputResult("/tmp/x/a.txt", "/tmp/x/b.txt", false, false)
	})
	assert.Contains(t, out, "a.txt")
	assert.Contains(t, out, "-->")
}

func TestOutputResult_plainInplace(t *testing.T) {
	defer withDisplayFlags(t, true, false, false)()
	out := captureStdout(t, func() {
		OutputResult("/x/old.txt", "/x/new.txt", true, true)
	})
	assert.Contains(t, out, "==>")
}

func TestOutputResult_richDiff(t *testing.T) {
	defer withDisplayFlags(t, false, false, false)()
	out := captureStdout(t, func() {
		OutputResult("ab", "ac", false, true)
	})
	assert.NotEmpty(t, out)
}

func TestOutputResult_richPretty(t *testing.T) {
	defer withDisplayFlags(t, false, true, true)()
	out := captureStdout(t, func() {
		OutputResult("wide▯x", "y", false, true)
	})
	assert.NotEmpty(t, out)
}

func TestOutputResult_richInsert(t *testing.T) {
	defer withDisplayFlags(t, false, false, true)()
	out := captureStdout(t, func() {
		OutputResult("ab", "abc", false, true)
	})
	assert.NotEmpty(t, out)
}

func TestOutputResult_richDelete(t *testing.T) {
	defer withDisplayFlags(t, false, false, true)()
	out := captureStdout(t, func() {
		OutputResult("abc", "ab", false, true)
	})
	assert.NotEmpty(t, out)
}

func TestOutputResult_richEqual(t *testing.T) {
	defer withDisplayFlags(t, false, false, true)()
	out := captureStdout(t, func() {
		OutputResult("same", "same", true, true)
	})
	assert.Contains(t, out, "same")
}

func TestAddRecord_createsNew(t *testing.T) {
	testIsolatedHome(t)
	conn, err := db.ConnectRDDB()
	if err != nil {
		t.Fatal(err)
	}
	to := "uniqnew_to"
	cur := "uniqnew_from"
	encPrev, err := utils.Encrypt(to, cur)
	if err != nil {
		t.Fatal(err)
	}
	rec := db.Record{
		EncryptedPreviousName: encPrev,
		HashedCurrentName:     utils.KeyHash(to),
		Count:                 1,
	}
	if err := AddRecord(conn, rec); err != nil {
		t.Fatal(err)
	}
	var got db.Record
	if err := conn.Where("hashed_current_name = ?", utils.KeyHash(to)).First(&got).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), got.Count)
}

func TestGetConfirm(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = r
	_, _ = w.WriteString("no\n")
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	v := GetConfirm()
	_ = r.Close()
	os.Stdin = old
	assert.Equal(t, No, v)
}

func Test_noEffectTip(t *testing.T) {
	noEffectTip()
}

func TestConfigTermWords_skipDuplicate(t *testing.T) {
	seedEmbeddedCfgDB(t)
	m := map[string]string{"dupkeyzz": "v1"}
	if err := ConfigTermWords(m); err != nil {
		t.Fatal(err)
	}
	if err := ConfigTermWords(m); err != nil {
		t.Fatal(err)
	}
}

func TestConfigToSepWords_skipDuplicate(t *testing.T) {
	seedEmbeddedCfgDB(t)
	w := []string{"dupsepzz"}
	if err := ConfigToSepWords(w); err != nil {
		t.Fatal(err)
	}
	if err := ConfigToSepWords(w); err != nil {
		t.Fatal(err)
	}
}

func TestConfigSeparator_skipDuplicate(t *testing.T) {
	seedEmbeddedCfgDB(t)
	if err := ConfigSeparator("|"); err != nil {
		t.Fatal(err)
	}
	if err := ConfigSeparator("|"); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_renameAndReverse(t *testing.T) {
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()

	orig := filepath.Join(dir, "My Test File.txt")
	if err := os.WriteFile(orig, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := FDNedFrom("My Test File")
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(dir, out+".txt")

	if err := FDNFile(orig, renamed, false); err != nil {
		t.Fatal(err)
	}
	assert.NoFileExists(t, orig)
	assert.FileExists(t, renamed)

	if err := FDNFile(renamed, orig, true); err != nil {
		t.Fatal(err)
	}
	assert.FileExists(t, orig)
	assert.NoFileExists(t, renamed)
}

func TestE2E_fullRenamePipeline(t *testing.T) {
	seedEmbeddedCfgDB(t)
	defer withDisplayFlags(t, true, false, false)()

	dir := t.TempDir()
	files := map[string]string{
		"Hello World.txt": "Hello_World.txt",
		"Foo  Bar.txt":    "Foo_Bar.txt",
	}
	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	allPaths, err := RetrievedAbsPaths([]string{dir}, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range allPaths {
		fn := filepath.Base(p)
		ext := filepath.Ext(fn)
		bn := fn[:len(fn)-len(ext)]
		if expected, ok := files[fn]; ok {
			fdned, err := FDNedFrom(bn)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, expected, fdned+ext)
		}
	}
}

func TestReplaceWords_spacesBecomeSeparator(t *testing.T) {
	seedEmbeddedCfgDB(t)
	out, err := ReplaceWords("Hello World Test")
	assert.NoError(t, err)
	assert.Equal(t, "Hello_World_Test", out)
}
