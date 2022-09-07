/*
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "0.0.0"

var onlyDirectory bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Names",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{"./"}
		}
		for _, arg := range args {
			if fileInfo, err := os.Stat(arg); err != nil {
				log.Error(err)
				continue
			} else {
				if fileInfo.IsDir() {
					paths, err := FilteredSubPaths(arg, onlyDirectory)
					if err != nil {
						log.Error(err)
					} else {
						fmt.Println("isDir:", paths)
					}
				} else if fileInfo.Mode().IsRegular() {
					fmt.Println("isRegular:", arg)
				} else {
					continue
				}
			}
		}
		fmt.Println("root called.", len(args), onlyDirectory)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&onlyDirectory, "directory", "d", false, "directory only")
}

// retrieve absolute paths
func FilteredSubPaths(dirPath string, onlyDirectory bool) ([]string, error) {
	var absolutePaths []string
	dirPath = filepath.Clean(dirPath)
	log.Trace(dirPath)
	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Trace(err)
			return err
		}
		if onlyDirectory && info.IsDir() {
			log.Trace("isDir:", path)
			if absPath, err := filepath.Abs(filepath.Join(dirPath, path)); err != nil {
				log.Error(err)
			} else {
				absolutePaths = append(absolutePaths, absPath)
			}
			return nil
		}
		if !onlyDirectory && info.Type().IsRegular() {
			if absPath, err := filepath.Abs(filepath.Join(dirPath, path)); err != nil {
				log.Error(err)
			} else {
				absolutePaths = append(absolutePaths, absPath)
			}
			return nil
		}
		log.Trace("skipped:", path)

		return nil
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return absolutePaths, nil
}
