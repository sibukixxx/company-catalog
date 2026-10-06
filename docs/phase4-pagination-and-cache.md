> Historical design/test notes for the pagination/cache implementation.
> Counts and coverage below describe that implementation stage, not a current
> release guarantee. Current behavior and non-goals are defined in README.md,
> SDK_REFERENCE.md, SOURCE_PACKAGE_CONTRACT.md and package GoDoc.

# Phase 4: 分割応答対応 + In-memory Cache 実装ログ

ロードマップ Phase 4 の前半 2 機能（**分割応答の自動結合** + **MemoryCache**）を
TDD で追加した際の作業記録。同じ手順で他の Phase（Retry backoff 等）を進める際の
リファレンスとして残す。

- **作業日**: 2026-05-17
- **対象**: `houjin` package（ルート）+ `internal/xmlapi`
- **方針**: t-wada 流 TDD（Red → Green → Refactor）、Tidy First で構造変更と動作変更を分離
- **テスト**: ユニット 13 本追加 + E2E 2 本追加。実 API 検証は手元で `make test-e2e`

---

## 0. 前提

### 必要なもの

| 項目 | 用途 |
|---|---|
| `HOUJIN_APP_ID` | E2E + `make seed` の実行に必須。Secret 管理 |
| Go 1.26.2 以上 | `go.mod` 準拠 |
| `make setup` 実行済み | golangci-lint v2.11.4 + gofumpt |

### スコープ確定の判断材料

ロードマップ Phase 4 は 3 機能（分割応答 / Cache / Retry backoff）。今回は前 2 つだけ採用。

| 候補 | 採用 | 理由 |
|---|---|---|
| 分割応答自動取得 | ✅ | README で外向きに「Phase 4 で対応予定」と明示済み。実害最大 |
| MemoryCache 実装 | ✅ | interface だけ用意済みの cache.go を埋める。Lookup ホットパスで効果大 |
| Retry backoff | ❌ | 429 を意図的に再現する E2E が国税庁規約的にグレー。別 PR |

---

## 1. 設計判断とトレードオフ

### 1.1 分割応答

**国税庁 API の仕様**:
- `<corporations divideNumber="N" divideSize="T">` で分割される
- `divideNumber=1` が初回応答。2 ページ目以降は `divide=N` クエリで再リクエスト
- 単一応答は `divideNumber=1, divideSize=1`

**設計判断**:

| 論点 | 採用案 | 却下案 |
|---|---|---|
| `xmlapi.Parse` のシグネチャ | 破壊的に `(Result, error)` に変更 | `ParseWithMeta` を新設して既存維持 |
| 取得上限 | 定数 `maxDividePages=20` でハードコード | `Config.MaxDividePages` で外部化 |
| 異常応答の扱い | `ErrorCodeAPIError` で打ち切り | 取得済みページだけ返す |

**理由**:
- `xmlapi` は internal package なので破壊変更可。新関数 = 2 つの呼び出しパスができて保守負債
- 取得上限は SDK の安全装置で、利用者が増やしたくなる状況はバグ。外部化は YAGNI
- 中途半端な結果を返すと利用側で気づけない。明示エラーが安全

### 1.2 MemoryCache

**設計判断**:

| 論点 | 採用案 | 却下案 |
|---|---|---|
| 配置 | ルート package の `cache_inmemory.go` | `internal/cache/inmemory/` サブパッケージ |
| 期限管理 | TTL のみ（LRU/サイズ制限なし） | LRU + 最大件数 |
| Search/Diff 結果のキャッシュ | しない（Lookup のみ） | 全 3 メソッドでキャッシュ |
| `not_found` のキャッシュ | しない | TTL 短めで保存 |

**理由**:
- リポジトリのGoライブラリ規約では、ルート package を公開 API、internal を実装詳細として扱う。
  `CompanyCache` interface の実装は公開 API 扱い
- LRU は YAGNI。プロセス内短命キャッシュ前提で十分。永続化や上限が必要なら
  `CompanyCache` interface を自前実装すれば良い
- Search/Diff はキー設計（normalized name、日付範囲）が Phase 3 と絡む。早い実装は無駄
- negative caching は「後から法人が登録されると古い not_found を返す」事故源。回避

---

## 2. 実装手順（TDD 13 ステップ）

各ステップは「Red 失敗確認 → Green 最小実装 → 既存テストが緑」のサイクル。

### Step 1-2: xmlapi.Parse の divide メタデータ対応

**Red**: `internal/xmlapi/parser_test.go` に追加
```go
func TestParseReturnsDivideMetadataFromXMLWhenPaginated(t *testing.T) { ... }
func TestParseReturnsDivideSizeOneForNonPaginatedResponse(t *testing.T) { ... }
```
コンパイルエラー = Red（`result.DivideSize` 未定義）。

