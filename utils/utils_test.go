/*
Package utils test
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package utils

import (
	"os"
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
