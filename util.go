package main

import (
	"errors"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}

func ParentDirs(path string) []string {
	var dirs []string
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" && dir != "" {
		dirs = append(dirs, dir)
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}

	return dirs
}

func DefaultExtForContentType(ctype string) string {
	if exts, err := mime.ExtensionsByType(ctype); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ""
}
