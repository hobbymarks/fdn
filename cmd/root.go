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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Names",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		onlyDirectory, err := cmd.Flags().GetBool("directory")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(onlyDirectory)
		notUnitTest, err := cmd.Flags().GetString("nut")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(notUnitTest)
		for _, arg := range args {
			if fileInfo, err := os.Stat(arg); err != nil {
				log.Error(err)
				continue
			} else {
				if fileInfo.IsDir() {
					fmt.Println("isDir:", arg)
				} else if fileInfo.Mode().IsRegular() {
					fmt.Println("isRegular:", arg)
				} else {
					continue
				}
			}
		}
		fmt.Println("root called.", len(args))
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fdn.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("directory", "d", false, "directory only")
	rootCmd.Flags().StringP("nut", "t", "NotUnitTest", "not unit test")
}

func DirFiles(dirPath string) ([]string, error) {
	var relativePaths []string
	dirPath = filepath.Clean(dirPath)
	log.Trace(dirPath)
	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Trace(err)
			return err
		}
		if info.IsDir() {
			log.Trace("isDir:", path)
			return nil
		}
		if info.Type().IsRegular() {
			relativePath := path
			if dirPath != "." {
				relativePath = path[len(dirPath)+1:]
			}
			relativePaths = append(relativePaths, filepath.ToSlash(relativePath))
		} else {
			log.Warn("skipped:", path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return relativePaths, nil
}
