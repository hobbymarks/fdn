/*
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/hobbymarks/fdn/pb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var version = "0.0.0"

var onlyDirectory bool
var inputPaths []string
var depthLevel int

var FDNConfigPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Names",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		if paths, err := RetrievedAbsPaths(inputPaths, depthLevel, onlyDirectory); err != nil {
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
	rootCmd.Flags().IntVarP(&depthLevel, "level", "l", 1, "maxdepth level")
	rootCmd.Flags().StringArrayVarP(&inputPaths, "path", "p", []string{"."}, "input paths")

	homeDir, err := os.UserHomeDir() //get home dir
	if err != nil {
		log.Fatal(err)
	}
	FDNConfigPath = filepath.Join(homeDir, ".fdn")
	if _, err := os.Lstat(FDNConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(FDNConfigPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	FDNConfigPath = filepath.Join(FDNConfigPath, "fdncfg")
	if _, err := os.Stat(FDNConfigPath); err != nil { //if not exist then create
		giatrds := pb.Fdnconfig{}
		if data, err := proto.Marshal(&giatrds); err != nil {
			log.Fatal(err)
		} else {
			if err := os.WriteFile(FDNConfigPath, data, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Paths form args by flag
func RetrievedAbsPaths(inputPaths []string, depthLevel int, onlyDirectory bool) ([]string, error) {
	var absolutePaths []string

	for _, path := range inputPaths {
		if fileInfo, err := os.Stat(path); err != nil {
			log.Error(err)
			continue
		} else {
			if fileInfo.IsDir() {
				paths, err := FilteredSubPaths(path, depthLevel, onlyDirectory)
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
func FilteredSubPaths(dirPath string, depthLevel int, OnlyDir bool) ([]string, error) {
	var absolutePaths []string

	dirPath = filepath.Clean(dirPath)
	log.Trace(dirPath)
	if depthLevel == -1 {
		err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				log.Trace(err)
				return err
			}
			if (OnlyDir && info.IsDir()) || (!OnlyDir && info.Type().IsRegular()) {
				log.Trace("isDir:", path)
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
	} else {
		if paths, err := DepthFiles(dirPath, depthLevel, OnlyDir); err != nil {
			log.Error(err)
			return nil, err
		} else {
			absolutePaths = paths
		}
	}
	return absolutePaths, nil
}

// Depth read dir
func DepthFiles(dirPath string, depthLevel int, onlyDirectory bool) ([]string, error) {
	var absolutePaths []string

	log.Debug(depthLevel)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, file := range files {
		if absPath, err := filepath.Abs(filepath.Join(dirPath, file.Name())); err != nil {
			log.Error(err)
		} else {
			if (onlyDirectory && file.IsDir()) || (!onlyDirectory && file.Type().IsRegular()) {
				absolutePaths = append(absolutePaths, absPath)
			}
			if depthLevel > 1 && file.Type().IsDir() {
				if files, err := DepthFiles(absPath, depthLevel-1, onlyDirectory); err != nil {
					log.Error(err)
				} else {
					absolutePaths = append(absolutePaths, files...)
				}
			}
		}
	}
	return absolutePaths, nil
}

//TODO:support directory and files
//TODO:dry run result buffered for next step
