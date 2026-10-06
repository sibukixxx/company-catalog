# SDK reference

公開Go APIの利用例です。初回実行は[README](../README.md)、fixtureとCLIは
[Examples](../examples/README.md)を参照してください。

## 使い方

### 1. 法人番号で 1 件取得

```go
company, err := client.LookupByCorporateNumber(ctx, "1180301018771")
```

- 入力は 13 桁の数字文字列。チェックディジットも検証される
- 該当法人がない場合は `ErrorCodeNotFound` を返す
- 入力が不正な場合は `ErrorCodeInvalidInput` を返す（API へのリクエストは飛ばない）

### 2. 法人名で候補を検索

```go
results, err := client.SearchByName(ctx, houjin.SearchByNameInput{
	Name:           "トヨタ自動車",
	PrefectureCode: "23",            // 愛知県（任意）
	CityCode:       "211",           // 豊田市（任意）
	CompanyKind:    "",              // 法人種別（任意）
	Limit:          50,              // クライアント側で先頭 N 件に絞る
	MatchMode:      houjin.NameMatchModePartial, // 部分一致
})
```

| `MatchMode` | 動作 |
| --- | --- |
| `NameMatchModeDefault` | API 側のデフォルト |
| `NameMatchModeExact`   | 完全一致 |
| `NameMatchModePartial` | 部分一致 |

