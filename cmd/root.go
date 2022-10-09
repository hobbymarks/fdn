/*
Package cmd root subcommand is the default
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"crypto/md5"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/hobbymarks/fdn/pb"
	"github.com/hobbymarks/go-difflib/difflib"
	"github.com/mattn/go-runewidth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var version = "0.0.0"

var onlyDirectory bool
var inputPaths []string
var depthLevel int
var inplace bool
var cfm bool
var reverse bool
var fullpath bool
var plainStyle bool

var pretty bool

//go:embed fdncfg
var defaultFDNCFGBytes []byte

// FDNConfigPath is the config file path
var FDNConfigPath string

// FDNRecordPath is the record file path
var FDNRecordPath string

var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Names",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		PrintTipFlag := false
		curNameHashPreNameMap := map[string]string{}
		if reverse {
			fdnrd, err := GetFDNRecord()
			if err != nil {
				log.Fatal(err)
			}
			for _, rd := range fdnrd.Records {
				curNameHashPreNameMap[rd.CurrentNameHash] = rd.PreviousName
			}
		}
		paths, err := RetrievedAbsPaths(inputPaths, depthLevel, onlyDirectory)
		if err != nil {
			log.Fatal(err)
		}
		sort.SliceStable(paths, func(i, j int) bool { return paths[i] > paths[j] })
		for _, path := range paths {
			path = filepath.Clean(path) //remove tailing slash if exist
			toPath := ""
			if reverse {
				preName, exist := curNameHashPreNameMap[KeyHash(filepath.Base(path))]
				if exist {
					toPath = filepath.Join(filepath.Dir(path), preName)
				}
			} else {
				ext := filepath.Ext(path) //ext empty if path is dir
				bn := strings.TrimSuffix(filepath.Base(path), ext)
				if fdned := FNDedFrom(bn); fdned != bn {
					toPath = filepath.Join(filepath.Dir(path), fdned+ext)
				}
			}
			if toPath == "" {
				continue
			}
			if inplace {
				if err := FDNFile(path, toPath, reverse); err != nil {
					log.Error(err)
				} else {
					OutputResult(path, toPath, inplace, fullpath)
				}
			} else {
				if cfm {
					switch confirm() {
					case A, All:
						inplace = true
						if err := FDNFile(path, toPath, reverse); err != nil {
							log.Error(err)
						} else {
							OutputResult(path, toPath, true, fullpath)
						}
					case Y, Yes:
						if err := FDNFile(path, toPath, reverse); err != nil {
							log.Error(err)
						} else {
							OutputResult(path, toPath, true, fullpath)
						}
					case N, No:
						OutputResult(path, toPath, false, fullpath)
						continue
					case Q, Quit:
						os.Exit(0)
					}
				} else {
					PrintTipFlag = true
					OutputResult(path, toPath, false, fullpath)
				}
			}
		}
		if PrintTipFlag {
			noEffectTip()
		}
	},
}

// Execute is the cmd entry
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&onlyDirectory, "directory", "d", false, "If enable,directory only.Default file only")
	rootCmd.Flags().IntVarP(&depthLevel, "level", "l", 1, "Maxdepth level")
	rootCmd.Flags().StringArrayVarP(&inputPaths, "path", "p", []string{"."}, "Input paths")
	rootCmd.Flags().BoolVarP(&inplace, "inplace", "i", false, "In-place")
	rootCmd.Flags().BoolVarP(&cfm, "confirm", "c", false, "Confirm")
	rootCmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "Reverse")
	rootCmd.Flags().BoolVarP(&fullpath, "fullpath", "f", false, "FullPath")
	rootCmd.Flags().BoolVarP(&plainStyle, "plainstyle", "s", false, "PlainStyle Output")
	rootCmd.Flags().BoolVarP(&pretty, "pretty", "e", false, "Pretty Display")

	homeDir, err := os.UserHomeDir() //get home dir
	if err != nil {
		log.Fatal(err)
	}
	//Process FDNConfig
	FDNConfigPath = filepath.Join(homeDir, ".fdn")
	if _, err := os.Lstat(FDNConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(FDNConfigPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	FDNConfigPath = filepath.Join(FDNConfigPath, "fdncfg")
	if _, err := os.Lstat(FDNConfigPath); errors.Is(err, os.ErrNotExist) {
		fdncfg := pb.Fdnconfig{}
		if err := proto.Unmarshal(defaultFDNCFGBytes, &fdncfg); err != nil {
			log.Error(err)
		}
		if err := SaveFDNConfig(&fdncfg); err != nil {
			log.Error(err)
		}
	}
	//Process FDNRecord
	FDNRecordPath = filepath.Join(homeDir, ".fdn", "fdnrd")
	if _, err := os.Lstat(FDNRecordPath); errors.Is(err, os.ErrNotExist) {
		fdnrd := pb.Fdnrecord{}
		if err := SaveFDNRecord(&fdnrd); err != nil {
			log.Error(err)
		}
	}
}

// RetrievedAbsPaths Paths form args by flag
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

// FilteredSubPaths retrieve absolute paths
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
		paths, err := DepthFiles(dirPath, depthLevel, OnlyDir)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		absolutePaths = paths
	}
	return absolutePaths, nil
}

// DepthFiles Depth read dir
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

// ConfigTermWords to config term words
func ConfigTermWords(keyValueMap map[string]string) error {
	fdncfg, err := GetFDNConfig()
	if err != nil {
		log.Error(err)
		return err
	}
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
	return nil
}

// DeleteTermWords delete term words in config file
func DeleteTermWords(keys []string) error {
	fdncfg, err := GetFDNConfig()
	if err != nil {
		log.Error(err)
		return err
	}
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
	return nil
}

// ConfigToSepWords config tosep words in config file
func ConfigToSepWords(words []string) error {
	fdncfg, err := GetFDNConfig()
	if err != nil {
		log.Error(err)
		return err
	}
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
	return nil
}

// DeleteToSepWords delete tosep words in config file
func DeleteToSepWords(keys []string) error {
	fdncfg, err := GetFDNConfig()
	if err != nil {
		log.Error(err)
		return err
	}
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
	return nil
}

// ConfigSeparator config separator in config file
func ConfigSeparator(separator string) error {
	fdncfg, err := GetFDNConfig()
	if err != nil {
		log.Error(err)
		return err
	}
	fdncfg.Separator = &pb.Separator{
		KeyHash: KeyHash(separator),
		Value:   separator}
	fdncfg.LastUpdated = timestamppb.Now()
	log.Trace(fdncfg.GetSeparator())
	if err := SaveFDNConfig(fdncfg); err != nil {
		return err
	}
	return nil
}

// GetFDNConfig return fdnconfig pointer and error info
func GetFDNConfig() (*pb.Fdnconfig, error) {
	fdncfg := pb.Fdnconfig{}
	data, err := os.ReadFile(FDNConfigPath)
	if err != nil {
		log.Error(err)
		return &fdncfg, err
	}
	if err := proto.Unmarshal(data, &fdncfg); err != nil {
		log.Error(err)
		return &fdncfg, err
	}
	return &fdncfg, nil
}

// SaveFDNConfig save fdn config to config file
func SaveFDNConfig(fdncfg *pb.Fdnconfig) error {
	data, err := proto.Marshal(fdncfg)
	if err != nil {
		log.Error(err)
		return err
	}
	if err := os.WriteFile(FDNConfigPath, data, 0644); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// GetFDNRecord return fdnrecord and error
func GetFDNRecord() (*pb.Fdnrecord, error) {
	fdnrd := pb.Fdnrecord{}
	data, err := os.ReadFile(FDNRecordPath)
	if err != nil {
		log.Error(err)
		return &fdnrd, err
	}
	if err := proto.Unmarshal(data, &fdnrd); err != nil {
		log.Error(err)
		return &fdnrd, err
	}
	return &fdnrd, nil
}

// SaveFDNRecord save fdn record to fdn record file
func SaveFDNRecord(fdnrd *pb.Fdnrecord) error {
	data, err := proto.Marshal(fdnrd)
	if err != nil {
		log.Error(err)
		return err
	}
	if err := os.WriteFile(FDNRecordPath, data, 0644); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// ReplaceWords process inputName string and return new string
func ReplaceWords(inputName string) string {
	outName := inputName
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
		sep := fdncfg.Separator.Value
		rpCNS := regexp.MustCompile("[" + sep + "]+")
		for idx, wd := range words {
			if !bools[idx] {
				for _, sw := range fdncfg.ToSepWords {
					wd = strings.Replace(wd, sw.Value, sep, -1)
				}
			}
			newWords = append(newWords, wd)
		}
		outName = strings.Join(newWords, "")
		//Process continous separator
		outName = rpCNS.ReplaceAllString(outName, sep)
	}

	return outName
}

// ProcessHeadTail process head and tail of input string
func ProcessHeadTail(inputName string) string {
	outName := inputName
	fdncfg, err := GetFDNConfig()
	if err != nil {
		log.Fatal(err)
	}
	sep := fdncfg.Separator.Value
	rpHTSeps := regexp.MustCompile("^" + sep + "+" + "|" + sep + "+" + "$")
	//Process Head and Tail Sepatrators
	outName = rpHTSeps.ReplaceAllString(outName, "")

	return outName
}

// ASCHead add ascii head if not startwith ascii
func ASCHead(inputName string) string {
	outName := inputName
	sa := []rune(outName)
	ascH := ""
	var proxCS = func(c rune) string {
		if c > 'Z' {
			return fmt.Sprintf("%c", int(c)-int(math.Ceil(float64(c-'Z')/26)*26))
		} else if c < 'A' {
			return fmt.Sprintf("%c", int(c)+int(math.Ceil(float64('A'-c)/26)*26))
		} else {
			return fmt.Sprintf("%c", int(c))
		}
	}
	if !((unicode.IsDigit(sa[0])) || (unicode.IsLower(sa[0])) || (unicode.IsUpper(sa[0]))) {
		ascH = proxCS(sa[0]) + proxCS(sa[1]) + proxCS(sa[2])
	}

	return ascH + outName
}

// KeyHash create hash from key and return string
func KeyHash(key string) string {
	data := []byte(key)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// ArrayContainsElemenet check element e if exist in array s
func ArrayContainsElemenet[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// FDNFile fdn a file
func FDNFile(currentPath string, toBePath string, reserve bool) error {
	if !reserve {
		fdnrd, err := GetFDNRecord()
		if err != nil {
			log.Error(err)
			return err
		}
		fdnrd.Records = append(fdnrd.Records, &pb.Record{
			PreviousName:    filepath.Base(currentPath),
			CurrentNameHash: KeyHash(filepath.Base(toBePath)),
			LastUpdated:     timestamppb.Now()})
		err = SaveFDNRecord(fdnrd)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	if err := os.Rename(currentPath, toBePath); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// FNDedFrom from input and return
func FNDedFrom(input string) string {
	output := input
	output = ReplaceWords(input)
	output = ProcessHeadTail(output)
	output = ASCHead(output)
	return output
}

func confirm() UserInput {
	var cmsg string

	fmt.Print("Please confirm (all,yes,no,quit):")
	fmt.Scan(&cmsg)

	return UserInput(strings.ToLower(cmsg))
}

// UserInput receive user interactive input
type UserInput string

const (
	//All for all
	All UserInput = "all"
	//A shorcut all
	A UserInput = "a"
	//Yes for Ok only valid for current
	Yes UserInput = "yes"
	//Y shortcut for Yes
	Y UserInput = "y"
	//No for refuse only valid for current
	No UserInput = "no"
	//N shortcut for NO
	N UserInput = "n"
	//Quit exit app
	Quit UserInput = "quit"
	//Q shortcut for Quit
	Q UserInput = "q"
)

func noEffectTip() {
	var tipsDivider string

	if term.IsTerminal(0) {
		tw, _, err := term.GetSize(0)
		if err != nil {
			log.Error(err)
			tipsDivider = strings.Repeat("*", 80)
		} else {
			tipsDivider = strings.Repeat("*", tw)
		}
		fmt.Println(tipsDivider)
		fmt.Println("--> 'will change to' ==> 'changed to',in order to take effect,add flag '-i' or '-c'")
	}
}

// OutputResult fdn processed result
func OutputResult(origin string, processed string, inplace bool, fullpath bool) {
	if !fullpath {
		origin = filepath.Base(origin)
		processed = filepath.Base(processed)
	}
	if plainStyle {
		fmt.Println("   ", origin)
		if inplace {
			fmt.Println("==>", processed)
		} else {
			fmt.Println("-->", processed)
		}
	} else {
		origin = strings.Replace(origin, " ", "▯", -1) //for display space
		processed = strings.Replace(processed, " ", "▯", -1)
		a := []string{}
		b := []string{}
		for _, c := range []rune(origin) {
			a = append(a, string(c))
		}
		for _, c := range []rune(processed) {
			b = append(b, string(c))
		}
		seqm := difflib.NewMatcher(a, b)
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		richOrigin := ""
		richProcessed := ""
		sw := runewidth.StringWidth //shortcut
		for _, opc := range seqm.GetOpCodes() {
			switch opc.Tag {
			case 'r':
				as := strings.Join(a[opc.I1:opc.I2], "")
				bs := strings.Join(b[opc.J1:opc.J2], "")
				log.Trace("R:" + as + bs)
				if pretty {
					if sw(as) > sw(bs) {
						richOrigin += red(as)
						richProcessed += green(bs) + strings.Repeat(" ", sw(as)-sw(bs))
					} else if sw(as) < sw(bs) {
						richOrigin += red(as) + strings.Repeat(" ", sw(bs)-sw(as))
						richProcessed += green(bs)
					} else {
						richOrigin += red(as)
						richProcessed += green(bs)
					}
				} else {
					richOrigin += red(as)
					richProcessed += green(bs)
				}
			case 'd':
				as := strings.Join(a[opc.I1:opc.I2], "")
				log.Trace("D:" + as)
				if pretty {
					richOrigin += red(as)
					richProcessed += strings.Repeat(" ", sw(as))
				} else {
					richOrigin += red(as)
				}
			case 'i':
				as := strings.Join(a[opc.I1:opc.I2], "") //empty string
				bs := strings.Join(b[opc.J1:opc.J2], "")
				log.Trace("I:" + as + bs)
				if pretty {
					richOrigin += as + strings.Repeat(" ", sw(bs))
					richProcessed += green(bs)
				} else {
					richOrigin += as
					richProcessed += green(bs)
				}
			case 'e':
				as := strings.Join(a[opc.I1:opc.I2], "")
				bs := strings.Join(b[opc.J1:opc.J2], "")
				log.Trace("E:" + as + bs)
				richOrigin += as
				richProcessed += bs
			}
		}
		fmt.Println("   ", richOrigin) //display space
		if inplace {
			fmt.Println("==>", richProcessed)
		} else {
			fmt.Println("-->", richProcessed)
		}
	}
}

//TODO:ascii head
//TODO:remove nosense word
//TODO:support directory and files
//TODO:dry run result buffered for next step
