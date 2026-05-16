/*
Package utils test
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenDB_LegalPath(t *testing.T) {
	dp := "fdn.db"
	if PathExist(dp) {
		t.Errorf("please remove file %s", dp)
		return
	}
	_, err := OpenDB(dp)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("%s not exist", dp)
		} else {
			t.Errorf("check %s err:%s", dp, err)
		}
	}
	os.Remove(dp)
}

func TestPathMaker(t *testing.T) {
	path, err := PathMaker("f")
	if err != nil {
		t.Fatal(err)
	}
	if !PathExist(path) {
		t.Errorf("'f' error:%s", path)
	} else {
		t.Logf("%s", path)
		os.RemoveAll(path)
	}
	dir, err := PathMaker("d")
	if err != nil {
		t.Fatal(err)
	}
	if !PathExist(dir) {
		t.Errorf("'d' error:%s", dir)
	} else {
		t.Logf("%s", dir)
		os.RemoveAll(dir)
	}
}

func TestExt(t *testing.T) {
	path, err := PathMaker("f")
	if err != nil {
		t.Fatal(err)
	}
	if ext := Ext(path); ext == "" {
		t.Errorf("ext '%s' error:%s", path, ext)
	} else {
		t.Logf("file '%s' ext:%s", path, ext)
		os.RemoveAll(path)
	}

	dir, err := PathMaker("d")
	if err != nil {
		t.Fatal(err)
	}
	if ext := Ext(dir); ext != "" {
		t.Errorf("ext '%s' error:%s", dir, ext)
	} else {
		t.Logf("dir '%s' ext:%s", dir, ext)
		os.RemoveAll(dir)
	}
}

func TestEncryDecry(t *testing.T) {
	key := ".......|.......|.......|.......|.......|.......|"
	plainText := ".... simple plain ...."

	encStr, err := Encrypt(key, plainText)
	if err != nil {
		t.Fatal(err)
	}
	decStr, err := Decrypt(key, encStr)
	if err != nil {
		t.Fatal(err)
	}

	if decStr != plainText {
		t.Errorf("%s not equal %s", decStr, plainText)
	} else {
		t.Logf("\ndecrypted==>%s\nplaintext==>%s", decStr, plainText)
	}
}

func TestKeyHash(t *testing.T) {
	h := KeyHash("test")
	if len(h) == 0 {
		t.Error("KeyHash returned empty string")
	}
	h2 := KeyHash("test")
	if h != h2 {
		t.Error("KeyHash should be deterministic")
	}
}

func TestPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	if !PathIsDirectory(dir) {
		t.Error("temp dir should be a directory")
	}
	f := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if PathIsDirectory(f) {
		t.Error("regular file should not be detected as directory")
	}
	if PathIsDirectory(filepath.Join(dir, "nonexistent")) {
		t.Error("nonexistent path should not be detected as directory")
	}
}

func TestFDNDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	dir, err := FDNDir()
	if err != nil {
		t.Fatal(err)
	}
	if !PathIsDirectory(dir) {
		t.Error("FDNDir should create and return a directory")
	}
}

func TestSameFiles(t *testing.T) {
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
	same, err := SameFiles(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !same {
		t.Error("identical files should be same")
	}

	c := filepath.Join(dir, "c.txt")
	if err := os.WriteFile(c, []byte("different"), 0o644); err != nil {
		t.Fatal(err)
	}
	same, err = SameFiles(a, c)
	if err != nil {
		t.Fatal(err)
	}
	if same {
		t.Error("different files should not be same")
	}
}

func TestFileMD5(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	md5s, err := FileMD5(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(md5s) == 0 {
		t.Error("FileMD5 returned empty")
	}

	_, err = FileMD5(filepath.Join(dir, "nope"))
	if err == nil {
		t.Error("FileMD5 should error on missing file")
	}
}

func TestSameFiles_missingPath(t *testing.T) {
	_, err := SameFiles("/nope/a", "/nope/b")
	if err == nil {
		t.Error("SameFiles should error on missing paths")
	}
}
