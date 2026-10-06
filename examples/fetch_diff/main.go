// Command fetch_diff は更新差分 (/diff) を取得して件数とサンプルを表示する例。
//
//	HOUJIN_APP_ID=xxx go run ./examples/fetch_diff [days] [prefecture_code]
//
// days を省略すると 7 日。prefecture_code を省略すると全国対象（pagination が
// 発生しやすい）。件数が 5000 を超えていれば SDK 内部で分割応答結合が走っている。
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	houjin "github.com/sibukixxx/ja-company-base"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	days := 7
	if len(os.Args) >= 2 {
		n, err := strconv.Atoi(os.Args[1])
		if err != nil || n <= 0 {
			return fmt.Errorf("invalid days argument: %q (must be positive integer)", os.Args[1])
		}
		days = n
	}
	prefectureCode := ""
	if len(os.Args) >= 3 {
		prefectureCode = os.Args[2]
	}

	appID := os.Getenv("HOUJIN_APP_ID")
	if appID == "" {
		return errors.New("HOUJIN_APP_ID is not set")
	}

	client, err := houjin.NewClient(houjin.Config{
		AppID:   appID,
		Timeout: 90 * time.Second, // pagination で複数往復になる可能性を見て長め
	})
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	now := time.Now()
	from := now.AddDate(0, 0, -days).Format("2006-01-02")
	to := now.Format("2006-01-02")

	fmt.Printf("=== /diff %s..%s (prefectureCode=%q) ===\n", from, to, prefectureCode)

	start := time.Now()
	results, err := client.FetchUpdatesByDateRange(ctx, houjin.FetchUpdatesInput{
		FromDate:       from,
		ToDate:         to,
		PrefectureCode: prefectureCode,
	})
	elapsed := time.Since(start)
	if err != nil {
		return fmt.Errorf("FetchUpdatesByDateRange: %w", err)
	}

	fmt.Printf("取得件数:     %d\n", len(results))
	fmt.Printf("経過時間:     %s\n", elapsed.Round(time.Millisecond))
	// 1 ページ最大 5000 件。それを超えていれば SDK 内部で divide=2..N の追加リクエストが走っている。
	if len(results) > 5000 {
		fmt.Println("分割応答:     検出 (件数 > 5000、SDK が divide=2..N を順次取得して結合済み)")
	} else {
		fmt.Println("分割応答:     未検出 (単一ページ応答だった可能性)")
	}

	if len(results) == 0 {
		fmt.Println("(0 件のため詳細出力なし)")
		return nil
	}

	sampleN := min(len(results), 5)
	fmt.Printf("\n=== 先頭 %d 件のサンプル ===\n", sampleN)
	for i := range sampleN {
		r := results[i]
		fmt.Printf("%2d. [%s] %s — %s (更新日 %s)\n",
			i+1, r.CorporateNumber, r.Name, r.Address.Full(), r.UpdateDate)
	}
	return nil
}
