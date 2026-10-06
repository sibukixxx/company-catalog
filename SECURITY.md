# Security Policy

## Supported versions

Security fixes are applied to the current `main` branch. Stable release support will be documented when versioned releases are published.

## Reporting a vulnerability

Please do **not** publish credentials, exploit details, or sensitive request/response data in a public issue.

Use GitHub's private security reporting feature when available for this repository. If private reporting is not available, open a minimal issue that contains no secret or exploit details and asks the maintainer for a private reporting channel.

## Credential handling

The Japanese National Tax Agency Corporate Number Web API AppID must be treated as a secret.

`ja-company-base` is designed so that:

- AppID values are supplied by the consuming application
- AppID values are not included in normal error messages
- examples read the AppID from `HOUJIN_APP_ID`
- live E2E tests are opt-in
- client-side/browser embedding of AppID values is discouraged

Never commit a real AppID, token, password, private key, or unredacted credential-bearing HTTP trace.

## Scope

Security reports about this SDK, its parsing, request construction, secret leakage, validation, or dependency handling are in scope.

Issues in the upstream National Tax Agency service itself should be reported to the appropriate upstream authority rather than this repository.

## Error and export boundaries

Client Config retains AppID for requests. APIError.Error() does not append the
transport Cause, but errors.Unwrap / APIError.Cause may expose a request URL
containing AppID. Log public error codes/messages only; do not dump Config,
unwrapped causes, raw HTTP traces, or custom transport logging without redaction.

Source Package does not include Config, AppID, CompanyProfile.Raw, or raw request
URLs. Its public reader checks provenance/schema/identities before downstream use.
Fixture data is explicitly synthetic and is not a private customer/outreach list.

## Release checks

Run current-tree secret scanning and govulncheck. A current snapshot or locally
initialized working Git repository is not the original project history. Before
Public visibility, the maintainer must obtain all original branches/tags/history
and run a full-history secret scan. See docs/OSS_RELEASE_GATE.md.
