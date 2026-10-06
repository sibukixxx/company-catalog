# Analysis export contract (`techvit.analysis.company/v1`)

`company-catalog` は国税庁 法人番号システム Web-API 専用の source adapter であり、
分析用途向けに `AnalysisRecord` を CSV / JSONL として出力する。本書はその出力ファイルの
契約を定める。downstream の分析ツール（Insight Lab 等）は Go module に依存せず、
**本書のファイル契約だけ** で出力を読めることを目的とする。

```text
NTA Corporate Number API (Web-API Ver.4.0)
  → typed Client (FetchUpdatesByDateRange)
  → CompanyProfile (normalization + SourceInfo provenance)
  → AnalysisRecord (ProjectForAnalysis)
  → CSV / JSONL + optional manifest JSON
```

## 1. Schema identifier

| 項目 | 値 |
| --- | --- |
| Schema ID | `techvit.analysis.company/v1` |
| Go 定数 | `houjin.AnalysisSchemaVersion` |
| 記録場所 | CSV の `schema_version` 列、JSONL 各行の `schema_version`、manifest の `schema` |

schema ID は全レコードに入る。manifest を失っても、行単体でどの契約で書かれたかを判定できる。

## 2. Fields

すべての値は文字列。値がないときは空文字（JSONL でも `null` ではなく `""`）。

| JSONL key | CSV column | 内容 | 空文字の意味 |
| --- | --- | --- | --- |
| `corporate_number` | `corporate_number` | 法人番号（13 桁） | 上流に値がない |
| `name` | `name` | 商号または名称 | 上流に値がない |
| `kind` | `kind` | 法人種別コード（例: `301`） | 上流に値がない |
| `prefecture_code` | `prefecture_code` | 都道府県コード | **常に空（v1）**。後述の推論禁止を参照 |
| `prefecture_name` | `prefecture_name` | 都道府県名 | 上流に値がない |
| `city_code` | `city_code` | 市区町村コード | **常に空（v1）**。後述の推論禁止を参照 |
| `city_name` | `city_name` | 市区町村名 | 上流に値がない |
| `assignment_date` | `assignment_date` | 法人番号指定年月日（`YYYY-MM-DD`） | 上流に値がない |
| `update_date` | `update_date` | 最終更新年月日 | 上流に値がない |
| `change_date` | `change_date` | 直近変更年月日 | 上流に値がない |
| `close_date` | `close_date` | 登記記録の閉鎖等年月日 | 閉鎖等の記録がない |
| `close_cause` | `close_cause` | 登記記録の閉鎖等の事由（上流の値をそのまま出力） | 閉鎖等の記録がない |
| `event_type` | `event_type` | `ASSIGNED` / `UPDATED` / `CHANGED` / `CLOSED` / `UNKNOWN` | 空にならない |
| `source.provider` | `source_provider` | 上流データソース。v1 では `houjin_bangou` | SDK 外で組み立てた record |
| `source.version` | `source_version` | 上流 API バージョン。v1 では `v4` | SDK 外で組み立てた record |
| `source.fetched_at` | `source_fetched_at` | SDK が上流応答を受信した時刻（UTC RFC3339） | SDK 外で組み立てた record |
| `schema_version` | `schema_version` | `techvit.analysis.company/v1` | 空にならない |

`event_type` は profile に含まれる日付の有無から `CLOSED > CHANGED > UPDATED > ASSIGNED`
の優先順で決める分析用の分類であり、設立・倒産などの事業イベントを推論したものではない。

`source.fetched_at` は分割応答ではページごとの受信時刻になる。`MemoryCache` から返った
profile は、キャッシュに格納された時点の値（実際に上流から取得した時刻）を保持する。

## 3. File formats

### CSV

- UTF-8、RFC 4180 準拠（`encoding/csv`）。カンマ・改行・`"` を含む値はクォートされる
- 1 行目はヘッダー。列順は以下で固定する（テストで検証）

```text
corporate_number,name,kind,prefecture_code,prefecture_name,city_code,city_name,assignment_date,update_date,change_date,close_date,close_cause,event_type,source_provider,source_version,source_fetched_at,schema_version
```

### JSONL

- UTF-8、1 行 1 JSON object、改行区切り
- provenance は `source` object にネストする

```json
{"corporate_number":"1180301018771","name":"...","kind":"301","prefecture_code":"","prefecture_name":"愛知県","city_code":"","city_name":"豊田市","assignment_date":"2015-10-05","update_date":"2024-01-01","change_date":"","close_date":"","close_cause":"","event_type":"UPDATED","source":{"provider":"houjin_bangou","version":"v4","fetched_at":"2026-09-13T00:00:00Z"},"schema_version":"techvit.analysis.company/v1"}
```

