package houjin

import (
	"errors"
	"testing"
)

func TestValidateCorporateNumberReturnsNilWhenValid(t *testing.T) {
	// Toyota Motor Corp の法人番号（実在の番号でチェックディジット = 1）
	if err := ValidateCorporateNumber("1180301018771"); err != nil {
		t.Errorf("ValidateCorporateNumber returned error for valid number: %v", err)
	}
}

func TestValidateCorporateNumberReturnsErrorWhenLengthIsNot13(t *testing.T) {
	cases := []string{"", "123", "12345678901234"}
	for _, c := range cases {
		if err := ValidateCorporateNumber(c); err == nil {
			t.Errorf("ValidateCorporateNumber(%q) should fail on length", c)
		}
	}
}

func TestValidateCorporateNumberReturnsErrorWhenContainsNonDigit(t *testing.T) {
	if err := ValidateCorporateNumber("118030101877a"); err == nil {
		t.Error("ValidateCorporateNumber should fail on non-digit input")
	}
}

func TestValidateCorporateNumberReturnsErrorWhenCheckDigitIsWrong(t *testing.T) {
	// 末尾を改竄: 1180301018771 -> 9180301018771 (先頭=チェックディジット)
	if err := ValidateCorporateNumber("9180301018771"); err == nil {
		t.Error("ValidateCorporateNumber should fail on wrong check digit")
	}
}

func TestValidateCorporateNumberReturnsAPIErrorWithInvalidInputCode(t *testing.T) {
	cases := map[string]string{
		"length too short":   "123",
		"length too long":    "12345678901234",
		"contains non-digit": "118030101877a",
		"wrong check digit":  "9180301018771",
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateCorporateNumber(input)
			if err == nil {
				t.Fatalf("ValidateCorporateNumber(%q) returned nil, want error", input)
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("ValidateCorporateNumber(%q) returned %T, want *APIError", input, err)
			}
			if apiErr.Code != ErrorCodeInvalidInput {
				t.Errorf("APIError.Code = %q, want %q", apiErr.Code, ErrorCodeInvalidInput)
			}
		})
	}
}

func TestComputeCheckDigitMatchesKnownValue(t *testing.T) {
	// 基礎番号 180301018771 → チェックディジット 1
	got, err := computeCheckDigit("180301018771")
	if err != nil {
		t.Fatalf("computeCheckDigit returned error: %v", err)
	}
	if got != 1 {
		t.Errorf("computeCheckDigit = %d, want 1", got)
	}
}

// computeCheckDigit と ValidateCorporateNumber の round-trip 性質検証
// 任意の 12 桁基礎番号 base について、computeCheckDigit(base) で得た c を
// 先頭に付けた 13 桁番号 c+base は ValidateCorporateNumber を必ず通る。
// → 「アルゴリズム実装が国税庁仕様と内部整合している」ことを示す property-based test
func TestValidateCorporateNumberAcceptsAllNumbersBuiltByComputeCheckDigit(t *testing.T) {
	bases := []string{
		"180301018771", // Toyota の基礎番号 (チェックディジット = 1)
		"000000000001",
		"123456789012",
		"999999999999",
		"100000000000",
		"010101010101",
	}

	for _, base := range bases {
		t.Run(base, func(t *testing.T) {
			c, err := computeCheckDigit(base)
			if err != nil {
				t.Fatalf("computeCheckDigit(%q): %v", base, err)
			}
			full := byte('0' + c) // c は 0..9 の範囲
			number := string(full) + base
			if err := ValidateCorporateNumber(number); err != nil {
				t.Errorf("ValidateCorporateNumber(%q) failed: %v", number, err)
			}
		})
	}
}

// computeCheckDigit が範囲外（13 桁、11 桁、空文字）を弾くことを検証
func TestComputeCheckDigitRejectsBaseOfWrongLength(t *testing.T) {
	cases := []string{"", "1", "12345678901", "1234567890123"}
	for _, base := range cases {
		t.Run("base="+base, func(t *testing.T) {
			if _, err := computeCheckDigit(base); err == nil {
				t.Errorf("computeCheckDigit(%q) returned nil, want error", base)
			}
		})
	}
}

// computeCheckDigit が非数字を弾くことを検証
func TestComputeCheckDigitRejectsNonDigits(t *testing.T) {
	if _, err := computeCheckDigit("18030101877a"); err == nil {
		t.Error("computeCheckDigit should fail on non-digit input")
	}
}
