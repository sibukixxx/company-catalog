package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	houjin "github.com/sibukixxx/company-catalog"
)

type fakeFetcher struct {
	profiles []houjin.CompanyProfile
	input    houjin.FetchUpdatesInput
}

func (f *fakeFetcher) FetchUpdatesByDateRange(_ context.Context, input houjin.FetchUpdatesInput) ([]houjin.CompanyProfile, error) {
	f.input = input
	return f.profiles, nil
}

func TestRunWritesCSVFromFetchedProfiles(t *testing.T) {
	fetcher := &fakeFetcher{profiles: []houjin.CompanyProfile{{
		CorporateNumber: "1234567890123",
		Name:            "株式会社テスト,研究所",
		Address:         houjin.Address{PrefectureName: "東京都", CityName: "千代田区"},
		AssignmentDate:  "2026-09-01",
		Source:          houjin.SourceInfo{Provider: "houjin_bangou", Version: "v4"},
	}}}
	var stdout bytes.Buffer
	err := run(context.Background(), []string{
		"-from", "2026-09-01", "-to", "2026-09-13", "-format", "csv",
		"-prefecture", "13", "-limit", "10",
	}, func(string) string { return "test-app-id" }, &stdout, &bytes.Buffer{}, func(houjin.Config) (updateFetcher, error) {
		return fetcher, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	rows, err := csv.NewReader(strings.NewReader(stdout.String())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][1] != "株式会社テスト,研究所" {
		t.Fatalf("unexpected CSV: %#v", rows)
	}
	if fetcher.input.PrefectureCode != "13" || fetcher.input.Limit != 10 {
		t.Fatalf("flags not forwarded: %#v", fetcher.input)
	}
}

func TestRunWritesJSONLToFile(t *testing.T) {
	fetcher := &fakeFetcher{profiles: []houjin.CompanyProfile{
		{CorporateNumber: "1", Name: "株式会社甲", AssignmentDate: "2026-09-01"},
		{CorporateNumber: "2", Name: "株式会社乙", CloseDate: "2026-09-02"},
	}}
	output := filepath.Join(t.TempDir(), "companies.jsonl")
	err := run(context.Background(), []string{
		"-from", "2026-09-01", "-to", "2026-09-13", "-format", "jsonl", "-output", output,
	}, func(string) string { return "test-app-id" }, &bytes.Buffer{}, &bytes.Buffer{}, func(houjin.Config) (updateFetcher, error) {
		return fetcher, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	var second houjin.AnalysisRecord
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatal(err)
	}
	if second.EventType != houjin.EventTypeClosed {
		t.Fatalf("event type = %q, want CLOSED", second.EventType)
	}
}

func TestRunWritesManifestWithProvenanceAndWithoutAppID(t *testing.T) {
	fetcher := &fakeFetcher{profiles: []houjin.CompanyProfile{
		{CorporateNumber: "1", Name: "株式会社甲"},
		{CorporateNumber: "2", Name: "株式会社乙"},
	}}
	dir := t.TempDir()
	output := filepath.Join(dir, "companies.jsonl")
	manifestPath := filepath.Join(dir, "companies.manifest.json")
	err := run(context.Background(), []string{
		"-from", "2026-09-01", "-to", "2026-09-13", "-prefecture", "13",
		"-format", "jsonl", "-output", output, "-manifest", manifestPath,
	}, func(string) string { return "secret-app-id" }, &bytes.Buffer{}, &bytes.Buffer{}, func(houjin.Config) (updateFetcher, error) {
		return fetcher, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("secret-app-id")) {
		t.Fatalf("manifest leaked the AppID: %s", data)
	}
	var manifest houjin.AnalysisManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != houjin.AnalysisSchemaVersion || manifest.Source != "houjin_bangou" || manifest.SourceVersion != "v4" {
		t.Fatalf("unexpected manifest provenance: %#v", manifest)
	}
	wantQuery := houjin.AnalysisManifestQuery{From: "2026-09-01", To: "2026-09-13", Prefecture: "13"}
	if manifest.Format != "jsonl" || manifest.Query != wantQuery || manifest.RecordCount != 2 {
		t.Fatalf("unexpected manifest: %#v", manifest)
	}
	if _, err := time.Parse(time.RFC3339, manifest.GeneratedAt); err != nil {
		t.Fatalf("generated_at is not RFC3339: %v", err)
	}
}

func TestRunRejectsMissingAppID(t *testing.T) {
	err := run(context.Background(), []string{"-from", "2026-09-01", "-to", "2026-09-13"}, func(string) string { return "" }, &bytes.Buffer{}, &bytes.Buffer{}, func(houjin.Config) (updateFetcher, error) {
		t.Fatal("client must not be created")
		return nil, nil
	})
	if err == nil || !strings.Contains(err.Error(), appIDEnv) {
		t.Fatalf("error = %v", err)
	}
}

func TestRunEndToEndFetchesNTACompatibleXMLAndExportsCSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/diff" {
			t.Errorf("path = %q, want /diff", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("from") != "2026-09-01" || query.Get("to") != "2026-09-13" {
			t.Errorf("date query = %q to %q", query.Get("from"), query.Get("to"))
		}
		if query.Get("id") != "test-app-id" {
			t.Errorf("app id not forwarded")
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<corporations id="1" lastUpdateDate="2026-09-13" count="1" divideNumber="1" divideSize="1">
  <corporation>
    <corporateNumber>1180301018771</corporateNumber>
    <name>動作確認株式会社</name>
    <kind>301</kind>
    <prefectureName>東京都</prefectureName>
    <cityName>西東京市</cityName>
    <assignmentDate>2026-09-01</assignmentDate>
    <updateDate>2026-09-13</updateDate>
  </corporation>
</corporations>`)
	}))
	defer server.Close()

	output := filepath.Join(t.TempDir(), "companies.csv")
	err := run(context.Background(), []string{
		"-from", "2026-09-01", "-to", "2026-09-13", "-output", output,
	}, func(string) string { return "test-app-id" }, &bytes.Buffer{}, &bytes.Buffer{}, func(cfg houjin.Config) (updateFetcher, error) {
		cfg.BaseURL = server.URL
		return houjin.NewClient(cfg)
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want header + 1 record", len(rows))
	}
	if rows[1][0] != "1180301018771" || rows[1][1] != "動作確認株式会社" || rows[1][12] != "UPDATED" {
		t.Fatalf("unexpected exported row: %#v", rows[1])
	}
}

func TestParseOptionsRejectsInvalidFormatLimitAndManifestPath(t *testing.T) {
	for _, args := range [][]string{
		{"-from", "2026-09-01", "-to", "2026-09-13", "-format", "yaml"},
		{"-from", "2026-09-01", "-to", "2026-09-13", "-limit", "-1"},
		{"-from", "2026-09-01", "-to", "2026-09-13", "-output", "out.csv", "-manifest", "out.csv"},
		{"-from", "2026-09-01", "-to", "2026-09-13", "-output", "out.csv", "-manifest", "-"},
	} {
		if _, err := parseOptions(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("parseOptions(%v) succeeded", args)
		}
	}
}

func TestRunWritesValidatedSourcePackageWithoutCredential(t *testing.T) {
	fetcher := &fakeFetcher{profiles: []houjin.CompanyProfile{{CorporateNumber: "0000000000001", Name: "Synthetic Contract Company", Source: houjin.SourceInfo{Provider: "houjin_bangou", Version: "v4", FetchedAt: "2026-01-01T00:00:00Z"}}}}
	var stdout bytes.Buffer
	err := run(context.Background(), []string{"-from", "2025-12-01", "-to", "2025-12-31", "-format", "json"}, func(string) string { return "credential-sentinel" }, &stdout, &bytes.Buffer{}, func(houjin.Config) (updateFetcher, error) { return fetcher, nil })
	if err != nil {
		t.Fatal(err)
	}
	p, err := houjin.ReadCompanySourcePackage(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if p.Manifest.RecordCount != 1 || strings.Contains(stdout.String(), "credential-sentinel") {
		t.Fatal("invalid/credential-bearing output")
	}
	fetcher.profiles[0].Source.Provider = ""
	stdout.Reset()
	if err := run(context.Background(), []string{"-from", "2025-12-01", "-to", "2025-12-31", "-format", "json"}, func(string) string { return "credential-sentinel" }, &stdout, &bytes.Buffer{}, func(houjin.Config) (updateFetcher, error) { return fetcher, nil }); err == nil || stdout.Len() != 0 {
		t.Fatal("invalid source exported")
	}
}
