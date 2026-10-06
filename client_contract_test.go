package houjin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

// 本ファイルは README で約束した API 契約を網羅的に検証する。
// すべて緑なら「README の Client / APIError / Config / 値型 の挙動契約は仕様通り」と言える。
//
// 詳細な契約 → テストの対応表は docs/API_CONTRACT_TESTS.md を参照。

// HTTP ステータスごとに正しい ErrorCode と StatusCode を返すか
// （README「エラーハンドリング」表の機械的検証）
func TestLookupByCorporateNumberMapsHTTPStatusToErrorCode(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantCode   ErrorCode
		wantStatus int
	}{
		{"400 BadRequest → InvalidInput", http.StatusBadRequest, ErrorCodeInvalidInput, 400},
		{"401 Unauthorized → Unauthorized", http.StatusUnauthorized, ErrorCodeUnauthorized, 401},
		{"403 Forbidden → Unauthorized", http.StatusForbidden, ErrorCodeUnauthorized, 403},
		{"404 NotFound → NotFound", http.StatusNotFound, ErrorCodeNotFound, 404},
		{"429 TooManyRequests → RateLimited", http.StatusTooManyRequests, ErrorCodeRateLimited, 429},
		{"500 InternalServerError → APIError", http.StatusInternalServerError, ErrorCodeAPIError, 500},
		{"503 ServiceUnavailable → APIError", http.StatusServiceUnavailable, ErrorCodeAPIError, 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			})
			c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

			_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.wantCode)
			}
			if apiErr.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantStatus)
			}
		})
	}
}

// context.Canceled が ErrorCodeTimeout に分類されることを検証
// （README「エラーハンドリング」: timeout = context.DeadlineExceeded 等）
func TestLookupByCorporateNumberReturnsTimeoutWhenContextCanceled(t *testing.T) {
	// blocker は handler を中断させるための shutdown 信号。
	// httptest.Server.Close は handler の終了を待つため、Cleanup を
	// newTestServer (= srv.Close 登録) の "後に" 登録して LIFO 順で
	// blocker → srv.Close の順に走るようにする。
	blocker := make(chan struct{})
	srv, hc := newTestServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		<-blocker
	})
	t.Cleanup(func() { close(blocker) })

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.LookupByCorporateNumber(ctx, "1180301018771")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != ErrorCodeTimeout {
		t.Errorf("Code = %q, want %q", apiErr.Code, ErrorCodeTimeout)
	}
}

// context.DeadlineExceeded が ErrorCodeTimeout に分類されることを検証
func TestLookupByCorporateNumberReturnsTimeoutWhenContextDeadlineExceeded(t *testing.T) {
	blocker := make(chan struct{})
	srv, hc := newTestServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		<-blocker
	})
	t.Cleanup(func() { close(blocker) })

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.LookupByCorporateNumber(ctx, "1180301018771")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != ErrorCodeTimeout {
		t.Errorf("Code = %q, want %q", apiErr.Code, ErrorCodeTimeout)
	}
}

// XML パース失敗時に ErrorCodeParseError を返し、Cause に元エラーが入ることを検証
// （README「エラーハンドリング」: parse_error = XML 解析失敗）
func TestLookupByCorporateNumberReturnsParseErrorOnMalformedXML(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "<not-xml<<<")
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != ErrorCodeParseError {
		t.Errorf("Code = %q, want %q", apiErr.Code, ErrorCodeParseError)
	}
	if apiErr.Cause == nil {
		t.Error("Cause should be populated with the underlying parse error")
	}
}

