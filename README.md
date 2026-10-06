# company-catalog

<!-- oss-readme-intro:v1 -->
国税庁の法人番号Web-APIから日本企業データを取得し、**型付き・出典付き・再利用可能なデータ**として扱うGo SDK / CLIです。

### こんな人向け

- 法人番号や法人名から企業情報を取得したい
- CRM・営業・分析へ渡す前に、企業データを正規化して出典を残したい
- CSV / JSONL / versioned JSON packageとして再利用したい
- 実APIを呼ばずにfixtureやmockで組み込みを試したい

### できること

法人番号・法人名・更新期間での取得、pagination、XML解析、取得時刻とprovenanceの保持、Source Package v1 / CSV / JSONL export、offline fixture / mockによるcontract検証を提供します。CRM、営業判断、曖昧企業名matchingや企業情報の自動補完はこのツールの責務ではありません。

### 最短で試す

```sh
git clone https://github.com/sibukixxx/company-catalog.git
cd company-catalog
go run ./examples/read_source_package docs/examples/company-source-package.v1.json
```

実APIを使う場合は `HOUJIN_APP_ID` を環境変数に設定します。SDKとして使う場合は現在の公開tagを指定してください。

```sh
go get github.com/sibukixxx/company-catalog@v0.1.0
```

### 使いどころ

「公的な企業データを取得して、そのまま業務ロジックへ混ぜず、**出典付きの入力データとして他システムへ渡したい**」ときに使います。

詳しいAPI、export contract、security、release手順は [SDK reference](docs/SDK_REFERENCE.md)、[Source Package contract](docs/SOURCE_PACKAGE_CONTRACT.md)、[Documentation](docs/) を参照してください。
<!-- /oss-readme-intro -->



<!-- ecosystem-expansion:v1 -->

## Expansion contract

This repository is one capability owner in the TechVit ecosystem. New ideas should extend the existing graph before creating another silo.

**Canonical responsibility:** 日本企業データのcanonical producer/export foundation

When new work appears:

1. **Extend here** when the new behavior belongs to this repository's canonical responsibility.
2. **Integrate, don't copy** when another repository owns the concept. Exchange versioned API/artifact/event references; do not share databases or import another product's internals.
3. **Promote reusable capability only after real reuse appears.** Keep the first vertical slice close to its product; extract a common core when there are at least two real consumers or a deliberate public/private boundary.
4. **Create a new repository only** for a distinct product/vertical with its own lifecycle, deployment/security boundary, or OSS/commercial boundary. Its README must declare upstream/downstream owners from day one.
5. **Experiments may end without losing the asset.** Move validated rules, contracts, fixtures, and learnings into the canonical owner, then freeze/archive the duplicate implementation.

**Route new work:** 新しい企業属性・出典・export contractはここへ追加し、営業判断や送信履歴はdownstreamへ渡す。

Cross-cutting owners: TechVit Admin = company operations; TechVit Platform = shared cloud/bootstrap/identity; Insight = research semantics; techvit-insight = managed research; Relayboard = work execution; AI Allowance Guard = privileged-action authority; EC Hub = commerce; AffiliateMediaOS/VideoForge = content production; ja-company-* = company-data/outreach producers; public/product sites = acquisition surfaces.

If ownership is unclear, implement the smallest vertical slice in the nearest existing owner first. Extract only when actual reuse proves the boundary.
<!-- /ecosystem-expansion:v1 -->
<!-- current-implementation:2026-10-06 -->
## Current implementation snapshot — 2026-10-06

- The implemented core is a typed Go client for Japan's NTA Corporate Number Web-API v4 with validation, pagination, XML parsing, structured errors and source timestamps.
- `CompanyProfile` data can be exported as validated Source Package v1 or as versioned CSV/JSONL analytical records with provenance and manifests.
- Offline synthetic fixtures, mock clients and contract examples allow the public API and export boundary to be verified without an AppID; live verification remains explicitly credential-gated.
- This repository is a **company-data producer**, not a CRM or enrichment product: fuzzy matching, website/revenue enrichment, outreach decisions and sales history remain downstream concerns.
- OSS preparation (MIT, security/contribution docs, release gate and examples) is present, but the current repository stays private while history/identity cleanup and the `company-catalog` migration are completed.
<!-- /current-implementation -->