## 4. Manifest（任意）

`company-catalog -manifest <path>` を指定すると、データ本体の書き込み完了後に manifest
JSON を別ファイルとして出力する。データの書き込みに失敗した場合、manifest は書かれない。

```json
{
  "schema": "techvit.analysis.company/v1",
  "source": "houjin_bangou",
  "source_version": "v4",
  "format": "jsonl",
  "query": {
    "from": "2026-01-01",
    "to": "2026-06-30",
    "prefecture": "13"
  },
  "generated_at": "2026-07-01T00:00:00Z",
  "record_count": 1234
}
```

| key | 内容 |
| --- | --- |
| `schema` | データ本体の schema ID |
| `source` / `source_version` | 上流データソースと API バージョン |
| `format` | `csv` または `jsonl` |
| `query` | 上流 API に渡した条件。`from` / `to` は常に出力、`prefecture` / `city` / `kind` / `limit` は指定時のみ |
| `generated_at` | export を実行した時刻（UTC RFC3339） |
| `record_count` | データ本体のレコード数（CSV ヘッダー行を含まない） |

manifest には AppID などの credential を **絶対に含めない**（テストで検証）。
同じ内容のサンプルを [examples/analysis-manifest.example.json](examples/analysis-manifest.example.json) に置いている。

export 時刻は manifest だけに記録する。レコードに入れると、同じデータでも export の度に
差分が出てしまうため。

## 5. Source provenance と推論禁止

- 本 repo は上流 API が返した値だけを出力する。住所文字列から `prefecture_code` /
  `city_code` を推論・補完しない。v1 でこれらの列が空なのは意図した挙動である
- コードが必要な場合は、downstream 側で出典を明示した別 source と結合する
- 上流 API が将来コードを返すようになった場合は、その値を埋めることを v1 の範囲内の
  変更として扱う（列の意味は変わらない）

## 6. Compatibility policy

downstream reader に求めること:

- CSV は **ヘッダー名で列を引く**（列番号で引かない）
- JSONL / manifest の **未知の key は無視する**
- `event_type` の未知の値は `UNKNOWN` と同様に扱う
- `schema_version` / `schema` の major（`/v1`）を確認し、未知の major は読まない

同じ schema ID（`/v1`）のまま行ってよい変更:

- 新しい列 / key の追加。CSV では末尾に追加する
- `event_type` など列挙値への新しい値の追加
- manifest への key の追加
- これまで空だった列に、上流が返すようになった値を入れること

新しい schema ID（`/v2`）が必要な変更:

- 列 / key の削除・rename、CSV 列順の変更
- 既存の値の意味・形式の変更（日付形式、`event_type` の判定規則など）
- `source` object の構造変更

## 7. Schema migration policy

1. 破壊的変更は新しい schema ID（例: `techvit.analysis.company/v2`）として追加する
2. 移行期間中は旧 schema の出力を残し、利用者が旧 schema を選べるようにする
3. 旧 schema の廃止は GitHub Release の release note で事前告知し、少なくとも 1 minor release の
   移行期間を置く
4. schema ID を変えずに破壊的変更を入れることはしない

## 8. Downstream integration（Insight Lab 等）

downstream は Go module に依存せず、ファイルだけを受け取る。

```bash
HOUJIN_APP_ID=xxx go run ./cmd/company-catalog \
  -from 2026-01-01 -to 2026-06-30 -prefecture 13 \
  -format jsonl -output companies.jsonl \
  -manifest companies.manifest.json
```

読み込み手順:

1. `companies.manifest.json` を読み、`schema` が `techvit.analysis.company/v1` であることを確認する
2. `record_count` と `query` を取り込み記録（lineage）に残す
3. データ本体を 1 行ずつ読み、各行の `schema_version` を確認する
4. `source.fetched_at` と manifest の `generated_at` を provenance として分析結果に引き継ぐ

## 9. Scope and non-goals

- 本 repo は **国税庁 法人番号データ専用** の source adapter である
- e-Stat、自治体オープンデータ、gBizINFO などの別データソースは本 repo に追加しない。
  それらは別の adapter / repo が担う
- 汎用 Open Data SDK や、複数 source を横断する ingestion framework にはしない
- 分析・スコアリング・名寄せは downstream の責務とする
