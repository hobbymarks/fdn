package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_KVPrint(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = buf.ReadFrom(r)
		close(done)
	}()

	KVPrint("Title", []kvEntry{{Key: "k1", Source: "user", Value: "v1"}, {Key: "k2", Source: "builtin", Value: "v2"}})

	_ = w.Close()
	os.Stdout = old
	<-done
	_ = r.Close()

	out := buf.String()
	assert.Contains(t, out, "Title")
	assert.Contains(t, out, "k1")
	assert.Contains(t, out, "v1")
}

func executeRoot(t *testing.T, args ...string) {
	t.Helper()
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs(args)
	t.Cleanup(func() { rootCmd.SetArgs(nil) })
	assert.NoError(t, rootCmd.Execute())
}

func Test_configCmd_listSeparator(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "list", "sep")
}

func Test_configCmd_listTermWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "list", "twl")
}

func Test_configCmd_listToSepWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "list", "swl")
}

func Test_configCmd_listAliases(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "list", "separator")
	executeRoot(t, "config", "list", "termkey_colon_termword_list")
	executeRoot(t, "config", "list", "to_separator_word_list")
}

func Test_configCmd_configSeparator(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "set", "separator", "_")
}

func Test_configCmd_configSeparatorAlias(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "set", "separator", "z")
}

func Test_configCmd_configToSepWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "add", "sepword", "§")
}

func Test_configCmd_configTermWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "add", "term", "mykey:myval")
}

func Test_configCmd_configTermWordsMultiple(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "add", "term", "k1:v1", "k2:v2")
}

func Test_configCmd_deleteToSepWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	_ = ConfigToSepWords([]string{"delme_sep_word"})
	executeRoot(t, "config", "delete", "sepword", "delme_sep_word")
}

func Test_configCmd_deleteTermWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	_ = ConfigTermWords(map[string]string{"delterm": "v"})
	executeRoot(t, "config", "delete", "term", "delterm")
}

func Test_configCmd_configTermWordsLongAlias(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "add", "term", "longaliaskey:longaliasval")
}

func Test_configCmd_listLongAliases(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "list", "separator")
	executeRoot(t, "config", "list", "to_separator_word_list")
}

func Test_configCmd_addSepwordThenListSep(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "add", "sepword", "+")
	executeRoot(t, "config", "list", "sep")
}

func Test_normalizeConfigKind_invalid(t *testing.T) {
	_, err := normalizeConfigKind("nope")
	assert.Error(t, err)
}

func Test_runConfigReset(t *testing.T) {
	seedEmbeddedCfgDB(t)
	executeRoot(t, "config", "reset")
	executeRoot(t, "config", "list", "sep")
}
