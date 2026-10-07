// Command seed は国税庁の法人番号 Web-API から実 XML レスポンスを取得し、
// testdata/seed/ 配下にゴールデンファイルとして保存する。
//
// オフラインでパーサや enrichment 層のテストを書く際の入力ソースとして使う。
//
//	HOUJIN_APP_ID=xxx make seed
//
// 既知の公開情報のみを保存する。AppID やレスポンスヘッダーは保存しない。
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.houjin-bangou.nta.go.jp/4"
	outputDir      = "testdata/seed"
	httpTimeout    = 15 * time.Second
)

type seedSpec struct {
	name string // ファイル名（拡張子除く）
	path string // /num, /name, /diff
	q    url.Values
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	appID := strings.TrimSpace(os.Getenv("HOUJIN_APP_ID"))
	if appID == "" {
		return errors.New("HOUJIN_APP_ID is not set")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create seed output directory: %w", err)
	}

	specs := []seedSpec{
		{
			name: "lookup_toyota",
			path: "/num",
			q: url.Values{
				"id":      {appID},
				"number":  {"1180301018771"}, // トヨタ自動車株式会社
				"type":    {"12"},
				"history": {"0"},
			},
		},
		{
			name: "lookup_sony",
			path: "/num",
			q: url.Values{
				"id":      {appID},
				"number":  {"7010401088415"}, // ソニーグループ株式会社
				"type":    {"12"},
				"history": {"0"},
			},
		},
		{
			name: "search_toyota_name",
			path: "/name",
			q: url.Values{
				"id":      {appID},
				"name":    {"トヨタ自動車"},
				"address": {"23"}, // 愛知県
				"type":    {"12"},
				"history": {"0"},
				"mode":    {"2"}, // partial
			},
		},
	}

	hc := &http.Client{Timeout: httpTimeout}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, spec := range specs {
		if err := fetchAndSave(ctx, hc, spec); err != nil {
			return fmt.Errorf("fetch %s: %w", spec.name, err)
		}
		fmt.Printf("OK   %s -> %s/%s.xml\n", spec.name, outputDir, spec.name)
	}
	return nil
}
func fetchAndSave(ctx context.Context, hc *http.Client, spec seedSpec) error {
	endpoint := defaultBaseURL + spec.path + "?" + spec.q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.New("build request failed")
	}

	resp, err := hc.Do(req)
	if err != nil {
		return errors.New("HTTP request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("upstream returned non-success status")
	}

	out := filepath.Join(outputDir, spec.name+".xml")
	if err := os.WriteFile(out, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	return nil
}
