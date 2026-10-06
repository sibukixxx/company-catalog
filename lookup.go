package houjin

import (
	"context"
	"net/url"
)

func (c *httpClient) LookupByCorporateNumber(ctx context.Context, corporateNumber string) (*CompanyProfile, error) {
	if err := ValidateCorporateNumber(corporateNumber); err != nil {
		return nil, &APIError{
			Code:    ErrorCodeInvalidInput,
			Message: err.Error(),
			Cause:   err,
		}
	}

	if c.cacheEnabled() {
		if hit, ok, err := c.cfg.Cache.GetByCorporateNumber(ctx, corporateNumber); err == nil && ok && hit != nil {
			return hit, nil
		}
	}

	q := url.Values{}
	q.Set("id", c.cfg.AppID)
	q.Set("number", corporateNumber)
	q.Set("type", "12") // 12 = XML / UTF-8
	q.Set("history", "0")

	profiles, err := c.fetchProfiles(ctx, "/num", q)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, &APIError{
			Code:    ErrorCodeNotFound,
			Message: "no corporation matched",
		}
	}
	if c.cacheEnabled() {
		// Set のエラーはキャッシュ書き込み失敗にすぎず、本来の応答を失敗にしない。
		_ = c.cfg.Cache.SetByCorporateNumber(ctx, profiles[0])
	}
	return &profiles[0], nil
}

func (c *httpClient) cacheEnabled() bool {
	return c.cfg.Cache != nil && c.cfg.CachePolicy == CachePolicyReadThrough
}
