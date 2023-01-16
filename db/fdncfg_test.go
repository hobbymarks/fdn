package db

import (
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/hobbymarks/fdn/utils"
)

func TestConnectCFGDB_NoParam(t *testing.T) {
	dp := filepath.Join(utils.FDNDir(), "cfg.db")
	if utils.PathExist(dp) {
		t.Errorf("please remove file %s", dp)
		return
	}
	ConnectCFGDB()
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

func TestConnectCFGDB_AParam_Exist(t *testing.T) {
	f, err := os.CreateTemp("", "cfg*.db")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("Temp file %s", f.Name())
	defer os.Remove(f.Name())

	dp := f.Name()
	ConnectCFGDB(dp)
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

	ConnectCFGDB(dp)
	_, err := os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("%s not exist", dp)
		} else {
			t.Errorf("check %s err:%s", dp, err)
		}
	}
}
