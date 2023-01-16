package cmd

import (
	"testing"
)

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
