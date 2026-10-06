# Company Source Package v1

The producer acquires NTA Corporate Number Web API v4 responses through Client,
normalizes CompanyProfile and exports the existing AnalysisRecord v1 projection.
It has no DB. AnalysisRecord's name is historical: these records are source facts,
not an AnalyticalArtifact, research conclusion or company founding judgment.

## Run

Go requirement: see go.mod (currently Go 1.26.8). Offline clone verification:

```sh
make test
make build
make lint
# Explicitly synthetic contract fixture; no API credential needed:
go run ./examples/insight_handoff docs/examples/company-source-package.v1.json
python -m pip install jsonschema
python scripts/verify_public_contract.py
```

Live export (AppID supplied via HOUJIN_APP_ID, never in parameters or output):

```sh
go run ./cmd/ja-company-export -from 2026-09-01 -to 2026-09-30 \
  -prefecture 13 -format json -output companies.json
```

CSV/JSONL and their existing optional manifest remain unchanged. `-format json`
emits one validated CompanySourcePackage with embedded AnalysisManifest. It
rejects a separate `-manifest`. Consumers call ReadCompanySourcePackage or
implement this public JSON contract; no downstream DB is consulted.

## Identities and provenance

- schema_version: techvit.company.source-package/v1. Unsupported major versions fail.
- manifest: existing analysis schema, provider/API version, typed query filters,
  export time and exact record count.
- content_hash: SHA-256 of Go encoding/json's compact JSON array of sorted **JSON
  strings** of AnalysisRecords with source.fetched_at set to empty. Duplicates are
  retained. This is a semantic content hash, **not** the file byte checksum.
- dataset_definition_revision: SHA-256 of Go compact JSON with fields Schema,
  Provider, Version, Normalization, Query (in that order), describing the analysis
  schema, source version, normalization version and existing manifest query.
- snapshot_revision: SHA-256 of compact JSON [definition revision, content hash].
- package_id: `company-` followed by snapshot_revision.
- source_fetched_at: successful acquisition completion time (also available for
  a zero-row response); each record retains its page's original source.fetched_at.
- generated_at/generator_version: export execution metadata, excluded from content
  and definition identity. Use generator build/release version when calling the
  library; the current CLI identifies its generator as source-package-cli/v1.
- provenance: fixed provider API URL, terms URL and FetchUpdatesByDateRange path.
  Acquisition query excludes AppID; replay requires one's own authorized AppID.

Reacquisition may yield different records; provenance supports replay of the
conditions, not a promise the mutable upstream reproduces old bytes. Preserve
package files to reproduce a downstream result.

Records are canonically sorted for export; record order alone never changes a
content/snapshot identity. Observation times can change while identities stay
constant. Reader/writer recompute identities, record count and metadata and
reject missing record provenance or modified content before use.

## Missingness and completeness

The existing record v1 represents unavailable optional strings as empty strings;
it does not preserve the distinction between an upstream omitted tag and an empty
tag. Neither is interpreted as a known-empty measurement or numeric zero.
CompanyKind empty and EventType UNKNOWN remain unknown. Record count 0 represents
an observed successful empty result; missing/failed acquisition cannot create it.
Prefecture/city codes are unavailable in this projection even though some upstream
responses contain them. Industry, employees, turnover, website and contacts are
unsupported. Warnings carry these limitations. A future richer missingness model
requires a deliberate record contract revision, not inferred values.

Completeness COMPLETE means successful completion of the requested API traversal,
**not** coverage of all Japanese companies or proof of upstream truth. Any positive
query.limit marks LIMITED conservatively. Partial/malformed responses return an
error and no source package. API split-page inconsistency is rejected. The
20-page safety limit remains. No geography/industry completeness is inferred.

## Public consumer fixture

`docs/examples/company-source-package.v1.json` contains two explicitly synthetic
contract records with invalid, deliberately reserved test identifiers, no personal
address/contact information, and illustrative times. It is not a seed/golden API
response and must not be cited as evidence about actual companies.

`examples/insight_handoff` is a file-only downstream adapter example. It validates
source packages then emits Public Engine v1 AddEvidenceRequest using source=dataset
and the entire package as document content. It creates no Observation,
Recommendation or Research semantics. Schema validation runs against the public
schema vendored from sibukixxx/insight at the commit recorded in
`docs/contracts/README.md`. This proves public request conformance, not live engine
execution or analytical ingestion. That remains downstream's integration step.

Outreach maps available fields to its existing Targets and Sales handoff;
website/contact fields and eligibility stay there. The pre-existing
adapters/saleslead compatibility package is retained to avoid breaking existing
users; no new sales policy is added to it. New producer contracts are consumer-neutral.

## Time semantics (public consumer reference)

| Field | Meaning | Not interchangeable with |
| --- | --- | --- |
| record.source.fetched_at / CompanyProfile.Source.FetchedAt | original page reception time, UTC RFC3339 | source record update/change date |
| source_fetched_at | successful acquisition completion time supplied to constructor; CLI captures after fetch | every page reception time or generator time |
| generated_at / manifest.generated_at | export execution time | acquisition time |
| record.update_date / change_date / assignment_date / close_date | upstream YYYY-MM-DD markers | SDK observation timestamp or inferred business event |
| provider_updated_at | not exported in current public model | do not substitute FetchedAt or GeneratedAt |

Same data re-exported preserves content/snapshot identity. Same query re-fetched
may update observation times alone, or may change content if source records differ.
A declared source/API/schema/normalization change alters definition revision;
a silent upstream definition change remains unknown without authoritative metadata.
Row order alone is irrelevant. For future comparisons see SNAPSHOT_REVISIONS.md.

Missing or empty upstream optional values collapse to unavailable strings in the
legacy record model; the package cannot reconstruct a distinction the projection
has already lost. UNKNOWN lifecycle/kind markers are not zero. A successful
record_count=0 is observed zero; partial/failed acquisition is an error or LIMITED,
not observed zero. Not-applicable is not explicitly encoded in v1 and must not be
inferred from an empty date or absent close marker.

The legacy CSV/JSONL writers do not validate provenance by themselves. The source
package constructor/reader/writer do, and are the boundary for validated exports.
Use examples/read_source_package and examples/export_source_package to inspect
or rebuild the existing synthetic fixture without credentials. See examples/README.md
for fixture locations and docs/DATA_SOURCE_AND_ATTRIBUTION.md for licensing.
