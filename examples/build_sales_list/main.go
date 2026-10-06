// Optional legacy downstream example, retained for existing consumers.
// New sales policy belongs to outreach/Sales, not this source SDK.
// Command build_sales_list は案件 (Case) ベースで営業リストを生成するサンプル。
//
//	HOUJIN_APP_ID=xxx go run ./examples/build_sales_list <case_id> <name> [prefecture_code]
//
// 例:
//
//	HOUJIN_APP_ID=xxx go run ./examples/build_sales_list case-2025q2 "トヨタ" 23
//
// SearchByName で取得した CompanyProfile を案件条件で絞り込み、
// 営業リスト (SalesProspect) として標準出力する。
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	houjin "github.com/sibukixxx/ja-company-base"
	"github.com/sibukixxx/ja-company-base/adapters/saleslead"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: %s <case_id> <name> [prefecture_code]", os.Args[0])
	}
	caseID := os.Args[1]
	name := os.Args[2]
	prefectureCode := ""
	if len(os.Args) >= 4 {
		prefectureCode = os.Args[3]
	}

	appID := os.Getenv("HOUJIN_APP_ID")
	if appID == "" {
		return errors.New("HOUJIN_APP_ID is not set")
	}

	client, err := houjin.NewClient(houjin.Config{
		AppID:   appID,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	profiles, err := client.SearchByName(ctx, houjin.SearchByNameInput{
		Name:           name,
		PrefectureCode: prefectureCode,
		Limit:          50,
		MatchMode:      houjin.NameMatchModePartial,
	})
	if err != nil {
		return fmt.Errorf("SearchByName: %w", err)
	}

	c := saleslead.Case{
		ID:           caseID,
		Name:         "営業リスト: " + name,
		CompanyKinds: []string{"301"}, // 株式会社のみ対象 (例)
	}
	prospects := saleslead.BuildList(c, profiles)

	if len(prospects) == 0 {
		fmt.Println("対象法人は見つかりませんでした")
		return nil
	}

	fmt.Printf("=== 案件 %s : %d 件の営業対象 ===\n", c.ID, len(prospects))
	for i, p := range prospects {
		fmt.Printf("%2d. [%s] %s — %s (%s)\n",
			i+1, p.CorporateNumber, p.Name, p.Address.Full(), p.Status)
	}
	return nil
}
