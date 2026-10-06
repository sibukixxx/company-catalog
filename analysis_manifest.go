package houjin

import (
	"encoding/json"
	"io"
	"time"
)

// AnalysisManifest describes one analysis export file without repeating the
// records. It is written next to the CSV / JSONL data so downstream tools can
// check the schema and provenance before reading rows. It must never contain
// the AppID or any other credential.
type AnalysisManifest struct {
	Schema        string                `json:"schema"`
	Source        string                `json:"source"`
	SourceVersion string                `json:"source_version"`
	Format        string                `json:"format"`
	Query         AnalysisManifestQuery `json:"query"`
	GeneratedAt   string                `json:"generated_at"`
	RecordCount   int                   `json:"record_count"`
}

// AnalysisManifestQuery records the upstream filters used for the export.
// Empty filters are omitted.
type AnalysisManifestQuery struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Prefecture string `json:"prefecture,omitempty"`
	City       string `json:"city,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// NewAnalysisManifest fills the schema and source fields owned by this
// package. generatedAt is the export time and is stored in UTC RFC3339.
func NewAnalysisManifest(format string, query AnalysisManifestQuery, generatedAt time.Time, recordCount int) AnalysisManifest {
	return AnalysisManifest{
		Schema:        AnalysisSchemaVersion,
		Source:        sourceProvider,
		SourceVersion: sourceVersion,
		Format:        format,
		Query:         query,
		GeneratedAt:   generatedAt.UTC().Format(time.RFC3339),
		RecordCount:   recordCount,
	}
}

// WriteAnalysisManifest writes the manifest as indented JSON.
func WriteAnalysisManifest(w io.Writer, manifest AnalysisManifest) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}