国税庁の法人番号Web-API v4を、型付きのGo clientと、出典を追跡できるデータとして扱うSDKです。
法人番号・法人名・更新期間から取得した情報を`CompanyProfile`へ変換し、
任意のconsumerへSource Package v1やCSV/JSONLで渡せます。コードは[MIT](LICENSE)です。

XMLのパース、分割応答、構造化エラー、取得時刻の記録をSDKにまとめています。
AppIDがなくても、合成fixtureとMockClientでcontractを試せます。
最初は以下のオフライン手順を実行し、API接続が必要になったらlive手順へ進んでください。

## Scope — 取得・正規化・出力を担当します

NTA Corporate Number API → company-catalog → typed Company / versioned Source Package → 任意のconsumer

| SDKが担当すること | consumerが担当すること |
| --- | --- |
| NTA API client、validation、pagination | CRM、outreach、sales scoring、lead管理 |
| source fieldのprojection、provenance | billing、recommendation、business decision |
| typed contract、deterministic export | cross-source enrichment、generic Open Data ingestion |
| versioned Source Package、mock/test | InsightのResearch semantics |

正規化はupstream fieldを型へ写す範囲です。法人名の曖昧matching、業種・売上・websiteの補完は実装していません。
既存`adapters/saleslead`は互換性のため残すoptional downstream adapterです。新しい営業機能は追加しません。

## Quick start — credentialなしで試す

Goの必要versionは[go.mod](go.mod)に記載しています（現在Go 1.26.8以上）。
初回のtoolchain取得にはnetworkが必要になる場合があります。以下のExamples自体はnetworkを使いません。
Public化前のcloneには、このrepoへのアクセス権が必要です。

```sh
git clone https://github.com/sibukixxx/company-catalog.git
cd company-catalog
go run ./examples/read_source_package docs/examples/company-source-package.v1.json
```

schema、package ID、provider、取得/生成時刻、`records: 2 (COMPLETE)`と合成法人2件を表示します。
fixtureはcontract検証用です。実在法人や営業リストとして使用しないでください。

同じfixtureをpublic producer APIで再出力できます。

```sh
go run ./examples/export_source_package docs/examples/company-source-package.v1.json > /tmp/company-package.json
go run ./examples/read_source_package /tmp/company-package.json
```

入力と同じ固定時刻・recordsを使うため、出力bytesも再現できます。
[Examples一覧](examples/README.md)にはlive API、export、public consumer handoffの実行方法と出力を記載しています。

## Quick start — NTA APIを試す

