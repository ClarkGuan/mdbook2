package main

import (
	"os"
	"path/filepath"
)

func getRoot() (string, error) {
	path := os.Args[len(os.Args)-1]
	if isRoot(path) {
		return path, nil
	}
	current := absPath(".")
	if isRoot(current) {
		return current, nil
	}
	return "", os.ErrExist
}

func isRoot(path string) bool {
	return dirExist(path) &&
		dirExist(filepath.Join(path, "src")) &&
		fileExist(filepath.Join(path, "book.toml")) &&
		fileExist(filepath.Join(path, "src", summaryTemplateName))
}

func fileExist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

func dirExist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func absPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}
