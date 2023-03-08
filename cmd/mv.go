/*
Copyright © 2023 hobbymarks ihobbymarks@gmail.com
*/package cmd

import (
	"fmt"

	"github.com/hobbymarks/fdn/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:   "mv",
	Short: "move files",
	Long:  `the utility renames the file named by the source operand to the destination path named by the target operand or moves each file named by a source operand to a destination file in the existing directory named by the directory perand.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			return
		}
		orgPath := args[0]
		tgtPath := args[1]
		if !utils.PathExist(orgPath) {
			fmt.Println("NotExist:", orgPath)
			return
		}
		if utils.PathExist(tgtPath) {
			fmt.Println("Target Exist:", tgtPath)
			return
		}
		err := FDNFile(orgPath, tgtPath, false)
		if err != nil {
			log.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
