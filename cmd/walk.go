package cmd

import (
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

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
