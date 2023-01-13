package db

import (
	"os"
	"testing"
)

func TestOpenDB(t *testing.T) {
	_, got := OpenDB()
	if got != nil {
		t.Errorf("OpenDB failed:%s", got)
	}
	_, err := os.Stat("fdn.db")
	if err != nil {
		if os.IsNotExist(err) {
			t.Error("fdb.db not exist")
		} else {
			t.Errorf("check fdn.db err:%s", err)
		}
	}
}
