package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

var (
	linkRegx      = regexp.MustCompile(`(\s*)-\s+\[INCLUDE]\s*\((.+)\)`)
	plainLinkRegx = regexp.MustCompile(`(\s*)-\s+\[.+]\s*\((.+)\)`)
)

type Replacer struct {
	// summary.md 所在目录的绝对路径
	baseDir  string
	readBuf  bytes.Buffer
	writeBuf bytes.Buffer
}

func NewReplacer(path string) (*Replacer, error) {
	// path -> SUMMARY.tpl.md 的路径
	r := new(Replacer)
	r.baseDir = filepath.Dir(absPath(path))
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if _, err := r.readBuf.Write(content); err != nil {
		return nil, errors.WithStack(err)
	}
	return r, nil
}

func (r *Replacer) swap() {
	tmp := r.readBuf
	r.readBuf = r.writeBuf
	r.writeBuf = tmp
	r.writeBuf.Reset()
}

func (r *Replacer) String() string {
	return r.writeBuf.String()
}

func (r *Replacer) Bytes() []byte {
	return r.writeBuf.Bytes()
}

func (r *Replacer) Start() error {
	for {
		notOver, err := r.read()
		if err != nil {
			return err
		}
		if !notOver {
			break
		}
		r.swap()
	}
	return nil
}

func (r *Replacer) read() (bool, error) {
	reader := bufio.NewReader(&r.readBuf)
	found := false
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return false, errors.WithStack(err)
		}
		if indexes := linkRegx.FindSubmatchIndex(line); len(indexes) >= 6 {
			prefix := line[indexes[2]:indexes[3]]
			link := line[indexes[4]:indexes[5]]
			target := filepath.Join(r.baseDir, string(link))
			if err := r.readFile(prefix, target); err != nil {
				return false, err
			}
			// 找到了替换项
			found = true
		} else {
			if err := write(&r.writeBuf, line); err != nil {
				return false, err
			}
		}
		if err == io.EOF {
			break
		}
	}
	return found, nil
}

func (r *Replacer) readFile(prefix []byte, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer closeQuietly(file)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return errors.WithStack(err)
		}
		if err := write(&r.writeBuf, prefix); err != nil {
			return err
		}
		if indexes := plainLinkRegx.FindSubmatchIndex(line); len(indexes) >= 6 {
			link := line[indexes[4]:indexes[5]]
			newLink, err := filepath.Rel(r.baseDir, absPath(filepath.Join(filepath.Dir(path), string(link))))
			if err != nil {
				return errors.WithStack(err)
			}
			if err := write(&r.writeBuf, line[:indexes[4]]); err != nil {
				return err
			}
			if err := write(&r.writeBuf, []byte(newLink)); err != nil {
				return err
			}
			if err := write(&r.writeBuf, line[indexes[5]:]); err != nil {
				return err
			}
		} else {
			if err := write(&r.writeBuf, line); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}
