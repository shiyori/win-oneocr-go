package oneocr

import "testing"

func TestValidateRGBA(t *testing.T) {
	if err := validateRGBA(2, 2, 8, make([]byte, 16)); err != nil {
		t.Fatalf("validateRGBA valid data returned %v", err)
	}
	if err := validateRGBA(2, 2, 7, make([]byte, 16)); err == nil {
		t.Fatalf("expected invalid stride error")
	}
	if err := validateRGBA(2, 2, 8, make([]byte, 15)); err == nil {
		t.Fatalf("expected invalid length error")
	}
	if err := validateRGBA(0, 2, 8, make([]byte, 16)); err == nil {
		t.Fatalf("expected invalid size error")
	}
}
