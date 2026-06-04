package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"strings"
	"time"

	oneocr "github.com/shiyori/win-oneocr-go"
)

const (
	maxServeHeaderBytes  = 1 << 20
	maxServePayloadBytes = 512 << 20
)

type serveRequest struct {
	ID            string `json:"id,omitempty"`
	Type          string `json:"type"`
	Input         string `json:"input,omitempty"`
	Path          string `json:"path,omitempty"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	Stride        int    `json:"stride,omitempty"`
	PayloadLength int    `json:"payloadLength,omitempty"`
}

type serveResponse struct {
	ID         string        `json:"id,omitempty"`
	Type       string        `json:"type,omitempty"`
	OK         bool          `json:"ok"`
	Text       string        `json:"text,omitempty"`
	Lines      []oneocr.Line `json:"lines,omitempty"`
	DurationMs int64         `json:"durationMs,omitempty"`
	Error      *cliError     `json:"error,omitempty"`
}

type serveRecognizer interface {
	Recognize(ctx context.Context, img image.Image) (*oneocr.Result, error)
	RecognizeRGBA(ctx context.Context, width, height int, rgba []byte) (*oneocr.Result, error)
	RecognizeRGBAWithStride(ctx context.Context, width, height, stride int, rgba []byte) (*oneocr.Result, error)
}

func runServe(args []string) int {
	flags := flagSet("serve")
	snippingTool := flags.String("snipping-tool", "", "SnippingTool directory")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	ctx := context.Background()
	engine, err := oneocr.New(ctx, optionsFromFlags(*snippingTool)...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer engine.Close()
	return serveLoop(ctx, engine, os.Stdin, os.Stdout)
}

func serveLoop(ctx context.Context, engine serveRecognizer, input io.Reader, output io.Writer) int {
	reader := bufio.NewReader(input)
	writer := bufio.NewWriter(output)
	for {
		req, payload, err := readServeRequest(reader)
		if err != nil {
			if err == io.EOF {
				return 0
			}
			_ = writeServeJSONFrame(writer, serveError("", "read_failed", err.Error()))
			return 1
		}
		resp, stop := handleServeRequest(ctx, engine, req, payload)
		if err := writeServeJSONFrame(writer, resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if stop {
			return 0
		}
	}
}

func handleServeRequest(ctx context.Context, engine serveRecognizer, req serveRequest, payload []byte) (serveResponse, bool) {
	switch req.Type {
	case "ping":
		return serveResponse{ID: req.ID, Type: "pong", OK: true}, false
	case "shutdown":
		return serveResponse{ID: req.ID, Type: "shutdown", OK: true}, true
	case "recognize":
		started := time.Now()
		result, err := recognizeServeInput(ctx, engine, req, payload)
		if err != nil {
			resp := serveError(req.ID, "recognize_failed", err.Error())
			resp.DurationMs = time.Since(started).Milliseconds()
			return resp, false
		}
		resp := serveResponse{ID: req.ID, OK: true, DurationMs: time.Since(started).Milliseconds()}
		if result != nil {
			resp.Text = result.Text
			resp.Lines = result.Lines
		}
		return resp, false
	default:
		return serveError(req.ID, "bad_request", "unsupported request type"), false
	}
}

func recognizeServeInput(ctx context.Context, engine serveRecognizer, req serveRequest, payload []byte) (*oneocr.Result, error) {
	if engine == nil {
		return nil, oneocr.ErrUnavailable
	}
	switch req.Input {
	case "path":
		if strings.TrimSpace(req.Path) == "" {
			return nil, fmt.Errorf("path input requires path")
		}
		img, err := decodeImageFile(req.Path)
		if err != nil {
			return nil, err
		}
		return engine.Recognize(ctx, img)
	case "image_bytes":
		if len(payload) == 0 {
			return nil, fmt.Errorf("image_bytes input requires payload")
		}
		img, _, err := image.Decode(bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("decode image bytes: %w", err)
		}
		return engine.Recognize(ctx, img)
	case "rgba":
		if req.Width <= 0 || req.Height <= 0 {
			return nil, fmt.Errorf("rgba input requires positive width and height")
		}
		if len(payload) == 0 {
			return nil, fmt.Errorf("rgba input requires payload")
		}
		if req.Stride > 0 {
			return engine.RecognizeRGBAWithStride(ctx, req.Width, req.Height, req.Stride, payload)
		}
		return engine.RecognizeRGBA(ctx, req.Width, req.Height, payload)
	default:
		return nil, fmt.Errorf("unsupported recognize input: %s", req.Input)
	}
}

func readServeRequest(reader *bufio.Reader) (serveRequest, []byte, error) {
	var size [4]byte
	if _, err := io.ReadFull(reader, size[:]); err != nil {
		return serveRequest{}, nil, err
	}
	headerLength := binary.LittleEndian.Uint32(size[:])
	if headerLength == 0 {
		return serveRequest{}, nil, fmt.Errorf("empty serve header")
	}
	if headerLength > maxServeHeaderBytes {
		return serveRequest{}, nil, fmt.Errorf("serve header too large: %d", headerLength)
	}
	header := make([]byte, int(headerLength))
	if _, err := io.ReadFull(reader, header); err != nil {
		return serveRequest{}, nil, err
	}
	var req serveRequest
	if err := json.Unmarshal(header, &req); err != nil {
		return serveRequest{}, nil, fmt.Errorf("decode serve header: %w", err)
	}
	if req.PayloadLength < 0 {
		return serveRequest{}, nil, fmt.Errorf("negative payload length: %d", req.PayloadLength)
	}
	if req.PayloadLength > maxServePayloadBytes {
		return serveRequest{}, nil, fmt.Errorf("serve payload too large: %d", req.PayloadLength)
	}
	payload := make([]byte, req.PayloadLength)
	if req.PayloadLength > 0 {
		if _, err := io.ReadFull(reader, payload); err != nil {
			return serveRequest{}, nil, err
		}
	}
	return req, payload, nil
}

func writeServeJSONFrame(writer *bufio.Writer, value serveResponse) error {
	if err := writeServeFrame(writer, value, nil); err != nil {
		return err
	}
	return writer.Flush()
}

func writeServeFrame(writer io.Writer, header any, payload []byte) error {
	data, err := json.Marshal(header)
	if err != nil {
		return err
	}
	if len(data) == 0 || len(data) > maxServeHeaderBytes {
		return fmt.Errorf("invalid serve header length: %d", len(data))
	}
	if len(payload) > maxServePayloadBytes {
		return fmt.Errorf("serve payload too large: %d", len(payload))
	}
	var size [4]byte
	binary.LittleEndian.PutUint32(size[:], uint32(len(data)))
	if _, err := writer.Write(size[:]); err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if len(payload) > 0 {
		_, err = writer.Write(payload)
	}
	return err
}

func serveError(id, code, message string) serveResponse {
	return serveResponse{
		ID:    id,
		OK:    false,
		Error: &cliError{Code: code, Message: message},
	}
}

func flagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	return flags
}
