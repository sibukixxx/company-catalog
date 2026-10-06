# Changelog

Go module release versionとfile schema versionは別に管理します。
release手順・互換性方針は[RELEASING](docs/RELEASING.md)を参照してください。

## Unreleased — next candidate v0.2.0

- AnalysisRecord v1 CSV/JSONL、manifest、FetchedAtを保持するSource Package v1。
- content hash、definition/snapshot revision、provenance、completenessの検証。
- malformed XML、reported count不足、ページmetadata不整合の拒否。
- Insight public JSON schemaに適合するfile-only handoff fixture。
- govulncheckの標準ライブラリ検出に対応し、最低Go versionを1.26.8へ更新。
- OSS向けREADME、Examples一覧、source/license attribution、package GoDoc。
- credential不要のSource Package reader / deterministic producer Examples。
- salesleadをoptional legacy compatibility adapterとして明記（型・import pathを維持）。

これらはcurrent mainと今回の公開整備PRを対象とする未releaseの内容です。
新しいtagやGitHub Releaseはまだ作成していません。

## Initial public release — pending maintainer cutover

公開前にfull original git historyのsecret scan、Public visibility、public Actionsとpkg.go.dev確認を行います。
既存v0.1.0 tagを動かさず、review済みcommitにv0.2.0を付ける案です。
GitHub Releaseには上記の変更と、ライブAPI/E2Eの実施状況・既知の制約を記載してください。

## v0.1.0 — existing tag (2026-07-28)

repositoryに既存のannotated tagがあり、core Lookup/Search/Diff、法人番号check digit、pagination、
in-memory cache、SalesLead DTO adapterをtag messageに記載しています。
GitHub Releaseは2026-10-03の確認時点でありません。後から追加されたSource Package機能を
この既存tagで利用できるとは記載しません。
