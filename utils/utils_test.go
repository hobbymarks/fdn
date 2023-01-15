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
