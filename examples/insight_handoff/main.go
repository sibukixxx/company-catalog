// An external file-only consumer example. The producer has no Insight dependency.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	houjin "github.com/sibukixxx/company-catalog"
)

func handoff(r io.Reader, w io.Writer) error {
	p, err := houjin.ReadCompanySourcePackage(r)
	if err != nil {
		return err
	}
	content, err := json.Marshal(p)
	if err != nil {
		return err
	}
	// Public Engine v1 EvidenceDocument is a source document, not an
	// AnalyticalArtifact. Full source provenance survives inside content.
	request := map[string]any{"contractVersion": "1", "idempotencyKey": p.PackageID + "-" + p.SourceFetchedAt, "documents": []any{map[string]any{"externalRef": p.PackageID, "source": "dataset", "title": "NTA company source snapshot", "content": string(content), "metadata": map[string]any{"snapshotRevision": p.SnapshotRevision, "datasetDefinitionRevision": p.DatasetDefinitionRevision, "contentHash": p.ContentHash, "sourceFetchedAt": p.SourceFetchedAt}}}}
	return json.NewEncoder(w).Encode(request)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: insight_handoff source-package.json")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = handoff(f, os.Stdout)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
