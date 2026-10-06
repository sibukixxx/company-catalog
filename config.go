package houjin

import (
	"errors"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.houjin-bangou.nta.go.jp/4"

// Config configures acquisition only. AppID must be provided by the API user;
// never log Config because it contains that secret. See WithDefaults for defaults.
type Config struct {
	// AppID is required; the SDK retains it to authenticate requests.
	AppID string

	// BaseURL defaults to the NTA v4 endpoint; override for test servers.
	BaseURL string

	Timeout time.Duration
	// MaxRetries is reserved and currently unused; it does not enable retries.
	MaxRetries int

	// HTTPClient overrides Timeout; the caller owns its transport safety.
	HTTPClient *http.Client

	Cache       CompanyCache
	CachePolicy CachePolicy
}

// Validate checks that AppID is nonempty without exposing its value.
func (c Config) Validate() error {
	if c.AppID == "" {
		return errors.New("houjin: app id is required")
	}
	return nil
}

// WithDefaults returns a copy with the NTA v4 URL, a 10-second timeout,
// a default HTTPClient and disabled cache. MaxRetries defaults to 2 but is unused.
func (c Config) WithDefaults() Config {
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 2
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: c.Timeout}
	}
	if c.CachePolicy == "" {
		c.CachePolicy = CachePolicyDisabled
	}
	return c
}

// CachePolicy controls the optional Lookup read-through cache.
type CachePolicy string

const (
	CachePolicyDisabled    CachePolicy = "disabled"
	CachePolicyReadThrough CachePolicy = "read_through"
)
