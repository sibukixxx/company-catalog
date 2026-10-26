# API 契約とテストの対応表

`make test` が緑なら、本ライブラリが README で公開している契約は仕様通りに動作する
ことが機械的に検証されている。本ドキュメントは「**どの契約**」が「**どのテスト**」で
証明されているかを 1:1 で示すマップである。

> 上流の国税庁 Web-API そのものに対する疎通は本表に含まない。
> それは `test/e2e/` 配下の `make test-e2e` (build tag `e2e`, 要 `HOUJIN_APP_ID`) で検証する。

## 1. Client 公開メソッドの基本動作

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `LookupByCorporateNumber` で 1 件取得・正しい URL/クエリ送出 (`/num`, `id`, `number`, `type=12`) | `TestLookupByCorporateNumberSendsExpectedRequest` | client_test.go |
| `LookupByCorporateNumber` 該当なし → `ErrorCodeNotFound` | `TestLookupByCorporateNumberReturnsNotFoundWhenNoCorporation` | client_test.go |
| `LookupByCorporateNumber` 入力不正 → `ErrorCodeInvalidInput` | `TestLookupByCorporateNumberValidatesInput` | client_test.go |
| `SearchByName` で正しい URL/クエリ送出 (`/name`, `address`) | `TestSearchByNameSendsExpectedRequest` | search_test.go |
| `SearchByName` 空 Name → `ErrorCodeInvalidInput` | `TestSearchByNameReturnsErrorOnEmptyName` | search_test.go |
| `SearchByName` 0 件は空スライスを返しエラーにしない | `TestSearchByNameReturnsEmptySliceOnNoResults` | search_test.go |
| `FetchUpdatesByDateRange` で正しい URL/クエリ送出 (`/diff`, `from`, `to`) | `TestFetchUpdatesByDateRangeSendsExpectedRequest` | search_test.go |
| `FetchUpdatesByDateRange` From/To 未指定 → `ErrorCodeInvalidInput` | `TestFetchUpdatesByDateRangeRequiresFromAndToDate` | search_test.go |

## 2. CompanyProfile / Address データモデル

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `CompanyProfile` の全フィールド (`Name`, `NameKana`, `NameEn`, `Kind`, `Address`, `Assignment/Update/ChangeDate`, `ChangeReason`, `CloseDate`, `CloseCause`, `Source`) が XML から正しくマップされる | `TestLookupByCorporateNumberMapsAllFieldsFromXML` | client_contract_test.go |
| `Address.Full()` が PrefectureName + CityName + StreetNumber を順に連結する | `TestAddressFullConcatenatesParts` | client_contract_test.go |
| `Address.Full()` の空文字スキップ・全空文字パターン | `TestAddressFullSkipsEmptyPartsWhenSomeFieldsMissing`, `TestAddressFullReturnsEmptyWhenAllFieldsEmpty`, `TestAddressFullReturnsConcatenatedAddressWhenAllFieldsSet` | address_test.go |
| `SourceInfo.Provider="houjin_bangou"`, `Version="v4"`, `FetchedAt` = 受信時刻 (UTC RFC3339) | 上記 `TestLookupByCorporateNumberMapsAllFieldsFromXML` 内 | client_contract_test.go |
| 分割応答では各ページの受信時刻を `FetchedAt` に記録する | `TestFetchUpdatesByDateRangeStampsFetchedAtPerPageInUTC` | pagination_test.go |

## 3. NameMatchMode → API クエリ "mode" マッピング

| README 契約 | 検証テスト |
|------------|-----------|
| `NameMatchModeDefault` → mode を送らない | `TestSearchByNameMapsMatchModeToQueryParam/Default は mode を送らない` |
| `NameMatchModeExact` → `mode=1` | `TestSearchByNameMapsMatchModeToQueryParam/Exact → mode=1` |
| `NameMatchModePartial` → `mode=2` | `TestSearchByNameMapsMatchModeToQueryParam/Partial → mode=2` |

すべて search_test.go の `TestSearchByNameMapsMatchModeToQueryParam` 配下のサブテスト。

## 4. Limit / CompanyKind の挙動

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `SearchByName` の `Limit` が結果を先頭 N 件に切り詰める | `TestSearchByNameTruncatesResultsToLimit` | search_test.go |
| `SearchByName` の `Limit=0` で全件返す | `TestSearchByNameReturnsAllResultsWhenLimitIsZero` | search_test.go |
| `FetchUpdatesByDateRange` の `Limit` が結果を切り詰め、`CompanyKind` が `kind` クエリに乗る | `TestFetchUpdatesByDateRangeAppliesLimitAndCompanyKind` | search_test.go |

## 5. APIError 構造化エラー（README「エラーハンドリング」表）

