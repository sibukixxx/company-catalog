package houjin

import (
	"fmt"
)

const corporateNumberLength = 13

// ValidateCorporateNumber checks 13 numeric digits and the NTA check digit.
// It performs no network lookup and cannot prove the number has been assigned.
func ValidateCorporateNumber(s string) error {
	if len(s) != corporateNumberLength {
		return newInvalidInputError("corporate number must be %d digits, got %d", corporateNumberLength, len(s))
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return newInvalidInputError("corporate number must contain only digits")
		}
	}

	want, err := computeCheckDigit(s[1:])
	if err != nil {
		return err
	}
	got := int(s[0] - '0')
	if got != want {
		return newInvalidInputError("invalid check digit: got %d, want %d", got, want)
	}
	return nil
}

// computeCheckDigit は基礎番号（12桁）からチェックディジットを算出する。
//
// Why: 国税庁の法人番号仕様で定義された式
//
//	C = 9 - (Σ(P_n × Q_n) mod 9)
//
// Q_n は右から数えて偶数桁=2、奇数桁=1。
func computeCheckDigit(base string) (int, error) {
	if len(base) != corporateNumberLength-1 {
		return 0, newInvalidInputError("base number must be %d digits", corporateNumberLength-1)
	}
	sum := 0
	n := len(base)
	for i := range n {
		ch := base[i]
		if ch < '0' || ch > '9' {
			return 0, newInvalidInputError("base number must contain only digits")
		}
		digit := int(ch - '0')
		// 右から数えた位置 = n - i
		fromRight := n - i
		q := 1
		if fromRight%2 == 0 {
			q = 2
		}
		sum += digit * q
	}
	return 9 - (sum % 9), nil
}

func newInvalidInputError(format string, args ...any) *APIError {
	return &APIError{
		Code:    ErrorCodeInvalidInput,
		Message: fmt.Sprintf(format, args...),
	}
}
