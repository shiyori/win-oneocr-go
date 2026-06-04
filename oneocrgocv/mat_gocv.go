//go:build gocv

package oneocrgocv

import (
	"context"
	"fmt"

	oneocr "github.com/shiyori/win-oneocr-go"
	"gocv.io/x/gocv"
)

func RecognizeMat(ctx context.Context, engine *oneocr.Engine, mat gocv.Mat) (*oneocr.Result, error) {
	if engine == nil {
		return nil, oneocr.ErrUnavailable
	}
	if mat.Empty() {
		return &oneocr.Result{}, nil
	}
	channels := mat.Channels()
	if channels != 4 {
		return nil, fmt.Errorf("oneocrgocv requires RGBA Mat with 4 channels, got %d", channels)
	}
	width, height := mat.Cols(), mat.Rows()
	bytes, err := mat.DataPtrUint8()
	if err != nil {
		return nil, err
	}
	return engine.RecognizeRGBAWithStride(ctx, width, height, matStride(mat, width), bytes)
}

func matStride(mat gocv.Mat, width int) int {
	if step := mat.Step(); step > 0 {
		return step
	}
	return width * 4
}
