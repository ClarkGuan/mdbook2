package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
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
		fileExist(filepath.Join(path, "src", "SUMMARY.tpl.md"))
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

func closeQuietly(c io.Closer) {
	if c != nil {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}
}

func write(w io.Writer, content []byte) error {
	if _, err := w.Write(content); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
