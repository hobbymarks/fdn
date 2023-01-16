/*
Package cmd root subcommand is the default
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/package cmd

import (
	"embed"
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
	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
	"github.com/hobbymarks/go-difflib/difflib"
	"github.com/mattn/go-runewidth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
var overwrite bool

//go:embed cfg.db
var defaultCFG embed.FS

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
			var rds []db.Record
			_db := db.ConnectRDDB()
			_db.Find(&rds)

			for _, rd := range rds {
				curNameHashPreNameMap[rd.CurrentNameHash] = rd.PreviousName
			}
		}
		paths, err := RetrievedAbsPaths(inputPaths, depthLevel, onlyDirectory)
		if err != nil {
			log.Fatal(err)
		}
		paths = RemoveHidden(paths)
		sort.SliceStable(
			paths,
			func(i, j int) bool { return paths[i] > paths[j] },
		)
		for _, path := range paths {
			path = filepath.Clean(path) //remove tailing slash if exist
			toPath := ""
			if reverse {
				preName, exist := curNameHashPreNameMap[utils.KeyHash(filepath.Base(path))]
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
				CheckDoFDN(path, toPath, reverse, overwrite)
			} else {
				if cfm {
					switch confirm() {
					case A, All:
						inplace = true
						CheckDoFDN(path, toPath, reverse, overwrite)
					case Y, Yes:
						CheckDoFDN(path, toPath, reverse, overwrite)
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
	rootCmd.Flags().
		BoolVarP(&onlyDirectory, "directory", "d", false, "If enable,directory only.Default file")
	rootCmd.Flags().IntVarP(&depthLevel, "level", "l", 1, "Maxdepth level")
	rootCmd.Flags().
		StringArrayVarP(&inputPaths, "path", "p", []string{"."}, "Input paths")
	rootCmd.Flags().BoolVarP(&inplace, "inplace", "i", false, "In-place")
	rootCmd.Flags().BoolVarP(&cfm, "confirm", "c", false, "Confirm")
	rootCmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "Reverse")
	rootCmd.Flags().BoolVarP(&fullpath, "fullpath", "f", false, "FullPath")
	rootCmd.Flags().
		BoolVarP(&plainStyle, "plainstyle", "s", false, "PlainStyle Output")
	rootCmd.Flags().BoolVarP(&pretty, "pretty", "e", false, "Pretty Display")
	rootCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite")

	//Process FDNConfig
	fdnDir := utils.FDNDir()
	FDNConfigPath = filepath.Join(fdnDir, "cfg.db")
	if _, err := os.Lstat(FDNConfigPath); errors.Is(err, os.ErrNotExist) {
		contents, err := defaultCFG.ReadFile("cfg.db")
		if err != nil {
			log.Errorf("read default config error:%s", err)
		}
		err = os.WriteFile(FDNConfigPath, contents, 0644)
		if err != nil {
			log.Errorf("copy default config error:%s", err)
		}
	}
}

// RetrievedAbsPaths Paths form args by flag
func RetrievedAbsPaths(
	inputPaths []string,
	depthLevel int,
	onlyDirectory bool,
) ([]string, error) {
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

// RemoveHidden remove all hidden files
func RemoveHidden(abspaths []string) []string {
	results := []string{}
	for _, apath := range abspaths {
		hidden, err := IsHidden(apath)
		if err != nil {
			log.Error(err)
		} else {
			if !hidden {
				results = append(results, apath)
			}
		}
		log.Trace("IsHidden:", hidden, apath)
	}
	return results
}

// FilteredSubPaths retrieve absolute paths
func FilteredSubPaths(
	dirPath string,
	depthLevel int,
	OnlyDir bool,
) ([]string, error) {
	var absolutePaths []string

	dirPath = filepath.Clean(dirPath)
	log.Trace(dirPath)
	if depthLevel == -1 {
		err := filepath.WalkDir(
			dirPath,
			func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					log.Trace(err)
					return err
				}
				if (OnlyDir && info.IsDir()) ||
					(!OnlyDir && info.Type().IsRegular()) {
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
			},
		)
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
func DepthFiles(
	dirPath string,
	depthLevel int,
	onlyDirectory bool,
) ([]string, error) {
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
	_db := db.ConnectCFGDB()
	for key, value := range keyValueMap {
		_key := strings.ToLower(key)
		termWord := db.TermWord{
			KeyHash:       utils.KeyHash(_key),
			OriginalLower: _key,
			TargetWord:    value,
		}
		_db.Create(&termWord)
	}
	return nil
}

// DeleteTermWords delete term words in config file
func DeleteTermWords(keys []string) error {
	_db := db.ConnectCFGDB()
	for _, key := range keys {
		_key := utils.KeyHash(key)
		_db.Delete(&db.TermWord{}, _key)
	}
	return nil
}

// ConfigToSepWords config tosep words in config file
func ConfigToSepWords(words []string) error {
	_db := db.ConnectCFGDB()
	for _, key := range words {
		_key := utils.KeyHash(key)
		toSepWord := db.ToSepWord{KeyHash: _key, Value: key}
		_db.Create(&toSepWord)
	}
	return nil
}

// DeleteToSepWords delete tosep words in config file
func DeleteToSepWords(keys []string) error {
	_db := db.ConnectCFGDB()
	for _, key := range keys {
		_key := utils.KeyHash(key)
		_db.Delete(&db.ToSepWord{}, _key)
	}
	return nil
}

// ConfigSeparator config separator in config file
func ConfigSeparator(separator string) error {
	_db := db.ConnectCFGDB()
	_sep := db.Separator{KeyHash: utils.KeyHash(separator), Value: separator}
	_db.Create(&_sep)
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
		var termWords []db.TermWord
		_db := db.ConnectRDDB()
		rlt := _db.Find(&termWords)
		if rlt.Error != nil {
			log.Fatalf("retrive TermWord error %s", rlt.Error)
		}

		pts := []string{}
		for _, twd := range termWords {
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

	words, wordMasks := mask(inputName)
	if len(words) != len(wordMasks) {
		log.Fatal("words not equal wordMasks")
	}

	newWords := []string{}

	var sep db.Separator
	var termWords []db.TermWord
	var toSepWords []db.ToSepWord
	_db := db.ConnectRDDB()
	rlt := _db.First(&sep)
	if rlt.Error != nil {
		log.Fatalf("retrieve Separator error %s", rlt.Error)
	}
	_sep := sep.Value
	rlt = _db.Find(&termWords)
	if rlt.Error != nil {
		log.Fatalf("retrieve TermWord error %s", rlt.Error)
	}
	rlt = _db.Find(&toSepWords)
	if rlt.Error != nil {
		log.Fatalf("retrieve ToSepWord error %s", rlt.Error)
	}

	rpCNS := regexp.MustCompile("[" + _sep + "]+")
	termWordMap := make(map[string]string)
	for _, twd := range termWords {
		termWordMap[twd.OriginalLower] = twd.TargetWord
	}
	for idx, wd := range words {
		if !wordMasks[idx] {
			for _, sw := range toSepWords { //replaced by separator
				wd = strings.Replace(wd, sw.Value, _sep, -1)
			}
		}
		newWords = append(newWords, wd)
	}
	outName = strings.Join(newWords, "")
	//Process continous separator
	outName = rpCNS.ReplaceAllString(outName, _sep)
	//
	newWords = []string{}

	for _, wd := range strings.Split(outName, _sep) {
		if v, exist := termWordMap[wd]; exist {
			wd = v
		}
		newWords = append(newWords, wd)
	}
	outName = strings.Join(newWords, _sep)
	// }

	return outName
}

// ProcessHeadTail process head and tail of input string
func ProcessHeadTail(inputName string) string {
	outName := inputName

	var sep db.Separator
	_db := db.ConnectRDDB()
	rlt := _db.First(&sep)
	if rlt.Error != nil {
		log.Fatalf("retrieve Separator error %s", rlt.Error)
	}
	_sep := sep.Value

	rpHTSeps := regexp.MustCompile("^" + _sep + "+" + "|" + _sep + "+" + "$")
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
			return fmt.Sprintf(
				"%c",
				int(c)-int(math.Ceil(float64(c-'Z')/26)*26),
			)
		} else if c < 'A' {
			return fmt.Sprintf("%c", int(c)+int(math.Ceil(float64('A'-c)/26)*26))
		} else {
			return fmt.Sprintf("%c", int(c))
		}
	}
	if len(sa) >= 1 {
		if !((unicode.IsDigit(sa[0])) || (unicode.IsLower(sa[0])) || (unicode.IsUpper(sa[0]))) {
			uL := 3
			if len(sa) < 3 {
				uL = len(sa)
			}
			for i := 0; i < uL; i++ {
				ascH += proxCS(sa[i])
			}
		}
	}
	return ascH + outName
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
		rd := db.Record{
			PreviousName:    filepath.Base(currentPath),
			CurrentNameHash: utils.KeyHash(filepath.Base(toBePath)),
		}
		_db := db.ConnectRDDB()
		_db.Create(&rd)
	}
	if err := os.Rename(currentPath, toBePath); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// CheckDoFDN check and do fdn
func CheckDoFDN(
	currentPath string,
	toBePath string,
	reserve bool,
	overwrite bool,
) error {
	if utils.PathExist(toBePath) && !overwrite {
		fmt.Println("[EXIST]Skip:", toBePath)
		fmt.Println("You can add 'overwrite' or 'o' flag to force do")
	} else {
		if err := FDNFile(currentPath, toBePath, reverse); err != nil {
			log.Error(err)
			return err
		}
		OutputResult(currentPath, toBePath, true, fullpath)
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
		fmt.Println(
			"--> 'will to' ==> 'to',add flag '-i' or '-c' to take effect",
		)
	}
}

// OutputResult fdn processed result
func OutputResult(
	origin string,
	processed string,
	inplace bool,
	fullpath bool,
) {
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

//TODO:support temp words
//TODO:add dry run for config
//TODO:remove nosense word
//TODO:support directory and files
//TODO:dry run result buffered for next step
//TODO:support add prefix or postfix by private order
