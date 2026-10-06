package houjin

import (
	"testing"
	"time"
)

func TestConfigValidateReturnsErrorWhenAppIDIsEmpty(t *testing.T) {
	c := Config{}
	if err := c.Validate(); err == nil {
		t.Error("Validate() should fail when AppID is empty")
	}
}

func TestConfigValidateReturnsNilWhenAppIDIsSet(t *testing.T) {
	c := Config{AppID: "test-app-id"}
	if err := c.Validate(); err != nil {
		t.Errorf("Validate() returned error: %v", err)
	}
}

func TestConfigWithDefaultsFillsBaseURLWhenEmpty(t *testing.T) {
	c := Config{AppID: "x"}.WithDefaults()
	if c.BaseURL == "" {
		t.Error("WithDefaults should set BaseURL")
	}
}

func TestConfigWithDefaultsKeepsExplicitBaseURL(t *testing.T) {
	c := Config{AppID: "x", BaseURL: "https://example.test/4"}.WithDefaults()
	if c.BaseURL != "https://example.test/4" {
		t.Errorf("WithDefaults overrode BaseURL: %q", c.BaseURL)
	}
}

func TestConfigWithDefaultsSetsTimeoutAndRetries(t *testing.T) {
	c := Config{AppID: "x"}.WithDefaults()
	if c.Timeout <= 0 {
		t.Errorf("Timeout should be positive, got %v", c.Timeout)
	}
	if c.MaxRetries < 1 {
		t.Errorf("MaxRetries should be >= 1, got %d", c.MaxRetries)
	}
}

func TestConfigWithDefaultsKeepsCustomTimeout(t *testing.T) {
	c := Config{AppID: "x", Timeout: 3 * time.Second}.WithDefaults()
	if c.Timeout != 3*time.Second {
		t.Errorf("WithDefaults overrode Timeout: %v", c.Timeout)
	}
}

func TestConfigWithDefaultsCreatesHTTPClientWhenNil(t *testing.T) {
	c := Config{AppID: "x"}.WithDefaults()
	if c.HTTPClient == nil {
		t.Error("WithDefaults should create HTTPClient")
	}
}
