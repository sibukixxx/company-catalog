package houjin

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// Cache が設定され、CachePolicyReadThrough のとき:
//   - 1 回目の Lookup で API を叩き、結果が cache に保存される
//   - 2 回目の Lookup で API は叩かれず、cache から返る
func TestLookupByCorporateNumberServesFromCacheOnSecondCallWhenReadThrough(t *testing.T) {
	var apiCalls atomic.Int32
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalls.Add(1)
		_, _ = io.WriteString(w, sampleXMLOneCorp)
	})

	cache := NewMemoryCache(MemoryCacheOptions{TTL: time.Minute})
	c, err := NewClient(Config{
		AppID:       "test",
		BaseURL:     srv.URL,
		HTTPClient:  hc,
		Cache:       cache,
		CachePolicy: CachePolicyReadThrough,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()
	first, err := c.LookupByCorporateNumber(ctx, "1180301018771")
	if err != nil {
		t.Fatalf("1st Lookup: %v", err)
	}
	second, err := c.LookupByCorporateNumber(ctx, "1180301018771")
	if err != nil {
		t.Fatalf("2nd Lookup: %v", err)
	}

	if got := apiCalls.Load(); got != 1 {
		t.Errorf("apiCalls = %d, want 1 (2nd call should hit cache)", got)
	}
	if first.CorporateNumber != second.CorporateNumber || first.Name != second.Name {
		t.Errorf("cached profile differs: first=%+v, second=%+v", first, second)
	}
}

// CachePolicyDisabled (デフォルト) では、Cache が設定されていてもバイパスして毎回 API を叩く。
func TestLookupByCorporateNumberBypassesCacheWhenPolicyDisabled(t *testing.T) {
	var apiCalls atomic.Int32
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalls.Add(1)
		_, _ = io.WriteString(w, sampleXMLOneCorp)
	})

	cache := NewMemoryCache(MemoryCacheOptions{TTL: time.Minute})
	c, _ := NewClient(Config{
		AppID:       "test",
		BaseURL:     srv.URL,
		HTTPClient:  hc,
		Cache:       cache,
		CachePolicy: CachePolicyDisabled,
	})

	ctx := context.Background()
	_, _ = c.LookupByCorporateNumber(ctx, "1180301018771")
	_, _ = c.LookupByCorporateNumber(ctx, "1180301018771")

	if got := apiCalls.Load(); got != 2 {
		t.Errorf("apiCalls = %d, want 2 (cache should be bypassed when policy=disabled)", got)
	}
}

// Cache 未設定 (Config.Cache == nil) のときは、CachePolicyReadThrough が指定されていても
// 単に API を毎回叩くだけで panic などせず安全に動くこと。
func TestLookupByCorporateNumberWorksWithoutCacheEvenWhenPolicyReadThrough(t *testing.T) {
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sampleXMLOneCorp)
	})
	c, _ := NewClient(Config{
		AppID:       "test",
		BaseURL:     srv.URL,
		HTTPClient:  hc,
		CachePolicy: CachePolicyReadThrough,
	})

	got, err := c.LookupByCorporateNumber(context.Background(), "1180301018771")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.CorporateNumber != "1180301018771" {
		t.Errorf("CorporateNumber = %q, want %q", got.CorporateNumber, "1180301018771")
	}
}

// not_found のレスポンスは cache に保存してはいけない (negative caching を避ける)。
// 後から法人が登録された場合に古い not_found を返し続けてしまうため。
func TestLookupByCorporateNumberDoesNotCacheNotFoundResults(t *testing.T) {
	var apiCalls atomic.Int32
	srv, hc := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalls.Add(1)
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?><corporations count="0"></corporations>`)
	})

	cache := NewMemoryCache(MemoryCacheOptions{TTL: time.Minute})
	c, _ := NewClient(Config{
		AppID:       "test",
		BaseURL:     srv.URL,
		HTTPClient:  hc,
		Cache:       cache,
		CachePolicy: CachePolicyReadThrough,
	})

	ctx := context.Background()
	_, _ = c.LookupByCorporateNumber(ctx, "1180301018771")
	_, _ = c.LookupByCorporateNumber(ctx, "1180301018771")

	if got := apiCalls.Load(); got != 2 {
		t.Errorf("apiCalls = %d, want 2 (not_found must not be cached)", got)
	}
}
