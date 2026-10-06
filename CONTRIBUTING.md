# Contributing

Thank you for considering a contribution to `company-catalog`.

This project is intentionally a small Go SDK around the Japanese National Tax Agency Corporate Number Web API. Keep changes focused on that boundary.

## Scope

Good contributions include:

- API compatibility and parsing fixes
- typed client improvements
- pagination, caching, validation, and error-handling improvements
- deterministic tests and fixtures
- documentation and examples
- security hardening that does not expose credentials

Out of scope for this repository:

- CRM or lead management
- outreach automation
- billing or invoicing workflows
- company scoring or enrichment from unrelated providers
- storing or distributing an NTA AppID

Those belong in downstream applications.

## Development

Requirements follow `go.mod`.

```bash
make lint
make test
make build
```

Before opening a pull request:

```bash
gofmt -w .
go mod tidy
make test
```

Do not require the live NTA API for unit tests. Use local HTTP test servers, fixtures, or `MockClient`.

Live E2E paths must remain explicitly gated and must read credentials from environment variables.

## Secrets and test data

Never commit:

- `HOUJIN_APP_ID` values
- credentials, tokens, or private keys
- private customer/company datasets
- raw HTTP traces containing secret query parameters

Synthetic fixtures and information already publicly available through the Corporate Number system are preferred.

## Compatibility

Public Go types and behavior are part of the SDK contract. Prefer additive changes. Document intentional breaking changes clearly.

## Pull requests

Keep pull requests small enough to review. Include:

- the problem being solved
- tests for behavior changes
- documentation changes when the public contract changes

By submitting a contribution, you agree that it is licensed under the repository's MIT License.

## Offline examples / public documentation

Run the examples in examples/README.md without HOUJIN_APP_ID before proposing a
release. Source Package examples reuse the existing synthetic documentation
fixture; acquired seed/golden files must still come through the documented live
seed path. Do not silently substitute synthetic data for an acquired golden.

Public GoDoc includes a runnable ExampleReadCompanySourcePackage. Keep README,
examples and public contracts paired with behavior. See docs/RELEASING.md for
module/schema versioning and CHANGELOG.md for the fresh-repository initial release candidate v0.1.0. Do not move tags or change visibility as part of a code PR.

The third-party Insight public schema retains Apache-2.0 in docs/contracts;
NTA data terms remain separate from the SDK's MIT license. See
 docs/DATA_SOURCE_AND_ATTRIBUTION.md.
