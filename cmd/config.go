/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	_ "embed"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:embed fdncfg
var configData []byte

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config fdn",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatal(err)
		}
		log.Trace(cfg)
		if cfg == "twl" || cfg == "termkey_colon_termword_list" {
			data := map[string]string{}
			for _, arg := range args {
				kvs := strings.Split(arg, ":")
				data[kvs[0]] = kvs[1]
			}
			if err := ConfigTerm(data); err != nil {
				log.Error(err)
			}
			log.Trace("✓ConfigTerm")
		} else if cfg == "swl" || cfg == "to_separator_word_list" {
			if err := ConfigToSepWords(args); err != nil {
				log.Error(err)
			}
			log.Trace("✓ConfigToBeSeparatorWord") /*✕*/
		} else if cfg == "sep" || cfg == "separator" {
			if err := ConfigSeparator(args[0]); err != nil {
				log.Error(err)
			}
			log.Trace("✓ConfigSeparator")
		}

		lst, err := cmd.Flags().GetString("list")
		if err != nil {
			log.Fatal(err)
		}
		fdncfg, err := GetFDNConfig()
		if err != nil {
			log.Fatal(err)
		}
		log.Trace(lst)
		if lst == "sep" || lst == "separator" {
			fmt.Println(fdncfg.Separator)
		} else if lst == "twl" || lst == "termkey_colon_termword_list" {
			kvs := map[string]string{}
			for _, tw := range fdncfg.TermWords {
				kvs[tw.KeyHash] = tw.OriginalLower + ":" + tw.TargetWord
			}
			fmt.Println("TermWords:")
			for k, v := range kvs {
				fmt.Println(k, v)
			}
		} else if lst == "swl" || lst == "to_separator_word_list" {
			sws := map[string]string{}
			for _, sw := range fdncfg.ToSepWords {
				sws[sw.KeyHash] = sw.Value
			}
			fmt.Println("ToBeSepWords:")
			for k, v := range sws {
				fmt.Println(k, v)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringP("config", "c", "", `Config
	separator                   sep,
	termkey_colon_termword_list twl,
	to_separator_word_list      swl`)
	configCmd.Flags().StringP("list", "l", "", `List
	separator                   sep,
	termkey_colon_termword_list twl,
	to_separator_word_list      swl`)
}
