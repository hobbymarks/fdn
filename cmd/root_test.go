/*
Package cmd root test
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrievedAbsPaths(t *testing.T) {
	assert := assert.New(t)
	_, err := RetrievedAbsPaths(
		[]string{"/Users/mm/Downloads/Moore K. Flutter Apprentice 3ed 2022"},
		1,
		true,
	)
	assert.NoError(err, "should no error")
}

func TestFDNedFrom(t *testing.T) {
	seedEmbeddedCfgDB(t)
	out, err := FDNedFrom("123 456 789")
	assert.NoError(t, err)
	assert.Equal(t, "123_456_789", out)
}
