package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

var (
	verbose = true
)

func main() {
	if _, err := exec.LookPath("mdbook"); err != nil {
		log.Fatalln("mdbook not found")
	}

	removes := make([]string, 0)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build", "serve":
			root, err := getRoot()
			if err != nil {
				log.Fatalf("%+v\n", err)
			}
			if err := transform(filepath.Join(root, "src", summaryTemplateName), &removes, true); err != nil {
				log.Fatalf("%+v\n", err)
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

	for _, r := range removes {
		if err := os.Remove(r); err != nil {
			log.Fatalf("%+v\n", errors.WithStack(err))
		}
	}
}
