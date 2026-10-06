package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestFileOnlyHandoff(t *testing.T) {
	data, err := os.ReadFile("../../docs/examples/company-source-package.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := handoff(bytes.NewReader(data), &out); err != nil {
		t.Fatal(err)
	}
	var request struct {
		ContractVersion string `json:"contractVersion"`
		Documents       []struct {
			Content string `json:"content"`
			Source  string `json:"source"`
		} `json:"documents"`
	}
	if err := json.Unmarshal(out.Bytes(), &request); err != nil {
		t.Fatal(err)
	}
	if request.ContractVersion != "1" || len(request.Documents) != 1 || request.Documents[0].Source != "dataset" {
		t.Fatal("invalid public envelope")
	}
	var p map[string]any
	if err := json.Unmarshal([]byte(request.Documents[0].Content), &p); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"snapshot_revision", "content_hash", "provenance", "source_fetched_at", "manifest"} {
		if p[key] == nil {
			t.Fatal("lost", key)
		}
	}
	if err := handoff(strings.NewReader(strings.ReplaceAll(string(data), "techvit.company.source-package/v1", "unsupported/v2")), &out); err == nil {
		t.Fatal("accepted unsupported version")
	}
}
