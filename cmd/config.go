/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
		cfg, err := cmd.Flags().GetStringArray("config")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(cfg)
		data := ""
		for _, arg := range args {
			data += arg
		}
		fmt.Println(data)
		// bytes, err := os.ReadFile(cfg)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// var result map[string]interface{}
		// json.Unmarshal(bytes, &result)

		// for key, _ := range result {
		// 	fmt.Println(key)
		// }
		// fmt.Println(result["BeReplacedCharDictionary"])
		// cmds := "ls -l"
		// out, err := exec.Command(strings.Fields(cmds)[0], strings.Fields(cmds)[1:]...).Output()
		// if err != nil {
		// 	log.Error(err)
		// } else {
		// 	fmt.Println(string(out))
		// }
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	configCmd.Flags().StringArrayP("config", "c", []string{""}, "Config")
}

func UpdateConfigTerm(kvs map[string]string) error {
	return nil
}
