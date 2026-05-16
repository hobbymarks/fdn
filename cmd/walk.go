package cmd

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func RetrievedAbsPaths(
	inputPaths []string,
	depthLevel int,
	onlyDir bool,
) ([]string, error) {
	var absolutePaths []string
	var errs []error

	for _, path := range inputPaths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			slog.Error(err.Error())
			errs = append(errs, err)
			continue
		} else {
			if fileInfo.IsDir() {
				paths, err := FilteredSubPaths(path, depthLevel, onlyDir)
				if err != nil {
					slog.Error(err.Error())
					errs = append(errs, err)
				} else {
					absolutePaths = append(absolutePaths, paths...)
				}
			} else if !onlyDir && fileInfo.Mode().IsRegular() {
				absPath, err := filepath.Abs(path)
				if err != nil {
					slog.Error(err.Error())
					errs = append(errs, err)
				} else {
					absolutePaths = append(absolutePaths, absPath)
				}
			} else {
				slog.Debug("skipped:" + path)
			}
		}
	}
	if len(absolutePaths) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return absolutePaths, nil
}

func RemoveHidden(abspaths []string) []string {
	results := []string{}
	for _, apath := range abspaths {
		hidden, err := IsHidden(apath)
		if err != nil {
			slog.Error(err.Error())
		} else {
			if !hidden {
				results = append(results, apath)
			}
		}
		slog.Debug("IsHidden:" + apath)
	}
	return results
}

func FilteredSubPaths(
	dirPath string,
	depthLevel int,
	onlyDir bool,
) ([]string, error) {
	var absolutePaths []string

	dirPath = filepath.Clean(dirPath)
	slog.Debug(dirPath)
	if depthLevel == -1 {
		err := filepath.WalkDir(
			dirPath,
			func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					slog.Debug(err.Error())
					return err
				}
				if (onlyDir && info.IsDir()) ||
					(!onlyDir && info.Type().IsRegular()) {
					slog.Debug("isDir:" + path)
					absPath, err := filepath.Abs(filepath.Join(dirPath, path))
					if err != nil {
						slog.Error(err.Error())
					} else {
						absolutePaths = append(absolutePaths, absPath)
					}
					return nil
				}
				slog.Debug("skipped:" + path)

				return nil
			},
		)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
	} else {
		paths, err := DepthFiles(dirPath, depthLevel, onlyDir)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		absolutePaths = paths
	}
	if onlyDir && depthLevel != -1 {
		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			slog.Error(err.Error())
		} else {
			absolutePaths = append(absolutePaths, absPath)
		}
	}
	return absolutePaths, nil
}

func DepthFiles(
	dirPath string,
	depthLevel int,
	onlyDir bool,
) ([]string, error) {
	var absolutePaths []string

	slog.Debug("depth", "level", depthLevel)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	for _, file := range files {
		absPath, err := filepath.Abs(filepath.Join(dirPath, file.Name()))
		if err != nil {
			slog.Error(err.Error())
		} else {
			if (onlyDir && file.IsDir()) ||
				(!onlyDir && file.Type().IsRegular()) {
				absolutePaths = append(absolutePaths, absPath)
			}
			if depthLevel > 1 && file.Type().IsDir() {
				files, err := DepthFiles(absPath, depthLevel-1, onlyDir)
				if err != nil {
					slog.Error(err.Error())
				} else {
					absolutePaths = append(absolutePaths, files...)
				}
			}
		}
	}
	return absolutePaths, nil
}
