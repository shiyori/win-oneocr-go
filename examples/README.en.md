# Examples

[中文](README.md) | [English](README.en.md)

This directory contains runnable examples and an embedded sample image with the text `ONEOCR 123`.

## Go SDK Examples

Install the SDK first:

```powershell
go get github.com/shiyori/win-oneocr-go
```

Run the Engine and RGBA examples:

```powershell
go run ./examples/engine
go run ./examples/rgba
```

Export the embedded sample image:

```powershell
go run ./examples/write_sample --out oneocr-sample.png
```

## CLI / Serve Examples

Install or build the CLI first:

```powershell
go install github.com/shiyori/win-oneocr-go/cmd/oneocr@latest
```

After installation, use the single-image command directly:

```powershell
oneocr image .\oneocr-sample.png --json
```

The `serve_client` example needs an executable path. It starts `oneocr serve` and sends an `image_bytes` request using the embedded sample image:

```powershell
go build -o oneocr.exe ./cmd/oneocr
go run ./examples/serve_client --oneocr .\oneocr.exe
```

All OCR examples require Windows and an installed Snipping Tool OneOCR runtime.