| HTTP / 条件 | 期待 ErrorCode | 検証テスト |
|------------|--------------|----------|
| `400 Bad Request` | `ErrorCodeInvalidInput` | `TestLookupByCorporateNumberMapsHTTPStatusToErrorCode/400 …` |
| `401 Unauthorized` | `ErrorCodeUnauthorized` | 同上 `/401 …` |
| `403 Forbidden` | `ErrorCodeUnauthorized` | 同上 `/403 …` |
| `404 Not Found` | `ErrorCodeNotFound` | 同上 `/404 …` |
| `429 Too Many Requests` | `ErrorCodeRateLimited` | 同上 `/429 …` |
| `500 / 503` その他 5xx | `ErrorCodeAPIError` | 同上 `/500 …`, `/503 …` |
| `context.Canceled` | `ErrorCodeTimeout` | `TestLookupByCorporateNumberReturnsTimeoutWhenContextCanceled` |
| `context.DeadlineExceeded` | `ErrorCodeTimeout` | `TestLookupByCorporateNumberReturnsTimeoutWhenContextDeadlineExceeded` |
| 不正な XML レスポンス | `ErrorCodeParseError`, `Cause` 設定済み | `TestLookupByCorporateNumberReturnsParseErrorOnMalformedXML` |
| 不正な法人番号文字列 (length / 非数字 / wrong check digit) | `ErrorCodeInvalidInput` | `TestValidateCorporateNumberReturnsAPIErrorWithInvalidInputCode` |

すべて client_contract_test.go / corporate_number_test.go。

| README 契約 | 検証テスト |
|------------|-----------|
| `apiErr.StatusCode` に HTTP ステータスが入る (HTTP 由来時) | `TestLookupByCorporateNumberMapsHTTPStatusToErrorCode` 内 `wantStatus` チェック |
| `apiErr.Code` 7 つの定数値 (`invalid_input`, `unauthorized`, `not_found`, `rate_limited`, `timeout`, `parse_error`, `api_error`) | `TestErrorCodeConstantsAreExposed` |
| `errors.As(err, &apiErr)` で型判定可能 | `TestAPIErrorWorksWithErrorsAs` (errors_test.go) |
| `errors.Unwrap(apiErr)` で `Cause` を取り出せる | `TestAPIErrorUnwrapReturnsCause` (errors_test.go) |
| `apiErr.Error()` が code と message を含む | `TestAPIErrorErrorMessageContainsCodeAndMessage` (errors_test.go) |

## 6. セキュリティ契約（README「セキュリティ・運用上の注意」）

| 契約 | 検証テスト |
|------|----------|
| `APIError.Error()` に AppID（Secret）が漏れない（HTTP 4xx / 5xx / parse error の全エラー経路） | `TestAPIErrorDoesNotLeakAppIDOnAnyErrorPath` (client_contract_test.go) |
| `LookupByCorporateNumber` 内部でも AppID が error メッセージに混ざらない | `TestLookupByCorporateNumberDoesNotLogAppID` (client_test.go) |

## 7. Config 契約

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| AppID 未設定の `NewClient` はエラー | `TestNewClientReturnsErrorWhenAppIDIsEmpty` / `TestConfigValidateReturnsErrorWhenAppIDIsEmpty` | client_test.go / config_test.go |
| `BaseURL` 省略時はデフォルト `https://api.houjin-bangou.nta.go.jp/4` | `TestConfigWithDefaultsFillsBaseURLWhenEmpty` | config_test.go |
| `BaseURL` 明示時は上書きしない | `TestConfigWithDefaultsKeepsExplicitBaseURL` | config_test.go |
| `Timeout` 省略時はポジティブな値、明示時は保持 | `TestConfigWithDefaultsSetsTimeoutAndRetries`, `TestConfigWithDefaultsKeepsCustomTimeout` | config_test.go |
| `MaxRetries` 省略時 ≥1 | `TestConfigWithDefaultsSetsTimeoutAndRetries` | config_test.go |
| `HTTPClient` nil 時に内部生成 | `TestConfigWithDefaultsCreatesHTTPClientWhenNil` | config_test.go |

## 8. ValidateCorporateNumber / 13 桁チェックディジット

| README 契約 | 検証テスト |
|------------|-----------|
| 正しい 13 桁番号は通る (Toyota: `1180301018771`) | `TestValidateCorporateNumberReturnsNilWhenValid` |
| 13 桁でない長さは弾く (空, 短い, 長い) | `TestValidateCorporateNumberReturnsErrorWhenLengthIsNot13` |
| 数字以外を含むものは弾く | `TestValidateCorporateNumberReturnsErrorWhenContainsNonDigit` |
| チェックディジットが合わないものは弾く | `TestValidateCorporateNumberReturnsErrorWhenCheckDigitIsWrong` |
| 失敗時は `*APIError{Code: ErrorCodeInvalidInput}` | `TestValidateCorporateNumberReturnsAPIErrorWithInvalidInputCode` |
| **アルゴリズムの round-trip 性質**: `computeCheckDigit(base)` で計算した値を base の先頭に付けると `ValidateCorporateNumber` が必ず通る（property-based） | `TestValidateCorporateNumberAcceptsAllNumbersBuiltByComputeCheckDigit` |

