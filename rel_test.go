package main

import (
	"log"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	base := "/home/clark/work/blog/test/src"
	target := "/home/clark/work/blog/test/src/第二章/第一节.md"
	rel, err := filepath.Rel(base, target)
	if err != nil {
		t.Fatalf("%+v\n", err)
	}
	log.Println(rel)
}
