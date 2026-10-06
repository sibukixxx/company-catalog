package houjin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sibukixxx/ja-company-base/internal/xmlapi"
)

// 国税庁 Web-API の分割応答に対する安全装置。
// 異常な divideSize を返されても暴走しないよう、SDK 側の取得上限を設ける。
// 実運用で 1 リクエスト = 1MB 弱、20 ページ = 20MB 超を想定すれば十分な余裕。
const maxDividePages = 20

// SourceInfo と AnalysisManifest が記録する上流データソースの識別子。
const (
	sourceProvider = "houjin_bangou"
	sourceVersion  = "v4"
)

func (c *httpClient) fetchProfiles(ctx context.Context, path string, q url.Values) ([]CompanyProfile, error) {
	first, err := c.fetchPage(ctx, path, q, 0)
	if err != nil {
		return nil, err
	}
	if first.DivideNumber != 1 {
		return nil, errors.New("unexpected first page number")
	}
	profiles := profilesFromCorporations(first.Corporations, c.fetchedAt())

	if first.DivideSize <= 1 {
		return profiles, nil
	}
	if first.DivideSize > maxDividePages {
		return nil, &APIError{
			Code:    ErrorCodeAPIError,
			Message: fmt.Sprintf("upstream returned divideSize=%d exceeding safety limit %d", first.DivideSize, maxDividePages),
		}
	}

	for page := 2; page <= first.DivideSize; page++ {
		next, err := c.fetchPage(ctx, path, q, page)
		if err != nil {
			return nil, err
		}
		if next.DivideNumber != page || next.DivideSize != first.DivideSize || (first.TotalCount != nil && next.TotalCount != nil && *first.TotalCount != *next.TotalCount) {
			return nil, errors.New("inconsistent pagination metadata")
		}
		profiles = append(profiles, profilesFromCorporations(next.Corporations, c.fetchedAt())...)
	}
	if first.TotalCount != nil && *first.TotalCount != len(profiles) {
		return nil, errors.New("incomplete upstream response count")
	}
	return profiles, nil
}

func (c *httpClient) fetchPage(ctx context.Context, path string, base url.Values, page int) (xmlapi.Result, error) {
	q := base
	if page >= 2 {
		// url.Values は map なので Clone してから divide を書き換える。
		q = cloneValues(base)
		q.Set("divide", strconv.Itoa(page))
	}
	body, err := c.do(ctx, path, q)
	if err != nil {
		return xmlapi.Result{}, err
	}
	result, perr := xmlapi.Parse(bytes.NewReader(body))
	if perr != nil {
		return xmlapi.Result{}, &APIError{Code: ErrorCodeParseError, Message: perr.Error(), Cause: perr}
	}
	return result, nil
}

// fetchedAt はページ受信直後の時刻を UTC RFC3339 で返す。
func (c *httpClient) fetchedAt() string {
	return c.now().UTC().Format(time.RFC3339)
}

func profilesFromCorporations(corps []xmlapi.Corporation, fetchedAt string) []CompanyProfile {
	out := make([]CompanyProfile, 0, len(corps))
	for _, corp := range corps {
		profile := corporationToProfile(corp)
		profile.Source.FetchedAt = fetchedAt
		out = append(out, profile)
	}
	return out
}

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, vs := range v {
		cp := make([]string, len(vs))
		copy(cp, vs)
		out[k] = cp
	}
	return out
}

func corporationToProfile(c xmlapi.Corporation) CompanyProfile {
	return CompanyProfile{
		CorporateNumber: c.CorporateNumber,
		Name:            c.Name,
		NameKana:        c.Furigana,
		NameEn:          c.EnName,
		Kind:            CompanyKind(c.Kind),
		Address: Address{
			PostalCode:     c.PostCode,
			PrefectureName: c.PrefectureName,
			CityName:       c.CityName,
			StreetNumber:   c.StreetNumber,
			AddressImageID: c.AddressImageID,
		},
		AssignmentDate: c.AssignmentDate,
		UpdateDate:     c.UpdateDate,
		ChangeDate:     c.ChangeDate,
		ChangeReason:   c.ChangeCause,
		CloseDate:      c.CloseDate,
		CloseCause:     c.CloseCause,
		Source: SourceInfo{
			Provider: sourceProvider,
			Version:  sourceVersion,
		},
	}
}
