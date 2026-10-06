package houjin

import (
	"context"
	"net/url"
	"strings"
)

// SearchByNameInput selects upstream name matching and optional region/kind
// filters. CityCode requires PrefectureCode. Limit is applied after acquisition.
type SearchByNameInput struct {
	Name string

	PrefectureCode string
	CityCode       string
	CompanyKind    string

	Limit int

	MatchMode NameMatchMode
}

// NameMatchMode selects NTA default, exact or partial name matching.
type NameMatchMode string

const (
	NameMatchModeDefault NameMatchMode = ""
	NameMatchModeExact   NameMatchMode = "exact"
	NameMatchModePartial NameMatchMode = "partial"
)

func (c *httpClient) SearchByName(ctx context.Context, input SearchByNameInput) ([]CompanyProfile, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, &APIError{Code: ErrorCodeInvalidInput, Message: "name is required"}
	}

	q := url.Values{}
	q.Set("id", c.cfg.AppID)
	q.Set("name", input.Name)
	q.Set("type", "12")
	q.Set("history", "0")

	if addr := buildAddressFilter(input.PrefectureCode, input.CityCode); addr != "" {
		q.Set("address", addr)
	}
	if input.CompanyKind != "" {
		q.Set("kind", input.CompanyKind)
	}
	switch input.MatchMode {
	case NameMatchModeExact:
		q.Set("mode", "1")
	case NameMatchModePartial:
		q.Set("mode", "2")
	}

	profiles, err := c.fetchProfiles(ctx, "/name", q)
	if err != nil {
		return nil, err
	}
	return applyLimit(profiles, input.Limit), nil
}
