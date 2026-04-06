/*
Package cmd root subcommand is the default
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

// version is 0.0.0 for local builds; release binaries get the tag from GoReleaser ldflags (.goreleaser.yaml).
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
				bn := strings.TrimSuffix(filepath.Base(path), ext)
				fdned, err := FDNedFrom(bn)
				if err != nil {
					log.Fatal(err)
				}
				if fdned != bn {
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
	if err := prepareFDNDataDir(); err != nil {
		log.Fatal(err)
	}
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func prepareFDNDataDir() error {
	fdnDir := utils.FDNDir()
	if err := db.MigrateLegacyFDNDatabases(fdnDir); err != nil {
		log.Errorf("migrate legacy fdn databases: %s", err)
	}
	FDNConfigPath = filepath.Join(fdnDir, db.FDNDBFileName)
	FDNRecordPath = FDNConfigPath
	if err := db.EnsureDefaultCFG(FDNConfigPath); err != nil {
		return err
	}
	return nil
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
