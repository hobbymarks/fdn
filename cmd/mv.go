/*
Package cmd mv subcommand
Copyright © 2023 hobbymarks ihobbymarks@gmail.com
*/
package cmd

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
	Long:  `the utility rename the file or directory named by the source operand to the destination path named by the target operand or moves each file or directory named by a source operand to a destination file or directory in the existing directory named by the directory operand.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			return
		}
		orgPath := args[0]
		tgtPath := args[1]

		// org not exit
		if !utils.PathExist(orgPath) {
			// TODO(hm): Should print error and return if org is not exist
			fmt.Println("NotExist:", orgPath)
			return
		}

		// then org exist
		// tgt exist
		if utils.PathExist(tgtPath) {
			// TODO(hm): Should check if the target is a directory and the source is a file ,and then move the file to the directory
			fmt.Println("Target Exist:", tgtPath)
			return
		}
		// TODO(hm): Should check if the target is a file and the source is a directory,and then print error and return
		// tgt not exist
		err := FDNFile(orgPath, tgtPath, false)
		if err != nil {
			log.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
