package oneocr

import (
	"fmt"
	"image"
	"image/draw"
)

func imageToRGBA(img image.Image) ([]byte, int, int, int, error) {
	if img == nil {
		return nil, 0, 0, 0, nil
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, 0, 0, 0, nil
	}
	switch src := img.(type) {
	case *image.RGBA:
		offset := src.PixOffset(bounds.Min.X, bounds.Min.Y)
		if offset >= 0 && offset < len(src.Pix) {
			return src.Pix[offset:], width, height, src.Stride, nil
		}
	case *image.NRGBA:
		offset := src.PixOffset(bounds.Min.X, bounds.Min.Y)
		if offset >= 0 && offset < len(src.Pix) {
			return src.Pix[offset:], width, height, src.Stride, nil
		}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	return rgba.Pix, width, height, rgba.Stride, nil
}

func validateRGBA(width, height, stride int, rgba []byte) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid image size %dx%d", width, height)
	}
	if stride <= 0 {
		stride = width * 4
	}
	if stride < width*4 {
		return fmt.Errorf("invalid RGBA stride %d for width %d", stride, width)
	}
	required := (height-1)*stride + width*4
	if required <= 0 || len(rgba) < required {
		return fmt.Errorf("invalid RGBA data length: got %d want at least %d", len(rgba), required)
	}
	return nil
}
