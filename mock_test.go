package houjin

import (
	"context"
	"errors"
	"testing"
)

func TestMockClientReturnsConfiguredResponses(t *testing.T) {
	want := &CompanyProfile{CorporateNumber: "1180301018771", Name: "Test Inc."}
	mock := &MockClient{
		LookupByCorporateNumberFn: func(ctx context.Context, n string) (*CompanyProfile, error) {
			if n != "1180301018771" {
				t.Errorf("got %q", n)
			}
			return want, nil
		},
	}

	got, err := mock.LookupByCorporateNumber(context.Background(), "1180301018771")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// 3 メソッドすべてで未設定時に *APIError(ErrorCodeAPIError) を返す契約
// （README「テスト用 Mock」: 設定していない関数を呼ぶと ErrorCodeAPIError が返る）
func TestMockClientReturnsAPIErrorWhenFunctionNotSet(t *testing.T) {
	tests := []struct {
		name string
		call func(*MockClient) error
	}{
		{
			"LookupByCorporateNumber",
			func(m *MockClient) error {
				_, err := m.LookupByCorporateNumber(context.Background(), "1180301018771")
				return err
			},
		},
		{
			"SearchByName",
			func(m *MockClient) error {
				_, err := m.SearchByName(context.Background(), SearchByNameInput{Name: "x"})
				return err
			},
		},
		{
			"FetchUpdatesByDateRange",
			func(m *MockClient) error {
				_, err := m.FetchUpdatesByDateRange(context.Background(), FetchUpdatesInput{FromDate: "a", ToDate: "b"})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(&MockClient{})
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.Code != ErrorCodeAPIError {
				t.Errorf("Code = %q, want %q", apiErr.Code, ErrorCodeAPIError)
			}
		})
	}
}

// 設定された関数が呼ばれ、引数が透過することを検証（SearchByName / FetchUpdatesByDateRange）
func TestMockClientSearchByNamePassesInputThrough(t *testing.T) {
	wantInput := SearchByNameInput{Name: "ABC", PrefectureCode: "13", Limit: 5, MatchMode: NameMatchModeExact}
	called := false
	mock := &MockClient{
		SearchByNameFn: func(_ context.Context, in SearchByNameInput) ([]CompanyProfile, error) {
			called = true
			if in != wantInput {
				t.Errorf("input = %+v, want %+v", in, wantInput)
			}
			return []CompanyProfile{{CorporateNumber: "1180301018771"}}, nil
		},
	}

	results, err := mock.SearchByName(context.Background(), wantInput)
	if err != nil {
		t.Fatalf("SearchByName: %v", err)
	}
	if !called {
		t.Error("SearchByNameFn was not invoked")
	}
	if len(results) != 1 {
		t.Errorf("len(results) = %d, want 1", len(results))
	}
}

func TestMockClientFetchUpdatesByDateRangePassesInputThrough(t *testing.T) {
	wantInput := FetchUpdatesInput{FromDate: "2024-01-01", ToDate: "2024-01-31", CompanyKind: "301"}
	called := false
	mock := &MockClient{
		FetchUpdatesByDateRangeFn: func(_ context.Context, in FetchUpdatesInput) ([]CompanyProfile, error) {
			called = true
			if in != wantInput {
				t.Errorf("input = %+v, want %+v", in, wantInput)
			}
			return nil, nil
		},
	}

	if _, err := mock.FetchUpdatesByDateRange(context.Background(), wantInput); err != nil {
		t.Fatalf("FetchUpdatesByDateRange: %v", err)
	}
	if !called {
		t.Error("FetchUpdatesByDateRangeFn was not invoked")
	}
}

func TestMockClientImplementsClientInterface(t *testing.T) {
	var _ Client = &MockClient{}
}
