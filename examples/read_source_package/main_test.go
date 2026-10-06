package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadFixtureOffline(t *testing.T) {
	t.Setenv("HOUJIN_APP_ID", "")
	var out bytes.Buffer
	if err := run([]string{"../../docs/examples/company-source-package.v1.json"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"schema: techvit.company.source-package/v1", "records: 2 (COMPLETE)", "Synthetic Contract Company A", "source: houjin_bangou v4", "fetched_at=2026-01-01T00:00:00Z"} {
		if !strings.Contains(out.String(), part) {
			t.Fatal("missing output", part)
		}
	}
}
