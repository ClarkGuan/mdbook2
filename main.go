package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	summaryTemplateName = "SUMMARY.tpl.md"
	summaryName         = "SUMMARY.md"
)

var (
	linkRegx = regexp.MustCompile(`^(\s*)-\s+\[INCLUDE]\s*\(.+\)$`)

	verbose = false
)

func main() {
	if _, err := exec.LookPath("mdbook"); err != nil {
		log.Fatalln("mdbook not found")
	}
	root, err := getRoot()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	newContent, err := transform(filepath.Join(root, "src", summaryTemplateName))
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	if verbose {
		fmt.Println(newContent)
	}
	if err := os.WriteFile(filepath.Join(root, "src", summaryName), []byte(newContent), 0666); err != nil {
		log.Fatalf("%+v\n", errors.WithStack(err))
	}

	cmd := exec.Command("mdbook", os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(cmd.Env, os.Environ()...)
	if err := cmd.Run(); err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func getRoot() (string, error) {
	path := os.Args[len(os.Args)-1]
	if isRoot(path) {
		return path, nil
	}
	current, err := filepath.Abs(".")
	if err != nil {
		return "", errors.WithStack(err)
	}
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

func transform(path string) (string, error) {
	content, err := os.Open(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer content.Close()
	reader := bufio.NewReader(content)
	retBuf := new(strings.Builder)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return "", errors.WithStack(err)
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if indexes := linkRegx.FindSubmatchIndex(line); len(indexes) > 5 {
			prefix := line[indexes[2]:indexes[3]]
			link := line[indexes[4]:indexes[5]]
			target := filepath.Join(filepath.Dir(path), string(link))
			if verbose {
				log.Printf("line: %q\nprefix=%q\nlink:%q\ntarget:%q\n", line, prefix, link, target)
			}
			fmt.Fprintln(retBuf, "copy from", target)
			if err := embedFragment(retBuf, prefix, target); err != nil {
				return "", err
			}
		} else {
			if _, err := retBuf.Write(line); err != nil {
				return "", errors.WithStack(err)
			}
			if err := retBuf.WriteByte('\n'); err != nil {
				return "", errors.WithStack(err)
			}
		}
		if err == io.EOF {
			break
		}
	}
	return retBuf.String(), nil
}

func embedFragment(w *strings.Builder, prefix []byte, path string) error {
	reader, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer reader.Close()
	bufReader := bufio.NewReader(reader)
	for {
		line, err := bufReader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return errors.WithStack(err)
		}
		if _, err := w.Write(prefix); err != nil {
			return errors.WithStack(err)
		}
		if _, err := w.Write(line); err != nil {
			return errors.WithStack(err)
		}
		if err == io.EOF {
			break
		}
	}
	if err := w.WriteByte('\n'); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
