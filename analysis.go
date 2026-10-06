package houjin

// EventType classifies the most specific lifecycle marker present in the
// National Tax Agency company profile. It does not infer business events such
// as founding, business commencement, or bankruptcy.
type EventType string

const (
	EventTypeUnknown  EventType = "UNKNOWN"
	EventTypeAssigned EventType = "ASSIGNED"
	EventTypeUpdated  EventType = "UPDATED"
	EventTypeChanged  EventType = "CHANGED"
	EventTypeClosed   EventType = "CLOSED"
)

// AnalysisSchemaVersion identifies the AnalysisRecord export contract written
// to every CSV row and JSONL line. Additive changes keep this value; renaming,
// removing, or changing the meaning of a field requires a new schema version.
// See docs/ANALYSIS_EXPORT_CONTRACT.md.
const AnalysisSchemaVersion = "techvit.analysis.company/v1"

// AnalysisSource is the provenance carried by each exported record. It is a
// dedicated export type so that the file contract does not change when
// SourceInfo evolves inside the SDK.
type AnalysisSource struct {
	Provider  string `json:"provider"`
	Version   string `json:"version"`
	FetchedAt string `json:"fetched_at"`
}

// AnalysisRecord is a projection of CompanyProfile intended for downstream
// analytical tools. Empty geographic codes are intentional: CompanyProfile
// currently does not carry authoritative municipality codes, so this package
// does not infer them from address strings.
type AnalysisRecord struct {
	CorporateNumber string         `json:"corporate_number"`
	Name            string         `json:"name"`
	Kind            CompanyKind    `json:"kind"`
	PrefectureCode  string         `json:"prefecture_code"`
	PrefectureName  string         `json:"prefecture_name"`
	CityCode        string         `json:"city_code"`
	CityName        string         `json:"city_name"`
	AssignmentDate  string         `json:"assignment_date"`
	UpdateDate      string         `json:"update_date"`
	ChangeDate      string         `json:"change_date"`
	CloseDate       string         `json:"close_date"`
	CloseCause      string         `json:"close_cause"`
	EventType       EventType      `json:"event_type"`
	Source          AnalysisSource `json:"source"`
	SchemaVersion   string         `json:"schema_version"`
}

// ProjectForAnalysis converts a CompanyProfile without network access or
// inference beyond fields explicitly present in the profile.
func ProjectForAnalysis(company CompanyProfile) AnalysisRecord {
	return AnalysisRecord{
		CorporateNumber: company.CorporateNumber,
		Name:            company.Name,
		Kind:            company.Kind,
		PrefectureName:  company.Address.PrefectureName,
		CityName:        company.Address.CityName,
		AssignmentDate:  company.AssignmentDate,
		UpdateDate:      company.UpdateDate,
		ChangeDate:      company.ChangeDate,
		CloseDate:       company.CloseDate,
		CloseCause:      company.CloseCause,
		EventType:       analysisEventType(company),
		Source: AnalysisSource{
			Provider:  company.Source.Provider,
			Version:   company.Source.Version,
			FetchedAt: company.Source.FetchedAt,
		},
		SchemaVersion: AnalysisSchemaVersion,
	}
}

// analysisEventType uses explicit marker precedence. CLOSED wins over CHANGED,
// UPDATED, and ASSIGNED because it is the most specific marker represented by
// the current profile. This is a classification for analysis only.
func analysisEventType(company CompanyProfile) EventType {
	switch {
	case company.CloseDate != "":
		return EventTypeClosed
	case company.ChangeDate != "":
		return EventTypeChanged
	case company.UpdateDate != "":
		return EventTypeUpdated
	case company.AssignmentDate != "":
		return EventTypeAssigned
	default:
		return EventTypeUnknown
	}
}
