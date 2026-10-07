package houjin

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client acquires normalized NTA profiles. Lookup returns ErrorCodeNotFound for
// no match; Search and Diff return an empty slice for a successful empty result.
// Context cancellation is honored; split responses are joined before limits apply.
type Client interface {
	LookupByCorporateNumber(ctx context.Context, corporateNumber string) (*CompanyProfile, error)
	SearchByName(ctx context.Context, input SearchByNameInput) ([]CompanyProfile, error)
	FetchUpdatesByDateRange(ctx context.Context, input FetchUpdatesInput) ([]CompanyProfile, error)
}

type httpClient struct {
	cfg Config
	// now は SourceInfo.FetchedAt の打刻に使う。テストから固定時刻を注入する。
	now func() time.Time
}

// NewClient validates the required AppID and copies configuration with defaults.
// It does not contact the API until a Client method is called.
func NewClient(cfg Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &httpClient{cfg: cfg.WithDefaults(), now: time.Now}, nil
}

func buildAddressFilter(prefectureCode, cityCode string) string {
	if prefectureCode == "" {
		return ""
	}
	return prefectureCode + cityCode
}

func applyLimit(profiles []CompanyProfile, limit int) []CompanyProfile {
	if limit > 0 && len(profiles) > limit {
		return profiles[:limit]
	}
	return profiles
}

func (c *httpClient) do(ctx context.Context, path string, q url.Values) ([]byte, error) {
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, &APIError{Code: ErrorCodeAPIError, Message: "build request failed", Cause: err}
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, &APIError{Code: ErrorCodeTimeout, Message: "request timed out", Cause: err}
		}
		return nil, &APIError{Code: ErrorCodeAPIError, Message: "http request failed", Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	body, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		return nil, &APIError{Code: ErrorCodeAPIError, Message: "read body failed", Cause: rerr}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}
	return nil, mapStatusToAPIError(resp.StatusCode)
}

func mapStatusToAPIError(status int) error {
	code := ErrorCodeAPIError
	switch status {
	case http.StatusBadRequest:
		code = ErrorCodeInvalidInput
	case http.StatusUnauthorized, http.StatusForbidden:
		code = ErrorCodeUnauthorized
	case http.StatusNotFound:
		code = ErrorCodeNotFound
	case http.StatusTooManyRequests:
		code = ErrorCodeRateLimited
	}
	// AppID をログに出さないため、レスポンスボディは含めず status のみ。
	return &APIError{
		Code:       code,
		StatusCode: status,
		Message:    "upstream returned status " + strconv.Itoa(status),
	}
}
