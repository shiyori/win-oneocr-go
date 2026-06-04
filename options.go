package oneocr

import (
	"strings"
)

type Option func(*options)

type options struct {
	snippingToolDir string
	windowsAppsRoot string
}

func defaultOptions() options {
	return options{
		windowsAppsRoot: DefaultWindowsAppsDir,
	}
}

func WithSnippingToolDir(dir string) Option {
	return func(o *options) {
		o.snippingToolDir = strings.TrimSpace(dir)
	}
}

func WithWindowsAppsRoot(root string) Option {
	return func(o *options) {
		o.windowsAppsRoot = strings.TrimSpace(root)
	}
}

func applyOptions(opts []Option) options {
	cfg := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.windowsAppsRoot == "" {
		cfg.windowsAppsRoot = DefaultWindowsAppsDir
	}
	return cfg
}
