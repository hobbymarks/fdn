/*
Package cmd config subcommand
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or change separator, term replacements, and separator-like substrings",
	Long: `Reads and writes naming configuration in ~/.fdn/fdn.db: separator character, per-term
replacements (original:target), and literal substrings normalized to that separator.

Subcommands:
  list              print one category (sep, twl, or swl)
  set separator     set the separator string
  add term          add one or more original:target pairs (keys stored lowercased)
  add sepword       register substrings to replace with the separator (outside masked terms)
  delete term       remove terms by original key as configured
  delete sepword    remove separator-words by the same literal strings used when adding`,
	Example: `  fdn config list sep
  fdn config list twl
  fdn config set separator _
  fdn config add term "MyBrand:mybrand" "wiki:wikipedia"
  fdn config add sepword "·" "—"
  fdn config delete term mybrand
  fdn config delete sepword "·"`,
}

var configListCmd = &cobra.Command{
	Use:       "list",
	Short:     "Print configured separator, term mappings, or separator-words",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"sep", "twl", "swl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		kind, err := normalizeConfigKind(args[0])
		if err != nil {
			return err
		}
		return runConfigList(kind)
	},
}

var configSetSeparatorCmd = &cobra.Command{
	Use:   "separator <value>",
	Short: "Set the separator used in FDN names",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ConfigSeparator(args[0]); err != nil {
			log.Error(err)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
}

var configAddTermCmd = &cobra.Command{
	Use:   "term <original:target>...",
	Short: "Add or skip-existing term replacements (case-insensitive match on original)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data := make(map[string]string, len(args))
		for _, arg := range args {
			key, val, ok := strings.Cut(arg, ":")
			if !ok || key == "" || val == "" {
				return fmt.Errorf("invalid term pair %q: want original:target", arg)
			}
			data[key] = val
		}
		if err := ConfigTermWords(data); err != nil {
			log.Error(err)
		}
		return nil
	},
}

var configAddSepwordCmd = &cobra.Command{
	Use:   "sepword <substring>...",
	Short: "Add separator-like substrings (skipped if already present)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ConfigToSepWords(args); err != nil {
			log.Error(err)
		}
		return nil
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add term mappings or separator-words",
}

var configDeleteTermCmd = &cobra.Command{
	Use:   "term <original>...",
	Short: "Delete term mappings by original key",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := DeleteTermWords(args); err != nil {
			log.Error(err)
		}
		return nil
	},
}

var configDeleteSepwordCmd = &cobra.Command{
	Use:   "sepword <substring>...",
	Short: "Delete separator-words by literal substring",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := DeleteToSepWords(args); err != nil {
			log.Error(err)
		}
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove term mappings or separator-words",
}

func normalizeConfigKind(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sep", "separator":
		return "sep", nil
	case "twl", "term", "termkey_colon_termword_list":
		return "twl", nil
	case "swl", "sepword", "to_separator_word_list":
		return "swl", nil
	default:
		return "", fmt.Errorf("unknown kind %q: use sep, twl, or swl", s)
	}
}

func runConfigList(kind string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)

	switch kind {
	case "sep":
		var sep db.Separator
		if rlt := _db.First(&sep); rlt.Error != nil {
			return fmt.Errorf("retrieve Separator: %w", rlt.Error)
		}
		_KVPrint("Separator", map[string]string{sep.KeyHash: sep.Value})
		return nil
	case "twl":
		var termWords []db.TermWord
		if rlt := _db.Find(&termWords); rlt.Error != nil {
			return fmt.Errorf("retrieve TermWord: %w", rlt.Error)
		}
		kvs := map[string]string{}
		for _, tw := range termWords {
			kvs[tw.KeyHash] = tw.OriginalLower + ":" + tw.TargetWord
		}
		_KVPrint("TermWords", kvs)
		return nil
	case "swl":
		var toSepWords []db.ToSepWord
		if rlt := _db.Find(&toSepWords); rlt.Error != nil {
			return fmt.Errorf("retrieve ToSepWord: %w", rlt.Error)
		}
		sws := map[string]string{}
		for _, sw := range toSepWords {
			sws[sw.KeyHash] = sw.Value
		}
		_KVPrint("ToBeSepWords", sws)
		return nil
	default:
		return fmt.Errorf("unknown kind %q", kind)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configListCmd, configSetCmd, configAddCmd, configDeleteCmd)
	configSetCmd.AddCommand(configSetSeparatorCmd)
	configAddCmd.AddCommand(configAddTermCmd, configAddSepwordCmd)
	configDeleteCmd.AddCommand(configDeleteTermCmd, configDeleteSepwordCmd)
}

func _KVPrint(title string, kvs map[string]string) {
	t := table.NewWriter()
	t.SetAutoIndex(true)
	t.SetOutputMirror(os.Stdout)
	t.SetTitle(title)
	t.AppendHeader(table.Row{"KeyID", "Value"})
	for k, v := range kvs {
		t.AppendRow(table.Row{k, v})
	}
	t.AppendSeparator()
	t.Render()
}
