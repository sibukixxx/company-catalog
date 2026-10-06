package houjin

// CompanyKind is an upstream NTA kind code; the empty value means unknown.
type CompanyKind string

const (
	// CompanyKindUnknown represents an unavailable kind, not a numeric zero.
	CompanyKindUnknown CompanyKind = ""
)

func (k CompanyKind) String() string {
	return string(k)
}

// SourceInfo identifies the NTA provider/version and page acquisition time.
// FetchedAt is UTC RFC3339; it is not the upstream record update date.
type SourceInfo struct {
	Provider  string
	Version   string
	FetchedAt string
}

// CompanyProfile projects fields supplied by NTA without company scoring or
// enrichment. Optional empty strings mean unavailable; dates are source markers,
// not inferred founding/operating-status judgments. Raw is currently unpopulated.
type CompanyProfile struct {
	CorporateNumber string

	Name     string
	NameKana string
	NameEn   string

	Kind CompanyKind

	Address Address

	AssignmentDate string
	UpdateDate     string
	ChangeDate     string

	ChangeReason string
	CloseDate    string
	CloseCause   string

	Source SourceInfo

	Raw map[string]string
}
