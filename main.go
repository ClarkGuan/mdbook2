package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
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
			r, err := NewReplacer(filepath.Join(root, "src", "SUMMARY.tpl.md"))
			if err != nil {
				log.Fatalf("%+v\n", err)
			}
			if err := r.Start(); err != nil {
				log.Fatalf("%+v\n", err)
			}
			if err := os.WriteFile(filepath.Join(root, "src", "SUMMARY.md"), r.Bytes(), 0666); err != nil {
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
