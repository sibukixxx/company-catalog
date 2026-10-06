// Package adapters contains optional legacy downstream projection adapters.
//
// The existing saleslead package is retained for compatibility. These adapters
// are not the canonical responsibility of the source SDK. New CRM, billing,
// lead management, scoring, enrichment and business policy belong to consumers.
// The producer's public CompanyProfile and source-package contracts remain
// consumer-neutral; adapters do not use downstream databases or private modules.
package adapters
