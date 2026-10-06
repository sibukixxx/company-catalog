package houjin

import "testing"

func TestAddressFullReturnsConcatenatedAddressWhenAllFieldsSet(t *testing.T) {
	addr := Address{
		PrefectureName: "東京都",
		CityName:       "千代田区",
		StreetNumber:   "霞が関３－１－１",
	}

	got := addr.Full()
	want := "東京都千代田区霞が関３－１－１"

	if got != want {
		t.Errorf("Full() = %q, want %q", got, want)
	}
}

func TestAddressFullReturnsEmptyWhenAllFieldsEmpty(t *testing.T) {
	addr := Address{}

	if got := addr.Full(); got != "" {
		t.Errorf("Full() = %q, want empty string", got)
	}
}

func TestAddressFullSkipsEmptyPartsWhenSomeFieldsMissing(t *testing.T) {
	addr := Address{
		PrefectureName: "東京都",
		StreetNumber:   "霞が関３－１－１",
	}

	got := addr.Full()
	want := "東京都霞が関３－１－１"

	if got != want {
		t.Errorf("Full() = %q, want %q", got, want)
	}
}