// CompanyProfile の全フィールドが XML から正しくマップされることを検証
// （README「データモデル」CompanyProfile / Address の全フィールド契約）
func TestLookupByCorporateNumberMapsAllFieldsFromXML(t *testing.T) {
	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<corporations id="1" lastUpdateDate="2024-01-01" count="1" divideNumber="1" divideSize="1">
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>1180301018771</corporateNumber>
    <name>トヨタ自動車株式会社</name>
    <furigana>トヨタジドウシャ</furigana>
    <enName>Toyota Motor Corporation</enName>
    <kind>301</kind>
    <prefectureName>愛知県</prefectureName>
    <cityName>豊田市</cityName>
    <streetNumber>トヨタ町１番地</streetNumber>
    <addressImageId>img-001</addressImageId>
    <postCode>4718571</postCode>
    <assignmentDate>2015-10-05</assignmentDate>
    <updateDate>2024-01-01</updateDate>
    <changeDate>2023-12-15</changeDate>
    <changeCause>商号変更</changeCause>
    <closeDate>2024-03-31</closeDate>
    <closeCause>合併</closeCause>
  </corporation>
</corporations>`

	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, xml)
	})
	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})
	c.(*httpClient).now = func() time.Time {
		return time.Date(2026, 9, 13, 9, 30, 0, 0, time.FixedZone("JST", 9*60*60))
	}

	got, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	if err != nil {
		t.Fatalf("LookupByCorporateNumber: %v", err)
	}

	want := CompanyProfile{
		CorporateNumber: "1180301018771",
		Name:            "トヨタ自動車株式会社",
		NameKana:        "トヨタジドウシャ",
		NameEn:          "Toyota Motor Corporation",
		Kind:            CompanyKind("301"),
		Address: Address{
			PostalCode:     "4718571",
			PrefectureName: "愛知県",
			CityName:       "豊田市",
			StreetNumber:   "トヨタ町１番地",
			AddressImageID: "img-001",
		},
		AssignmentDate: "2015-10-05",
		UpdateDate:     "2024-01-01",
		ChangeDate:     "2023-12-15",
		ChangeReason:   "商号変更",
		CloseDate:      "2024-03-31",
		CloseCause:     "合併",
		Source:         SourceInfo{Provider: "houjin_bangou", Version: "v4", FetchedAt: "2026-09-13T00:30:00Z"},
	}
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("CompanyProfile mismatch:\n got = %+v\nwant = %+v", *got, want)
	}
}

// Address.Full() が PrefectureName + CityName + StreetNumber を順に連結することを検証
func TestAddressFullConcatenatesParts(t *testing.T) {
	a := Address{
		PrefectureName: "愛知県",
		CityName:       "豊田市",
		StreetNumber:   "トヨタ町１番地",
	}
	const want = "愛知県豊田市トヨタ町１番地"
	if got := a.Full(); got != want {
		t.Errorf("Full() = %q, want %q", got, want)
	}
}

// 7 つの ErrorCode 定数が README の表と一致することを検証
// （README「エラーハンドリング」表の各値が export されていること）
func TestErrorCodeConstantsAreExposed(t *testing.T) {
	want := map[ErrorCode]string{
		ErrorCodeInvalidInput: "invalid_input",
		ErrorCodeUnauthorized: "unauthorized",
		ErrorCodeNotFound:     "not_found",
		ErrorCodeRateLimited:  "rate_limited",
		ErrorCodeAPIError:     "api_error",
		ErrorCodeParseError:   "parse_error",
		ErrorCodeTimeout:      "timeout",
	}
	for code, str := range want {
		if string(code) != str {
			t.Errorf("ErrorCode constant = %q, want %q", string(code), str)
		}
	}
}

// APIError.Error() に AppID（Secret）が漏れない契約を、
// HTTP エラー時とパースエラー時の両方で検証
// （README「セキュリティ・運用上の注意」: AppID をログに出してはいけない）
func TestAPIErrorDoesNotLeakAppIDOnAnyErrorPath(t *testing.T) {
	const secret = "SECRET-APP-ID-9876"

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"HTTP 500", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}},
		{"HTTP 401", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}},
		{"malformed XML", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "<<<broken")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, hc := newTestServer(t, tt.handler)
			c, _ := NewClient(Config{AppID: secret, BaseURL: srv.URL, HTTPClient: hc})

			_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
			if err == nil {
				t.Fatal("expected error")
			}
			if strings.Contains(err.Error(), secret) {
				t.Errorf("APIError.Error() leaks AppID: %v", err)
			}
		})
	}
}
