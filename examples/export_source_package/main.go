// Command export_source_package rebuilds the small synthetic contract fixture.
// It demonstrates the public producer API without inventing new golden data.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	houjin "github.com/sibukixxx/company-catalog"
)

func run(args []string, out io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: export_source_package synthetic-package.json")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	fixture, readErr := houjin.ReadCompanySourcePackage(f)
	closeErr := f.Close()
	if readErr != nil {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	fetched, err := time.Parse(time.RFC3339, fixture.SourceFetchedAt)
	if err != nil {
		return err
	}
	generated, err := time.Parse(time.RFC3339, fixture.GeneratedAt)
	if err != nil {
		return err
	}
	// Reuse existing records/conditions and fixed times for deterministic output.
	p, err := houjin.NewCompanySourcePackage(fixture.Records, fixture.Manifest.Query, fetched, generated, fixture.GeneratorVersion)
	if err != nil {
		return err
	}
	return houjin.WriteCompanySourcePackage(out, p)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
