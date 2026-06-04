package oneocr

import (
	"context"
	"image"
	"testing"
)

type fakeEngineImpl struct {
	width  int
	height int
	stride int
	rgba   []byte
}

func (f *fakeEngineImpl) RecognizeRGBA(_ context.Context, width, height, stride int, rgba []byte) (*Result, error) {
	f.width = width
	f.height = height
	f.stride = stride
	f.rgba = rgba
	return &Result{Text: "ok"}, nil
}

func (f *fakeEngineImpl) Close() error {
	return nil
}

func TestEngineRecognizeRGBAPassesImageBytes(t *testing.T) {
	impl := &fakeEngineImpl{}
	engine := &Engine{impl: impl}
	rgba := []byte{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	result, err := engine.RecognizeRGBA(context.Background(), 2, 2, rgba)
	if err != nil {
		t.Fatalf("RecognizeRGBA returned error: %v", err)
	}
	if result.Text != "ok" {
		t.Fatalf("Text = %q, want ok", result.Text)
	}
	if impl.width != 2 || impl.height != 2 || impl.stride != 8 {
		t.Fatalf("received geometry = %dx%d stride=%d", impl.width, impl.height, impl.stride)
	}
	if len(impl.rgba) != len(rgba) || &impl.rgba[0] != &rgba[0] {
		t.Fatalf("RecognizeRGBA did not pass through original byte slice")
	}
}

func TestEngineRecognizeRGBAWithStridePassesCustomStride(t *testing.T) {
	impl := &fakeEngineImpl{}
	engine := &Engine{impl: impl}
	rgba := make([]byte, 24)
	result, err := engine.RecognizeRGBAWithStride(context.Background(), 2, 2, 12, rgba)
	if err != nil {
		t.Fatalf("RecognizeRGBAWithStride returned error: %v", err)
	}
	if result.Text != "ok" {
		t.Fatalf("Text = %q, want ok", result.Text)
	}
	if impl.width != 2 || impl.height != 2 || impl.stride != 12 {
		t.Fatalf("received geometry = %dx%d stride=%d", impl.width, impl.height, impl.stride)
	}
}

func TestEngineRecognizeUsesRGBAFastPath(t *testing.T) {
	impl := &fakeEngineImpl{}
	engine := &Engine{impl: impl}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	result, err := engine.Recognize(context.Background(), img)
	if err != nil {
		t.Fatalf("Recognize returned error: %v", err)
	}
	if result.Text != "ok" {
		t.Fatalf("Text = %q, want ok", result.Text)
	}
	if len(impl.rgba) == 0 || &impl.rgba[0] != &img.Pix[0] {
		t.Fatalf("Recognize did not use RGBA Pix fast path")
	}
	if impl.stride != img.Stride {
		t.Fatalf("stride = %d, want %d", impl.stride, img.Stride)
	}
}

func TestEngineRecognizeUsesNRGBAFastPath(t *testing.T) {
	impl := &fakeEngineImpl{}
	engine := &Engine{impl: impl}
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	result, err := engine.Recognize(context.Background(), img)
	if err != nil {
		t.Fatalf("Recognize returned error: %v", err)
	}
	if result.Text != "ok" {
		t.Fatalf("Text = %q, want ok", result.Text)
	}
	if len(impl.rgba) == 0 || &impl.rgba[0] != &img.Pix[0] {
		t.Fatalf("Recognize did not use NRGBA Pix fast path")
	}
	if impl.stride != img.Stride {
		t.Fatalf("stride = %d, want %d", impl.stride, img.Stride)
	}
}
