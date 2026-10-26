# OSS release gate

This is a source-specific MIT Go library/CLI (the copied Insight public schema
retains its Apache-2.0 license in docs/contracts/INSIGHT_LICENSE), not a CRM, research engine or
complete enterprise-information platform. It requires no private repo/module,
consumer database or live credentials for ordinary build/tests/contract fixtures.

## Repository gate

- LICENSE, SECURITY.md, CONTRIBUTING.md exist; Go requirements are in go.mod.
- `make test` includes race and vet; `make build` compiles library, CLI and examples.
- `make lint` uses pinned golangci-lint. Public contract example is offline.
- Samples: contract fixtures are explicitly synthetic and contain no personal
  address/contact data. Existing Toyota API test samples are public corporate
  data. Live seed data is separate and is never silently replaced by synthetic
  records. API fixtures contain test credentials only.
- Gitleaks v8.30.0 found no leaks in the current working tree (2026-10-03). AppID references use test sentinels/environment reads. This is not a full-history secret audit.
  Normal export includes no AppID field, raw HTTP URL or CompanyProfile.Raw data.
  A sentinel regression test verifies the CLI output contains no AppID.
- Full-history scan is a separate gate: a connector-exported working snapshot
  does not contain original Git history and cannot prove historical safety.
- Existing SalesLead adapter remains only as backward-compatible public API;
  new lead/scoring/CRM semantics belong to outreach/Sales, never source records.

## Source terms and attribution

Only NTA Corporate Number Web API v4 is supported. MIT covers SDK code, not a
new license grant over upstream data. Users must review and comply with:

- [Web API terms](https://www.houjin-bangou.nta.go.jp/webapi/riyokiyaku.html)
- [Corporate Number publication site terms](https://www.houjin-bangou.nta.go.jp/riyokiyaku/)
- [API portal and application ID procedure](https://www.houjin-bangou.nta.go.jp/webapi/)

As checked on 2026-10-03, API terms article 6 requires source/non-endorsement
notice for services using the API; article 7 delegates content-use conditions
to the publication-site terms. Preserve provenance and explain normalization
and missing fields in derived exports. An AppID is personal to the authorized
API user and must never be included in packages or shared with consumers.
Do not label normalized source markers as inferred business events.

## Maintainer-only cutover (#7 remains open)

1. Authenticated clone including all branches/tags/history. Run gitleaks/trufflehog
   full-history scan and review findings (never publish leaked values).
2. Review final samples, terms/attribution and release changes.
3. GitHub Settings → General → visibility Public. No automated visibility change
   is part of this implementation.
4. Verify workflows on the public repo and pkg.go.dev availability; add description
   and topics; optionally enable private vulnerability reporting.
5. Follow [RELEASING.md](RELEASING.md). This fresh repository does not inherit historical tags. The recommended first public release candidate is v0.1.0 after the release gate passes. Outreach can then replace its existing sibling
   module replace with the published version in a separate downstream change.

Live API/E2E checks require an authorized HOUJIN_APP_ID and explicit E2E enablement;
report them separately from offline checks. Original history and live API access
were unavailable for this connector-based implementation.

## Verified preparation (2026-10-03)

Main at `4d1467ac16a8a17c707d0152f8753885f393ed6c` already contains Source
Package v1 and the pinned Insight public-contract fixture. Both repositories
passed test/build/lint before the release-preparation changes. the implemented Source Package/export work
are evaluated against that main, independently of the preparation PR.

The preparation changes add an offline-first README, examples index, executable
read/rebuild examples, GoDoc, source/license notes and release/version policy.
The offline package rebuild is byte-identical to the checked-in synthetic
fixture. Outreach reads it without producer DB access and exports two rows.
No public API signature, source-package schema or Sales handoff field changes.

The former Go 1.26.2 minimum produced ten reachable standard-library findings
in govulncheck. The minimum is now Go 1.26.10; govulncheck reports no
vulnerabilities for this tree. Consumers building this revision also need
Go 1.26.10 or newer. This scan result is dated evidence, not a continuing
vulnerability-free guarantee.

Repository preparation does not complete #7: the PR must be reviewed/merged,
original full history must be scanned, and maintainers must perform visibility,
public CI/pkg.go.dev and release checks. No visibility or release settings
were changed. Live API tests were not run without authorized credentials.
