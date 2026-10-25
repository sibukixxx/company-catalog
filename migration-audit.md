# Migration audit — company-catalog

監査日: 2026-10-25。作業中の記録。未実施・未完了を成功扱いしない。

## 正本と比較対象

- 正本: sibukixxx/company-catalog, main `804249e0bf6dac000708c7e269e308b973fc17fb`
- 旧実装: sibukixxx/ja-company-base, main `d40bd8dd2d38a5f4f1eaae987174d6a11d74ac38`
- 比較単位: 機能・API・入出力契約。名称変更後のファイルだけを機械的に欠落判定しない。

国税庁v4の取得・法人番号/法人名検索・差分取得・CSV/JSONL・Source Package v1のコアを保持。新正本の認証情報を隠すエラー処理を旧版で上書きしない。営業リストadapterと実利用consumerの分離は未完了。

## 今回の修正と検証

Goの最低版を1.26.10へ更新。既知の標準ライブラリ脆弱性に対応。既存CIは保持し、startup_failureをコードテスト失敗と混同しない。Cloudflare CI移行はplatform #41/#60の検証後に別途確定する。

- テスト・vet・lint・build: 実行中。結果はこのファイルに追記する。
- govulncheck: 実行予定。
- 現行ツリー/公開予定全Git履歴のsecret scan、author表示名・メール確認: 公開前の再確認が必要。
- README Quick Start/外部consumer/ライセンス/リリースtag: 最終確認が必要。
- Public化、匿名clone、Release: 未実施。ゲート未完了のまま公開しない。
- 旧リポジトリ: 保全。削除・archive・force pushは行わない。

環境リセット前の検証結果は現在の作業ツリーの成功証拠として流用しない。
