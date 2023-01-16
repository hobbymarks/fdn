package db

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hobbymarks/fdn/utils"
)

func TestConnectRDDB_NoParam(t *testing.T) {
	dp := filepath.Join(utils.FDNDir(), "rd.db")
	if utils.PathExist(dp) {
		t.Errorf("please remove file %s", dp)
		return
	}
	ConnectRDDB()
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

func TestConnectRDDB_AParam_Exist(t *testing.T) {
	f, err := os.CreateTemp("", "rd*.db")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("Temp file %s", f.Name())
	defer os.Remove(f.Name())

	dp := f.Name()
	ConnectRDDB(dp)
	_, err = os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("%s not exist", dp)
		} else {
			t.Errorf("check %s err:%s", dp, err)
		}
	}
}

func TestConnectRDDB_AParam_NotExist(t *testing.T) {
	dp := filepath.Join(utils.RandStr(32), "rd"+utils.RandStr(9)+".db")
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

	ConnectRDDB(dp)
	_, err := os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("%s not exist", dp)
		} else {
			t.Errorf("check %s err:%s", dp, err)
		}
	}
}
