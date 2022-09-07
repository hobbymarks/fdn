/*
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "0.0.0"

var onlyDirectory bool
var inputPaths []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Names",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		if paths, err := RetrievedAbsPaths(inputPaths, onlyDirectory); err != nil {
			log.Fatal(err)
		} else {
			sort.SliceStable(paths, func(i, j int) bool { return paths[i] > paths[j] })
			for _, path := range paths {
				fmt.Println(path)
			}
		}
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
	rootCmd.Flags().BoolVarP(&onlyDirectory, "directory", "d", false, "if enable,directory only.Default file only")
	rootCmd.Flags().StringArrayVarP(&inputPaths, "path", "p", []string{"./"}, "input paths")
}

// Paths form args by flag
func RetrievedAbsPaths(inputPaths []string, onlyDirectory bool) ([]string, error) {
	var absolutePaths []string

	for _, path := range inputPaths {
		if fileInfo, err := os.Stat(path); err != nil {
			log.Error(err)
			continue
		} else {
			if fileInfo.IsDir() {
				paths, err := FilteredSubPaths(path, onlyDirectory)
				if err != nil {
					log.Error(err)
				} else {
					absolutePaths = append(absolutePaths, paths...)
				}
			} else if !onlyDirectory && fileInfo.Mode().IsRegular() {
				if absPath, err := filepath.Abs(path); err != nil {
					log.Error(err)
				} else {
					absolutePaths = append(absolutePaths, absPath)
				}
			} else {
				log.Trace("skipped:", path)
			}
		}
	}

	return absolutePaths, nil
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

//TODO:dry run result buffered for next step
