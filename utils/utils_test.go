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
	OpenDB(dp)
	_, err := os.Stat(dp)
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
	_f := PathMaker("f")
	if !PathExist(_f) {
		t.Errorf("'f' error:%s", _f)
	} else {
		t.Logf("%s", _f)
		os.RemoveAll(_f)
	}
	_d := PathMaker("d")
	if !PathExist(_d) {
		t.Errorf("'d' error:%s", _d)
	} else {
		t.Logf("%s", _d)
		os.RemoveAll(_d)
	}
}

func TestExt(t *testing.T) {
	_f := PathMaker("f")
	if _ext := Ext(_f); _ext == "" {
		t.Errorf("ext '%s' error:%s", _f, _ext)
	} else {
		t.Logf("file '%s' ext:%s", _f, _ext)
		os.RemoveAll(_f)
	}

	_d := PathMaker("d")
	if _ext := Ext(_d); _ext != "" {
		t.Errorf("ext '%s' error:%s", _d, _ext)
	} else {
		t.Logf("dir '%s' ext:%s", _d, _ext)
		os.RemoveAll(_d)
	}
}

func TestEncryDecry(t *testing.T) {
	key := ".......|.......|.......|.......|.......|.......|"
	plainText := ".... simple plain ...."

	_encry := Encrypt(key, plainText)
	_decry := Decrypt(key, _encry)

	if _decry != plainText {
		t.Errorf("%s not equal %s", _decry, plainText)
	} else {
		t.Logf("\ndecrypted==>%s\nplaintext==>%s", _decry, plainText)
	}
}
