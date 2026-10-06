// Command read_source_package validates a source package without credentials.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	houjin "github.com/sibukixxx/ja-company-base"
)

func run(args []string, out io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: read_source_package package.json")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	p, readErr := houjin.ReadCompanySourcePackage(f)
	closeErr := f.Close()
	if readErr != nil {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	if _, err := fmt.Fprintf(out, "schema: %s\npackage: %s\nsource: %s %s\nsource_fetched_at: %s\ngenerated_at: %s\nrecords: %d (%s)\n", p.SchemaVersion, p.PackageID, p.Manifest.Source, p.Manifest.SourceVersion, p.SourceFetchedAt, p.GeneratedAt, p.Manifest.RecordCount, p.Completeness); err != nil {
		return err
	}
	for _, r := range p.Records {
		if _, err := fmt.Fprintf(out, "%s: %s (fetched_at=%s)\n", r.CorporateNumber, r.Name, r.Source.FetchedAt); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
