package testimage

import "testing"

func TestImageDecodes(t *testing.T) {
	img := Image()
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid sample image bounds: %v", img.Bounds())
	}
}
