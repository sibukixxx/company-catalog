package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	houjin "github.com/sibukixxx/ja-company-base"
)

const appIDEnv = "HOUJIN_APP_ID"

type updateFetcher interface {
	FetchUpdatesByDateRange(context.Context, houjin.FetchUpdatesInput) ([]houjin.CompanyProfile, error)
}

type clientFactory func(houjin.Config) (updateFetcher, error)

type options struct {
	fromDate       string
	toDate         string
	prefectureCode string
	cityCode       string
	companyKind    string
	format         string
	output         string
	manifest       string
	limit          int
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr, func(cfg houjin.Config) (updateFetcher, error) {
		return houjin.NewClient(cfg)
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer, newClient clientFactory) error {
	opts, err := parseOptions(args, stderr)
	if err != nil {
		return err
	}
	appID := getenv(appIDEnv)
	if appID == "" {
		return fmt.Errorf("%s is required", appIDEnv)
	}

	client, err := newClient(houjin.Config{AppID: appID})
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	profiles, err := client.FetchUpdatesByDateRange(ctx, houjin.FetchUpdatesInput{
		FromDate:       opts.fromDate,
		ToDate:         opts.toDate,
		PrefectureCode: opts.prefectureCode,
		CityCode:       opts.cityCode,
		CompanyKind:    opts.companyKind,
		Limit:          opts.limit,
	})
	if err != nil {
		return fmt.Errorf("fetch updates: %w", err)
	}

	var sourcePackage houjin.CompanySourcePackage
	if opts.format == "json" {
		records := make([]houjin.AnalysisRecord, 0, len(profiles))
		for _, profile := range profiles {
			records = append(records, houjin.ProjectForAnalysis(profile))
		}
		now := time.Now().UTC()
		sourcePackage, err = houjin.NewCompanySourcePackage(records, houjin.AnalysisManifestQuery{From: opts.fromDate, To: opts.toDate, Prefecture: opts.prefectureCode, City: opts.cityCode, Kind: opts.companyKind, Limit: opts.limit}, now, now, "source-package-cli/v1")
		if err != nil {
			return err
		}
	}
	destination, closeDestination, err := openDestination(opts.output, stdout)
	if err != nil {
		return err
	}
	if closeDestination != nil {
		defer func() {
			if closeDestination != nil {
				_ = closeDestination()
			}
		}()
	}

	if opts.format == "json" {
		if err := houjin.WriteCompanySourcePackage(destination, sourcePackage); err != nil {
			return err
		}
		if closeDestination != nil {
			err := closeDestination()
			closeDestination = nil
			return err
		}
		return nil
	}
	writer, err := newAnalysisWriter(opts.format, destination)
	if err != nil {
		return err
	}
	recordCount := 0
	for _, profile := range profiles {
		if err := writer.Write(houjin.ProjectForAnalysis(profile)); err != nil {
			return fmt.Errorf("write analysis record: %w", err)
		}
		recordCount++
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("flush analysis output: %w", err)
	}
	if closeDestination != nil {
		if err := closeDestination(); err != nil {
			return fmt.Errorf("close output: %w", err)
		}
		closeDestination = nil
	}
	if opts.manifest != "" {
		if err := writeManifest(opts, recordCount, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

// writeManifest is called only after the data file is fully written, so a
// manifest never describes a partial export.
func writeManifest(opts options, recordCount int, generatedAt time.Time) error {
	manifest := houjin.NewAnalysisManifest(opts.format, houjin.AnalysisManifestQuery{
		From:       opts.fromDate,
		To:         opts.toDate,
		Prefecture: opts.prefectureCode,
		City:       opts.cityCode,
		Kind:       opts.companyKind,
		Limit:      opts.limit,
	}, generatedAt, recordCount)
	file, err := os.Create(opts.manifest)
	if err != nil {
		return fmt.Errorf("create manifest %q: %w", opts.manifest, err)
	}
	if err := houjin.WriteAnalysisManifest(file, manifest); err != nil {
		_ = file.Close()
		return fmt.Errorf("write manifest: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close manifest: %w", err)
	}
	return nil
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var opts options
	fs := flag.NewFlagSet("ja-company-export", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&opts.fromDate, "from", "", "first update date (YYYY-MM-DD, required)")
	fs.StringVar(&opts.toDate, "to", "", "last update date (YYYY-MM-DD, required)")
	fs.StringVar(&opts.prefectureCode, "prefecture", "", "NTA prefecture code")
	fs.StringVar(&opts.cityCode, "city", "", "NTA city code")
	fs.StringVar(&opts.companyKind, "kind", "", "NTA company kind")
	fs.StringVar(&opts.format, "format", "csv", "output format: csv, jsonl or json (source package)")
	fs.StringVar(&opts.output, "output", "-", "output file, or - for stdout")
	fs.StringVar(&opts.manifest, "manifest", "", "optional manifest JSON file describing the export")
	fs.IntVar(&opts.limit, "limit", 0, "maximum records; 0 means no limit")
	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if opts.fromDate == "" || opts.toDate == "" {
		return options{}, errors.New("-from and -to are required")
	}
	if opts.format != "csv" && opts.format != "jsonl" && opts.format != "json" {
		return options{}, fmt.Errorf("unsupported format %q: use csv, jsonl or json", opts.format)
	}
	if opts.limit < 0 {
		return options{}, errors.New("-limit must be zero or greater")
	}
	if opts.format == "json" && opts.manifest != "" {
		return options{}, errors.New("json source package already contains its manifest")
	}
	if opts.manifest == "-" {
		return options{}, errors.New("-manifest must be a file path")
	}
	if opts.manifest != "" && opts.manifest == opts.output {
		return options{}, errors.New("-manifest must differ from -output")
	}
	return opts, nil
}

func newAnalysisWriter(format string, destination io.Writer) (houjin.AnalysisWriter, error) {
	switch format {
	case "csv":
		return houjin.NewCSVAnalysisWriter(destination)
	case "jsonl":
		return houjin.NewJSONLAnalysisWriter(destination), nil
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func openDestination(path string, stdout io.Writer) (io.Writer, func() error, error) {
	if path == "-" {
		return stdout, nil, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("create output %q: %w", path, err)
	}
	return file, file.Close, nil
}
