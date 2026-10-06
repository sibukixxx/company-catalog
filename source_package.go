package houjin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"
)

const (
	// SourcePackageSchemaVersion identifies the consumer-neutral JSON envelope.
	SourcePackageSchemaVersion = "techvit.company.source-package/v1"
	// NormalizationVersion identifies current source projection semantics.
	NormalizationVersion = "company-projection/v1"
)

// CompanySourcePackage wraps the existing AnalysisRecord contract. It is raw
// source data, not an analytical result. IDs exclude observation/export time.
type CompanySourcePackage struct {
	SchemaVersion             string            `json:"schema_version"`
	PackageID                 string            `json:"package_id"`
	DatasetDefinitionRevision string            `json:"dataset_definition_revision"`
	SnapshotRevision          string            `json:"snapshot_revision"`
	ContentHash               string            `json:"content_hash"`
	GeneratedAt               string            `json:"generated_at"`
	SourceFetchedAt           string            `json:"source_fetched_at"`
	GeneratorVersion          string            `json:"generator_version"`
	NormalizationVersion      string            `json:"normalization_version"`
	Manifest                  AnalysisManifest  `json:"manifest"`
	Provenance                PackageProvenance `json:"provenance"`
	Completeness              string            `json:"completeness"`
	Warnings                  []string          `json:"warnings"`
	Records                   []AnalysisRecord  `json:"records"`
}

// PackageProvenance records the public API/terms URL and credential-free
// acquisition operation; it does not grant a license over upstream data.
type PackageProvenance struct {
	SourceURL   string `json:"source_url"`
	TermsURL    string `json:"terms_url"`
	Acquisition string `json:"acquisition"`
}

func hashJSON(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// NewCompanySourcePackage accepts only successfully fetched records. fetchedAt
// is required even for an empty successful response. limit>0 is conservatively
// marked LIMITED; it must never be interpreted as a complete population.
func NewCompanySourcePackage(records []AnalysisRecord, query AnalysisManifestQuery, fetchedAt, generatedAt time.Time, generatorVersion string) (CompanySourcePackage, error) {
	if fetchedAt.IsZero() || generatedAt.IsZero() || generatorVersion == "" {
		return CompanySourcePackage{}, errors.New("acquisition/export time and generator version are required")
	}
	from, err := time.Parse(time.DateOnly, query.From)
	if err != nil {
		return CompanySourcePackage{}, fmt.Errorf("invalid query from: %w", err)
	}
	to, err := time.Parse(time.DateOnly, query.To)
	if err != nil || to.Before(from) || query.Limit < 0 {
		return CompanySourcePackage{}, errors.New("invalid query range or limit")
	}
	rows := append([]AnalysisRecord{}, records...)
	content := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.SchemaVersion != AnalysisSchemaVersion || r.Source.Provider != sourceProvider || r.Source.Version != sourceVersion || r.CorporateNumber == "" || r.Name == "" {
			return CompanySourcePackage{}, errors.New("record lacks supported schema, identity or source provenance")
		}
		if _, err := time.Parse(time.RFC3339, r.Source.FetchedAt); err != nil {
			return CompanySourcePackage{}, fmt.Errorf("record fetched_at: %w", err)
		}
		r.Source.FetchedAt = ""
		b, err := json.Marshal(r)
		if err != nil {
			return CompanySourcePackage{}, err
		}
		content = append(content, string(b))
	}
	sort.Strings(content)
	sort.Slice(rows, func(i, j int) bool {
		a, _ := json.Marshal(rows[i])
		b, _ := json.Marshal(rows[j])
		return string(a) < string(b)
	})
	definition := hashJSON(struct {
		Schema, Provider, Version, Normalization string
		Query                                    AnalysisManifestQuery
	}{AnalysisSchemaVersion, sourceProvider, sourceVersion, NormalizationVersion, query})
	hash := hashJSON(content)
	snapshot := hashJSON([]string{definition, hash})
	completeness := "COMPLETE"
	warnings := []string{"Empty optional record strings mean unavailable/not supplied; no known-empty or zero is inferred.", "Geographic codes and industry classification are unavailable in the current projection.", "NTA update markers are not business founding or operating-status judgments."}
	if query.Limit > 0 {
		completeness = "LIMITED"
		warnings = append(warnings, "Requested record limit; do not infer population additions/removals.")
	}
	return CompanySourcePackage{SourcePackageSchemaVersion, "company-" + snapshot, definition, snapshot, hash, generatedAt.UTC().Format(time.RFC3339), fetchedAt.UTC().Format(time.RFC3339), generatorVersion, NormalizationVersion, NewAnalysisManifest("json", query, generatedAt, len(rows)), PackageProvenance{"https://www.houjin-bangou.nta.go.jp/webapi/", "https://www.houjin-bangou.nta.go.jp/webapi/riyokiyaku.html", "FetchUpdatesByDateRange"}, completeness, warnings, rows}, nil
}

// Validate verifies the envelope before any downstream use, including hashes.
func (p CompanySourcePackage) Validate() error {
	if p.SchemaVersion != SourcePackageSchemaVersion || p.NormalizationVersion != NormalizationVersion {
		return errors.New("unsupported source package version")
	}
	fetched, err := time.Parse(time.RFC3339, p.SourceFetchedAt)
	if err != nil {
		return err
	}
	generated, err := time.Parse(time.RFC3339, p.GeneratedAt)
	if err != nil {
		return err
	}
	expected, err := NewCompanySourcePackage(p.Records, p.Manifest.Query, fetched, generated, p.GeneratorVersion)
	if err != nil {
		return err
	}
	if p.PackageID != expected.PackageID || p.ContentHash != expected.ContentHash || p.SnapshotRevision != expected.SnapshotRevision || p.DatasetDefinitionRevision != expected.DatasetDefinitionRevision || p.Manifest != expected.Manifest || p.Provenance != expected.Provenance || p.Completeness != expected.Completeness || hashJSON(p.Warnings) != hashJSON(expected.Warnings) {
		return errors.New("source package metadata/content mismatch")
	}
	return nil
}

// WriteCompanySourcePackage validates before writing one JSON envelope.
// It does not close w; callers must check their file Close errors as well.
func WriteCompanySourcePackage(w io.Writer, p CompanySourcePackage) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(p)
}

// ReadCompanySourcePackage decodes one envelope, rejects trailing JSON, and
// validates versions, provenance, identities, counts and metadata. It consumes
// the reader through EOF; use bounded inputs supplied by a trusted file source.
func ReadCompanySourcePackage(r io.Reader) (CompanySourcePackage, error) {
	var p CompanySourcePackage
	dec := json.NewDecoder(r)
	if err := dec.Decode(&p); err != nil {
		return p, err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return p, errors.New("trailing source package content")
	}
	return p, p.Validate()
}
