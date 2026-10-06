package houjin

import (
	"context"
	"sync"
	"time"
)

// MemoryCacheOptions は NewMemoryCache の挙動を制御する。
//
// TTL: 各エントリの有効期限。ゼロ値の場合は無期限 (アプリ寿命と同じ)。
// Now: 現在時刻を取得する関数。テスト用に差し替え可能。nil なら time.Now を使う。
type MemoryCacheOptions struct {
	TTL time.Duration
	Now func() time.Time
}

// NewMemoryCache は in-memory な CompanyCache 実装を返す。
//
// sync.RWMutex + map ベースで、サイズ制限・LRU は持たない。プロセス内の短命キャッシュとして
// 「同じ法人番号を短時間に複数回引く」ユースケースの API 呼び出しを削減する用途を想定。
// 大量保存・永続化が必要なら Redis 等の外部キャッシュ実装に差し替える。
//
// SearchByNormalizedName は Phase 3 (法人名正規化) と同時に実装するため、現状は
// 常に空スライスを返す no-op。
func NewMemoryCache(opts MemoryCacheOptions) *MemoryCache {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	return &MemoryCache{
		ttl:     opts.TTL,
		now:     now,
		entries: make(map[string]memoryCacheEntry),
	}
}

// MemoryCache synchronizes entries in process; no size limit or persistence is
// provided. Returned profiles are shallow copies, including any shared Raw map.
type MemoryCache struct {
	ttl time.Duration
	now func() time.Time

	mu      sync.RWMutex
	entries map[string]memoryCacheEntry
}

type memoryCacheEntry struct {
	profile   CompanyProfile
	expiresAt time.Time // ゼロ値は「無期限」
}

func (c *MemoryCache) GetByCorporateNumber(_ context.Context, corporateNumber string) (*CompanyProfile, bool, error) {
	c.mu.RLock()
	entry, ok := c.entries[corporateNumber]
	c.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	if !entry.expiresAt.IsZero() && c.now().After(entry.expiresAt) {
		return nil, false, nil
	}
	profile := entry.profile
	return &profile, true, nil
}

func (c *MemoryCache) SetByCorporateNumber(_ context.Context, company CompanyProfile) error {
	var expiresAt time.Time
	if c.ttl > 0 {
		expiresAt = c.now().Add(c.ttl)
	}
	c.mu.Lock()
	c.entries[company.CorporateNumber] = memoryCacheEntry{
		profile:   company,
		expiresAt: expiresAt,
	}
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) SearchByNormalizedName(_ context.Context, _ string, _ int) ([]CompanyProfile, error) {
	return nil, nil
}
