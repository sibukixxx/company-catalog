// Package houjin provides a typed, testable client and source-data producer for
// Japan's National Tax Agency Corporate Number Web API v4.
//
// NewClient requires an authorized NTA AppID supplied by the caller. No credential
// is distributed with this SDK. MockClient, ProjectForAnalysis and source-package
// readers/writers work offline without an AppID. The producer owns acquisition,
// field projection, validation, pagination and provenance; CRM, sales policy,
// cross-source enrichment and research conclusions belong to downstream consumers.
//
// The HTTP client may be shared across goroutines if its supplied HTTP transport
// and cache are safe for concurrent use. Config is copied at construction; callers
// must not mutate shared returned profiles or Raw maps concurrently. MemoryCache
// synchronizes its own entries but does not deep-copy profile maps. Configure
// MockClient function fields before use and synchronize any mutable mock state.
// Export writers are single-owner objects and require external synchronization
// for concurrent use. Config.MaxRetries is reserved; HTTP retries are not applied.
//
// APIError.Error omits the HTTP URL and AppID. Its unwrapped Cause may contain a
// credential-bearing transport URL: do not print raw causes, Config or HTTP traces.
// Source-package validation uses ordinary errors; do not assume every SDK error
// is an APIError. See docs/SOURCE_PACKAGE_CONTRACT.md for identity, time and
// missingness semantics. The existing AnalysisRecord name denotes a source
// projection, not an analytical conclusion or business lifecycle inference.
package houjin
