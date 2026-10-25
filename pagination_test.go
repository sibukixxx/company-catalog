package houjin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// 国税庁 Web-API は応答が大きい場合 divideSize 分割で返す。
// クライアントは divideNumber=1 の応答を見て、2..divideSize までを
// `divide=<N>` クエリで追加リクエストし、全件結合する責務を持つ。

func paginatedXML(divideNumber, divideSize int, corporateNumbers ...string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		`<?xml version="1.0" encoding="UTF-8"?>`+
			`<corporations id="1" lastUpdateDate="2024-01-01" count="%d" divideNumber="%d" divideSize="%d">`,
		divideSize*len(corporateNumbers), divideNumber, divideSize)
	for i, n := range corporateNumbers {
		fmt.Fprintf(&sb,
			`<corporation><sequenceNumber>%d</sequenceNumber><corporateNumber>%s</corporateNumber><name>会社%d</name></corporation>`,
			i+1, n, i+1)
	}
	sb.WriteString(`</corporations>`)
	return sb.String()
}

// 3 ページに分割された /diff 応答を、SDK が全ページ取得して結合することを検証。
func TestFetchUpdatesByDateRangeAggregatesAllPagesWhenResponseIsDivided(t *testing.T) {
	var (
		mu          sync.Mutex
		divideQuery []string
	)
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		divideQuery = append(divideQuery, r.URL.Query().Get("divide"))
		mu.Unlock()
		switch r.URL.Query().Get("divide") {
		case "", "1":
			_, _ = io.WriteString(w, paginatedXML(1, 3, "1000000000001", "1000000000002"))
		case "2":
			_, _ = io.WriteString(w, paginatedXML(2, 3, "1000000000003", "1000000000004"))
		case "3":
			_, _ = io.WriteString(w, paginatedXML(3, 3, "1000000000005", "1000000000006"))
		default:
			t.Errorf("unexpected divide query: %q", r.URL.Query().Get("divide"))
		}
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	})
	if err != nil {
		t.Fatalf("FetchUpdatesByDateRange: %v", err)
	}
	if len(results) != 6 {
		t.Errorf("len(results) = %d, want 6 (3 pages * 2 corporations)", len(results))
	}
	// 1 ページ目は divide なし、以降 "2", "3"
	wantQueries := []string{"", "2", "3"}
	if diff := stringSliceDiff(wantQueries, divideQuery); diff != "" {
		t.Errorf("divide queries mismatch:\n%s", diff)
	}
}

// 分割応答では各ページを受信した時刻を、そのページ由来の法人の FetchedAt に UTC RFC3339 で記録する。
func TestFetchUpdatesByDateRangeStampsFetchedAtPerPageInUTC(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("divide") {
		case "":
			_, _ = io.WriteString(w, paginatedXML(1, 2, "1000000000001"))
		case "2":
			_, _ = io.WriteString(w, paginatedXML(2, 2, "1000000000002"))
		}
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})
	clock := time.Date(2026, 9, 13, 9, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	c.(*httpClient).now = func() time.Time {
		clock = clock.Add(time.Minute)
		return clock
	}

	results, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	})
	if err != nil {
		t.Fatalf("FetchUpdatesByDateRange: %v", err)
	}
	want := []string{"2026-09-13T00:01:00Z", "2026-09-13T00:02:00Z"}
	if len(results) != len(want) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(want))
	}
	for i, profile := range results {
		if profile.Source.FetchedAt != want[i] {
			t.Errorf("results[%d].Source.FetchedAt = %q, want %q", i, profile.Source.FetchedAt, want[i])
		}
	}
}

// SearchByName でも同じく分割応答を結合する。
// さらに Limit は結合後の全件に対して適用される（先頭 N 件で切り詰め）。
func TestSearchByNameAggregatesAllPagesAndAppliesLimitAfterAggregation(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("divide") {
		case "", "1":
			_, _ = io.WriteString(w, paginatedXML(1, 2, "1000000000001", "1000000000002"))
		case "2":
			_, _ = io.WriteString(w, paginatedXML(2, 2, "1000000000003", "1000000000004"))
		default:
			t.Errorf("unexpected divide query: %q", r.URL.Query().Get("divide"))
		}
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	results, err := c.SearchByName(context.Background(), SearchByNameInput{
		Name:  "会社",
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("len(results) = %d, want 3 (4件中先頭3件)", len(results))
	}
	// 結合後の先頭から 3 件: 1, 2, 3
	wantNumbers := []string{"1000000000001", "1000000000002", "1000000000003"}
	for i, r := range results {
		if r.CorporateNumber != wantNumbers[i] {
			t.Errorf("results[%d].CorporateNumber = %q, want %q", i, r.CorporateNumber, wantNumbers[i])
		}
	}
}

// 異常応答による無限ループを防ぐ安全装置:
// サーバが divideSize を返し続けても、SDK は上限 (maxDividePages) 以降を取得しない。
// テストでは divideSize=100 を返すサーバを立てて、リクエスト回数が上限内に収まることを確認。
func TestFetchProfilesStopsAtMaxDivideToAvoidInfiniteLoop(t *testing.T) {
	var (
		mu      sync.Mutex
		callCnt int
	)
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCnt++
		mu.Unlock()
		page := 1
		if v := r.URL.Query().Get("divide"); v != "" {
			fmt.Sscanf(v, "%d", &page)
		}
		_, _ = io.WriteString(w, paginatedXML(page, 100, "100000000000"+strconv.Itoa(page%10)))
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	_, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	})
	if err == nil {
		t.Fatal("expected error when divideSize exceeds safety limit")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != ErrorCodeAPIError {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeAPIError)
	}
	mu.Lock()
	defer mu.Unlock()
	// サーバは divideSize=100 を返し続けるので、安全装置が無ければ 100 回呼ばれる。
	// 安全装置で打ち切られていれば、現実的な上限内 (50 回未満) に収まるはず。
	if callCnt >= 50 {
		t.Errorf("callCnt = %d, safety limit appears to be missing or too loose", callCnt)
	}
}

// 期待 / 実測のスライスを順序込みで diff する小さなヘルパー。
func stringSliceDiff(want, got []string) string {
	if len(want) != len(got) {
		return fmt.Sprintf("len: want=%d, got=%d\n  want=%v\n  got=%v", len(want), len(got), want, got)
	}
	for i := range want {
		if want[i] != got[i] {
			return fmt.Sprintf("[%d]: want=%q, got=%q\n  full want=%v\n  full got=%v", i, want[i], got[i], want, got)
		}
	}
	return ""
}

func TestFetchRejectsIncompleteAggregate(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		page := 1
		if r.URL.Query().Get("divide") == "2" {
			page = 2
		}
		_, _ = io.WriteString(w, strings.Replace(paginatedXML(page, 2, "1000000000001"), `count="2"`, `count="3"`, 1))
	})
	c, err := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := c.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{FromDate: "2024-01-01", ToDate: "2024-01-31"})
	if err == nil || rows != nil {
		t.Fatal("partial aggregate returned as success")
	}
}
