package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	fragmentTemplateName = "fragment.tpl.md"
	summaryTemplateName  = "SUMMARY.tpl.md"
	fragmentName         = "fragment.md"
	summaryName          = "SUMMARY.md"
)

var (
	templateSuffix = []byte(".tpl.md")
	newline        = []byte("\n")
)

var (
	linkRegx      = regexp.MustCompile(`(\s*)-\s+\[INCLUDE]\s*\((.+)\)`)
	plainLinkRegx = regexp.MustCompile(`(\s*)-\s+\[.+]\s*\((.+)\)`)
)

var (
	errNotTlpFile = errors.New("并不是模板文件")
)

func transform(source string, removes *[]string) error {
	if !isTemplate(source) {
		return errNotTlpFile
	}

	file, err := os.Open(source)
	if err != nil {
		return errors.WithStack(err)
	}
	defer closeQuietly(file)

	reader := bufio.NewReader(file)
	targetFile, err := os.OpenFile(targetPath(source), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return errors.WithStack(err)
	}
	defer closeQuietly(targetFile)
	writer := bufio.NewWriter(targetFile)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return errors.WithStack(err)
		}

		if indexes := linkRegx.FindSubmatchIndex(line); len(indexes) >= 6 {
			prefix := line[indexes[2]:indexes[3]]
			link := line[indexes[4]:indexes[5]]
			target := absPath(filepath.Join(filepath.Dir(source), string(link)))

			if bytes.HasSuffix(link, templateSuffix) {
				// 递归处理
				if err := transform(target, removes); err != nil {
					return err
				}
				// 处理完后生成新的文件，名称如下
				target = target[:len(target)-len(templateSuffix)] + ".md"
				*removes = append(*removes, target)
			}

			// 开始内嵌文件内容
			if err := embedFragment(writer, prefix, absPath(source), target); err != nil {
				return err
			}
		} else {
			if _, err := writer.Write(line); err != nil {
				return errors.WithStack(err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func embedFragment(w io.Writer, prefix []byte, base, path string) error {
	reader, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer closeQuietly(reader)
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
	if _, err := w.Write(newline); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func targetPath(source string) string {
	dir := filepath.Dir(source)
	if strings.HasSuffix(source, fragmentTemplateName) {
		return filepath.Join(dir, fragmentName)
	} else if strings.HasSuffix(source, summaryTemplateName) {
		return filepath.Join(dir, summaryName)
	}
	panic("not predefined file name")
}

func closeQuietly(c io.Closer) {
	if c != nil {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}
}

func isTemplate(path string) bool {
	baseName := filepath.Base(path)
	switch baseName {
	case fragmentTemplateName, summaryTemplateName:
		return true
	default:
		return false
	}
}

func replace(w io.Writer, line []byte, base, current string) error {
	if indexes := plainLinkRegx.FindSubmatchIndex(line); len(indexes) >= 6 {
		target := line[indexes[4]:indexes[5]]
		relative := absPath(filepath.Join(filepath.Dir(current), string(target)))
		rep, err := filepath.Rel(filepath.Dir(base), relative)
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
		return write(w, line)
	}
}

func write(w io.Writer, content []byte) error {
	if _, err := w.Write(content); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
