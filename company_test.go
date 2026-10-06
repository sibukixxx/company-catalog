package houjin

import "testing"

func TestCompanyKindStringReturnsUnderlyingValue(t *testing.T) {
	k := CompanyKind("301")

	if got := k.String(); got != "301" {
		t.Errorf("String() = %q, want %q", got, "301")
	}
}

func TestCompanyKindUnknownIsEmpty(t *testing.T) {
	if CompanyKindUnknown != "" {
		t.Errorf("CompanyKindUnknown = %q, want empty string", CompanyKindUnknown)
	}
}

func TestSourceInfoZeroValueHasEmptyFields(t *testing.T) {
	var s SourceInfo

	if s.Provider != "" || s.Version != "" || s.FetchedAt != "" {
		t.Errorf("SourceInfo zero value should have empty fields, got %+v", s)
	}
}