**Green**: `parser.go`
```go
type Result struct {
    Corporations []Corporation
    DivideNumber int
    DivideSize   int
}

func Parse(r io.Reader) (Result, error) {
    var doc corporations
    ...
    return Result{
        Corporations: doc.Corporations,
        DivideNumber: max(doc.DivideNumber, 1),
        DivideSize:   max(doc.DivideSize, 1),
    }, nil
}
```
呼び出し側 `client.go` の `fetchProfiles` を `result.Corporations` 経由に追従。

検証: `go test -race -vet=all ./internal/xmlapi/...`

### Step 3-4: 分割応答自動結合

**Red**: `pagination_test.go` を新規作成
- `TestFetchUpdatesByDateRangeAggregatesAllPagesWhenResponseIsDivided`
  - httptest で 3 ページ応答を返し、`divide=` クエリの順序を検証
- `TestSearchByNameAggregatesAllPagesAndAppliesLimitAfterAggregation`
  - 結合後に `Limit` 適用される順序
- `TestFetchProfilesStopsAtMaxDivideToAvoidInfiniteLoop`
  - `divideSize=100` を返し続けるサーバへの安全装置

**Green**: `client.go` の `fetchProfiles` を書き換え
```go
const maxDividePages = 20

func (c *httpClient) fetchProfiles(ctx, path, q) ([]CompanyProfile, error) {
    first, err := c.fetchPage(ctx, path, q, 0)
    if first.DivideSize <= 1 { return ... }
    if first.DivideSize > maxDividePages { return nil, &APIError{...} }
    for page := 2; page <= first.DivideSize; page++ {
        next, err := c.fetchPage(ctx, path, q, page)
        ...
    }
}

func (c *httpClient) fetchPage(ctx, path, base, page) (xmlapi.Result, error) {
    q := base
    if page >= 2 {
        q = cloneValues(base)
        q.Set("divide", strconv.Itoa(page))
    }
    ...
}
```

`url.Values` は map なのでクローン必須（共有すると別ページのリクエストで `divide` が上書きされる）。

### Step 5-6: MemoryCache 実装

**Red**: `cache_inmemory_test.go`
- `Returns` / `Miss` / `Expires` / `TTLZero` / `Concurrent` / `ImplementsInterface`

**Green**: `cache_inmemory.go`
```go
func NewMemoryCache(opts MemoryCacheOptions) *MemoryCache { ... }

type MemoryCacheOptions struct {
    TTL time.Duration
    Now func() time.Time // テスト差し替え用
}
```

`Now` を opts に持たせるのは時計依存テストのため。default は `time.Now`。

### Step 7-8: Lookup の cache read-through 統合

**Red**: `cache_integration_test.go`
- `ServesFromCacheOnSecondCall` — 2 回目は HTTP 呼ばれない（`atomic.Int32` でカウント）
- `BypassesCacheWhenPolicyDisabled` — disabled 時は毎回 HTTP
- `WorksWithoutCacheEvenWhenPolicyReadThrough` — Cache==nil でも panic しない
- `DoesNotCacheNotFoundResults` — negative caching しない

**Green**: `client.go` の `LookupByCorporateNumber`
```go
if c.cacheEnabled() {
    if hit, ok, err := c.cfg.Cache.GetByCorporateNumber(ctx, n); err == nil && ok && hit != nil {
        return hit, nil
    }
}
// ... API 呼び出し ...
if c.cacheEnabled() {
    _ = c.cfg.Cache.SetByCorporateNumber(ctx, profiles[0]) // Set 失敗は応答に影響させない
}

func (c *httpClient) cacheEnabled() bool {
    return c.cfg.Cache != nil && c.cfg.CachePolicy == CachePolicyReadThrough
}
```

### Step 9: Refactor

今回は不要だった。`fetchPage` / `profilesFromCorporations` / `cloneValues` で重複は解消済み。`make fmt` / `make vet` のみ通す。

### Step 10: Docs 更新

- `README.md`: 「3. 取得期間指定で差分取得」の注記、「In-memory キャッシュ」セクション追加、ロードマップ Phase 4 を 🚧 進行中
- `docs/API_CONTRACT_TESTS.md`: §12 / §13 / §14 を追加

### Step 11-12: E2E テスト追加

`test/e2e/` 配下、`//go:build e2e` 隔離。

| ファイル | テスト | 検証内容 |
|---|---|---|
| `pagination_e2e_test.go` | `TestE2EFetchUpdatesByDateRangeAggregatesDividedPages` | 過去 7 日間の `/diff` で全結合 + Corporation の sanity check |
| `cache_e2e_test.go` | `TestE2EMemoryCacheHitAvoidsSecondAPICall` | `http.RoundTripper` をラップして HTTP 回数カウント、2 回目が 1 回のままを検証 |

