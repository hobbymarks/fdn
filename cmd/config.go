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
		if cfg == "kcvl" || cfg == "key_colon_value_list" {
			data := map[string]string{}
			for _, arg := range args {
				kvs := strings.Split(arg, ":")
				data[kvs[0]] = kvs[1]
			}
			if err := ConfigTerm(data); err != nil {
				log.Error(err)
			}
			log.Trace("✓ConfigTerm")
		} else if cfg == "tbswl" || cfg == "to_be_separator_word_list" {
			if err := ConfigToBeSepWords(args); err != nil {
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
		} else if lst == "tws" || lst == "termwords" {
			kvs := map[string]string{}
			for _, tw := range fdncfg.TermWords {
				kvs[tw.OriginalLower] = tw.TargetWord
			}
			fmt.Println("TermWords:", kvs)
		} else if lst == "sws" || lst == "sepwords" {
			fmt.Println("ToBeSepWords:", fdncfg.ToBeSepWords)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringP("config", "c", "", `Config
	separator                     sep,
	termkey_colon_termvalue_list  kcvl,
	to_be_separator_word_list     tbswl`)
	configCmd.Flags().StringP("list", "l", "", `List
	separator sep,
	termwords tws
	sepwords  sws`)
}
