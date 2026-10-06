# Release policy

`company-catalog` is a fresh repository. Historical tags from the former repository are not carried forward.

## Versioning

- The first public release candidate is **v0.1.0**.
- Pre-1.0 minor releases may include breaking Go API changes; document them in CHANGELOG.
- File/schema versions (AnalysisRecord, Source Package, normalization semantics) evolve independently from the Go module version.
- Published tags are immutable.

## Before a release

1. Run the offline test/race/vet/build/lint suite on the exact commit to be tagged.
2. Run a current-tree secret scan and Go vulnerability scan.
3. Review LICENSE, third-party licenses, NTA terms/attribution and public fixtures.
4. Verify README installation and offline quick-start commands from a clean clone.
5. Verify the OSS release gate in [OSS_RELEASE_GATE.md](OSS_RELEASE_GATE.md).
6. After the repository is public, confirm CI and pkg.go.dev resolution.
7. Tag the reviewed commit and create the GitHub Release.

Live API/E2E checks require an authorized `HOUJIN_APP_ID`. If they are not run, state that explicitly in the release notes.

## Release-note limitations

- An NTA AppID is not bundled with the SDK.
- This is a source-specific company-data producer; it does not provide industry, website, contact or revenue enrichment.
- Missing upstream values remain missing; they are not coerced to zero or known-empty.
- `COMPLETE` means the requested API traversal completed, not that the upstream source is a perfect census.
- Retry remains unimplemented where documented; cache coverage is intentionally limited.
- Public Engine fixtures validate contract compatibility only; they do not prove a live research execution.
- Snapshot delta comparison remains tracked separately and is not required for the first public release.
