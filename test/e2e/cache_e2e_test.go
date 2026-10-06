//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	houjin "github.com/sibukixxx/ja-company-base"
)

// 実 API + MemoryCache + ReadThrough の組み合わせで、
// 同じ法人番号への 2 回目の Lookup が API を叩かないことを確認する。
//
// HTTP リクエスト回数をカウントするために、http.RoundTripper をラップした
// HTTPClient を Config に注入する。
func TestE2EMemoryCacheHitAvoidsSecondAPICall(t *testing.T) {
	appID := skipIfNotEnabled(t)

	counter := &countingTransport{base: http.DefaultTransport}
	cache := houjin.NewMemoryCache(houjin.MemoryCacheOptions{TTL: time.Minute})

	c, err := houjin.NewClient(houjin.Config{
		AppID:       appID,
		Timeout:     30 * time.Second,
		HTTPClient:  &http.Client{Transport: counter, Timeout: 30 * time.Second},
		Cache:       cache,
		CachePolicy: houjin.CachePolicyReadThrough,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	first, err := c.LookupByCorporateNumber(ctx, toyotaCorporateNumber)
	if err != nil {
		t.Fatalf("1st Lookup: %v", err)
	}
	afterFirst := counter.calls.Load()

	second, err := c.LookupByCorporateNumber(ctx, toyotaCorporateNumber)
	if err != nil {
		t.Fatalf("2nd Lookup: %v", err)
	}
	afterSecond := counter.calls.Load()

	if afterFirst != 1 {
		t.Errorf("HTTP calls after 1st Lookup = %d, want 1", afterFirst)
	}
	if afterSecond != 1 {
		t.Errorf("HTTP calls after 2nd Lookup = %d, want 1 (cache should serve 2nd)", afterSecond)
	}
	if first.CorporateNumber != second.CorporateNumber {
		t.Errorf("cached profile mismatch: first=%q, second=%q",
			first.CorporateNumber, second.CorporateNumber)
	}
	if os.Getenv("HOUJIN_E2E_VERBOSE") == "true" {
		t.Logf("verified cache: 1st=%s (HTTP), 2nd=%s (cache)", first.Name, second.Name)
	}
}

type countingTransport struct {
	base  http.RoundTripper
	calls atomic.Int32
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.calls.Add(1)
	return t.base.RoundTrip(req)
}
