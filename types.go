package oneocr

import "errors"

const (
	// DefaultWindowsAppsDir is the default Microsoft Store app root that contains ScreenSketch packages.
	DefaultWindowsAppsDir = `C:\Program Files\WindowsApps`
)

var (
	ErrUnsupported = errors.New("oneocr is only available on Windows")
	ErrUnavailable = errors.New("oneocr runtime is unavailable")
)

type Installation struct {
	Available       bool     `json:"available"`
	Path            string   `json:"path,omitempty"`
	Version         string   `json:"version,omitempty"`
	Message         string   `json:"message,omitempty"`
	RequiredFiles   []string `json:"requiredFiles,omitempty"`
	MissingFiles    []string `json:"missingFiles,omitempty"`
	PackageName     string   `json:"packageName,omitempty"`
	WindowsAppsRoot string   `json:"windowsAppsRoot,omitempty"`
}

type Point struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type BoundingBox struct {
	TopLeft     Point `json:"topLeft"`
	TopRight    Point `json:"topRight"`
	BottomRight Point `json:"bottomRight"`
	BottomLeft  Point `json:"bottomLeft"`
}

type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Word struct {
	Text        string      `json:"text"`
	BoundingBox BoundingBox `json:"boundingBox"`
	Rect        Rect        `json:"rect"`
	Confidence  float32     `json:"confidence"`
}

type Line struct {
	Text        string      `json:"text"`
	BoundingBox BoundingBox `json:"boundingBox"`
	Rect        Rect        `json:"rect"`
	Confidence  float32     `json:"confidence"`
	Words       []Word      `json:"words,omitempty"`
}

type Result struct {
	Text  string `json:"text"`
	Lines []Line `json:"lines"`
}
