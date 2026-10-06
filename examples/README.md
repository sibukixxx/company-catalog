# Examples — 最初はオフラインでcontractを試す

以下のコマンドはrepo rootから実行します。Go requirementはgo.modに従います。
Go toolchainの初回downloadを除き、offline例はAPI・network・AppIDを使いません。
SDKの型を確認したい場合は先頭2本を実行し、live APIやdownstream接続は必要な例を選んでください。

| 分類 / Example | 学べること | AppID | 実行時network | expected output |
| --- | --- | --- | --- | --- |
| Offline: `read_source_package` | schema/hash/provenanceを検証して読む | 不要 | 不要 | package ID、時刻、2 records、合成法人名 |
| Producer: `export_source_package` | public constructor/writerで既存fixtureを再出力 | 不要 | 不要 | Source Package JSON（fixtureと同じbytes） |
| Downstream: `insight_handoff` |公開EvidenceDocumentへprovenanceごと渡す | 不要 | 不要 | contractVersion=1のAddEvidenceRequest JSON |
| Live: `lookup_by_number` | 13桁番号で1件取得 | 必要 | 必要 | 法人番号、名称、住所、更新日 |
| Live: `search_by_name` | upstream name matchingと地域filter | 必要 | 必要 | 候補件数・番号・名称・住所 |
| Live: `fetch_diff` | 指定期間のupstream更新情報を読む | 必要 | 必要 | 期間、件数、先頭最大5件 |
| Live export: `cmd/company-catalog` | CSV/JSONL/Source Packageとquery metadata | 必要 | 必要 | 指定形式のfile/標準出力 |
| Optional legacy: `build_sales_list` | 既存SalesLead adapterとの互換利用例 | 必要 | 必要 | filtered SalesProspect一覧 |

## Offline / producer

```sh
go run ./examples/read_source_package docs/examples/company-source-package.v1.json
go run ./examples/export_source_package docs/examples/company-source-package.v1.json > /tmp/company-package.json
go run ./examples/read_source_package /tmp/company-package.json
```

fixtureの固定取得/生成時刻は2026-01-01T00:00:00Z、providerはhoujin_bangou v4、件数は2です。
末尾JSON・version不一致・hash/count不整合・provenance欠落はerrorになります。
fixtureは合成データで、実在法人の証拠ではありません。入力とは別のfileへ書き出してください。

## Public downstream contract

```sh
go run ./examples/insight_handoff docs/examples/company-source-package.v1.json
python -m pip install jsonschema
python scripts/verify_public_contract.py
```

handoff自身はfile-onlyです。schema validatorの初回installにはnetworkが必要です。
requestのdocuments[0].sourceはdatasetで、contentにSource Package全体を保持します。
schema検証の成功表示は`Public Engine v1 schema: PASS`です。
HTTP送信・Insight DB操作・Research実行は行いません。受信engineでの実行検証はconsumer側の作業です。

## Live NTA API

自分のAppIDを環境変数HOUJIN_APP_IDへ安全に設定してから実行します。
以下の実在法人は1件の公開API利用例です。結果のsnapshotは固定していません。

```sh
go run ./examples/lookup_by_number 1180301018771
go run ./examples/search_by_name トヨタ自動車 23
go run ./examples/fetch_diff 1 13
```

lookupは引数に法人番号、searchは名称と任意のprefecture code、fetch_diffは日数と任意のprefecture codeを受け取ります。
広い期間・地域は20ページの安全上限に達する場合があります。件数制限は取得後に適用され、API負荷を減らす保証ではありません。
取得情報・query・住所を共有するときは必要な範囲だけに絞り、AppIDやraw URLを公開しないでください。

## Export

```sh
go run ./cmd/company-catalog -from 2026-09-01 -to 2026-09-30 \
  -prefecture 13 -format json -output /tmp/companies.json
go run ./cmd/company-catalog -from 2026-09-01 -to 2026-09-30 \
  -prefecture 13 -format jsonl -output /tmp/companies.jsonl \
  -manifest /tmp/companies.manifest.json
```

JSONはmanifest内包のSource Package、JSONLは1行1AnalysisRecordです。CSVは`-format csv`に切り替えます。
query/date・record count・completenessをconsumerへ渡し、LIMITEDを人口全体とみなさないでください。

## Optional legacy SalesLead example

```sh
go run ./examples/build_sales_list example-case トヨタ 23
```

`adapters/saleslead`とこの例は既存consumerを壊さないために保持しています（Option B）。
canonical SDKの責務は法人データの取得・正規化・出力です。新しい営業判断・workflowはoutreach/Sales側に置きます。
この例は送信を行いません。SDKが連絡先や営業の適格性を保証することもありません。

## Fixturesの配置

| 配置 | 用途 |
| --- | --- |
| `docs/examples/` | public contractの小さな表示例。Source Package fixtureはtestとExamplesも共有 |
| `testdata/` | parser/client用fixture。実API seedは`testdata/seed/`へ保存しgitignore |
| Go test内のXML literals | HTTP/pagination/parserの再現入力。test AppIDのみ使用 |
| `docs/contracts/` | unchanged third-party public schemaとそのsource/license |

同じSource Package fixtureを別名で複製せず再利用します。合成fixtureは取得済みgolden APIデータの代用品ではありません。
エラー報告・質問はroot READMEのIssues/SECURITY案内へ進んでください。

analysis-manifest.example.jsonの件数1234も表示用の例で、実データ件数の主張ではありません。
