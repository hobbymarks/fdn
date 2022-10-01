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

func ConfigTermWords(keyValueMap map[string]string) error {
	if fdncfg, err := GetFDNConfig(); err != nil {
		log.Error(err)
		return err
	} else {
		ekhs := []string{}
		for _, tw := range fdncfg.TermWords {
			ekhs = append(ekhs, tw.KeyHash)
		}
		for key, value := range keyValueMap {
			kh := KeyHash(key)
			if ArrayContainsElemenet(ekhs, kh) {
				continue
			} else {
				fdncfg.TermWords = append(fdncfg.TermWords, &pb.TermWord{
					KeyHash:       kh,
					OriginalLower: key,
					TargetWord:    value})
				ekhs = append(ekhs, kh)
			}
		}
		fdncfg.LastUpdated = timestamppb.Now()
		log.Trace(fdncfg.GetToSepWords())
		if err := SaveFDNConfig(fdncfg); err != nil {
			return err
		}
	}
	return nil
}

func DeleteTermWords(keys []string) error {
	if fdncfg, err := GetFDNConfig(); err != nil {
		log.Error(err)
		return err
	} else {
		log.Trace(keys)
		tws := []*pb.TermWord{}
		for _, tw := range fdncfg.TermWords {
			if ArrayContainsElemenet(keys, tw.KeyHash) {
				log.Trace("-:" + tw.OriginalLower)
				continue
			} else {
				tws = append(tws, tw)
			}
		}
		fdncfg.TermWords = tws
		fdncfg.LastUpdated = timestamppb.Now()
		if err := SaveFDNConfig(fdncfg); err != nil {
			return err
		}
	}
	return nil
}

func ConfigToSepWords(words []string) error {
	if fdncfg, err := GetFDNConfig(); err != nil {
		log.Error(err)
		return err
	} else {
		ekhs := []string{}
		for _, sw := range fdncfg.ToSepWords {
			ekhs = append(ekhs, sw.KeyHash)
		}
		for _, wd := range words {
			kh := KeyHash(wd)
			if ArrayContainsElemenet(ekhs, kh) {
				continue
			} else {
				fdncfg.ToSepWords = append(fdncfg.ToSepWords, &pb.ToSepWord{
					KeyHash: kh,
					Value:   wd})
				ekhs = append(ekhs, kh)
			}
		}
		fdncfg.LastUpdated = timestamppb.Now()
		log.Trace(fdncfg.GetToSepWords())
		if err := SaveFDNConfig(fdncfg); err != nil {
			return err
		}
	}
	return nil
}

func DeleteToSepWords(keys []string) error {
	if fdncfg, err := GetFDNConfig(); err != nil {
		log.Error(err)
		return err
	} else {
		log.Trace(keys)
		sws := []*pb.ToSepWord{}
		for _, sw := range fdncfg.ToSepWords {
			if ArrayContainsElemenet(keys, sw.KeyHash) {
				log.Trace("-:" + sw.Value)
				continue
			} else {
				sws = append(sws, sw)
			}
		}
		fdncfg.ToSepWords = sws
		fdncfg.LastUpdated = timestamppb.Now()
		if err := SaveFDNConfig(fdncfg); err != nil {
			return err
		}
	}
	return nil
}

func ConfigSeparator(separator string) error {
	if fdncfg, err := GetFDNConfig(); err != nil {
		log.Error(err)
		return err
	} else {
		fdncfg.Separator = &pb.Separator{
			KeyHash: KeyHash(separator),
			Value:   separator}
		fdncfg.LastUpdated = timestamppb.Now()
		log.Trace(fdncfg.GetSeparator())
		if err := SaveFDNConfig(fdncfg); err != nil {
			return err
		}
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

func SaveFDNConfig(fdncfg *pb.Fdnconfig) error {
	if data, err := proto.Marshal(fdncfg); err != nil {
		log.Error(err)
		return err
	} else {
		if err := os.WriteFile(FDNConfigPath, data, 0644); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
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
	if len(words) != len(bools) {
		log.Fatal("words not equal bools")
	}

	newWords := []string{}
	if fdncfg, err := GetFDNConfig(); err != nil {
		log.Fatal(err)
	} else {
		for idx, wd := range words {
			if !bools[idx] {
				for _, sw := range fdncfg.ToSepWords {
					wd = strings.Replace(wd, sw.Value, fdncfg.Separator.Value, -1)
				}
			}
			newWords = append(newWords, wd)
		}
	}

	outName := strings.Join(newWords, "")

	return outName
}

func KeyHash(key string) string {
	data := []byte(key)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func ArrayContainsElemenet[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

//TODO:support directory and files
//TODO:dry run result buffered for next step
