package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"github.com/shiyori/win-oneocr-go"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type cliError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type cliSource struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

type cliResponse struct {
	OK         bool                `json:"ok"`
	Runtime    oneocr.Installation `json:"runtime,omitempty"`
	Source     *cliSource          `json:"source,omitempty"`
	Text       string              `json:"text,omitempty"`
	Lines      []oneocr.Line       `json:"lines,omitempty"`
	DurationMs int64               `json:"durationMs,omitempty"`
	Error      *cliError           `json:"error,omitempty"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}
	switch args[0] {
	case "detect":
		return runDetect(args[1:])
	case "image":
		return runImage(args[1:])
	case "serve":
		return runServe(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		writeError("usage", fmt.Sprintf("unknown command: %s", args[0]))
		return 2
	}
}

func runDetect(args []string) int {
	flags := flag.NewFlagSet("detect", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	snippingTool := flags.String("snipping-tool", "", "SnippingTool directory")
	_ = flags.Bool("json", true, "output JSON")
	if err := flags.Parse(args); err != nil {
		writeError("usage", err.Error())
		return 2
	}
	opts := optionsFromFlags(*snippingTool)
	runtime, err := oneocr.Detect(context.Background(), opts...)
	resp := cliResponse{OK: err == nil && runtime.Available, Runtime: runtime}
	if err != nil {
		resp.Error = &cliError{Code: "detect_failed", Message: err.Error()}
	}
	return writeJSON(resp, err == nil)
}

func runImage(args []string) int {
	flagArgs, imagePath, err := splitImageArgs(args)
	if err != nil {
		writeError("usage", err.Error())
		return 2
	}
	flags := flag.NewFlagSet("image", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	snippingTool := flags.String("snipping-tool", "", "SnippingTool directory")
	_ = flags.Bool("json", true, "output JSON")
	if err := flags.Parse(flagArgs); err != nil {
		writeError("usage", err.Error())
		return 2
	}
	if imagePath == "" {
		writeError("usage", "image command requires exactly one image path")
		return 2
	}
	started := time.Now()
	opts := optionsFromFlags(*snippingTool)
	ctx := context.Background()
	result, err := recognizeImageFile(ctx, imagePath, opts...)
	runtime, _ := oneocr.Detect(context.Background(), optionsFromFlags(*snippingTool)...)
	resp := cliResponse{
		OK:         err == nil,
		Runtime:    runtime,
		Source:     &cliSource{Type: "image", Path: imagePath},
		DurationMs: time.Since(started).Milliseconds(),
	}
	if result != nil {
		resp.Text = result.Text
		resp.Lines = result.Lines
	}
	if err != nil {
		resp.Error = &cliError{Code: "recognize_failed", Message: err.Error()}
	}
	return writeJSON(resp, err == nil)
}

func recognizeImageFile(ctx context.Context, path string, opts ...oneocr.Option) (*oneocr.Result, error) {
	img, err := decodeImageFile(path)
	if err != nil {
		return nil, err
	}
	engine, err := oneocr.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	defer engine.Close()
	return engine.Recognize(ctx, img)
}

func decodeImageFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}

func splitImageArgs(args []string) ([]string, string, error) {
	flagArgs := make([]string, 0, len(args))
	imagePath := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			if flagConsumesNextValue(arg) {
				if i+1 >= len(args) {
					return nil, "", fmt.Errorf("flag needs an argument: %s", arg)
				}
				i++
				flagArgs = append(flagArgs, args[i])
			}
			continue
		}
		if imagePath != "" {
			return nil, "", fmt.Errorf("image command requires exactly one image path")
		}
		imagePath = arg
	}
	return flagArgs, imagePath, nil
}

func flagConsumesNextValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	switch arg {
	case "--snipping-tool", "-snipping-tool":
		return true
	default:
		return false
	}
}

func optionsFromFlags(snippingTool string) []oneocr.Option {
	opts := make([]oneocr.Option, 0, 1)
	if strings.TrimSpace(snippingTool) != "" {
		opts = append(opts, oneocr.WithSnippingToolDir(snippingTool))
	}
	return opts
}

func writeJSON(resp cliResponse, success bool) int {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if success {
		return 0
	}
	return 1
}

func writeError(code, message string) {
	_ = json.NewEncoder(os.Stdout).Encode(cliResponse{
		OK:    false,
		Error: &cliError{Code: code, Message: message},
	})
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  oneocr detect [--json] [--snipping-tool DIR]
  oneocr image <path> [--json] [--snipping-tool DIR]
  oneocr serve [--snipping-tool DIR]`)
}
