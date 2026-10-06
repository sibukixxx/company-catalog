# Release policy — tagを動かさず互換性を記録する

既存のv0.1.0 tagを維持します。Source Packageと今回のOSS整備を含む次version候補はv0.2.0です。
これは提案で、tag/Release作成やvisibility変更はmaintainerが行います。

## Versioning

| 対象 | 方針 |
| --- | --- |
| Go module v0.x | APIは発展中。互換性を優先し、breaking changeはminor更新とCHANGELOGで明記 |
| Go module v1.x以降 | public Go APIのbreaking changeはmajor更新。additive featureはminor、fixはpatch |
| AnalysisRecord / Source Package schema | module versionと独立。field意味の変更・rename/removeは新schema version |
| normalization semantics | NormalizationVersionとdefinition revisionを変更。観測データの変化と区別 |
| tag | immutable。既存versionを差し替えない |

現在のAnalysisSchemaVersionのadditive-change方針はANALYSIS_EXPORT_CONTRACT.mdを優先します。
Source Package readerは認識するschema/normalization versionのみ受け付けます。
SDKとschemaのversionが同じ番号になる保証はありません。

## Release前の確認

1. review済みPRをmainへ反映し、mainでtest/race/vet、build、lint、offline Examplesとpublic schema検証を実行。
2. current-tree secret scanとGo vulnerability scan。原Git履歴の全branch/tagを取得しfull-history scan。
3. LICENSE、third-party license、source terms/attribution、fixtureの出典を確認。
4. CHANGELOGの対象commitと既存tagとの差分を確認し、release notesへknown limitationsを記載。
5. OSS_RELEASE_GATE.mdのmaintainer checklistを終え、公開後のCI/pkg.go.devを確認。
6. review済みcommitへv0.2.0（提案）をtagし、GitHub Releaseを作成。

この手順は認証情報や取得データの公開を承認するものではありません。
通常testはofflineです。live API/E2Eはauthorized AppIDと明示enableがある場合にだけ実行し、未実施ならrelease notesへ記載します。

## Release notesへ残す制約

- AppIDは別途取得。SDKにcredentialは付属しない。
- source-specificなprojectionであり、industry/website/contact/売上のenrichmentはない。
- optional upstream値の省略/空tagをrecord v1は区別できない。unavailableを0やknown-emptyへ補完しない。
- COMPLETEは指定したAPI traversalの成功。全日本法人の完全な母集団や上流データの真偽を保証しない。
- record limitは取得後に適用。retryは未実装、cacheはLookupのみ。
- Public Engine fixtureはrequest schema適合で、live research実行の証明ではない。
- Delta algorithm (#8)は未実装で、OSS公開の必須条件にはしない。
