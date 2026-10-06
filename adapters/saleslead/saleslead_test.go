package saleslead_test

import (
	"testing"

	houjin "github.com/sibukixxx/ja-company-base"
	"github.com/sibukixxx/ja-company-base/adapters/saleslead"
)

func sampleProfile(corp, name, kind, closeDate string) houjin.CompanyProfile {
	return houjin.CompanyProfile{
		CorporateNumber: corp,
		Name:            name,
		NameKana:        name + "カナ",
		Kind:            houjin.CompanyKind(kind),
		Address: houjin.Address{
			PostalCode:     "1000001",
			PrefectureName: "東京都",
			CityName:       "千代田区",
			StreetNumber:   "1-1-1",
		},
		CloseDate: closeDate,
		Source: houjin.SourceInfo{
			Provider: "houjin_bangou",
			Version:  "v4",
		},
	}
}

func TestFromCompanyProfileSetsCaseIDAndDefaultsToNew(t *testing.T) {
	c := saleslead.Case{ID: "case-123", Name: "Q2 関東"}
	p := sampleProfile("1180301018771", "テスト株式会社", "301", "")

	got := saleslead.FromCompanyProfile(c, p)

	if got.CaseID != "case-123" {
		t.Errorf("CaseID = %q, want %q", got.CaseID, "case-123")
	}
	if got.Status != saleslead.ProspectStatusNew {
		t.Errorf("Status = %q, want %q", got.Status, saleslead.ProspectStatusNew)
	}
	if got.Closed {
		t.Errorf("Closed = true, want false")
	}
	if got.CorporateNumber != p.CorporateNumber {
		t.Errorf("CorporateNumber = %q, want %q", got.CorporateNumber, p.CorporateNumber)
	}
}

func TestFromCompanyProfileMarksClosedAsExcluded(t *testing.T) {
	c := saleslead.Case{ID: "c1", IncludeClosed: true}
	p := sampleProfile("1234567890123", "閉鎖株式会社", "301", "2020-04-01")

	got := saleslead.FromCompanyProfile(c, p)

	if !got.Closed {
		t.Errorf("Closed = false, want true")
	}
	if got.Status != saleslead.ProspectStatusExcluded {
		t.Errorf("Status = %q, want %q", got.Status, saleslead.ProspectStatusExcluded)
	}
}

func TestBuildListExcludesClosedByDefault(t *testing.T) {
	c := saleslead.Case{ID: "c1"}
	profiles := []houjin.CompanyProfile{
		sampleProfile("1111111111111", "現役株式会社", "301", ""),
		sampleProfile("2222222222222", "閉鎖株式会社", "301", "2020-04-01"),
	}

	got := saleslead.BuildList(c, profiles)

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].CorporateNumber != "1111111111111" {
		t.Errorf("CorporateNumber = %q, want %q", got[0].CorporateNumber, "1111111111111")
	}
}

func TestBuildListIncludesClosedWhenConfigured(t *testing.T) {
	c := saleslead.Case{ID: "c1", IncludeClosed: true}
	profiles := []houjin.CompanyProfile{
		sampleProfile("1111111111111", "現役株式会社", "301", ""),
		sampleProfile("2222222222222", "閉鎖株式会社", "301", "2020-04-01"),
	}

	got := saleslead.BuildList(c, profiles)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[1].Status != saleslead.ProspectStatusExcluded {
		t.Errorf("got[1].Status = %q, want %q", got[1].Status, saleslead.ProspectStatusExcluded)
	}
}

func TestBuildListFiltersByCompanyKinds(t *testing.T) {
	c := saleslead.Case{ID: "c1", CompanyKinds: []string{"301"}}
	profiles := []houjin.CompanyProfile{
		sampleProfile("1111111111111", "株式会社A", "301", ""),
		sampleProfile("2222222222222", "合同会社B", "302", ""),
		sampleProfile("3333333333333", "株式会社C", "301", ""),
	}

	got := saleslead.BuildList(c, profiles)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for _, p := range got {
		if p.Kind.String() != "301" {
			t.Errorf("Kind = %q, want 301", p.Kind)
		}
	}
}

func TestBuildListDeduplicatesByCorporateNumber(t *testing.T) {
	c := saleslead.Case{ID: "c1"}
	dup := sampleProfile("1111111111111", "重複株式会社", "301", "")
	profiles := []houjin.CompanyProfile{
		dup,
		dup,
		sampleProfile("2222222222222", "別会社", "301", ""),
	}

	got := saleslead.BuildList(c, profiles)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].CorporateNumber != "1111111111111" || got[1].CorporateNumber != "2222222222222" {
		t.Errorf("unexpected order: %+v", got)
	}
}

func TestCaseMatchesReturnsFalseForClosedWhenNotIncluded(t *testing.T) {
	c := saleslead.Case{ID: "c1"}
	p := sampleProfile("1111111111111", "閉鎖", "301", "2020-04-01")

	if c.Matches(p) {
		t.Errorf("Matches = true, want false")
	}
}

func TestCaseMatchesReturnsFalseWhenKindNotInList(t *testing.T) {
	c := saleslead.Case{ID: "c1", CompanyKinds: []string{"301"}}
	p := sampleProfile("1111111111111", "合同会社", "302", "")

	if c.Matches(p) {
		t.Errorf("Matches = true, want false")
	}
}

func TestCaseMatchesReturnsTrueForEmptyFilters(t *testing.T) {
	c := saleslead.Case{ID: "c1"}
	p := sampleProfile("1111111111111", "株式会社", "301", "")

	if !c.Matches(p) {
		t.Errorf("Matches = false, want true")
	}
}

func TestProspectStatusConstantsAreStable(t *testing.T) {
	cases := []struct {
		status saleslead.ProspectStatus
		want   string
	}{
		{saleslead.ProspectStatusNew, "new"},
		{saleslead.ProspectStatusContacted, "contacted"},
		{saleslead.ProspectStatusInProgress, "in_progress"},
		{saleslead.ProspectStatusWon, "won"},
		{saleslead.ProspectStatusLost, "lost"},
		{saleslead.ProspectStatusExcluded, "excluded"},
	}
	for _, tc := range cases {
		if string(tc.status) != tc.want {
			t.Errorf("ProspectStatus %q != %q", tc.status, tc.want)
		}
	}
}
