package cmd

import (
	"bytes"
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

	_KVPrint("Title", map[string]string{"k1": "v1", "k2": "v2"})

	_ = w.Close()
	os.Stdout = old
	<-done
	_ = r.Close()

	out := buf.String()
	assert.Contains(t, out, "Title")
	assert.Contains(t, out, "k1")
	assert.Contains(t, out, "v1")
}

func Test_configCmd_listSeparator(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("list", "sep"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", ""))
}

func Test_configCmd_listTermWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("list", "twl"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", ""))
}

func Test_configCmd_listToSepWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("list", "swl"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", ""))
}

func Test_configCmd_listAliases(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("list", "separator"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", "termkey_colon_termword_list"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", "to_separator_word_list"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", ""))
}

func Test_configCmd_configSeparator(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "sep"))
	configCmd.Run(configCmd, []string{"_"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
}

func Test_configCmd_configSeparatorAlias(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "separator"))
	configCmd.Run(configCmd, []string{"z"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
}

func Test_configCmd_configToSepWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "swl"))
	configCmd.Run(configCmd, []string{"§"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
}

func Test_configCmd_configTermWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "twl"))
	configCmd.Run(configCmd, []string{"mykey:myval"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
}

func Test_configCmd_configTermWordsMultiple(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "twl"))
	configCmd.Run(configCmd, []string{"k1:v1", "k2:v2"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
}

func Test_configCmd_deleteToSepWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	_ = ConfigToSepWords([]string{"delme_sep_word"})
	assert.NoError(t, configCmd.Flags().Set("delete", "swl"))
	configCmd.Run(configCmd, []string{"delme_sep_word"})
	assert.NoError(t, configCmd.Flags().Set("delete", ""))
}

func Test_configCmd_deleteTermWords(t *testing.T) {
	seedEmbeddedCfgDB(t)
	_ = ConfigTermWords(map[string]string{"delterm": "v"})
	assert.NoError(t, configCmd.Flags().Set("delete", "twl"))
	configCmd.Run(configCmd, []string{"delterm"})
	assert.NoError(t, configCmd.Flags().Set("delete", ""))
}

func Test_configCmd_configTermWordsLongAlias(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "termkey_colon_termword_list"))
	configCmd.Run(configCmd, []string{"longaliaskey:longaliasval"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
}

func Test_configCmd_listLongAliases(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("list", "separator"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", "to_separator_word_list"))
	configCmd.Run(configCmd, nil)
	assert.NoError(t, configCmd.Flags().Set("list", ""))
}

func Test_configCmd_combinedFlagsNoExit(t *testing.T) {
	seedEmbeddedCfgDB(t)
	assert.NoError(t, configCmd.Flags().Set("config", "swl"))
	assert.NoError(t, configCmd.Flags().Set("list", "sep"))
	configCmd.Run(configCmd, []string{"+"})
	assert.NoError(t, configCmd.Flags().Set("config", ""))
	assert.NoError(t, configCmd.Flags().Set("list", ""))
}
