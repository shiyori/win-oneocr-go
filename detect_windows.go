//go:build windows

package oneocr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	screenSketchPrefix = "Microsoft.ScreenSketch_"
	screenSketchSuffix = "_8wekyb3d8bbwe"
)

var (
	requiredRuntimeFiles = []string{"oneocr.dll", "oneocr.onemodel", "onnxruntime.dll"}
	versionPattern       = regexp.MustCompile(`^Microsoft\.ScreenSketch_([0-9]+(?:\.[0-9]+)*)_`)
)

func Detect(ctx context.Context, opts ...Option) (Installation, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Installation{}, err
	}
	cfg := applyOptions(opts)
	if strings.TrimSpace(cfg.snippingToolDir) != "" {
		return detectManualSnippingToolDir(cfg.snippingToolDir)
	}
	return detectInRoot(ctx, cfg.windowsAppsRoot)
}

func detectManualSnippingToolDir(dir string) (Installation, error) {
	dir = filepath.Clean(strings.TrimSpace(dir))
	if dir == "." || dir == "" {
		return Installation{Available: false, Message: "SnippingTool directory is empty", RequiredFiles: requiredRuntimeFiles}, ErrUnavailable
	}
	if runtimeFilesExist(filepath.Join(dir, "SnippingTool")).Available {
		dir = filepath.Join(dir, "SnippingTool")
	}
	status := runtimeFilesExist(dir)
	status.Path = dir
	status.Message = "manual SnippingTool OneOCR runtime"
	if !status.Available {
		status.Message = "manual SnippingTool OneOCR runtime is missing required files"
		return status, ErrUnavailable
	}
	return status, nil
}

func detectInRoot(ctx context.Context, root string) (Installation, error) {
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "." || root == "" {
		status := Installation{Available: false, WindowsAppsRoot: root, Message: "WindowsApps directory is empty", RequiredFiles: requiredRuntimeFiles}
		return status, ErrUnavailable
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		status := Installation{Available: false, WindowsAppsRoot: root, Message: fmt.Sprintf("cannot scan WindowsApps: %v", err), RequiredFiles: requiredRuntimeFiles}
		return status, err
	}
	candidates := make([]runtimeCandidate, 0)
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return Installation{}, err
		}
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, screenSketchPrefix) || !strings.HasSuffix(name, screenSketchSuffix) {
			continue
		}
		dir := filepath.Join(root, name, "SnippingTool")
		status := runtimeFilesExist(dir)
		if !status.Available {
			continue
		}
		candidates = append(candidates, runtimeCandidate{
			path:        dir,
			version:     packageVersion(name),
			packageName: name,
		})
	}
	if len(candidates) == 0 {
		status := Installation{
			Available:       false,
			WindowsAppsRoot: root,
			Message:         "Windows Snipping Tool OneOCR runtime files were not found",
			RequiredFiles:   requiredRuntimeFiles,
		}
		return status, ErrUnavailable
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return compareVersions(candidates[i].version, candidates[j].version) > 0
	})
	selected := candidates[0]
	return Installation{
		Available:       true,
		Path:            selected.path,
		Version:         selected.version,
		Message:         "Windows Snipping Tool OneOCR runtime is available",
		RequiredFiles:   requiredRuntimeFiles,
		PackageName:     selected.packageName,
		WindowsAppsRoot: root,
	}, nil
}

type runtimeCandidate struct {
	path        string
	version     string
	packageName string
}

func runtimeFilesExist(dir string) Installation {
	missing := make([]string, 0)
	for _, name := range requiredRuntimeFiles {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.IsDir() {
			missing = append(missing, name)
		}
	}
	return Installation{
		Available:     len(missing) == 0,
		Path:          dir,
		RequiredFiles: requiredRuntimeFiles,
		MissingFiles:  missing,
	}
}

func packageVersion(name string) string {
	matches := versionPattern.FindStringSubmatch(name)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func compareVersions(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	maxLen := len(leftParts)
	if len(rightParts) > maxLen {
		maxLen = len(rightParts)
	}
	for i := 0; i < maxLen; i++ {
		leftValue, rightValue := 0, 0
		if i < len(leftParts) {
			leftValue, _ = strconv.Atoi(leftParts[i])
		}
		if i < len(rightParts) {
			rightValue, _ = strconv.Atoi(rightParts[i])
		}
		if leftValue > rightValue {
			return 1
		}
		if leftValue < rightValue {
			return -1
		}
	}
	return 0
}
