/*
Package cmd root subcommand is the default
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gorm.io/gorm"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
	"github.com/hobbymarks/go-difflib/difflib"
)

var version = "0.0.0"

var (
	onlyDirectory bool
	inputPaths    []string
	depthLevel    int
	inplace       bool
	confirm       bool
	reverse       bool
	fullpath      bool
	plainStyle    bool
	pretty        bool
	overwrite     bool
)

var verbose bool

// FDNConfigPath is the unified FDN database path (config + rename records); kept for compatibility.
var FDNConfigPath string

// FDNRecordPath matches FDNConfigPath; both point at fdn.db.
var FDNRecordPath string

var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Name",
	Long:    `A Tool For Unify File Name and Directory Name`,
	Example: `  fdn mv ./a.txt ./b.txt
  fdn mv ./doc.pdf ./backup/
  fdn mv "My File.txt" ./inbox/My_File.txt`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logFormatter := new(log.TextFormatter)
		logFormatter.TimestampFormat = "15:04:05.000"
		logFormatter.FullTimestamp = true
		log.SetFormatter(logFormatter)
		if verbose {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("rootCmd executing ...")
		PrintTipFlag := false
		curHashEncryPre := map[string]string{}
		if reverse {
			var rds []db.Record
			_db := db.ConnectRDDB()
			defer utils.DBClose(_db)
			_db.Find(&rds)

			for _, rd := range rds {
				curHashEncryPre[rd.HashedCurrentName] = rd.EncryptedPreviousName
			}
		}
		log.Infof("search paths...")
		paths, err := RetrievedAbsPaths(inputPaths, depthLevel, onlyDirectory)
		log.Infof("search paths!")
		if err != nil {
			log.Fatal(err)
		}
		paths = RemoveHidden(paths)
		sort.SliceStable(
			paths,
			func(i, j int) bool { return paths[i] > paths[j] },
		)
		log.Info("loopthrough process path...")
		for _, path := range paths {
			log.Infof("process...:%s", path)
			path = filepath.Clean(path)
			// remove tailing slash if exist
			toPath := ""
			if reverse {
				curName := filepath.Base(path)
				encryptedPre, exist := curHashEncryPre[utils.KeyHash(curName)]
				if exist {
					preName := utils.Decrypt(curName, encryptedPre)
					toPath = filepath.Join(filepath.Dir(path), preName)
				}
			} else {
				ext := utils.Ext(path)
				// ext empty if path is dir
				bn := strings.TrimSuffix(filepath.Base(path), ext)
				if fdned := FDNedFrom(bn); fdned != bn {
					toPath = filepath.Join(filepath.Dir(path), fdned+ext)
				}
			}
			if toPath == "" {
				log.Infof("process cont!:%s", path)
				continue
			}
			if inplace {
				CheckDoFDN(path, toPath, reverse, overwrite)
			} else {
				if confirm {
					switch GetConfirm() {
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
			log.Infof("process!:%s", path)
		}
		log.Info("loopthrough process path!")
		if PrintTipFlag {
			noEffectTip()
		}
	},
}

// Execute is the cmd entry
func Execute() {
	prepareFDNDataDir()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func prepareFDNDataDir() {
	fdnDir := utils.FDNDir()
	if err := db.MigrateLegacyFDNDatabases(fdnDir); err != nil {
		log.Errorf("migrate legacy fdn databases: %s", err)
	}
	FDNConfigPath = filepath.Join(fdnDir, db.FDNDBFileName)
	FDNRecordPath = FDNConfigPath
	if err := db.EnsureDefaultCFG(FDNConfigPath); err != nil {
		log.Errorf("init default config error:%s", err)
	}
}

func init() {
	rootCmd.Flags().
		BoolVarP(&onlyDirectory, "directory", "d", false,
			"If enable,directory only.Default regular file only")
	rootCmd.Flags().IntVarP(&depthLevel, "level", "l", 1, "Maxdepth level")
	rootCmd.Flags().
		StringArrayVarP(&inputPaths, "path", "p", []string{"."}, "Input paths")
	rootCmd.Flags().BoolVarP(&inplace, "inplace", "i", false, "In-place")
	rootCmd.Flags().BoolVarP(&confirm, "confirm", "c", false, "Confirm")
	rootCmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "Reverse")
	rootCmd.Flags().BoolVarP(&fullpath, "fullpath", "f", false, "FullPath")
	rootCmd.Flags().
		BoolVarP(&plainStyle, "plainstyle", "s", false, "PlainStyle Output")
	rootCmd.Flags().BoolVarP(&pretty, "pretty", "e", false, "Pretty Display")
	rootCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite")

	rootCmd.Flags().
		BoolVarP(&verbose, "verbose", "V", false, "Print more verbose information")
}

// RetrievedAbsPaths Paths from args by flag
func RetrievedAbsPaths(
	inputPaths []string,
	depthLevel int,
	onlyDir bool,
) ([]string, error) {
	var absolutePaths []string

	for _, path := range inputPaths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Error(err)
			continue
		} else {
			if fileInfo.IsDir() {
				paths, err := FilteredSubPaths(path, depthLevel, onlyDir)
				if err != nil {
					log.Error(err)
				} else {
					absolutePaths = append(absolutePaths, paths...)
				}
			} else if !onlyDir && fileInfo.Mode().IsRegular() {
				absPath, err := filepath.Abs(path)
				if err != nil {
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
	onlyDir bool,
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
				if (onlyDir && info.IsDir()) ||
					(!onlyDir && info.Type().IsRegular()) {
					log.Trace("isDir:", path)
					absPath, err := filepath.Abs(filepath.Join(dirPath, path))
					if err != nil {
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
		paths, err := DepthFiles(dirPath, depthLevel, onlyDir)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		absolutePaths = paths
	}
	if onlyDir {
		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			log.Error(err)
		} else {
			absolutePaths = append(absolutePaths, absPath)
		}
	}
	return absolutePaths, nil
}

// DepthFiles Depth read dir
func DepthFiles(
	dirPath string,
	depthLevel int,
	onlyDir bool,
) ([]string, error) {
	var absolutePaths []string

	log.Debug(depthLevel)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, file := range files {
		absPath, err := filepath.Abs(filepath.Join(dirPath, file.Name()))
		if err != nil {
			log.Error(err)
		} else {
			if (onlyDir && file.IsDir()) ||
				(!onlyDir && file.Type().IsRegular()) {
				absolutePaths = append(absolutePaths, absPath)
			}
			if depthLevel > 1 && file.Type().IsDir() {
				files, err := DepthFiles(absPath, depthLevel-1, onlyDir)
				if err != nil {
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
	defer utils.DBClose(_db)
	for key, value := range keyValueMap {
		_key := strings.ToLower(key)
		_termWord := db.TermWord{
			KeyHash:       utils.KeyHash(_key),
			OriginalLower: _key,
			TargetWord:    value,
		}
		var _cnt int64 = 0
		_db.Model(&db.TermWord{}).
			Where("key_hash = ?", _termWord.KeyHash).
			Count(&_cnt)
		if _cnt == 0 {
			_rlt := _db.Create(&_termWord)
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
		} else {
			log.Debugf("skipped:%s", _termWord.OriginalLower)
		}
	}
	return nil
}

// DeleteTermWords delete term words in config file
func DeleteTermWords(keys []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, key := range keys {
		_key := utils.KeyHash(key)
		_rlt := _db.Unscoped().Delete(&db.TermWord{}, _key)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	}
	return nil
}

// ConfigToSepWords config tosep words in config file
func ConfigToSepWords(words []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, word := range words {
		_key := utils.KeyHash(word)
		_toSepWord := db.ToSepWord{KeyHash: _key, Value: word}
		var _cnt int64 = 0
		_db.Model(&db.ToSepWord{}).
			Where("key_hash = ?", _toSepWord.KeyHash).
			Count(&_cnt)
		if _cnt == 0 {
			_rlt := _db.Create(&_toSepWord)
			if _rlt.Error != nil {
				log.Error(_rlt.Error)
			}
		} else {
			log.Debugf("skipped:%s", _toSepWord.Value)
		}
	}
	return nil
}

// DeleteToSepWords delete tosep words in config file
func DeleteToSepWords(keys []string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	for _, key := range keys {
		_key := utils.KeyHash(key)
		_rlt := _db.Unscoped().Delete(&db.ToSepWord{}, _key)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	}
	return nil
}

// ConfigSeparator config separator in config file
func ConfigSeparator(separator string) error {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	_sep := db.Separator{KeyHash: utils.KeyHash(separator), Value: separator}
	var _cnt int64 = 0
	_db.Model(&db.Separator{}).
		Where("key_hash = ?", _sep.KeyHash).
		Count(&_cnt)
	if _cnt == 0 {
		_rlt := _db.Create(&_sep)
		if _rlt.Error != nil {
			log.Error(_rlt.Error)
		}
	} else {
		log.Debugf("skipped:%s", _sep.Value)
	}
	return nil
}

// ReplaceWords process inputName string and return new string
func ReplaceWords(inputName string) string {
	outName := inputName
	mask := func(s string) ([]string, []bool) {
		regescape := func(s string) string {
			s = strings.ReplaceAll(s, "+", "\\+")
			s = strings.ReplaceAll(s, "?", "\\?")
			s = strings.ReplaceAll(s, "*", "\\*")

			return s
		}

		words := []string{}
		wdmsk := []bool{}
		var termWords []db.TermWord
		_db := db.ConnectCFGDB()
		defer utils.DBClose(_db)
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
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
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
			for _, sw := range toSepWords {
				// replaced by separator
				wd = strings.ReplaceAll(wd, sw.Value, _sep)
			}
		}
		newWords = append(newWords, wd)
	}
	outName = strings.Join(newWords, "")
	// Process continous separator
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
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	rlt := _db.First(&sep)
	if rlt.Error != nil {
		log.Fatalf("retrieve Separator error %s", rlt.Error)
	}
	_sep := sep.Value

	rpHTSeps := regexp.MustCompile("^" + _sep + "+" + "|" + _sep + "+" + "$")
	// Process Head and Tail Sepatrators
	outName = rpHTSeps.ReplaceAllString(outName, "")

	return outName
}

// ASCHead add ascii head if not startwith ascii
func ASCHead(inputName string) string {
	outName := inputName
	sa := []rune(outName)
	var ascH strings.Builder
	proxCS := func(c rune) string {
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
		if !(unicode.IsDigit(sa[0])) && !(unicode.IsLower(sa[0])) && !(unicode.IsUpper(sa[0])) {
			uL := min(len(sa), 3)
			for i := range uL {
				ascH.WriteString(proxCS(sa[i]))
			}
		}
	}
	return ascH.String() + outName
}

// ArrayContainsElemenet check element e if exist in array s
func ArrayContainsElemenet[T comparable](s []T, e T) bool {
	return slices.Contains(s, e)
}

// FDNFile renames currentPath to toBePath, then updates the rename journal (record DB)
// from basenames. Filesystem and DB cannot be one true atomic transaction; order is
// rename-first so the journal never describes a rename that did not occur. If the journal
// step fails, the rename is rolled back when os.Rename can reverse it.
func FDNFile(currentPath string, toBePath string, reversed bool) error {
	_to := filepath.Base(toBePath)
	_cur := filepath.Base(currentPath)

	if err := os.Rename(currentPath, toBePath); err != nil {
		log.Error(err)
		return err
	}

	_db := db.ConnectRDDB()
	defer utils.DBClose(_db)
	var journalErr error
	if !reversed {
		_rd := db.Record{
			EncryptedPreviousName: utils.Encrypt(_to, _cur),
			HashedCurrentName:     utils.KeyHash(_to),
		}
		journalErr = AddRecord(_db, _rd)
	} else {
		_rd := db.Record{
			EncryptedPreviousName: utils.Encrypt(_cur, _to),
			HashedCurrentName:     utils.KeyHash(_cur),
		}
		journalErr = DeleteRecord(_db, _rd)
	}
	if journalErr != nil {
		if rb := os.Rename(toBePath, currentPath); rb != nil {
			return fmt.Errorf("%w; rename rollback failed: %v", journalErr, rb)
		}
		log.Error(journalErr)
		return journalErr
	}
	return nil
}

// AddRecord add a record in db
func AddRecord(_db *gorm.DB, _rd db.Record) error {
	var rd db.Record
	rlt := _db.First(
		&rd,
		"encrypted_previous_name = ? AND hashed_current_name = ?",
		_rd.EncryptedPreviousName,
		_rd.HashedCurrentName,
	)
	if rlt.Error != nil {
		if errors.Is(rlt.Error, gorm.ErrRecordNotFound) {
			return _db.Create(&_rd).Error
		}
		return rlt.Error
	}
	rd.Count++
	return _db.Save(&rd).Error
}

// DeleteRecord delete a record in db
func DeleteRecord(_db *gorm.DB, _rd db.Record) error {
	var rd db.Record
	rlt := _db.First(
		&rd,
		"encrypted_previous_name = ? AND hashed_current_name = ?",
		_rd.EncryptedPreviousName,
		_rd.HashedCurrentName,
	)
	if rlt.Error != nil {
		return rlt.Error
	}
	rd.Count--
	if rd.Count != 0 {
		return _db.Save(&rd).Error
	}
	rd.Count++
	return _db.Unscoped().Delete(&rd).Error
}

// CheckDoFDN renames or normalizes currentPath into toBePath via FDNFile when the destination
// rules allow it. If toBePath already exists and overwrite is false, it runs only when
// currentPath and toBePath refer to the same file; otherwise it prints a skip message and
// does nothing. Returns any error from FDNFile or related checks.
func CheckDoFDN(
	currentPath string,
	toBePath string,
	reverse bool,
	overwrite bool,
) error {
	if utils.PathExist(toBePath) {
		if overwrite {
			err := FDNFile(currentPath, toBePath, reverse)
			if err != nil {
				log.Error(err)
				return err
			}
			OutputResult(currentPath, toBePath, true, fullpath)
		} else {
			same, err := utils.SameFiles(currentPath, toBePath)
			if err != nil {
				fmt.Println("[ERROR]Skip:", currentPath)
			} else {
				if same {
					err := FDNFile(currentPath, toBePath, reverse)
					if err != nil {
						log.Error(err)
						return err
					}
					OutputResult(currentPath, toBePath, true, fullpath)
				} else {
					fmt.Println("[EXIST]Skip:", currentPath)
				}
			}
		}
	} else {
		err := FDNFile(currentPath, toBePath, reverse)
		if err != nil {
			log.Error(err)
			return err
		}
		OutputResult(currentPath, toBePath, true, fullpath)
	}
	return nil
}

// FDNedFrom returns the FDN-normalized form of input by applying, in order, ReplaceWords
// (configured term and separator rules from the config DB), ProcessHeadTail (strip leading
// and trailing separator runs), and ASCHead (synthetic ASCII prefix when the first rune is
// not a Latin letter or digit).
func FDNedFrom(input string) string {
	// TODO(hm): Optimize name
	output := input
	output = ReplaceWords(input)
	output = ProcessHeadTail(output)
	output = ASCHead(output)
	return output
}

// GetConfirm get user input confirmation
func GetConfirm() UserInput {
	var cmsg string

	fmt.Print("Please confirm (all,yes,no,quit):")
	fmt.Scan(&cmsg)

	return UserInput(strings.ToLower(cmsg))
}

// UserInput receive user interactive input
type UserInput string

const (
	// All for all
	All UserInput = "all"
	// A shorcut all
	A UserInput = "a"
	// Yes for Ok only valid for current
	Yes UserInput = "yes"
	// Y shortcut for Yes
	Y UserInput = "y"
	// No for refuse only valid for current
	No UserInput = "no"
	// N shortcut for NO
	N UserInput = "n"
	// Quit exit app
	Quit UserInput = "quit"
	// Q shortcut for Quit
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

// OutputResult prints a before/after line pair for an FDN rename: the original path and the
// processed path. When fullpath is false, only the final path components are shown. The
// second line uses "==>" when inplace is true (change applied) and "-->" when false (dry
// run). With plainStyle, output is plain text; otherwise spaces are shown as "▯" and a
// character-level diff is rendered with red/green highlighting (optionally width-aligned when
// pretty is set).
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
		// for display space
		origin = strings.ReplaceAll(origin, " ", "▯")
		processed = strings.ReplaceAll(processed, " ", "▯")

		// TODO(hm): Maybe can optimize bu avoid creating intermediate slices to save more memory
		a := strings.Split(origin, "")
		b := strings.Split(processed, "")

		seqm := difflib.NewMatcher(a, b)
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		richOrigin := ""
		richProcessed := ""
		sw := runewidth.StringWidth
		// shortcut
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
				as := strings.Join(a[opc.I1:opc.I2], "")
				// empty string
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
		fmt.Println("   ", richOrigin)
		// display space
		if inplace {
			fmt.Println("==>", richProcessed)
		} else {
			fmt.Println("-->", richProcessed)
		}
	}
}

// TODO(hm): Support ignore filename or filepath(add ignores,list ignores,delete
// ignore,force ignore ignores)
// TODO(hm): At bottom add dynamic revolved bar as not dead flag
// TODO(hm): Doc - multi args how to
// TODO(hm): Support temp words
// TODO(hm): Recursive query for change records
// TODO(hm): Add option for permanently delete record when count is 0 or soft delete
// TODO(hm): Add dry run for config
// TODO(hm): Remove nosense word
// TODO(hm): Support add prefix or postfix by private order
// TODO(hm): Optimize loop through files performance
