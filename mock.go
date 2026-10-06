package houjin

import "context"

// MockClient は Client interface の関数フィールド型モック。
// 利用側のユースケーステストで、必要な関数だけを設定して使う。
type MockClient struct {
	LookupByCorporateNumberFn func(ctx context.Context, corporateNumber string) (*CompanyProfile, error)
	SearchByNameFn            func(ctx context.Context, input SearchByNameInput) ([]CompanyProfile, error)
	FetchUpdatesByDateRangeFn func(ctx context.Context, input FetchUpdatesInput) ([]CompanyProfile, error)
}

func (m *MockClient) LookupByCorporateNumber(ctx context.Context, corporateNumber string) (*CompanyProfile, error) {
	if m.LookupByCorporateNumberFn == nil {
		return nil, &APIError{Code: ErrorCodeAPIError, Message: "MockClient.LookupByCorporateNumberFn is not set"}
	}
	return m.LookupByCorporateNumberFn(ctx, corporateNumber)
}

func (m *MockClient) SearchByName(ctx context.Context, input SearchByNameInput) ([]CompanyProfile, error) {
	if m.SearchByNameFn == nil {
		return nil, &APIError{Code: ErrorCodeAPIError, Message: "MockClient.SearchByNameFn is not set"}
	}
	return m.SearchByNameFn(ctx, input)
}

func (m *MockClient) FetchUpdatesByDateRange(ctx context.Context, input FetchUpdatesInput) ([]CompanyProfile, error) {
	if m.FetchUpdatesByDateRangeFn == nil {
		return nil, &APIError{Code: ErrorCodeAPIError, Message: "MockClient.FetchUpdatesByDateRangeFn is not set"}
	}
	return m.FetchUpdatesByDateRangeFn(ctx, input)
}
