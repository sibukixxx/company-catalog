// Package xmlapi は国税庁の法人番号 Web-API の XML レスポンスを
// パースする純粋関数だけを置く。ドメインモデルへの変換は親パッケージ側で行う。
package xmlapi

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Corporation は法人番号 Web-API の XML レスポンスにおける <corporation> 要素
// の生データを保持する。フィールド名は API のタグ名に対応させる。
type Corporation struct {
	SequenceNumber  string `xml:"sequenceNumber"`
	CorporateNumber string `xml:"corporateNumber"`
	Process         string `xml:"process"`
	Correct         string `xml:"correct"`
	UpdateDate      string `xml:"updateDate"`
	ChangeDate      string `xml:"changeDate"`
	Name            string `xml:"name"`
	NameImageID     string `xml:"nameImageId"`
	Kind            string `xml:"kind"`
	PrefectureName  string `xml:"prefectureName"`
	CityName        string `xml:"cityName"`
	StreetNumber    string `xml:"streetNumber"`
	AddressImageID  string `xml:"addressImageId"`
	PrefectureCode  string `xml:"prefectureCode"`
	CityCode        string `xml:"cityCode"`
	PostCode        string `xml:"postCode"`
	AddressOutside  string `xml:"addressOutside"`
	CloseDate       string `xml:"closeDate"`
	CloseCause      string `xml:"closeCause"`
	SuccessorNumber string `xml:"successorCorporateNumber"`
	ChangeCause     string `xml:"changeCause"`
	AssignmentDate  string `xml:"assignmentDate"`
	Latest          string `xml:"latest"`
	EnName          string `xml:"enName"`
	Furigana        string `xml:"furigana"`
	Hihyoji         string `xml:"hihyoji"`
}

type corporations struct {
	XMLName             xml.Name      `xml:"corporations"`
	LastUpdateDate      string        `xml:"lastUpdateDate,attr"`
	Count               *int          `xml:"count,attr"`
	CountElement        *int          `xml:"count"`
	DivideNumberElement int           `xml:"divideNumber"`
	DivideSizeElement   int           `xml:"divideSize"`
	DivideNumber        int           `xml:"divideNumber,attr"`
	DivideSize          int           `xml:"divideSize,attr"`
	Corporations        []Corporation `xml:"corporation"`
}

// Result は Parse の戻り値。法人一覧に加えて、国税庁 Web-API の
// 分割応答メタ (divideNumber / divideSize) を含む。
//
// divideSize > 1 のとき、応答は複数ページに分割されており、クライアントは
// `divide=2..divideSize` のクエリパラメータで追加リクエストを送る必要がある。
// 単一ページ応答では DivideSize=1, DivideNumber=1 となるよう正規化する。
type Result struct {
	TotalCount   *int
	Corporations []Corporation
	DivideNumber int
	DivideSize   int
}

// Parse は法人番号 Web-API の XML レスポンスを読み取り、Result を返す。
func Parse(r io.Reader) (Result, error) {
	var doc corporations
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&doc); err != nil {
		return Result{}, fmt.Errorf("xmlapi: parse: %w", err)
	}
	if doc.Count == nil {
		doc.Count = doc.CountElement
	} else if doc.CountElement != nil && *doc.Count != *doc.CountElement {
		return Result{}, errors.New("xmlapi: conflicting record counts")
	}
	if doc.DivideNumber == 0 {
		doc.DivideNumber = doc.DivideNumberElement
	} else if doc.DivideNumberElement != 0 && doc.DivideNumber != doc.DivideNumberElement {
		return Result{}, errors.New("xmlapi: conflicting page numbers")
	}
	if doc.DivideSize == 0 {
		doc.DivideSize = doc.DivideSizeElement
	} else if doc.DivideSizeElement != 0 && doc.DivideSize != doc.DivideSizeElement {
		return Result{}, errors.New("xmlapi: conflicting page totals")
	}
	if doc.Count != nil && (*doc.Count < 0 || (doc.DivideSize <= 1 && *doc.Count != len(doc.Corporations))) {
		return Result{}, errors.New("xmlapi: incomplete response count")
	}
	// Reject extra documents and invalid records rather than exporting a
	// successfully decoded prefix as a normal dataset.
	for {
		token, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("xmlapi: trailing content: %w", err)
		}
		if text, ok := token.(xml.CharData); ok && strings.TrimSpace(string(text)) == "" {
			continue
		}
		return Result{}, errors.New("xmlapi: unexpected trailing content")
	}
	if doc.DivideNumber < 0 || doc.DivideSize < 0 || (doc.DivideSize > 0 && doc.DivideNumber > doc.DivideSize) {
		return Result{}, errors.New("xmlapi: invalid pagination metadata")
	}
	for _, c := range doc.Corporations {
		if c.CorporateNumber == "" || c.Name == "" {
			return Result{}, errors.New("xmlapi: missing corporation identity/name")
		}
	}
	return Result{
		TotalCount:   doc.Count,
		Corporations: doc.Corporations,
		DivideNumber: max(doc.DivideNumber, 1),
		DivideSize:   max(doc.DivideSize, 1),
	}, nil
}
