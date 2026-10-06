package houjin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *http.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, srv.Client()
}

func TestNewClientReturnsErrorWhenAppIDIsEmpty(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Error("NewClient should fail without AppID")
	}
}

func TestLookupByCorporateNumberSendsExpectedRequest(t *testing.T) {
	var (
		gotPath  string
		gotQuery string
	)
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, sampleXMLOneCorp)
	})

	c, err := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	got, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	if err != nil {
		t.Fatalf("LookupByCorporateNumber: %v", err)
	}
	if got == nil || got.CorporateNumber != "1180301018771" {
		t.Errorf("unexpected company: %+v", got)
	}

	if !strings.HasSuffix(gotPath, "/num") {
		t.Errorf("path = %q, want suffix /num", gotPath)
	}
	for _, want := range []string{"id=test", "number=1180301018771", "type=12"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
}

func TestLookupByCorporateNumberReturnsNotFoundWhenNoCorporation(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?><corporations count="0"></corporations>`)
	})

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != ErrorCodeNotFound {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeNotFound)
	}
}

func TestLookupByCorporateNumberValidatesInput(t *testing.T) {
	c, _ := NewClient(Config{AppID: "test"})

	_, err := c.LookupByCorporateNumber(context.Background(), "abc")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != ErrorCodeInvalidInput {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeInvalidInput)
	}
}

func TestLookupByCorporateNumberMaps401ToUnauthorized(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != ErrorCodeUnauthorized {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeUnauthorized)
	}
}

func TestLookupByCorporateNumberMaps429ToRateLimited(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	c, _ := NewClient(Config{AppID: "test", BaseURL: srv.URL, HTTPClient: hc})

	_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != ErrorCodeRateLimited {
		t.Errorf("code = %q, want %q", apiErr.Code, ErrorCodeRateLimited)
	}
}

func TestLookupByCorporateNumberDoesNotLogAppID(t *testing.T) {
	// 確認: AppID は構造化ログ・エラーメッセージに出さない方針。
	// ここでは APIError.Error() に AppID が混ざらないことだけ確認。
	srv, hc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	c, _ := NewClient(Config{AppID: "SECRET-APP-ID", BaseURL: srv.URL, HTTPClient: hc})

	_, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "SECRET-APP-ID") {
		t.Errorf("error message leaks AppID: %v", err)
	}
}
