package houjin

import (
	"context"
	"net/url"
)

// FetchUpdatesInput selects an NTA update window (YYYY-MM-DD) and optional
// region/kind filters. Limit applies after acquisition; zero means no limit.
type FetchUpdatesInput struct {
	FromDate string
	ToDate   string

	PrefectureCode string
	CityCode       string
	CompanyKind    string

	Limit int
}

func (c *httpClient) FetchUpdatesByDateRange(ctx context.Context, input FetchUpdatesInput) ([]CompanyProfile, error) {
	if input.FromDate == "" || input.ToDate == "" {
		return nil, &APIError{Code: ErrorCodeInvalidInput, Message: "FromDate and ToDate are required"}
	}

	q := url.Values{}
	q.Set("id", c.cfg.AppID)
	q.Set("from", input.FromDate)
	q.Set("to", input.ToDate)
	q.Set("type", "12")

	if addr := buildAddressFilter(input.PrefectureCode, input.CityCode); addr != "" {
		q.Set("address", addr)
	}
	if input.CompanyKind != "" {
		q.Set("kind", input.CompanyKind)
	}

	profiles, err := c.fetchProfiles(ctx, "/diff", q)
	if err != nil {
		return nil, err
	}
	return applyLimit(profiles, input.Limit), nil
}
