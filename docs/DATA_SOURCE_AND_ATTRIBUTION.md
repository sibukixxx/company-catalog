# コードlicenseとデータの利用条件

このrepositoryのSDKコードは[MIT License](../LICENSE)です。
国税庁由来の情報を、このrepoが作成した著作物やMITで新たに許諾するデータとして扱いません。
API・上流データは、提供元の規約に従って利用してください。

## 国税庁 法人番号Web-API v4

| 項目 | 確認先 / 扱い |
| --- | --- |
| source/provider | 国税庁法人番号システム、SDK identifierはhoujin_bangou |
| APIとAppIDの手続 | https://www.houjin-bangou.nta.go.jp/webapi/ |
| API利用条件 | https://www.houjin-bangou.nta.go.jp/webapi/riyokiyaku.html |
| upstream contentの利用条件 | https://www.houjin-bangou.nta.go.jp/riyokiyaku/ |
| AppID | SDKに付属せず、API利用者が取得・管理する |
| dataの権利 | MITの対象であるSDKコードとは別に確認する |

2026-10-03に確認したAPI規約を根拠に、既存OSS_RELEASE_GATE.mdは第6条の取得元・非保証表示、
第7条の公表サイト規約への参照を記載しています。AppIDの第三者への譲渡・貸与・開示は第9条に禁止事項として記載されています。
consumerへAppIDを再配布せず、公開packageにも含めません。規約変更や具体的な二次利用条件は、利用・公開時に提供元で再確認してください。
不明な権利関係や第三者の権利は、SDKのMIT表記で解決したものと扱わず確認が必要です。

Source Packageはprovider/API version、公開source/terms URL、credentialを除くquery、
recordごとの取得時刻、packageの取得完了/生成時刻を保持します。加工・projection・missing fieldsはcontract docsとwarningsで説明します。
source URLが残っていても、必要なconsumer表示を自動で満たすとは限りません。

## Third-party public schema

`docs/contracts/insight-public-engine.v1.schema.json`は、Insight Labのpublic JSON schemaを無変更でコピーしています。
Copyright 2026 Insight Lab contributors、Apache License 2.0を維持します。
元commit/pathは[contracts README](contracts/README.md)、license全文は[INSIGHT_LICENSE](contracts/INSIGHT_LICENSE)にあります。
SDK本体のMITに置き換えません。Insightのprivate/internal packageやDBには依存しません。

## Example / fixtureの出典

Source Package fixtureの法人名・番号・時刻は、明示的な合成contractデータです。
実在法人の事実や営業リストとして引用しないでください。実在する大企業の少数API sampleは公開情報の利用例であり、
private customer、buyer/user、不要な個人事業者情報を集めたfixtureではありません。
実APIから取得するgolden seedは既存の取得経路で作り、合成データで置き換えません。
