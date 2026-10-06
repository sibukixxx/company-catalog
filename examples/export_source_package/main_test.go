package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRebuildFixtureDeterministically(t *testing.T) {
	t.Setenv("HOUJIN_APP_ID", "")
	path := "../../docs/examples/company-source-package.v1.json"
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var first, second bytes.Buffer
	if err := run([]string{path}, &first); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{path}, &second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), original) || !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("example changed fixture bytes")
	}
}
