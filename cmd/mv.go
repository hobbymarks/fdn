/*
Package cmd mv subcommand
Copyright © 2023 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"path/filepath"

	"github.com/hobbymarks/fdn/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// mv moves originPath to a resolved destination. Evaluation order:
//
//	origin exists?
//	  no  → error
//	  yes → target exists?
//	          no  → rename origin to targetPath (new basename or path)
//	          yes → target is directory?
//	                  yes → rename origin to filepath.Join(target, filepath.Base(origin))
//	                  no  → error (existing non-directory target)
var mvCmd = &cobra.Command{
	Use:   "mv SOURCE DEST",
	Short: "Move or rename a file or directory (updates FDN rename records)",
	Long: `Move or rename SOURCE to DEST using the same journal logic as the main fdn rename.

Behavior:
  • If DEST does not exist — SOURCE is renamed to DEST (new name or new path).
  • If DEST exists and is a directory — SOURCE is placed inside it as DEST/basename(SOURCE).
  • If DEST exists and is not a directory — the command fails (will not overwrite a file).

SOURCE must exist. Paths with spaces must be quoted in the shell.`,
	Example: `  fdn mv ./a.txt ./b.txt
  fdn mv ./doc.pdf ./backup/
  fdn mv "My File.txt" ./inbox/My_File.txt`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		originPath := args[0]
		targetPath := args[1]

		// origin path is not exist
		if !utils.PathExist(originPath) {
			log.Error("Origin path is not exist:", originPath)
			return
		}
		// then origin path is 	exist
		// and target path is also exist
		if utils.PathExist(targetPath) {
			if utils.PathIsDirectory(targetPath) {
				newTargetPath := filepath.Join(targetPath, filepath.Base(originPath))

				same, err := utils.SameFiles(newTargetPath, originPath)
				if err == nil && same {
					log.Warnf("Target path '%s' is the same as origin path '%s'", newTargetPath, originPath)
					return
				}

				err = FDNFile(originPath, newTargetPath, false)
				if err != nil {
					log.Error("Error when move file from", originPath, " to ", newTargetPath, ":", err)
					return
				}

				log.Info("Success move file from", originPath, " to ", newTargetPath)
				return
			}
			log.Error("Target path is not a directory:", targetPath)
			return
		}
		// target path is not exist, then move the origin path to the target path
		err := FDNFile(originPath, targetPath, false)
		if err != nil {
			log.Error("Error when move file from", originPath, " to ", targetPath, ":", err)
			return
		}
		log.Info("Success move file from", originPath, " to ", targetPath)
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
