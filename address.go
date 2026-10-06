package houjin

import "strings"

// Address carries upstream address strings; Full concatenates them without
// normalization or authoritative geographic-code inference.
type Address struct {
	PostalCode string

	PrefectureName string
	CityName       string
	StreetNumber   string

	AddressImageID string
}

func (a Address) Full() string {
	parts := []string{a.PrefectureName, a.CityName, a.StreetNumber}
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(p)
	}
	return b.String()
}
