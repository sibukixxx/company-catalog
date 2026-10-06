// Command lookup_by_number は法人番号 1 件から法人情報を取得する最小例。
//
//	HOUJIN_APP_ID=xxx go run ./examples/lookup_by_number 1180301018771
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	houjin "github.com/sibukixxx/ja-company-base"
)

var errNotFound = errors.New("not found")

func main() {
	if err := run(); err != nil {
		if errors.Is(err, errNotFound) {
			fmt.Println("見つかりませんでした")
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <corporate_number>", os.Args[0])
	}
	corpNumber := os.Args[1]

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

	company, err := client.LookupByCorporateNumber(ctx, corpNumber)
	if err != nil {
		var apiErr *houjin.APIError
		if errors.As(err, &apiErr) && apiErr.Code == houjin.ErrorCodeNotFound {
			return errNotFound
		}
		return fmt.Errorf("LookupByCorporateNumber: %w", err)
	}

	fmt.Printf("法人番号: %s\n", company.CorporateNumber)
	fmt.Printf("法人名:   %s\n", company.Name)
	fmt.Printf("ﾌﾘｶﾞﾅ:    %s\n", company.NameKana)
	fmt.Printf("住所:     〒%s %s\n", company.Address.PostalCode, company.Address.Full())
	fmt.Printf("更新日:   %s\n", company.UpdateDate)
	return nil
}
