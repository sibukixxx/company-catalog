package houjin

// 共通テストフィクスチャ。client / search の HTTP テストから参照される。

const sampleXMLOneCorp = `<?xml version="1.0" encoding="UTF-8"?>
<corporations xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" id="1" lastUpdateDate="2024-01-01" count="1" divideNumber="1" divideSize="1">
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>1180301018771</corporateNumber>
    <process>01</process>
    <correct>0</correct>
    <updateDate>2024-01-01</updateDate>
    <changeDate>2023-12-15</changeDate>
    <name>トヨタ自動車株式会社</name>
    <nameImageId></nameImageId>
    <kind>301</kind>
    <prefectureName>愛知県</prefectureName>
    <cityName>豊田市</cityName>
    <streetNumber>トヨタ町１番地</streetNumber>
    <addressImageId></addressImageId>
    <prefectureCode>23</prefectureCode>
    <cityCode>211</cityCode>
    <postCode>4718571</postCode>
    <addressOutside></addressOutside>
    <addressOutsideImageId></addressOutsideImageId>
    <closeDate></closeDate>
    <closeCause></closeCause>
    <successorCorporateNumber></successorCorporateNumber>
    <changeCause></changeCause>
    <assignmentDate>2015-10-05</assignmentDate>
    <latest>1</latest>
    <enName></enName>
    <enPrefectureName></enPrefectureName>
    <enCityName></enCityName>
    <enAddressOutside></enAddressOutside>
    <furigana>トヨタジドウシャ</furigana>
    <hihyoji>0</hihyoji>
  </corporation>
</corporations>`
