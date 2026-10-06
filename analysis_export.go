package houjin

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

// AnalysisWriter accepts analysis records one at a time so callers do not
// need to retain a complete dataset in memory.
type AnalysisWriter interface {
	Write(AnalysisRecord) error
	Close() error
}

// CSVAnalysisWriter emits the fixed AnalysisRecord v1 header and rows. It does
// not validate provenance; use CompanySourcePackage for validated packages.
type CSVAnalysisWriter struct {
	writer *csv.Writer
}

var analysisCSVHeader = []string{
	"corporate_number", "name", "kind", "prefecture_code", "prefecture_name",
	"city_code", "city_name", "assignment_date", "update_date", "change_date",
	"close_date", "close_cause", "event_type", "source_provider", "source_version",
	"source_fetched_at", "schema_version",
}

// NewCSVAnalysisWriter buffers the v1 header; Close must be called to flush
// and detect underlying writer errors. It does not close the supplied io.Writer.
func NewCSVAnalysisWriter(w io.Writer) (*CSVAnalysisWriter, error) {
	cw := csv.NewWriter(w)
	if err := cw.Write(analysisCSVHeader); err != nil {
		return nil, err
	}
	return &CSVAnalysisWriter{writer: cw}, nil
}

// Write buffers one legacy v1 CSV row; Close reports buffered I/O errors.
func (w *CSVAnalysisWriter) Write(record AnalysisRecord) error {
	return w.writer.Write([]string{
		record.CorporateNumber,
		record.Name,
		record.Kind.String(),
		record.PrefectureCode,
		record.PrefectureName,
		record.CityCode,
		record.CityName,
		record.AssignmentDate,
		record.UpdateDate,
		record.ChangeDate,
		record.CloseDate,
		record.CloseCause,
		string(record.EventType),
		record.Source.Provider,
		record.Source.Version,
		record.Source.FetchedAt,
		record.SchemaVersion,
	})
}

// Close flushes the CSV buffer without closing the underlying writer.
func (w *CSVAnalysisWriter) Close() error {
	w.writer.Flush()
	return w.writer.Error()
}

// JSONLAnalysisWriter emits one legacy AnalysisRecord JSON object per line.
// It does not validate source provenance and is not safe for concurrent writes.
type JSONLAnalysisWriter struct {
	encoder *json.Encoder
}

// NewJSONLAnalysisWriter wraps w without taking ownership of its lifetime.
func NewJSONLAnalysisWriter(w io.Writer) *JSONLAnalysisWriter {
	return &JSONLAnalysisWriter{encoder: json.NewEncoder(w)}
}

// Write encodes one record immediately and reports underlying I/O errors.
func (w *JSONLAnalysisWriter) Write(record AnalysisRecord) error {
	return w.encoder.Encode(record)
}

// Close is a no-op; it does not close the underlying writer.
func (w *JSONLAnalysisWriter) Close() error {
	return nil
}
