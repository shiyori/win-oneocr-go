//go:build windows

package oneocr

import (
	"context"
)

func New(ctx context.Context, opts ...Option) (*Engine, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	installation, err := Detect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	impl, err := newInProcessEngine(installation.Path)
	if err != nil {
		return nil, err
	}
	return &Engine{installation: installation, impl: impl}, nil
}
