// Command search_by_name は法人名から候補一覧を取得する例。
//
//	HOUJIN_APP_ID=xxx go run ./examples/search_by_name トヨタ自動車 23
//
// 第二引数の prefecture code は任意（省略可）。
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	houjin "github.com/sibukixxx/company-catalog"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <name> [prefecture_code]", os.Args[0])
	}
	name := os.Args[1]
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
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("NewClient: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	companies, err := client.SearchByName(ctx, houjin.SearchByNameInput{
		Name:           name,
		PrefectureCode: prefectureCode,
		Limit:          20,
		MatchMode:      houjin.NameMatchModePartial,
	})
	if err != nil {
		return fmt.Errorf("SearchByName: %w", err)
	}

	if len(companies) == 0 {
		fmt.Println("候補は見つかりませんでした")
		return nil
	}

	fmt.Printf("=== %d 件の候補 ===\n", len(companies))
	for i, c := range companies {
		fmt.Printf("%2d. [%s] %s — %s\n", i+1, c.CorporateNumber, c.Name, c.Address.Full())
	}
	return nil
}
