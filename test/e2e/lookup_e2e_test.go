//go:build e2e

// Package e2e は国税庁の法人番号 Web-API への実通信を伴うテストを置く。
//
// 通常の `go test ./...` では実行されない（build tag `e2e` で隔離）。
// 実行: HOUJIN_E2E_TEST_ENABLED=true HOUJIN_APP_ID=xxx make test-e2e
package e2e

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	houjin "github.com/sibukixxx/company-catalog"
)

const (
	// トヨタ自動車株式会社の法人番号（実在・公開情報）
	toyotaCorporateNumber = "1180301018771"
)

func skipIfNotEnabled(t *testing.T) string {
	t.Helper()
	if os.Getenv("HOUJIN_E2E_TEST_ENABLED") != "true" {
		t.Skip("HOUJIN_E2E_TEST_ENABLED is not 'true' — skipping e2e test")
	}
	appID := os.Getenv("HOUJIN_APP_ID")
	if appID == "" {
		t.Skip("HOUJIN_APP_ID is not set — skipping e2e test")
	}
	return appID
}

func newE2EClient(t *testing.T) houjin.Client {
	t.Helper()
	appID := skipIfNotEnabled(t)
	c, err := houjin.NewClient(houjin.Config{
		AppID:      appID,
		Timeout:    15 * time.Second,
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestE2ELookupByCorporateNumberReturnsValidProfile(t *testing.T) {
	c := newE2EClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got, err := c.LookupByCorporateNumber(ctx, toyotaCorporateNumber)
	if err != nil {
		t.Fatalf("LookupByCorporateNumber: %v", err)
	}
	if got.CorporateNumber != toyotaCorporateNumber {
		t.Errorf("CorporateNumber = %q, want %q", got.CorporateNumber, toyotaCorporateNumber)
	}
	if got.Name == "" {
		t.Error("Name should not be empty")
	}
	if got.Address.PrefectureName == "" {
		t.Error("PrefectureName should not be empty")
	}
}

func TestE2ELookupByCorporateNumberReturnsNotFoundForNonExistent(t *testing.T) {
	c := newE2EClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// チェックディジット式に従って計算した、実在しないであろう番号
	// 基礎番号 000000000001 の checkDigit は 8 -> 8000000000001
	_, err := c.LookupByCorporateNumber(ctx, "8000000000001")
	if err == nil {
		t.Fatal("expected error for non-existent corporate number")
	}
	var apiErr *houjin.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != houjin.ErrorCodeNotFound {
		t.Errorf("code = %q, want %q", apiErr.Code, houjin.ErrorCodeNotFound)
	}
}

func TestE2ESearchByNameReturnsCandidates(t *testing.T) {
	c := newE2EClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := c.SearchByName(ctx, houjin.SearchByNameInput{
		Name:           "トヨタ自動車",
		PrefectureCode: "23", // 愛知県
		Limit:          5,
	})
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}
