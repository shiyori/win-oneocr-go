//go:build windows

package oneocr

import "testing"

func TestSafeCStringCopiesBytes(t *testing.T) {
	buf := []byte{'h', 'e', 'l', 'l', 'o', 0}
	got := safeCString(&buf[0])
	buf[1] = 'a'
	if got != "hello" {
		t.Fatalf("safeCString = %q, want hello", got)
	}
}

func TestSafeCStringNil(t *testing.T) {
	if got := safeCString(nil); got != "" {
		t.Fatalf("safeCString(nil) = %q, want empty", got)
	}
}
