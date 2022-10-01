/*
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hobbymarks/fdn/pb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
				fmt.Println(ReplaceWords(filepath.Base(path)))
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

func ConfigTerm(kvs map[string]string) error {
	if data, err := os.ReadFile(FDNConfigPath); err != nil {
		log.Error(err)
		return err
	} else {
		fdncfg := pb.Fdnconfig{}
		if err := proto.Unmarshal(data, &fdncfg); err != nil {
			log.Error(err)
			return err
		}
		for key, value := range kvs { //FIXME:check exist
			fdncfg.TermWords = append(fdncfg.TermWords, &pb.TermWord{
				KeyHash:       KeyHash(key),
				OriginalLower: key,
				TargetWord:    value})
		}
		fdncfg.LastUpdated = timestamppb.Now()
		log.Trace(fdncfg.GetToSepWords())
		if data, err := proto.Marshal(&fdncfg); err != nil {
			log.Error(err)
			return err
		} else {
			if err := os.WriteFile(FDNConfigPath, data, 0644); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func ConfigToSepWords(words []string) error {
	if data, err := os.ReadFile(FDNConfigPath); err != nil {
		log.Error(err)
		return err
	} else {
		fdncfg := pb.Fdnconfig{}
		if err := proto.Unmarshal(data, &fdncfg); err != nil {
			log.Error(err)
			return err
		}
		//FIXME:check exist
		for _, wd := range words {
			fdncfg.ToSepWords = append(fdncfg.ToSepWords, &pb.ToSepWord{
				KeyHash: KeyHash(wd),
				Value:   wd})
		}
		fdncfg.LastUpdated = timestamppb.Now()
		log.Trace(fdncfg.GetToSepWords())
		if data, err := proto.Marshal(&fdncfg); err != nil {
			log.Error(err)
			return err
		} else {
			if err := os.WriteFile(FDNConfigPath, data, 0644); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func ConfigSeparator(separator string) error {
	if data, err := os.ReadFile(FDNConfigPath); err != nil {
		log.Error(err)
		return err
	} else {
		fdncfg := pb.Fdnconfig{}
		if err := proto.Unmarshal(data, &fdncfg); err != nil {
			log.Error(err)
			return err
		}
		//FIXME:check exist
		fdncfg.Separator = &pb.Separator{
			KeyHash: KeyHash(separator),
			Value:   separator}
		fdncfg.LastUpdated = timestamppb.Now()
		log.Trace(fdncfg.GetSeparator())
		if data, err := proto.Marshal(&fdncfg); err != nil {
			log.Error(err)
			return err
		} else {
			if err := os.WriteFile(FDNConfigPath, data, 0644); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func PrintFDNConfig() error {
	if data, err := os.ReadFile(FDNConfigPath); err != nil {
		log.Error(err)
		return err
	} else {
		fdncfg := pb.Fdnconfig{}
		if err := proto.Unmarshal(data, &fdncfg); err != nil {
			log.Error(err)
			return err
		} //TODO:pretty
		fmt.Println("Separator:", fdncfg.Separator)
		fmt.Println("ToBeSepWords:", fdncfg.ToSepWords)
		kvs := map[string]string{}
		for _, tw := range fdncfg.TermWords {
			kvs[tw.OriginalLower] = tw.TargetWord
		}
		fmt.Println("TermWords:", kvs)
	}
	return nil
}

func GetFDNConfig() (*pb.Fdnconfig, error) {
	fdncfg := pb.Fdnconfig{}
	if data, err := os.ReadFile(FDNConfigPath); err != nil {
		log.Error(err)
		return &fdncfg, err
	} else {
		if err := proto.Unmarshal(data, &fdncfg); err != nil {
			log.Error(err)
			return &fdncfg, err
		}
	}
	return &fdncfg, nil
}

func ReplaceWords(inputName string) string {
	var mask = func(s string) ([]string, []bool) {
		var regescape = func(s string) string {
			s = strings.Replace(s, "+", "\\+", -1)
			s = strings.Replace(s, "?", "\\?", -1)
			s = strings.Replace(s, "*", "\\*", -1)

			return s
		}

		words := []string{}
		wdmsk := []bool{}
		fdncfg, err := GetFDNConfig()
		if err != nil {
			log.Error(err)
		}
		pts := []string{}
		for _, twd := range fdncfg.TermWords {
			pts = append(pts, regescape(twd.OriginalLower))
		}
		for _, wd := range fdncfg.ToSepWords {
			pts = append(pts, regescape(wd.Value))
		}
		rp := regexp.MustCompile(strings.Join(pts, "|"))
		allSliceIndex := rp.FindAllStringIndex(s, -1)
		cur := 0
		for _, slice := range allSliceIndex {
			if slice[0] > cur {
				words = append(words, s[cur:slice[0]])
				wdmsk = append(wdmsk, false)
			}
			words = append(words, s[slice[0]:slice[1]])
			wdmsk = append(wdmsk, true)
			cur = slice[1]
		}
		if cur < len(s) {
			words = append(words, s[cur:])
			wdmsk = append(wdmsk, false)
		}
		return words, wdmsk
	}

	words, bools := mask(inputName)
	fmt.Println(words, bools)

	outName := inputName

	// for _, s := range inputName {

	// }
	return outName
}

func KeyHash(key string) string {
	data := []byte(key)
	return fmt.Sprintf("%x", md5.Sum(data))
}

//TODO:support directory and files
//TODO:dry run result buffered for next step
