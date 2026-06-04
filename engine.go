package oneocr

import (
	"context"
	"image"
	"sync"
)

type Engine struct {
	installation Installation
	impl         engineImpl
	once         sync.Once
	closeErr     error
}

type engineImpl interface {
	RecognizeRGBA(ctx context.Context, width, height, stride int, rgba []byte) (*Result, error)
	Close() error
}

func (e *Engine) Installation() Installation {
	if e == nil {
		return Installation{}
	}
	return e.installation
}

func (e *Engine) Recognize(ctx context.Context, img image.Image) (*Result, error) {
	rgba, width, height, stride, err := imageToRGBA(img)
	if err != nil {
		return nil, err
	}
	if width <= 0 || height <= 0 {
		return &Result{}, nil
	}
	return e.RecognizeRGBAWithStride(ctx, width, height, stride, rgba)
}

func (e *Engine) RecognizeRGBA(ctx context.Context, width, height int, rgba []byte) (*Result, error) {
	return e.RecognizeRGBAWithStride(ctx, width, height, width*4, rgba)
}

func (e *Engine) RecognizeRGBAWithStride(ctx context.Context, width, height, stride int, rgba []byte) (*Result, error) {
	if e == nil || e.impl == nil {
		return nil, ErrUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateRGBA(width, height, stride, rgba); err != nil {
		return nil, err
	}
	return e.impl.RecognizeRGBA(ctx, width, height, stride, rgba)
}

func (e *Engine) Close() error {
	if e == nil {
		return nil
	}
	e.once.Do(func() {
		if e.impl != nil {
			e.closeErr = e.impl.Close()
		}
	})
	return e.closeErr
}