**重要**: `skipIfNotEnabled` ヘルパーで二重ガード（`HOUJIN_E2E_TEST_ENABLED=true` + `HOUJIN_APP_ID`）。

### Step 13: ユーザー検証

```bash
HOUJIN_E2E_TEST_ENABLED=true HOUJIN_APP_ID=xxx make test-e2e
```

---

## 3. 追加した契約とテストの 1:1 対応

`docs/API_CONTRACT_TESTS.md` の §12 / §13 / §14 を参照。

| 契約区分 | 件数 | ファイル |
|---|---|---|
| 分割応答 (xmlapi) | 2 | `internal/xmlapi/parser_test.go` |
| 分割応答 (client) | 3 | `pagination_test.go` |
| MemoryCache | 6 | `cache_inmemory_test.go` |
| Cache 統合 | 4 | `cache_integration_test.go` |
| E2E pagination | 1 | `test/e2e/pagination_e2e_test.go` |
| E2E cache hit | 1 | `test/e2e/cache_e2e_test.go` |
| **合計** | **17** | |

---

## 4. 検証コマンド

### ユニットテスト（実 API 不要）

```bash
make test                              # 全ユニットテスト
make test TEST_RUN=TestMemoryCache     # MemoryCache だけ
make test TEST_RUN=TestFetchUpdates    # FetchUpdates 系
make test-cover                        # カバレッジ確認
```

期待カバレッジ:
- `cache_inmemory.go` 100% (SearchByNormalizedName の no-op を除く)
- `internal/xmlapi` 100%
- `client.go` の新 fetchPage / cacheEnabled / profilesFromCorporations / cloneValues 100%

### E2E（実 API、`HOUJIN_APP_ID` 必須）

```bash
# 全 E2E
HOUJIN_E2E_TEST_ENABLED=true HOUJIN_APP_ID=xxx make test-e2e

# 個別
HOUJIN_E2E_TEST_ENABLED=true HOUJIN_APP_ID=xxx \
    make test-e2e E2E_RUN=TestE2EFetchUpdatesByDateRangeAggregatesDividedPages

HOUJIN_E2E_TEST_ENABLED=true HOUJIN_APP_ID=xxx \
    make test-e2e E2E_RUN=TestE2EMemoryCacheHitAvoidsSecondAPICall
```

### Lint

```bash
make fmt && make vet && make lint
```

**注意**: 既存 7 件の lint warning（`errors_test.go`, `examples/*/main.go`, `corporate_number.go`）が origin/main に残っている。今回追加分は clean。別 PR で対処予定。

---

## 5. 同じ手順を別 Phase でやる場合のテンプレート

1. **スコープ確定**: ロードマップから 1〜2 機能を選ぶ。YAGNI。1 PR に詰め込まない
2. **設計判断の表を書く**: 「論点 / 採用案 / 却下案 / 理由」を明文化
3. **タスク分割**: Red → Green を機能ごとに 4〜8 タスクに分ける
4. **Red ファースト**: コンパイルエラーも Red。実装側を 1 行も書かない
5. **Green は最小**: 必要最小限。`if foo == nil { return }` で済むなら helper を抽出しない
6. **既存テストの後方互換確認**: 既存呼び出し側の追従だけで通すか、新関数を生やすかは早めに決める
7. **Refactor は緑のとき**: テストに触らず、構造変更だけ
8. **Docs を同 PR で**: README + 契約表 + プレイブック
9. **E2E は最後**: ガード（`//go:build e2e` + `skipIfNotEnabled`）を必ず通す
10. **ユーザーに E2E を依頼**: AppID 経路を分離

---

## 6. 既知の課題（次の作業候補）

| 項目 | 優先度 | 備考 |
|---|---|---|
| 既存 lint warning 7 件の修正 | 中 | `errorlint` / `gocritic exitAfterDefer` / `intrange` / `perfsprint`。Tidy First で構造変更コミットとして別 PR |
| Retry backoff (Phase 4 残り) | 中 | `Config.MaxRetries` が宣言だけで未使用。429/503 で exponential backoff |
| Search/Diff の cache 拡張 | 低 | Phase 3 法人名正規化と合わせて設計 |
| `Now` をテスト用 build tag で公開する仕組み | 低 | 今は `MemoryCacheOptions.Now` で露出している。internal にできるか検討 |
| MemoryCache の `Delete` / `Clear` メソッド | 低 | YAGNI。要望が出てから |

---

## 7. 関連ドキュメント

- [README.md](../README.md) — SDK 公開仕様
- [docs/API_CONTRACT_TESTS.md](API_CONTRACT_TESTS.md) — 契約とテストの 1:1 対応表
- [AGENTS.md](../AGENTS.md) — リポジトリの開発・検証規約
