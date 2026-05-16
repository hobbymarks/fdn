/*
Package db fdn cfg test
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package db

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/utils"
)

func TestConnectCFGDB_NoParam(t *testing.T) {
	resetSharedDBOnCleanup(t)
	defer ResetSharedDB()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	fdnDir, err := utils.FDNDir()
	if err != nil {
		t.Fatal(err)
	}
	dp := filepath.Join(fdnDir, FDNDBFileName)
	if utils.PathExist(dp) {
		t.Errorf("please remove file %s", dp)
		return
	}
	ResetSharedDB()
	if _, err := ConnectCFGDB(); err != nil {
		t.Fatal(err)
	}
	ResetSharedDB()
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

func TestConnectCFGDB_AParam_Exist(t *testing.T) {
	resetSharedDBOnCleanup(t)
	defer ResetSharedDB()
	f, err := os.CreateTemp("", "cfg*.db")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("Temp file %s", f.Name())
	defer os.Remove(f.Name())

	dp := f.Name()
	ResetSharedDB()
	if _, err := ConnectCFGDB(dp); err != nil {
		t.Fatal(err)
	}
	ResetSharedDB()
	_, err = os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("%s not exist", dp)
		} else {
			t.Errorf("check %s err:%s", dp, err)
		}
	}
}

func TestConnectCFGDB_AParam_NotExist(t *testing.T) {
	resetSharedDBOnCleanup(t)
	defer ResetSharedDB()
	dp := filepath.Join(
		utils.RandEnAlph(32),
		"cfg"+utils.RandEnAlph(9)+".db",
	)
	dpDir := filepath.Dir(dp)
	if utils.PathExist(dpDir) {
		t.Errorf("please remove directory %s", dpDir)
		return
	}
	if utils.PathExist(dp) {
		t.Errorf("please remove file %s", dp)
		return
	}

	t.Logf("CFG path %s", dp)
	defer os.RemoveAll(dpDir)

	ResetSharedDB()
	if _, err := ConnectCFGDB(dp); err != nil {
		t.Fatal(err)
	}
	ResetSharedDB()
	_, err := os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("%s not exist", dp)
		} else {
			t.Errorf("check %s err:%s", dp, err)
		}
	}
}
