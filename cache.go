package houjin

import "context"

// CompanyCache is a caller-supplied cache boundary. Only Lookup currently uses
// Get/Set; name search is reserved. Implementations shared by a Client must be
// safe for concurrent use. Cache errors do not replace successful API responses.
type CompanyCache interface {
	GetByCorporateNumber(
		ctx context.Context,
		corporateNumber string,
	) (*CompanyProfile, bool, error)

	SetByCorporateNumber(
		ctx context.Context,
		company CompanyProfile,
	) error

	SearchByNormalizedName(
		ctx context.Context,
		normalizedName string,
		limit int,
	) ([]CompanyProfile, error)
}
