package xmlapi

import (
	"strings"
	"testing"
)

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
    <closeDate></closeDate>
    <closeCause></closeCause>
    <successorCorporateNumber></successorCorporateNumber>
    <changeCause></changeCause>
    <assignmentDate>2015-10-05</assignmentDate>
    <latest>1</latest>
    <enName></enName>
    <furigana>トヨタジドウシャ</furigana>
    <hihyoji>0</hihyoji>
  </corporation>
</corporations>`

func TestParseReturnsSingleCorporation(t *testing.T) {
	result, err := Parse(strings.NewReader(sampleXMLOneCorp))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.Corporations) != 1 {
		t.Fatalf("got %d corporations, want 1", len(result.Corporations))
	}

	c := result.Corporations[0]
	if c.CorporateNumber != "1180301018771" {
		t.Errorf("CorporateNumber = %q", c.CorporateNumber)
	}
	if c.Name != "トヨタ自動車株式会社" {
		t.Errorf("Name = %q", c.Name)
	}
	if c.Furigana != "トヨタジドウシャ" {
		t.Errorf("Furigana = %q", c.Furigana)
	}
	if c.Kind != "301" {
		t.Errorf("Kind = %q", c.Kind)
	}
	if c.PostCode != "4718571" {
		t.Errorf("PostCode = %q", c.PostCode)
	}
	if c.PrefectureName != "愛知県" {
		t.Errorf("PrefectureName = %q", c.PrefectureName)
	}
	if c.CityName != "豊田市" {
		t.Errorf("CityName = %q", c.CityName)
	}
	if c.StreetNumber != "トヨタ町１番地" {
		t.Errorf("StreetNumber = %q", c.StreetNumber)
	}
	if c.AssignmentDate != "2015-10-05" {
		t.Errorf("AssignmentDate = %q", c.AssignmentDate)
	}
	if c.UpdateDate != "2024-01-01" {
		t.Errorf("UpdateDate = %q", c.UpdateDate)
	}
	if c.ChangeDate != "2023-12-15" {
		t.Errorf("ChangeDate = %q", c.ChangeDate)
	}
}

func TestParseReturnsEmptyWhenNoCorporation(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<corporations id="1" count="0"></corporations>`
	result, err := Parse(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.Corporations) != 0 {
		t.Errorf("expected 0 corporations, got %d", len(result.Corporations))
	}
}

func TestParseReturnsErrorOnMalformedXML(t *testing.T) {
	if _, err := Parse(strings.NewReader("<not-xml>")); err == nil {
		t.Error("expected error on malformed XML")
	}
}

func TestParseReturnsDivideMetadataFromXMLWhenPaginated(t *testing.T) {
	// 国税庁 Web-API は応答が大きい場合、divideSize 分割で返す。
	// divideNumber は現在ページ (1-indexed)。クライアント側は divideSize に達するまで
	// `divide=` クエリで追加リクエストを送る必要がある。
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<corporations id="1" lastUpdateDate="2024-01-01" count="2500" divideNumber="2" divideSize="3">
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>1180301018771</corporateNumber>
    <name>トヨタ自動車株式会社</name>
  </corporation>
</corporations>`
	result, err := Parse(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if result.DivideSize != 3 {
		t.Errorf("DivideSize = %d, want 3", result.DivideSize)
	}
	if result.DivideNumber != 2 {
		t.Errorf("DivideNumber = %d, want 2", result.DivideNumber)
	}
	if len(result.Corporations) != 1 {
		t.Errorf("len(Corporations) = %d, want 1", len(result.Corporations))
	}
}

func TestParseReturnsDivideSizeOneForNonPaginatedResponse(t *testing.T) {
	// 単一応答 (divide なし) は divideSize=1 として扱う。
	result, err := Parse(strings.NewReader(sampleXMLOneCorp))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if result.DivideSize != 1 {
		t.Errorf("DivideSize = %d, want 1", result.DivideSize)
	}
	if result.DivideNumber != 1 {
		t.Errorf("DivideNumber = %d, want 1", result.DivideNumber)
	}
}

func TestParseRejectsMalformedSuffixAndIncompleteRecords(t *testing.T) {
	for _, input := range []string{sampleXMLOneCorp + `<broken`, sampleXMLOneCorp + `<corporations/>`, `<corporations><corporation><name>partial</name></corporation></corporations>`, `<corporations divideNumber="2" divideSize="1"/>`} {
		if _, err := Parse(strings.NewReader(input)); err == nil {
			t.Fatal("accepted malformed/partial response")
		}
	}
}

func TestParseCountAndElementPagination(t *testing.T) {
	if _, err := Parse(strings.NewReader(`<corporations count="2"><corporation><corporateNumber>1</corporateNumber><name>A</name></corporation></corporations>`)); err == nil {
		t.Fatal("partial single-page accepted")
	}
	r, err := Parse(strings.NewReader(`<corporations><count>2</count><divideNumber>1</divideNumber><divideSize>2</divideSize><corporation><corporateNumber>1</corporateNumber><name>A</name></corporation></corporations>`))
	if err != nil || r.DivideSize != 2 || r.TotalCount == nil || *r.TotalCount != 2 {
		t.Fatal("lost element metadata", err)
	}
}
