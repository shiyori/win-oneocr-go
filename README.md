# win-oneocr-go

[中文](README.md) | [English](README.en.md)

Windows 截图工具 OneOCR 运行时的 Go SDK 和命令行工具。

## Go 集成

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

提示：如果宿主进程可能已经加载了冲突版 `onnxruntime.dll`，可参考下方“可选子进程隔离”，由调用方自行启动子进程隔离 OneOCR。

如果已经有紧密排列的 RGBA 图像 bytes，可以调用 `RecognizeRGBA`：

```go
result, err := engine.RecognizeRGBA(ctx, width, height, rgbaBytes)
```

如果每行存在 padding，可以使用显式 stride 版本：

```go
result, err := engine.RecognizeRGBAWithStride(ctx, width, height, stride, rgbaBytes)
```

### gocv.Mat

`oneocrgocv.RecognizeMat` 面向已经使用 `gocv.Mat` 的调用方。调用方项目已有 gocv/OpenCV 环境即可，核心 SDK 不依赖 OpenCV，也不要求只为本项目额外安装 gocv。

```go
// go build -tags gocv
result, err := oneocrgocv.RecognizeMat(ctx, engine, mat)
```

更多可运行示例和内嵌测试图例见 [examples](examples/README.md)。

## 可选子进程隔离

如果宿主进程可能已经加载了冲突的 `onnxruntime.dll`，调用方可以自行用子进程方式隔离 OneOCR。仓库提供 `cmd/oneocr` 命令行工具，但 SDK 不内置 helper、daemon 或隐藏 IPC。

可从 GitHub Releases 下载独立 `oneocr.exe`，也可以通过 Go 安装：

单图命令适合低频调用：

```powershell
go install github.com/shiyori/win-oneocr-go/cmd/oneocr@latest

oneocr image .\sample.png --json
```

高频调用可以使用常驻服务复用 OneOCR 初始化结果：

```powershell
oneocr serve
```

`oneocr serve` 从 stdin 读取请求 frame，向 stdout 写响应 frame；stderr 只用于日志。每个 frame 为：

```text
uint32 little-endian headerLength
JSON header bytes
optional raw payload bytes
```

路径请求不带 payload：

```json
{"id":"1","type":"recognize","input":"path","path":"sample.png"}
```

图片 bytes 请求后接 PNG/JPEG/GIF 原始 bytes：

```json
{"id":"2","type":"recognize","input":"image_bytes","payloadLength":12345}
```

RGBA bytes 请求后接 raw RGBA bytes；`stride` 可省略，默认 `width * 4`：

```json
{"id":"3","type":"recognize","input":"rgba","width":800,"height":600,"stride":3200,"payloadLength":1920000}
```

响应 frame 只包含 JSON：

```json
{"id":"1","ok":true,"text":"recognized text","lines":[],"durationMs":12}
```

控制请求支持 `ping` 和 `shutdown`：

```json
{"id":"p","type":"ping"}
{"id":"s","type":"shutdown"}
```

## 许可证

本仓库使用 `AGPL-3.0-only` 协议。

## 说明

本项目不重新分发微软文件，只读取用户本机已安装 Windows 截图工具中的 `oneocr.dll`、`oneocr.onemodel` 和 `onnxruntime.dll`。默认扫描路径为 `C:\Program Files\WindowsApps\Microsoft.ScreenSketch_*\SnippingTool`。
