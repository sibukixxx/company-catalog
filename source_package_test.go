package houjin

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func packageFixture(t *testing.T) CompanySourcePackage {
	t.Helper()
	// Explicitly synthetic contract records, not live seed/golden API data.
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	records := []AnalysisRecord{
		ProjectForAnalysis(CompanyProfile{CorporateNumber: "0000000000001", Name: "Synthetic Contract Company A", Source: SourceInfo{sourceProvider, sourceVersion, at.Format(time.RFC3339)}}),
		ProjectForAnalysis(CompanyProfile{CorporateNumber: "0000000000002", Name: "Synthetic Contract Company B", Kind: "301", Source: SourceInfo{sourceProvider, sourceVersion, at.Format(time.RFC3339)}}),
	}
	p, err := NewCompanySourcePackage(records, AnalysisManifestQuery{From: "2025-12-01", To: "2025-12-31", Prefecture: "13"}, at, at, "contract-fixture/v1")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSourcePackageIdentity(t *testing.T) {
	p := packageFixture(t)
	rows := append([]AnalysisRecord{}, p.Records...)
	rows[0], rows[1] = rows[1], rows[0]
	for i := range rows {
		rows[i].Source.FetchedAt = "2026-01-02T00:00:00Z"
	}
	later := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	rerun, err := NewCompanySourcePackage(rows, p.Manifest.Query, later, later, "new-build")
	if err != nil {
		t.Fatal(err)
	}
	if rerun.PackageID != p.PackageID || rerun.ContentHash != p.ContentHash {
		t.Fatal("time/order changed content identity")
	}
	rows[0].Name = "Changed"
	changed, err := NewCompanySourcePackage(rows, p.Manifest.Query, later, later, "new-build")
	if err != nil {
		t.Fatal(err)
	}
	if changed.ContentHash == p.ContentHash || changed.SnapshotRevision == p.SnapshotRevision {
		t.Fatal("missed record change")
	}
	q := p.Manifest.Query
	q.Prefecture = "14"
	filtered, err := NewCompanySourcePackage(p.Records, q, later, later, "new-build")
	if err != nil {
		t.Fatal(err)
	}
	if filtered.DatasetDefinitionRevision == p.DatasetDefinitionRevision || filtered.ContentHash != p.ContentHash {
		t.Fatal("definition/content confused")
	}
}

func TestSourcePackageValidation(t *testing.T) {
	for _, mutate := range []func(*CompanySourcePackage){
		func(p *CompanySourcePackage) { p.Records[0].Source.Provider = "" },
		func(p *CompanySourcePackage) { p.Records[0].Source.FetchedAt = "" },
		func(p *CompanySourcePackage) { p.SchemaVersion = "v2" },
		func(p *CompanySourcePackage) { p.Manifest.RecordCount++ },
		func(p *CompanySourcePackage) { p.Records[0].Name = "tampered" },
		func(p *CompanySourcePackage) { p.Provenance.SourceURL = "" },
	} {
		p := packageFixture(t)
		mutate(&p)
		var b bytes.Buffer
		if WriteCompanySourcePackage(&b, p) == nil || b.Len() != 0 {
			t.Fatal("invalid package exported")
		}
	}
	p := packageFixture(t)
	var b bytes.Buffer
	if err := WriteCompanySourcePackage(&b, p); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadCompanySourcePackage(&b); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadCompanySourcePackage(strings.NewReader(`{} {}`)); err == nil {
		t.Fatal("accepted trailing input")
	}
}

func TestSourcePackageEmptyAndLimited(t *testing.T) {
	p := packageFixture(t)
	at, _ := time.Parse(time.RFC3339, p.GeneratedAt)
	empty, err := NewCompanySourcePackage(nil, p.Manifest.Query, at, at, "test")
	if err != nil || empty.Manifest.RecordCount != 0 || empty.Records == nil {
		t.Fatal("zero confused with unknown", err)
	}
	q := p.Manifest.Query
	q.Limit = 1
	limited, err := NewCompanySourcePackage(p.Records[:1], q, at, at, "test")
	if err != nil || limited.Completeness != "LIMITED" {
		t.Fatal("limited mislabeled", err)
	}
	if p.Records[0].Kind != "" || p.Records[0].EventType != EventTypeUnknown {
		t.Fatal("missing/unknown inferred")
	}
}

func TestSourcePackageFixture(t *testing.T) {
	p := packageFixture(t)
	// Explicit update mode only; fixture is contract data, never a fetched golden.
	if os.Getenv("UPDATE_CONTRACT_FIXTURE") == "true" {
		f, err := os.Create("docs/examples/company-source-package.v1.json")
		if err != nil {
			t.Fatal(err)
		}
		err = WriteCompanySourcePackage(f, p)
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
	data, err := os.ReadFile("docs/examples/company-source-package.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ReadCompanySourcePackage(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if hashJSON(got) != hashJSON(p) {
		t.Fatal("fixture drift")
	}
}
