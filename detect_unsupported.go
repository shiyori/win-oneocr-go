//go:build !windows

package oneocr

import "context"

func Detect(_ context.Context, _ ...Option) (Installation, error) {
	return Installation{Available: false, Message: ErrUnsupported.Error()}, ErrUnsupported
}
