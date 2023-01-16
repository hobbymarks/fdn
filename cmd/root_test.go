package cmd

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func pathMaker(_type string) string {
	var f *os.File
	if _type == "d" {

	}
	if _type == "f" {
		_f, err := os.CreateTemp("", "rd*.db")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(_f.Name())

		f = _f

	}
	return f.Name()
}

func TestRetrievedAbsPaths(t *testing.T) {
	rlt, err := RetrievedAbsPaths(
		[]string{"/Users/mm/Downloads/Moore K. Flutter Apprentice 3ed 2022"},
		1,
		true,
	)
	t.Logf("Result:%s", rlt)
	if err != nil {
		t.Errorf("RetrievedAbsPaths error:%s", err)
	}
}
