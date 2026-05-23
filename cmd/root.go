/*
Package cmd root subcommand is the default
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

var version = "1.0.5"

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

var rootCmd = &cobra.Command{
	Use:     "fdn",
	Version: version,
	Short:   "A Tool For Unify File Name",
	Long:    `A Tool For Unify File Name and Directory Name`,
	Example: `  fdn mv ./a.txt ./b.txt
  fdn mv ./doc.pdf ./backup/
  fdn mv "My File.txt" ./inbox/My_File.txt`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level := slog.LevelWarn
		if verbose {
			level = slog.LevelInfo
		}
		opts := &slog.HandlerOptions{Level: level}
		handler := slog.NewTextHandler(os.Stderr, opts)
		slog.SetDefault(slog.New(handler))
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("rootCmd executing ...")
		printTipFlag := false
		curHashEncryPre := map[string]string{}
		if reverse {
			conn, err := db.ConnectCFGDB()
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			rows, err := conn.Model(&db.Record{}).Rows()
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			defer rows.Close()
			for rows.Next() {
				var rec db.Record
				if err := conn.ScanRows(rows, &rec); err != nil {
					slog.Error(err.Error())
					continue
				}
				curHashEncryPre[rec.HashedCurrentName] = rec.EncryptedPreviousName
			}
		}
		slog.Info("search paths...")
		paths, err := RetrievedAbsPaths(inputPaths, depthLevel, onlyDirectory)
		slog.Info("search paths!")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		paths = RemoveHidden(paths)
		sort.SliceStable(
			paths,
			func(i, j int) bool { return paths[i] > paths[j] },
		)
		slog.Info("loopthrough process path...")
		for _, path := range paths {
			slog.Info(fmt.Sprintf("process...:%s", path))
			path = filepath.Clean(path)
			toPath := ""
			if reverse {
				curName := filepath.Base(path)
				encryptedPre, exist := curHashEncryPre[utils.KeyHash(curName)]
				if exist {
					preName, err := utils.Decrypt(curName, encryptedPre)
					if err != nil {
						slog.Error(err.Error())
						os.Exit(1)
					}
					toPath = filepath.Join(filepath.Dir(path), preName)
				}
			} else {
				ext := utils.Ext(path)
				bn := strings.TrimSuffix(filepath.Base(path), ext)
				fdned, err := FDNedFrom(bn)
				if err != nil {
					slog.Error(err.Error())
					os.Exit(1)
				}
				if fdned != bn {
					toPath = filepath.Join(filepath.Dir(path), fdned+ext)
				}
			}
			if toPath == "" {
				slog.Info(fmt.Sprintf("process cont!:%s", path))
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
					printTipFlag = true
					OutputResult(path, toPath, false, fullpath)
				}
			}
			slog.Info(fmt.Sprintf("process!:%s", path))
		}
		slog.Info("loopthrough process path!")
		if printTipFlag {
			noEffectTip()
		}
	},
}

func Execute() {
	if err := prepareFDNDataDir(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			slog.Error(err.Error())
		}
	}()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func prepareFDNDataDir() error {
	fdnDir, err := utils.FDNDir()
	if err != nil {
		return err
	}
	if err := db.MigrateLegacyFDNDatabases(fdnDir); err != nil {
		slog.Error(fmt.Sprintf("migrate legacy fdn databases: %s", err))
	}
	dbPath := filepath.Join(fdnDir, db.FDNDBFileName)
	if err := db.InitDB(dbPath); err != nil {
		return err
	}
	if err := db.EnsureDefaultCFG(dbPath); err != nil {
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
