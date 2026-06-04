# win-oneocr-go

[中文](README.md) | [English](README.en.md)

Go SDK and CLI for the Windows Snipping Tool OneOCR runtime.

## Go Integration

```powershell
go get github.com/shiyori/win-oneocr-go
```

```go
package main

import (
	"context"
	"fmt"

	oneocr "github.com/shiyori/win-oneocr-go"
)

func main() {
	ctx := context.Background()
	engine, err := oneocr.New(ctx)
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	result, err := engine.Recognize(ctx, img)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Text)
}
```

Tip: if the host process may already have a conflicting `onnxruntime.dll` loaded, see "Optional Subprocess Isolation" below and let the caller isolate OneOCR in its own subprocess.

For tightly packed RGBA image bytes, call `RecognizeRGBA`:

```go
result, err := engine.RecognizeRGBA(ctx, width, height, rgbaBytes)
```

If each row has padding, use the explicit stride variant:

```go
result, err := engine.RecognizeRGBAWithStride(ctx, width, height, stride, rgbaBytes)
```

### gocv.Mat

`oneocrgocv.RecognizeMat` is intended for callers that already use `gocv.Mat`. If your project already has a working gocv/OpenCV setup, no extra install step is needed for this package. The core SDK does not depend on OpenCV.

```go
// go build -tags gocv
result, err := oneocrgocv.RecognizeMat(ctx, engine, mat)
```

See [examples](examples/README.md) for runnable examples and an embedded sample image.

## Optional Subprocess Isolation

If the host process may already have a conflicting `onnxruntime.dll` loaded, callers can isolate OneOCR in their own subprocess. This repository provides the `cmd/oneocr` CLI, but the SDK does not include a helper, daemon, or hidden IPC layer.

Download the standalone `oneocr.exe` from GitHub Releases, or install it with Go:

Use the single-image command for low-frequency calls:

```powershell
go install github.com/shiyori/win-oneocr-go/cmd/oneocr@latest

oneocr image .\sample.png --json
```

Use the persistent service for high-frequency calls that should reuse OneOCR initialization:

```powershell
oneocr serve
```

`oneocr serve` reads request frames from stdin and writes response frames to stdout; stderr is reserved for logs. Each frame is:

```text
uint32 little-endian headerLength
JSON header bytes
optional raw payload bytes
```

Path requests do not include a payload:

```json
{"id":"1","type":"recognize","input":"path","path":"sample.png"}
```

Image bytes requests are followed by raw PNG/JPEG/GIF bytes:

```json
{"id":"2","type":"recognize","input":"image_bytes","payloadLength":12345}
```

RGBA bytes requests are followed by raw RGBA bytes; `stride` may be omitted and defaults to `width * 4`:

```json
{"id":"3","type":"recognize","input":"rgba","width":800,"height":600,"stride":3200,"payloadLength":1920000}
```

Response frames contain JSON only:

```json
{"id":"1","ok":true,"text":"recognized text","lines":[],"durationMs":12}
```

Control requests support `ping` and `shutdown`:

```json
{"id":"p","type":"ping"}
{"id":"s","type":"shutdown"}
```

## License

This repository is licensed under `AGPL-3.0-only`.

## Notes

This project does not redistribute Microsoft files. It reads `oneocr.dll`, `oneocr.onemodel`, and `onnxruntime.dll` from an installed Windows Snipping Tool package under `C:\Program Files\WindowsApps\Microsoft.ScreenSketch_*\SnippingTool`.
