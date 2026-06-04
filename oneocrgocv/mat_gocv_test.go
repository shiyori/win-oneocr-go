//go:build gocv

package oneocrgocv

import (
	"context"
	"errors"
	"image"
	"testing"

	oneocr "github.com/shiyori/win-oneocr-go"
	"gocv.io/x/gocv"
)

func TestRecognizeMatRejectsNilEngine(t *testing.T) {
	mat := gocv.NewMat()
	defer mat.Close()
	_, err := RecognizeMat(context.Background(), nil, mat)
	if !errors.Is(err, oneocr.ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
}

func TestRecognizeMatEmpty(t *testing.T) {
	mat := gocv.NewMat()
	defer mat.Close()
	result, err := RecognizeMat(context.Background(), &oneocr.Engine{}, mat)
	if err != nil {
		t.Fatalf("RecognizeMat empty returned error: %v", err)
	}
	if result == nil || result.Text != "" || len(result.Lines) != 0 {
		t.Fatalf("unexpected empty result: %+v", result)
	}
}

func TestMatStrideUsesOpenCVStep(t *testing.T) {
	base := gocv.NewMatWithSize(2, 4, gocv.MatTypeCV8UC4)
	defer base.Close()
	region := base.Region(image.Rect(1, 0, 3, 2))
	defer region.Close()

	if got, want := matStride(region, region.Cols()), region.Step(); got != want {
		t.Fatalf("matStride = %d, want Step %d", got, want)
	}
	if region.Step() == region.Cols()*4 {
		t.Fatalf("test setup expected padded ROI stride, got step=%d width*4=%d", region.Step(), region.Cols()*4)
	}
}
