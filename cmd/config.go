/*
Package cmd config subcommand
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"log/slog"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
	"github.com/jedib0t/go-pretty/v6/table"
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
  delete sepword    remove separator-words by the same literal strings used when adding
  reset             reset all configuration to built-in defaults`,
	Example: `  fdn config list sep
  fdn config list twl
  fdn config set separator _
  fdn config add term "MyBrand:mybrand" "wiki:wikipedia"
  fdn config add sepword "·" "—"
  fdn config delete term mybrand
  fdn config delete sepword "·"
  fdn config reset`,
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
			slog.Error(err.Error())
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
	Short: "Add or update term replacements (case-insensitive match on original)",
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
			slog.Error(err.Error())
		}
		return nil
	},
}

var configAddSepwordCmd = &cobra.Command{
	Use:   "sepword <substring>...",
	Short: "Add or update separator-like substrings",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ConfigToSepWords(args); err != nil {
			slog.Error(err.Error())
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
			slog.Error(err.Error())
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
			slog.Error(err.Error())
		}
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove term mappings or separator-words",
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all configuration to built-in defaults",
	Long: `Removes all user-defined configuration entries and restores built-in defaults.
User-added term words, separator words, and separator settings will be lost.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigReset()
	},
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
	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}

	switch kind {
	case "sep":
		var sep db.Separator
		if result := conn.First(&sep); result.Error != nil {
			return fmt.Errorf("retrieve Separator: %w", result.Error)
		}
		KVPrint("Separator", []kvEntry{{Key: sep.KeyHash, Source: sep.Source, Value: sep.Value}})
		return nil
	case "twl":
		var termWords []db.TermWord
		if result := conn.Find(&termWords); result.Error != nil {
			return fmt.Errorf("retrieve TermWord: %w", result.Error)
		}
		entries := make([]kvEntry, len(termWords))
		for i, tw := range termWords {
			entries[i] = kvEntry{Key: tw.KeyHash, Source: tw.Source, Value: tw.OriginalLower + ":" + tw.TargetWord}
		}
		KVPrint("TermWords", entries)
		return nil
	case "swl":
		var toSepWords []db.ToSepWord
		if result := conn.Find(&toSepWords); result.Error != nil {
			return fmt.Errorf("retrieve ToSepWord: %w", result.Error)
		}
		entries := make([]kvEntry, len(toSepWords))
		for i, sw := range toSepWords {
			entries[i] = kvEntry{Key: sw.KeyHash, Source: sw.Source, Value: sw.Value}
		}
		KVPrint("ToBeSepWords", entries)
		return nil
	default:
		return fmt.Errorf("unknown kind %q", kind)
	}
}

func runConfigReset() error {
	fdnDir, err := utils.FDNDir()
	if err != nil {
		return err
	}
	dbPath := db.DefaultFDNDBPath()
	if err := db.ResetCFG(dbPath); err != nil {
		return fmt.Errorf("reset config: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Configuration reset to built-in defaults in %s\n", fdnDir)
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configListCmd, configSetCmd, configAddCmd, configDeleteCmd, configResetCmd)
	configSetCmd.AddCommand(configSetSeparatorCmd)
	configAddCmd.AddCommand(configAddTermCmd, configAddSepwordCmd)
	configDeleteCmd.AddCommand(configDeleteTermCmd, configDeleteSepwordCmd)
}

type kvEntry struct {
	Key    string
	Source string
	Value  string
}

func KVPrint(title string, entries []kvEntry) {
	t := table.NewWriter()
	t.SetAutoIndex(true)
	t.SetOutputMirror(os.Stdout)
	t.SetTitle(title)
	t.AppendHeader(table.Row{"KeyID", "Source", "Value"})
	for _, e := range entries {
		t.AppendRow(table.Row{e.Key, e.Source, e.Value})
	}
	t.AppendSeparator()
	t.Render()
}
