package utils

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	// PathDoesNotExistErr ...
	PathDoesNotExistErr = errors.Errorf("path does not exist")
)

// ReadPath ...
func ReadPath(path string, suffix string) ([]string, error) {
	files, err := glob(path, suffix)
	if err != nil {
		return nil, err
	}

	contents := make([]string, len(files))
	for i := range files {
		content, err := ioutil.ReadFile(files[i])
		if err != nil {
			return nil, err
		}

		contents[i] = string(content)
	}
	return contents, nil
}

// Return a list of SQL files in the listed paths. Only includes files ending
// in .sql. Omits hidden files, directories, and migrations.
func glob(path string, suffix string) ([]string, error) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, PathDoesNotExistErr
	}

	files := []string{}
	// listing
	if f.IsDir() {
		listing, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, f := range listing {
			files = append(files, filepath.Join(path, f.Name()))
		}
	} else {
		files = append(files, path)
	}

	// validate
	var sqlFiles []string
	for _, file := range files {
		if !strings.HasSuffix(file, suffix) {
			continue
		}
		if strings.HasPrefix(filepath.Base(file), ".") {
			continue
		}

		sqlFiles = append(sqlFiles, file)
	}

	return sqlFiles, nil
}
