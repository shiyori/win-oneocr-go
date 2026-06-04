# 示例

[中文](README.md) | [English](README.en.md)

本目录提供可直接运行的示例，并内嵌一张测试图例，图中文字为 `ONEOCR 123`。

## Go SDK 示例

需要先安装 SDK：

```powershell
go get github.com/shiyori/win-oneocr-go
```

运行 Engine 与 RGBA 示例：

```powershell
go run ./examples/engine
go run ./examples/rgba
```

导出内嵌测试图例：

```powershell
go run ./examples/write_sample --out oneocr-sample.png
```

## CLI / Serve 示例

需要先安装或构建命令行工具：

```powershell
go install github.com/shiyori/win-oneocr-go/cmd/oneocr@latest
```

安装后可直接使用单图命令：

```powershell
oneocr image .\oneocr-sample.png --json
```

`serve_client` 示例需要指定可执行文件路径，会启动 `oneocr serve` 并发送内嵌测试图例的 `image_bytes` 请求：

```powershell
go build -o oneocr.exe ./cmd/oneocr
go run ./examples/serve_client --oneocr .\oneocr.exe
```

所有 OCR 示例都需要 Windows 和本机已安装的截图工具 OneOCR 运行时。
