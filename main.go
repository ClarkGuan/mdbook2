package main

import (
	"bufio"
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
	linkRegx      = regexp.MustCompile(`(\s*)-\s+\[INCLUDE]\s*\((.+)\)`)
	plainLinkRegx = regexp.MustCompile(`(\s*)-\s+\[.+]\s*\((.+)\)`)

	verbose = true
)

func main() {
	if _, err := exec.LookPath("mdbook"); err != nil {
		log.Fatalln("mdbook not found")
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build", "serve":
			root, err := getRoot()
			if err != nil {
				log.Fatalf("%+v\n", err)
			}
			newContent, err := transform(filepath.Join(root, "src", summaryTemplateName))
			if err != nil {
				log.Fatalf("%+v\n", err)
			}
			if err := os.WriteFile(filepath.Join(root, "src", summaryName), []byte(newContent), 0666); err != nil {
				log.Fatalf("%+v\n", errors.WithStack(err))
			}
		}
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

func transform(path string) (string, error) {
	summaryTplAbsPath := absPath(path)
	content, err := os.Open(summaryTplAbsPath)
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
		if indexes := linkRegx.FindSubmatchIndex(line); len(indexes) >= 6 {
			prefix := line[indexes[2]:indexes[3]]
			link := line[indexes[4]:indexes[5]]
			target := absPath(filepath.Join(filepath.Dir(path), string(link)))
			if verbose {
				log.Printf("line: %q\nprefix=%q\nlink:%q\ntarget:%q\n", line, prefix, link, target)
			}
			if err := embedFragment(retBuf, prefix, summaryTplAbsPath, target); err != nil {
				return "", err
			}
		} else {
			if verbose {
				log.Printf("not match: %q  indexes:%v\n", line, indexes)
			}
			if _, err := retBuf.Write(line); err != nil {
				return "", errors.WithStack(err)
			}
		}
		if err == io.EOF {
			break
		}
	}
	return retBuf.String(), nil
}

func embedFragment(w *strings.Builder, prefix []byte, base, path string) error {
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
		if err := write(w, prefix); err != nil {
			return err
		}
		if err := replace(w, line, base, path); err != nil {
			return err
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

func replace(w *strings.Builder, line []byte, base, current string) error {
	if indexes := plainLinkRegx.FindSubmatchIndex(line); len(indexes) >= 6 {
		target := line[indexes[4]:indexes[5]]
		relTarget := absPath(filepath.Join(filepath.Dir(current), string(target)))
		rep, err := filepath.Rel(filepath.Dir(base), relTarget)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := write(w, line[:indexes[4]]); err != nil {
			return err
		}
		if err := write(w, []byte(rep)); err != nil {
			return err
		}
		if err := write(w, line[indexes[5]:]); err != nil {
			return err
		}
		return nil
	} else {
		if verbose {
			log.Printf("replace not match %q indexes:%v\n", line, indexes)
		}
		return write(w, line)
	}
}

func write(w *strings.Builder, content []byte) error {
	if _, err := w.Write(content); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
