# OSS migration verification — company-catalog

検証日: 2026-10-26。実装コードの検証対象head: `9412b575b8962277239c087e837c56a9044162bb`。本書の追記commitは資料のみ。

## 実装と実行結果

- `make test vet build`、golangci-lint、govulncheck: 成功（Go 1.26.10）。
- READMEのoffline読み取り・再export・Insight Public Engine v1 schema検証: 成功。
- private ja-company-outreachへsalesleadと例・回帰テストを移設。実consumerのtest/vet/build/lintと新producer Source Package→Sales CSV接続: 成功。SDKから営業実装を除去。
- 国税庁live API: AppID未設定につき未実施。offline検証と区別する。

## 公開前監査

- GitHub APIで全refs（main、作業branch、残存branch、PR head参照）と到達可能なcommitのtreeを列挙。列挙時点で77 commits / 142 unique blobsを取得、欠落・truncated treeなし。以降の当作業commitは現行tree検査と併せて確認。
- Gitleaks既定ruleで履歴の全blobとcommit metadataを検査。secret検出なし。
- PR最終treeは全fileを用意してGitleaks成功。localの検証対象sourceとGitHub PR headのblob hash一致を確認。
- 全履歴のauthorはYuichi Takada、メールは866555+sibukixxx@users.noreply.github.com。LICENSEはMIT。
- CI startup_failureはテスト実行前の起動失敗。CI成功とは記録せず、上記local gateを代替証拠とする。既存CIは削除していない。

## 未完了の公開操作

4正本はPrivate。Public化・Release作成・匿名clone/downloadは未実施。GitHub connectorはadministration/Release作成を提供せず、ブラウザは未ログイン。移行公開完了とは扱わない。公開後に匿名repository/README/LICENSE/Release/Quick Startを再検証して公式サイトをAvailableへ変更する。

予定初回tag: v0.1.0。旧repoは削除・archiveせず、共有履歴のforce pushなし。未完了Issueを閉じない。DNS・メール・他Workers/Pages・顧客環境は変更していない。
