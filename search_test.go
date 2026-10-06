package houjin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

const sampleXMLTwoCorps = `<?xml version="1.0" encoding="UTF-8"?>
<corporations id="1" lastUpdateDate="2024-01-01" count="2" divideNumber="1" divideSize="1">
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>1180301018771</corporateNumber>
    <name>トヨタ自動車株式会社</name>
    <kind>301</kind>
    <prefectureName>愛知県</prefectureName>
    <cityName>豊田市</cityName>
    <streetNumber>トヨタ町１番地</streetNumber>
    <postCode>4718571</postCode>
    <updateDate>2024-01-01</updateDate>
    <changeDate>2023-12-15</changeDate>
    <assignmentDate>2015-10-05</assignmentDate>
  </corporation>
  <corporation>
    <sequenceNumber>2</sequenceNumber>
    <corporateNumber>5010001008857</corporateNumber>
    <name>ソニー株式会社</name>
    <kind>301</kind>
    <prefectureName>東京都</prefectureName>
    <cityName>港区</cityName>
    <streetNumber>港南１丁目７番１号</streetNumber>
    <postCode>1080075</postCode>
    <updateDate>2024-01-01</updateDate>
    <changeDate>2023-11-20</changeDate>
    <assignmentDate>2015-10-05</assignmentDate>
  </corporation>
</corporations>`

func TestSearchByNameSendsExpectedRequest(t *testing.T) {
	var gotPath, gotQuery string
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, sampleXMLTwoCorps)
	})

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.SearchByName(context.Background(), SearchByNameInput{
		Name:           "トヨタ",
		PrefectureCode: "23",
		CityCode:       "211",
		Limit:          50,
		MatchMode:      NameMatchModePartial,
	})
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}

	if !strings.HasSuffix(gotPath, "/name") {
		t.Errorf("path = %q, want suffix /name", gotPath)
	}
	for _, want := range []string{"id=test", "name=", "address=23211", "type=12"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
}

func TestSearchByNameReturnsErrorOnEmptyName(t *testing.T) {
	c, _ := NewClient(Config{AppID: "test"})

	_, err := c.SearchByName(context.Background(), SearchByNameInput{Name: ""})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != ErrorCodeInvalidInput {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeInvalidInput)
	}
}

func TestSearchByNameReturnsEmptySliceOnNoResults(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?><corporations count="0"></corporations>`)
	})

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.SearchByName(context.Background(), SearchByNameInput{Name: "存在しない"})
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

// MatchMode が API クエリ "mode" にどう変換されるかの契約検証
// （README: NameMatchModeDefault → 送らない / Exact → 1 / Partial → 2）
func TestSearchByNameMapsMatchModeToQueryParam(t *testing.T) {
	tests := []struct {
		name        string
		mode        NameMatchMode
		wantPresent bool
		wantValue   string
	}{
		{"Default は mode を送らない", NameMatchModeDefault, false, ""},
		{"Exact → mode=1", NameMatchModeExact, true, "1"},
		{"Partial → mode=2", NameMatchModePartial, true, "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMode string
			var modePresent bool
			srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if vals, ok := r.URL.Query()["mode"]; ok {
					modePresent = true
					gotMode = vals[0]
				}
				_, _ = io.WriteString(w, sampleXMLTwoCorps)
			})

			c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})
			_, err := c.SearchByName(context.Background(), SearchByNameInput{
				Name:      "トヨタ",
				MatchMode: tt.mode,
			})
			if err != nil {
				t.Fatalf("SearchByName: %v", err)
			}

			if modePresent != tt.wantPresent {
				t.Errorf("mode present = %v, want %v (got value %q)", modePresent, tt.wantPresent, gotMode)
			}
			if tt.wantPresent && gotMode != tt.wantValue {
				t.Errorf("mode = %q, want %q", gotMode, tt.wantValue)
			}
		})
	}
}

// SearchByName の Limit が結果を先頭 N 件に切り詰めることを検証
// （README「2. 法人名で候補を検索」: Limit はクライアント側で先頭 N 件に絞る）
func TestSearchByNameTruncatesResultsToLimit(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		// sampleXMLTwoCorps は 2 件返す
		_, _ = io.WriteString(w, sampleXMLTwoCorps)
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.SearchByName(context.Background(), SearchByNameInput{
		Name:  "トヨタ",
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len(results) = %d, want 1 (Limit=1)", len(results))
	}
}

// Limit=0 は「無制限」として扱われ、全結果が返ることを検証
func TestSearchByNameReturnsAllResultsWhenLimitIsZero(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sampleXMLTwoCorps)
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.SearchByName(context.Background(), SearchByNameInput{Name: "トヨタ"})
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("len(results) = %d, want 2 (no limit)", len(results))
	}
}

// FetchUpdatesByDateRange の Limit / CompanyKind がクエリ・結果に反映されることを検証
func TestFetchUpdatesByDateRangeAppliesLimitAndCompanyKind(t *testing.T) {
	var gotKind string
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotKind = r.URL.Query().Get("kind")
		_, _ = io.WriteString(w, sampleXMLTwoCorps)
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{
		FromDate:    "2024-01-01",
		ToDate:      "2024-01-31",
		CompanyKind: "301",
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("FetchUpdatesByDateRange: %v", err)
	}
	if gotKind != "301" {
		t.Errorf("kind query = %q, want %q", gotKind, "301")
	}
	if len(results) != 1 {
		t.Errorf("len(results) = %d, want 1 (Limit=1)", len(results))
	}
}

func TestFetchUpdatesByDateRangeRequiresFromAndToDate(t *testing.T) {
	c, _ := NewClient(Config{AppID: "test"})

	_, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != ErrorCodeInvalidInput {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeInvalidInput)
	}
}

func TestFetchUpdatesByDateRangeSendsExpectedRequest(t *testing.T) {
	var gotPath, gotQuery string
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, sampleXMLTwoCorps)
	})

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{
		FromDate:       "2024-01-01",
		ToDate:         "2024-01-31",
		PrefectureCode: "13",
	})
	if err != nil {
		t.Fatalf("FetchUpdatesByDateRange: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
	if !strings.HasSuffix(gotPath, "/diff") {
		t.Errorf("path = %q, want suffix /diff", gotPath)
	}
	for _, want := range []string{"from=2024-01-01", "to=2024-01-31", "address=13"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
}