すべて corporate_number_test.go。

## 9. MockClient

| README 契約 | 検証テスト |
|------------|-----------|
| `MockClient` は `Client` interface を実装する | `TestMockClientImplementsClientInterface` |
| 設定した関数フィールド経由で値が返る | `TestMockClientReturnsConfiguredResponses` (Lookup), `TestMockClientSearchByNamePassesInputThrough`, `TestMockClientFetchUpdatesByDateRangePassesInputThrough` |
| 未設定関数を呼ぶと `*APIError(ErrorCodeAPIError)` が返る (3 メソッドすべて) | `TestMockClientReturnsAPIErrorWhenFunctionNotSet` のサブテスト |

すべて mock_test.go。

## 10. internal/xmlapi の境界

| 契約 | 検証 |
|------|------|
| `internal/xmlapi.Corporation` が外部 package に漏れない | `internal/` 配下は import パス上 root package と adapters/ から見える。client.go から `xmlapi` 経由で呼んでも、`Corporation` は `corporationToProfile` で `CompanyProfile` に変換されてから返るため公開 API には現れない |
| XML パーサ単体は 100% カバレッジ | `internal/xmlapi/parser_test.go`、`go test -cover` で確認 |

## 11. Downstream sales contract

Sales eligibility and list construction moved to private `ja-company-outreach/internal/saleslead`, including the former regression tests. The real exporter and Source Package consumer were tested against company-catalog.

## 12. 分割応答の自動結合（Phase 4）

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `xmlapi.Parse` が `divideNumber` / `divideSize` を Result として返す | `TestParseReturnsDivideMetadataFromXMLWhenPaginated` | internal/xmlapi/parser_test.go |
| 単一応答 (`divide` 属性なし) は `DivideSize=1, DivideNumber=1` に正規化される | `TestParseReturnsDivideSizeOneForNonPaginatedResponse` | internal/xmlapi/parser_test.go |
| `FetchUpdatesByDateRange` で分割応答を全ページ取得・結合する | `TestFetchUpdatesByDateRangeAggregatesAllPagesWhenResponseIsDivided` | pagination_test.go |
| 2 ページ目以降は `divide=2..N` クエリパラメータで送られる | 同上 (期待 query 列の検証) | pagination_test.go |
| `SearchByName` の `Limit` は分割応答結合後の全件に対して適用される | `TestSearchByNameAggregatesAllPagesAndAppliesLimitAfterAggregation` | pagination_test.go |
| 異常な `divideSize` (>20) に対する安全装置で取得を打ち切り `ErrorCodeAPIError` を返す | `TestFetchProfilesStopsAtMaxDivideToAvoidInfiniteLoop` | pagination_test.go |

## 13. MemoryCache (Phase 4)

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `NewMemoryCache` が `CompanyCache` インターフェースを実装する | `TestMemoryCacheImplementsCompanyCacheInterface` | cache_inmemory_test.go |
| `SetByCorporateNumber` で保存した値が `GetByCorporateNumber` で返る | `TestMemoryCacheReturnsStoredCompanyByCorporateNumber` | cache_inmemory_test.go |
| 未保存番号は `(nil, false, nil)` を返す | `TestMemoryCacheReturnsMissWhenCorporateNumberIsNotStored` | cache_inmemory_test.go |
| `TTL` 経過後はミス扱い | `TestMemoryCacheExpiresEntryAfterTTL` | cache_inmemory_test.go |
| `TTL=0` は無期限 | `TestMemoryCacheTTLZeroMeansNoExpiration` | cache_inmemory_test.go |
| 並行アクセスで race しない (`-race`) | `TestMemoryCacheIsSafeForConcurrentUse` | cache_inmemory_test.go |

## 14. Cache 統合 (Lookup x ReadThrough)

| README 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `Cache` 設定 + `CachePolicyReadThrough` のとき、2 回目の `LookupByCorporateNumber` は API を叩かずキャッシュから返す | `TestLookupByCorporateNumberServesFromCacheOnSecondCallWhenReadThrough` | cache_integration_test.go |
| `CachePolicyDisabled` のときは Cache 設定があってもバイパスして毎回 API を叩く | `TestLookupByCorporateNumberBypassesCacheWhenPolicyDisabled` | cache_integration_test.go |
| Cache 未設定 (`Config.Cache == nil`) + `CachePolicyReadThrough` でも panic せず動作する | `TestLookupByCorporateNumberWorksWithoutCacheEvenWhenPolicyReadThrough` | cache_integration_test.go |
| `not_found` 応答は cache に書き込まれない (negative caching を避ける) | `TestLookupByCorporateNumberDoesNotCacheNotFoundResults` | cache_integration_test.go |