都道府県コード・市区町村コードは [総務省「全国地方公共団体コード」](https://www.soumu.go.jp/denshijiti/code.html) 準拠。

### 3. 取得期間指定で差分取得

```go
updates, err := client.FetchUpdatesByDateRange(ctx, houjin.FetchUpdatesInput{
	FromDate:       "2024-01-01",
	ToDate:         "2024-01-31",
	PrefectureCode: "13",  // 東京都のみに絞る (任意)
})
```

- 取引先マスタの定期同期や、法人情報の差分キャッシュ更新に使う
- 国税庁 API の分割応答 (`divideSize` > 1) は SDK 内部で `divide=2..N` を順次リクエスト
  して全件結合済みのスライスを返す。安全装置として最大 20 ページまで取得し、それを
  超える場合は `ErrorCodeAPIError` を返す

---

## データモデル

### `CompanyProfile`

| フィールド | 型 | 内容 |
| --- | --- | --- |
| `CorporateNumber` | `string` | 法人番号（13 桁） |
| `Name` | `string` | 商号または名称 |
| `NameKana` | `string` | フリガナ |
| `NameEn` | `string` | 英語表記（登録があれば） |
| `Kind` | `CompanyKind` | 法人種別コード（例: `301` = 株式会社） |
| `Address` | `Address` | 所在地（後述） |
| `AssignmentDate` | `string` | 法人番号指定年月日（YYYY-MM-DD） |
| `UpdateDate` | `string` | 最終更新年月日 |
| `ChangeDate` | `string` | 直近変更年月日 |
| `ChangeReason` | `string` | 変更事由 |
| `CloseDate` | `string` | 登記閉鎖等年月日（解散等） |
| `CloseCause` | `string` | 登記閉鎖等の事由 |
| `Source` | `SourceInfo` | データソース情報（`Provider=houjin_bangou` / `Version=v4` / `FetchedAt`=上流応答の受信時刻 UTC RFC3339） |
| `Raw` | `map[string]string` | 将来の拡張用（現状は未使用） |

### `Address`

| フィールド | 内容 |
| --- | --- |
| `PostalCode` | 郵便番号（ハイフンなし、例: `4718571`） |
| `PrefectureName` | 都道府県名 |
| `CityName` | 市区町村名 |
| `StreetNumber` | 丁目・番地 |
| `AddressImageID` | 画像で表記される住所の ID（外字対応） |

`addr.Full()` で都道府県＋市区町村＋丁目番地を連結した文字列を取得可能。

---

## Config

```go
type Config struct {
	AppID       string         // 必須。国税庁の AppID
	BaseURL     string         // 省略時は https://api.houjin-bangou.nta.go.jp/4
	Timeout     time.Duration  // 省略時 10s。HTTPClient 未指定時のみ反映
	MaxRetries  int            // 予約field。既定値2だがretry処理は未実装
	HTTPClient  *http.Client   // 省略時は &http.Client{Timeout: Timeout}
	Cache       CompanyCache   // Lookupのread-through cacheで使用
	CachePolicy CachePolicy    // CachePolicyDisabled / CachePolicyReadThrough
}
```

`Validate()` で必須項目を検査し、`WithDefaults()` でゼロ値を埋める。`NewClient` は両者を内部で呼び出す。

---

## エラーハンドリング

HTTP・入力・パース由来のエラーは通常`*houjin.APIError`で判定できます。
ページ整合性・Source Package検証等には通常のerrorもあるため、
`errors.As`がfalseの経路も処理してください。

```go
var apiErr *houjin.APIError
if errors.As(err, &apiErr) {
	switch apiErr.Code {
	case houjin.ErrorCodeNotFound:
		// 候補なし — UI で「見つかりません」表示など
	case houjin.ErrorCodeUnauthorized:
		// AppID 設定ミス — 即座に運用に通知
	case houjin.ErrorCodeRateLimited:
		// レート制限 — backoff してリトライ
	case houjin.ErrorCodeTimeout:
		// タイムアウト — リトライ可能性高
	case houjin.ErrorCodeInvalidInput:
		// 入力ミス — クライアント側で修正
	case houjin.ErrorCodeParseError:
		// XML パース失敗 — API 仕様変更の可能性。issue 報告希望
	case houjin.ErrorCodeAPIError:
		// その他 5xx 等
	}
}
```

| ErrorCode | 発生条件 | 推奨対応 |
| --- | --- | --- |
| `invalid_input` | 入力バリデーション失敗 | クライアント側で修正 |
| `unauthorized` | HTTP 401 / 403 | AppID 確認 |
| `not_found` | HTTP 404 / 結果 0 件 | UI で表示 |
| `rate_limited` | HTTP 429 | backoff（数秒〜数分） |
| `timeout` | `context.DeadlineExceeded` 等 | リトライ |
| `parse_error` | XML 解析失敗 | バグ報告対象 |
| `api_error` | その他（500 等） | バグ報告 + リトライ |

`apiErr.StatusCode` には HTTP ステータスが入る（HTTP 由来でない場合は 0）。
`apiErr.Cause`は`errors.Unwrap`で取り出せますが、transport由来のURLにAppIDが
含まれることがあります。CauseやConfigをそのままログへ出さないでください。

---

## In-memory キャッシュ（任意）

同じ法人番号を短時間に何度も引くユースケース（CRM の画面で何度も同じ法人を参照する等）で
API 呼び出しを減らしたい場合、`MemoryCache` を `Config.Cache` に設定し、
`CachePolicyReadThrough` を指定する。

```go
import (
	"time"

	houjin "github.com/sibukixxx/company-catalog"
)

cache := houjin.NewMemoryCache(houjin.MemoryCacheOptions{
	TTL: 10 * time.Minute, // 0 を渡すと無期限
})

client, _ := houjin.NewClient(houjin.Config{
	AppID:       os.Getenv("HOUJIN_APP_ID"),
	Cache:       cache,
	CachePolicy: houjin.CachePolicyReadThrough,
})
```

挙動:

- `LookupByCorporateNumber` のヒット時のみキャッシュ参照（`SearchByName` / `FetchUpdatesByDateRange`
  は現状キャッシュしない）
- not_found はキャッシュしない（後から法人が登録された場合に古い結果を返さないため）
- `MemoryCache` は `sync.RWMutex` + `map` ベース。サイズ制限・LRU は持たない（プロセス内の
  短命キャッシュ用途）
- 永続化や複数プロセス共有が必要な場合は `CompanyCache` interface を自前実装して
  Redis 等を背後に置く

---

## テスト用 Mock

利用側のユースケーステストで実 API を叩かないために `MockClient` を使う。
`MockClient` は `Client` インターフェースを満たし、必要な関数だけ差し替えられる。

```go
mock := &houjin.MockClient{
	LookupByCorporateNumberFn: func(ctx context.Context, n string) (*houjin.CompanyProfile, error) {
		return &houjin.CompanyProfile{
			CorporateNumber: n,
			Name:            "テスト株式会社",
			Address: houjin.Address{
				PrefectureName: "東京都",
				CityName:       "千代田区",
			},
		}, nil
	},
}

// 本番コードと同じ Client 型として注入できる
var c houjin.Client = mock
```

設定していない関数を呼ぶと `ErrorCodeAPIError` が返る（テストで意図しない呼び出しを検知しやすい）。

---

