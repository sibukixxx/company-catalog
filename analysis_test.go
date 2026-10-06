package houjin

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func TestProjectForAnalysis(t *testing.T) {
	company := CompanyProfile{
		CorporateNumber: "1234567890123",
		Name:            "株式会社テスト,研究所",
		Kind:            CompanyKind("301"),
		Address: Address{
			PrefectureName: "東京都",
			CityName:       "千代田区",
		},
		AssignmentDate: "2020-01-02",
		ChangeDate:     "2025-04-01",
		Source:         SourceInfo{Provider: "houjin_bangou", Version: "v4", FetchedAt: "2026-09-13T00:00:00Z"},
	}

	record := ProjectForAnalysis(company)
	if record.CorporateNumber != company.CorporateNumber || record.Name != company.Name {
		t.Fatalf("identity fields not preserved: %#v", record)
	}
	if record.PrefectureName != "東京都" || record.CityName != "千代田区" {
		t.Fatalf("address fields not preserved: %#v", record)
	}
	if record.PrefectureCode != "" || record.CityCode != "" {
		t.Fatalf("geographic codes must not be inferred: %#v", record)
	}
	if record.EventType != EventTypeChanged {
		t.Fatalf("event type = %q, want %q", record.EventType, EventTypeChanged)
	}
	wantSource := AnalysisSource{Provider: "houjin_bangou", Version: "v4", FetchedAt: "2026-09-13T00:00:00Z"}
	if record.Source != wantSource {
		t.Fatalf("source provenance not preserved: %#v", record.Source)
	}
}

func TestAnalysisEventType(t *testing.T) {
	tests := []struct {
		name string
		in   CompanyProfile
		want EventType
	}{
		{"unknown", CompanyProfile{}, EventTypeUnknown},
		{"assigned", CompanyProfile{AssignmentDate: "2020-01-01"}, EventTypeAssigned},
		{"updated", CompanyProfile{AssignmentDate: "2020-01-01", UpdateDate: "2021-01-01"}, EventTypeUpdated},
		{"changed", CompanyProfile{UpdateDate: "2021-01-01", ChangeDate: "2021-02-01"}, EventTypeChanged},
		{"closed", CompanyProfile{UpdateDate: "2021-01-01", ChangeDate: "2021-02-01", CloseDate: "2021-03-01"}, EventTypeClosed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analysisEventType(tt.in); got != tt.want {
				t.Fatalf("analysisEventType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCSVAnalysisWriter(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewCSVAnalysisWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	records := []AnalysisRecord{
		{CorporateNumber: "1", Name: "株式会社A,研究所", PrefectureName: "東京都", EventType: EventTypeAssigned},
		{CorporateNumber: "2", Name: "株式会社\"B\"\n支店", PrefectureName: "長野県", EventType: EventTypeUnknown},
	}
	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	rows, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
	if err != nil {
		t.Fatalf("exported CSV is invalid: %v\n%s", err, buf.String())
	}
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	if rows[0][0] != "corporate_number" || rows[1][1] != records[0].Name || rows[2][1] != records[1].Name {
		t.Fatalf("unexpected CSV rows: %#v", rows)
	}
}

func TestJSONLAnalysisWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLAnalysisWriter(&buf)
	records := []AnalysisRecord{
		{CorporateNumber: "1", Name: "株式会社甲", EventType: EventTypeAssigned},
		{CorporateNumber: "2", Name: "株式会社乙", EventType: EventTypeUnknown},
	}
	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(records) {
		t.Fatalf("lines = %d, want %d", len(lines), len(records))
	}
	for i, line := range lines {
		var got AnalysisRecord
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d is invalid JSON: %v", i, err)
		}
		if got.CorporateNumber != records[i].CorporateNumber {
			t.Fatalf("line %d corporate number = %q, want %q", i, got.CorporateNumber, records[i].CorporateNumber)
		}
	}
}

func TestProjectForAnalysisSetsSchemaVersion(t *testing.T) {
	record := ProjectForAnalysis(CompanyProfile{CorporateNumber: "1234567890123"})
	if record.SchemaVersion != AnalysisSchemaVersion {
		t.Fatalf("schema version = %q, want %q", record.SchemaVersion, AnalysisSchemaVersion)
	}
	if AnalysisSchemaVersion != "techvit.analysis.company/v1" {
		t.Fatalf("AnalysisSchemaVersion = %q; changing it is a breaking contract change", AnalysisSchemaVersion)
	}
}

// docs/ANALYSIS_EXPORT_CONTRACT.md の CSV 列順を固定する。v1 の範囲では末尾追加のみ許される。
func TestCSVAnalysisHeaderMatchesSchemaV1ColumnOrder(t *testing.T) {
	want := []string{
		"corporate_number", "name", "kind", "prefecture_code", "prefecture_name",
		"city_code", "city_name", "assignment_date", "update_date", "change_date",
		"close_date", "close_cause", "event_type", "source_provider", "source_version",
		"source_fetched_at", "schema_version",
	}
	var buf bytes.Buffer
	w, err := NewCSVAnalysisWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write(ProjectForAnalysis(CompanyProfile{
		CorporateNumber: "1234567890123",
		Source:          SourceInfo{Provider: "houjin_bangou", Version: "v4", FetchedAt: "2026-09-13T00:00:00Z"},
	})); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rows[0], ",") != strings.Join(want, ",") {
		t.Fatalf("CSV header =\n%v\nwant\n%v", rows[0], want)
	}
	row := map[string]string{}
	for i, column := range rows[0] {
		row[column] = rows[1][i]
	}
	if row["schema_version"] != AnalysisSchemaVersion || row["source_fetched_at"] != "2026-09-13T00:00:00Z" {
		t.Fatalf("provenance columns not exported: %#v", row)
	}
}

// docs/ANALYSIS_EXPORT_CONTRACT.md の JSONL キー集合を固定する。
func TestJSONLAnalysisRecordKeysMatchSchemaV1(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLAnalysisWriter(&buf)
	if err := w.Write(ProjectForAnalysis(CompanyProfile{CorporateNumber: "1234567890123"})); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	var source map[string]json.RawMessage
	if err := json.Unmarshal(got["source"], &source); err != nil {
		t.Fatalf("source is not an object: %v", err)
	}

	wantTop := []string{
		"assignment_date", "change_date", "city_code", "city_name", "close_cause",
		"close_date", "corporate_number", "event_type", "kind", "name",
		"prefecture_code", "prefecture_name", "schema_version", "source", "update_date",
	}
	wantSource := []string{"fetched_at", "provider", "version"}
	if gotKeys := sortedKeys(got); strings.Join(gotKeys, ",") != strings.Join(wantTop, ",") {
		t.Fatalf("JSONL keys = %v, want %v", gotKeys, wantTop)
	}
	if gotKeys := sortedKeys(source); strings.Join(gotKeys, ",") != strings.Join(wantSource, ",") {
		t.Fatalf("JSONL source keys = %v, want %v", gotKeys, wantSource)
	}
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