## 15. Analysis export contract (`techvit.analysis.company/v1`)

契約本文は [ANALYSIS_EXPORT_CONTRACT.md](ANALYSIS_EXPORT_CONTRACT.md)。

| 契約 | 検証テスト | ファイル |
|------------|-----------|---------|
| `ProjectForAnalysis` は識別子・住所名・provenance を保持し、地域コードを推論しない | `TestProjectForAnalysis` | analysis_test.go |
| 全レコードに `schema_version = AnalysisSchemaVersion` が入る | `TestProjectForAnalysisSetsSchemaVersion` | analysis_test.go |
| `event_type` の優先順 `CLOSED > CHANGED > UPDATED > ASSIGNED > UNKNOWN` | `TestAnalysisEventType` | analysis_test.go |
| CSV の列順が v1 の定義と一致し、provenance 列が出力される | `TestCSVAnalysisHeaderMatchesSchemaV1ColumnOrder` | analysis_test.go |
| JSONL のキー集合（`source.*` を含む）が v1 の定義と一致する | `TestJSONLAnalysisRecordKeysMatchSchemaV1` | analysis_test.go |
| manifest が schema / source / source_version / generated_at (UTC) を埋める | `TestNewAnalysisManifestFillsSchemaAndSourceProvenance` | analysis_manifest_test.go |
| manifest は未指定の query 条件を出力しない | `TestWriteAnalysisManifestOmitsEmptyQueryFilters` | analysis_manifest_test.go |
| `docs/examples/analysis-manifest.example.json` が契約と一致する | `TestAnalysisManifestExampleMatchesContract` | analysis_manifest_test.go |
| `company-catalog -manifest` が record_count / query を記録し、AppID を含めない | `TestRunWritesManifestWithProvenanceAndWithoutAppID` | cmd/company-catalog/main_test.go |
| `-manifest` は `-` や `-output` と同じパスを拒否する | `TestParseOptionsRejectsInvalidFormatLimitAndManifestPath` | cmd/company-catalog/main_test.go |

## カバレッジ指標（参考）

```
$ make test-cover
github.com/sibukixxx/company-catalog                     coverage: 95.4%
github.com/sibukixxx/company-catalog/internal/xmlapi     coverage: 100.0%
```

カバーされていない残り 4.6%（client.go の `do` の retry/HTTP 構築失敗パス等）は、
HTTP クライアントの内部失敗を再現するため固有の test double を要する領域で、
本ライブラリの公開 API 契約の検証としては付加価値が低い。

## 検証コマンド

```bash
# 全契約テストを実行
make test

# カバレッジ確認
make test-cover

# 上流 API への E2E（要 HOUJIN_APP_ID）
make test-e2e
```

## メンテナンスポリシー

- README に新しい公開 API・新しいエラーコード・新しい挙動を追加した場合は、
  本表に行を追加してから（または同時に）テストを書く
- `*_test.go` のテスト関数名を変更したら本表のテスト名カラムも更新する
- 上流 API 仕様が変わって解釈が変わる場合は、`internal/xmlapi/parser_test.go` と
  `client_contract_test.go` の `TestLookupByCorporateNumberMapsAllFieldsFromXML` を
  優先的に更新する

## OSS source-package and example contract coverage

| Public behavior | Verification |
| --- | --- |
| identity independent of row order and acquisition/export time | TestSourcePackageIdentity |
| source/provenance/schema/hash/count validation before output/use | TestSourcePackageValidation |
| observed zero distinct from failed/limited acquisition and unknown fields | TestSourcePackageEmptyAndLimited |
| synthetic documentation fixture equals current producer output | TestSourcePackageFixture |
| JSON source-package CLI excludes AppID | TestRunWritesValidatedSourcePackageWithoutCredential |
| file-only public consumer retains source package; rejects schema mismatch | TestFileOnlyHandoff; scripts/verify_public_contract.py |
| offline reader runs with no AppID | TestReadFixtureOffline; ExampleReadCompanySourcePackage |
| offline producer repeats fixture bytes | TestRebuildFixtureDeterministically |
| malformed suffix / incomplete records / reported counts rejected | TestParseRejectsMalformedSuffixAndIncompleteRecords; TestParseCountAndElementPagination; TestFetchRejectsIncompleteAggregate |

Config.MaxRetries is reserved, not an implemented retry guarantee. Public client
methods can return ordinary pagination-integrity errors; Source Package validation
also uses ordinary errors. APIError.Error hides its transport Cause, while Unwrap
may expose a credential-bearing URL. See SECURITY.md and package GoDoc.
