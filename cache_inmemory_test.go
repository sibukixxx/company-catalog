package houjin

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryCacheReturnsStoredCompanyByCorporateNumber(t *testing.T) {
	cache := NewMemoryCache(MemoryCacheOptions{TTL: time.Minute})
	ctx := context.Background()

	want := CompanyProfile{
		CorporateNumber: "1180301018771",
		Name:            "トヨタ自動車株式会社",
	}
	if err := cache.SetByCorporateNumber(ctx, want); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok, err := cache.GetByCorporateNumber(ctx, "1180301018771")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if got == nil {
		t.Fatal("got nil profile on cache hit")
	}
	if got.CorporateNumber != want.CorporateNumber || got.Name != want.Name {
		t.Errorf("got = %+v, want %+v", got, want)
	}
}

func TestMemoryCacheReturnsMissWhenCorporateNumberIsNotStored(t *testing.T) {
	cache := NewMemoryCache(MemoryCacheOptions{TTL: time.Minute})

	got, ok, err := cache.GetByCorporateNumber(context.Background(), "9999999999999")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ok {
		t.Errorf("expected miss, got hit: %+v", got)
	}
	if got != nil {
		t.Errorf("expected nil profile on miss, got %+v", got)
	}
}

func TestMemoryCacheExpiresEntryAfterTTL(t *testing.T) {
	cache := NewMemoryCache(MemoryCacheOptions{TTL: 10 * time.Millisecond})
	ctx := context.Background()

	_ = cache.SetByCorporateNumber(ctx, CompanyProfile{CorporateNumber: "1180301018771"})

	// TTL 経過前はヒット
	if _, ok, _ := cache.GetByCorporateNumber(ctx, "1180301018771"); !ok {
		t.Fatal("expected hit immediately after Set")
	}

	time.Sleep(20 * time.Millisecond)

	// TTL 経過後はミス
	if _, ok, _ := cache.GetByCorporateNumber(ctx, "1180301018771"); ok {
		t.Error("expected miss after TTL elapsed")
	}
}

func TestMemoryCacheTTLZeroMeansNoExpiration(t *testing.T) {
	// TTL=0 は明示的に「無期限」として扱う（ゼロ値の有用性）。
	cache := NewMemoryCache(MemoryCacheOptions{TTL: 0})
	ctx := context.Background()

	_ = cache.SetByCorporateNumber(ctx, CompanyProfile{CorporateNumber: "1180301018771"})
	time.Sleep(5 * time.Millisecond)
	if _, ok, _ := cache.GetByCorporateNumber(ctx, "1180301018771"); !ok {
		t.Error("expected hit when TTL is 0 (no expiration)")
	}
}

func TestMemoryCacheIsSafeForConcurrentUse(t *testing.T) {
	cache := NewMemoryCache(MemoryCacheOptions{TTL: time.Minute})
	ctx := context.Background()

	const goroutines = 32
	const iterations = 100

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for range iterations {
				num := "1000000000001"
				_ = cache.SetByCorporateNumber(ctx, CompanyProfile{CorporateNumber: num, Name: "test"})
				_, _, _ = cache.GetByCorporateNumber(ctx, num)
			}
		}(g)
	}
	wg.Wait()
}

func TestMemoryCacheImplementsCompanyCacheInterface(t *testing.T) {
	// コンパイル時の型互換性確認。NewMemoryCache の戻り値が CompanyCache を満たすこと。
	var _ CompanyCache = NewMemoryCache(MemoryCacheOptions{})
}
