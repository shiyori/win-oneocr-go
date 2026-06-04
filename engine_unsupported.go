//go:build !windows

package oneocr

import "context"

func New(_ context.Context, _ ...Option) (*Engine, error) {
	return nil, ErrUnsupported
}