live APIには、自分で取得したAppIDとnetwork接続が必要です。AppIDは付属しません。
申請・利用条件は[国税庁の案内](https://www.houjin-bangou.nta.go.jp/webapi/)を確認してください。
AppIDは環境変数へ安全に設定し、shell履歴・ログ・repoへ保存しないでください。

```sh
# HOUJIN_APP_ID is already set in the environment.
go run ./examples/lookup_by_number 1180301018771
go run ./examples/search_by_name トヨタ自動車 23
go run ./examples/fetch_diff 1 13
```

Lookupは法人番号・法人名・所在地・更新日を、Searchは候補一覧を、Diffは件数と先頭サンプルを表示します。
結果は取得時点によって変わります。Diffの更新markerは、会社設立や事業開始を意味するとは限りません。

自分のGo moduleから利用する場合:

```sh
# Public化前は認証済みprivate module設定が必要です。
go get github.com/sibukixxx/company-catalog@v0.1.0
```

`v0.1.0`は既存tagのcore APIを指します。このREADMEのSource Packageと新しいExamplesは現mainの機能です。
公開前の新version候補は[release方針](docs/RELEASING.md)を参照してください。

```go
import houjin "github.com/sibukixxx/company-catalog"

client, err := houjin.NewClient(houjin.Config{AppID: os.Getenv("HOUJIN_APP_ID")})
// Check err before calling client methods.
company, err := client.LookupByCorporateNumber(ctx, "1180301018771")
```

完全な実行例は`examples/lookup_by_number`、他の設定・error処理・cache・mockは[SDK reference](docs/SDK_REFERENCE.md)を参照してください。

## Data provenance — どこから、いつ取得したかを保持します

`CompanyProfile.Source`はprovider、API version、ページ受信時刻`FetchedAt`を保持します。
Source Packageは取得条件、取得完了時刻、出力時刻、records、件数、内容hash、definition/snapshot revisionをまとめます。
DBを参照せず、`ReadCompanySourcePackage`でversion・provenance・hash・件数を検証できます。

```sh
# Live export; HOUJIN_APP_ID is already set.
go run ./cmd/company-catalog -from 2026-09-01 -to 2026-09-30 \
  -prefecture 13 -format json -output companies.json
```

`-format csv`/`jsonl`は既存AnalysisRecord v1出力です。`-manifest`を指定するとmetadataを別JSONへ出します。
`-format json`はmanifest内包のSource Packageなので、別`-manifest`は指定しません。
`-limit`が正ならLIMITEDと表示し、全体の法人増減を判断する入力にはしません。

内容hashはrecord順序・取得時刻・出力時刻で変わりません。条件・source/schema・正規化versionの変更はdefinition revisionに反映します。
空のoptional値は取得できない情報として扱い、0や既知の空値へ補完しません。
AnalysisRecordの名称は既存API名であり、分析結果やResearch結論を出す機能ではありません。

## Contracts — fileとGo APIの境界

| 必要な情報 | 参照先 |
| --- | --- |
| Source Package v1、identity、time、missingness | [Source Package contract](docs/SOURCE_PACKAGE_CONTRACT.md) |
| CSV/JSONLの列・キーとmanifest | [Analysis export contract](docs/ANALYSIS_EXPORT_CONTRACT.md) |
| public APIと対応test | [API contract tests](docs/API_CONTRACT_TESTS.md) |
| Insight公開schemaへのfile-only handoff | [Examples](examples/README.md)、[schema出典](docs/contracts/README.md) |
| snapshotと将来のDeltaの区別 | [Snapshot revisions](docs/SNAPSHOT_REVISIONS.md) |
| package/API reference | `go doc .`、公開後の[pkg.go.dev](https://pkg.go.dev/github.com/sibukixxx/company-catalog) |

## Security — AppIDとraw transport情報を公開しない

AppIDはClientのConfigに保持されます。Source PackageにはAppIDやraw request URLを含めません。
`APIError.Error()`はtransport Causeを表示しませんが、`errors.Unwrap`したCauseには秘密情報を含むURLが残ることがあります。
Config・raw Cause・HTTP traceをログへ出さないでください。custom transportのloggingもconsumerが管理します。
脆弱性報告とcredential管理は[SECURITY.md](SECURITY.md)を参照してください。

## Development — offlineで検証できます

```sh
make test    # race + vet=all; live APIを呼ばない
make vet
make build
make lint
```

lintはpinned開発toolの取得にnetworkが必要な場合があります。SDKのgo.modに外部runtime dependencyはありません。
`Config.MaxRetries`は予約fieldで、retry処理は未実装です。cacheはLookupのみで利用し、name search cacheは未実装です。
live E2E/seedはcredentialを必要とする別経路です。詳しくは[CONTRIBUTING](CONTRIBUTING.md)を参照してください。

## Licenseと公開release

SDKコードはMIT。NTA由来データ/APIの利用条件は提供元に従います。
コピーしたInsight public schemaはApache-2.0を維持しています。
[出典・licenseの整理](docs/DATA_SOURCE_AND_ATTRIBUTION.md)、[CHANGELOG](CHANGELOG.md)、
[release方針](docs/RELEASING.md)、[OSS release gate](docs/OSS_RELEASE_GATE.md)を参照してください。
Public化・初回公開Releaseはmaintainerの作業として残しています。

利用上の質問・再現可能な不具合は[Issues](https://github.com/sibukixxx/company-catalog/issues)へ。
秘密情報を含む報告は公開Issueへ貼らず、SECURITY.mdのprivate報告手順を使ってください。
