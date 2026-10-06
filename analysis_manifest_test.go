package houjin

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestNewAnalysisManifestFillsSchemaAndSourceProvenance(t *testing.T) {
	generatedAt := time.Date(2026, 9, 13, 9, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	query := AnalysisManifestQuery{From: "2026-09-01", To: "2026-09-13", Prefecture: "13"}

	got := NewAnalysisManifest("jsonl", query, generatedAt, 42)

	want := AnalysisManifest{
		Schema:        AnalysisSchemaVersion,
		Source:        "houjin_bangou",
		SourceVersion: "v4",
		Format:        "jsonl",
		Query:         query,
		GeneratedAt:   "2026-09-13T00:00:00Z",
		RecordCount:   42,
	}
	if got != want {
		t.Fatalf("manifest = %#v, want %#v", got, want)
	}
}

func TestWriteAnalysisManifestOmitsEmptyQueryFilters(t *testing.T) {
	var buf bytes.Buffer
	manifest := NewAnalysisManifest("csv", AnalysisManifestQuery{From: "2026-09-01", To: "2026-09-13"}, time.Unix(0, 0), 0)
	if err := WriteAnalysisManifest(&buf, manifest); err != nil {
		t.Fatal(err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("manifest is invalid JSON: %v\n%s", err, buf.String())
	}
	for _, key := range []string{"schema", "source", "source_version", "format", "query", "generated_at", "record_count"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("manifest missing %q: %s", key, buf.String())
		}
	}
	var query map[string]json.RawMessage
	if err := json.Unmarshal(got["query"], &query); err != nil {
		t.Fatal(err)
	}
	if len(query) != 2 {
		t.Fatalf("query should only contain from/to when no filters are set: %s", got["query"])
	}
}

// docs/examples の manifest はドキュメント用のサンプルであり、上流から取得した
// seed/golden データではない。コードの契約とずれないことだけを検証する。
func TestAnalysisManifestExampleMatchesContract(t *testing.T) {
	data, err := os.ReadFile("docs/examples/analysis-manifest.example.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest AnalysisManifest
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("example manifest does not match AnalysisManifest: %v", err)
	}
	if manifest.Schema != AnalysisSchemaVersion {
		t.Fatalf("example schema = %q, want %q", manifest.Schema, AnalysisSchemaVersion)
	}
	if manifest.Source != "houjin_bangou" || manifest.SourceVersion != "v4" {
		t.Fatalf("example source provenance = %q/%q", manifest.Source, manifest.SourceVersion)
	}
	if _, err := time.Parse(time.RFC3339, manifest.GeneratedAt); err != nil {
		t.Fatalf("example generated_at is not RFC3339: %v", err)
	}
}
