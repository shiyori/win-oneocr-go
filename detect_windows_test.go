//go:build windows

package oneocr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSelectsLatestScreenSketchPackage(t *testing.T) {
	root := t.TempDir()
	writeRuntime(t, filepath.Join(root, "Microsoft.ScreenSketch_11.2409.25.0_x64__8wekyb3d8bbwe", "SnippingTool"))
	writeRuntime(t, filepath.Join(root, "Microsoft.ScreenSketch_11.2602.47.0_x64__8wekyb3d8bbwe", "SnippingTool"))
	if err := os.MkdirAll(filepath.Join(root, "Microsoft.ScreenSketch_99.0.0.0_neutral_split.scale-150_8wekyb3d8bbwe"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Detect(context.Background(), WithWindowsAppsRoot(root))
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if !got.Available {
		t.Fatalf("expected available runtime, got %+v", got)
	}
	if got.Version != "11.2602.47.0" {
		t.Fatalf("Version = %q, want 11.2602.47.0", got.Version)
	}
	wantPath := filepath.Join(root, "Microsoft.ScreenSketch_11.2602.47.0_x64__8wekyb3d8bbwe", "SnippingTool")
	if got.Path != wantPath {
		t.Fatalf("Path = %q, want %q", got.Path, wantPath)
	}
}

func TestDetectManualSnippingToolPathRequiresAllFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "oneocr.dll"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Detect(context.Background(), WithSnippingToolDir(dir))
	if err == nil {
		t.Fatalf("expected error")
	}
	if got.Available {
		t.Fatalf("expected unavailable runtime, got %+v", got)
	}
	if len(got.MissingFiles) != 2 {
		t.Fatalf("MissingFiles = %+v, want two missing files", got.MissingFiles)
	}
}

func writeRuntime(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range requiredRuntimeFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
