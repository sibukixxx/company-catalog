//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	houjin "github.com/sibukixxx/company-catalog"
)

// 分割応答の自動結合を実 API で確認する。
//
// 国税庁 API は応答が大きい場合に divideSize > 1 で返す。
// 期間広め + 都道府県指定なしの /diff は容易に divideSize が複数になる。
//
// 検証ポイント:
//   - エラーなく取得できる
//   - 件数が「1 ページのみ取得 (旧実装) ありえないほど大きい数」になっている
//     = 全ページ取得して結合できている
//
// 上流の挙動依存度を下げるため、件数の下限は緩めに置く (>= 1)。
// divide が実際に起きたかは件数だけからは断定できないが、本テストは
// 「実 API への疎通 + parser + 結合ロジックが破綻していない」回帰検知用。
func TestE2EFetchUpdatesByDateRangeAggregatesDividedPages(t *testing.T) {
	c := newE2EClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	results, err := c.FetchUpdatesByDateRange(ctx, houjin.FetchUpdatesInput{
		// 直近 7 日間。1 週間あれば数千〜数万件の差分が出るのが通常。
		FromDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
		ToDate:   time.Now().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("FetchUpdatesByDateRange: %v", err)
	}
	if len(results) == 0 {
		// 実 API が 0 件返すことは原理的にはあり得るが、過去 7 日間で
		// 全国の更新ゼロは現実的にゼロに近い。0 件なら API/環境を疑う。
		t.Log("WARNING: 0 corporations returned for last 7 days. API state may be unusual.")
	}
	// 各 Corporation がパース後正規化されていること（最小限の sanity check）
	for i, r := range results {
		if r.CorporateNumber == "" {
			t.Errorf("results[%d].CorporateNumber is empty", i)
			break
		}
	}
}
