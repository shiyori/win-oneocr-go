package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"strings"
	"testing"

	oneocr "github.com/shiyori/win-oneocr-go"
)

type fakeServeEngine struct {
	recognizeCalls int
	rgbaCalls      int
	lastBounds     image.Rectangle
	lastWidth      int
	lastHeight     int
	lastStride     int
	lastRGBA       []byte
}

func (f *fakeServeEngine) Recognize(_ context.Context, img image.Image) (*oneocr.Result, error) {
	f.recognizeCalls++
	f.lastBounds = img.Bounds()
	return &oneocr.Result{Text: "image ok"}, nil
}

func (f *fakeServeEngine) RecognizeRGBA(ctx context.Context, width, height int, rgba []byte) (*oneocr.Result, error) {
	return f.RecognizeRGBAWithStride(ctx, width, height, width*4, rgba)
}

func (f *fakeServeEngine) RecognizeRGBAWithStride(_ context.Context, width, height, stride int, rgba []byte) (*oneocr.Result, error) {
	f.rgbaCalls++
	f.lastWidth = width
	f.lastHeight = height
	f.lastStride = stride
	f.lastRGBA = rgba
	return &oneocr.Result{Text: "rgba ok"}, nil
}

func TestServeFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	wantPayload := []byte{1, 2, 3}
	wantReq := serveRequest{ID: "1", Type: "recognize", Input: "rgba", Width: 1, Height: 1, PayloadLength: len(wantPayload)}
	if err := writeServeFrame(&buf, wantReq, wantPayload); err != nil {
		t.Fatalf("writeServeFrame returned error: %v", err)
	}
	gotReq, gotPayload, err := readServeRequest(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("readServeRequest returned error: %v", err)
	}
	if gotReq.ID != wantReq.ID || gotReq.Input != wantReq.Input || gotReq.PayloadLength != wantReq.PayloadLength {
		t.Fatalf("request = %+v, want %+v", gotReq, wantReq)
	}
	if !bytes.Equal(gotPayload, wantPayload) {
		t.Fatalf("payload = %v, want %v", gotPayload, wantPayload)
	}
}

func TestReadServeRequestRejectsLargePayload(t *testing.T) {
	var buf bytes.Buffer
	req := serveRequest{ID: "1", Type: "recognize", Input: "rgba", PayloadLength: maxServePayloadBytes + 1}
	if err := writeServeFrame(&buf, req, nil); err != nil {
		t.Fatalf("writeServeFrame returned error: %v", err)
	}
	_, _, err := readServeRequest(bufio.NewReader(&buf))
	if err == nil || !strings.Contains(err.Error(), "payload too large") {
		t.Fatalf("error = %v, want payload too large", err)
	}
}

func TestHandleServeRGBA(t *testing.T) {
	engine := &fakeServeEngine{}
	payload := make([]byte, 12)
	resp, stop := handleServeRequest(context.Background(), engine, serveRequest{
		ID:            "rgba",
		Type:          "recognize",
		Input:         "rgba",
		Width:         1,
		Height:        2,
		Stride:        8,
		PayloadLength: len(payload),
	}, payload)
	if stop {
		t.Fatalf("unexpected stop")
	}
	if !resp.OK || resp.Text != "rgba ok" {
		t.Fatalf("response = %+v, want rgba ok", resp)
	}
	if engine.rgbaCalls != 1 || engine.lastWidth != 1 || engine.lastHeight != 2 || engine.lastStride != 8 {
		t.Fatalf("engine geometry = calls:%d %dx%d stride=%d", engine.rgbaCalls, engine.lastWidth, engine.lastHeight, engine.lastStride)
	}
	if len(engine.lastRGBA) != len(payload) || &engine.lastRGBA[0] != &payload[0] {
		t.Fatalf("rgba payload was not passed through")
	}
}

func TestHandleServeImageBytes(t *testing.T) {
	engine := &fakeServeEngine{}
	payload := testPNGBytes(t)
	resp, stop := handleServeRequest(context.Background(), engine, serveRequest{
		ID:            "img",
		Type:          "recognize",
		Input:         "image_bytes",
		PayloadLength: len(payload),
	}, payload)
	if stop {
		t.Fatalf("unexpected stop")
	}
	if !resp.OK || resp.Text != "image ok" {
		t.Fatalf("response = %+v, want image ok", resp)
	}
	if engine.recognizeCalls != 1 || engine.lastBounds.Dx() != 2 || engine.lastBounds.Dy() != 1 {
		t.Fatalf("image bounds = %v calls=%d, want 2x1 one call", engine.lastBounds, engine.recognizeCalls)
	}
}

func TestHandleServePath(t *testing.T) {
	engine := &fakeServeEngine{}
	file := tempPNGFile(t)
	resp, stop := handleServeRequest(context.Background(), engine, serveRequest{
		ID:    "path",
		Type:  "recognize",
		Input: "path",
		Path:  file,
	}, nil)
	if stop {
		t.Fatalf("unexpected stop")
	}
	if !resp.OK || resp.Text != "image ok" {
		t.Fatalf("response = %+v, want image ok", resp)
	}
	if engine.recognizeCalls != 1 {
		t.Fatalf("recognize calls = %d, want 1", engine.recognizeCalls)
	}
}

func TestServeLoopPingShutdown(t *testing.T) {
	var input bytes.Buffer
	if err := writeServeFrame(&input, serveRequest{ID: "p", Type: "ping"}, nil); err != nil {
		t.Fatalf("write ping returned error: %v", err)
	}
	if err := writeServeFrame(&input, serveRequest{ID: "s", Type: "shutdown"}, nil); err != nil {
		t.Fatalf("write shutdown returned error: %v", err)
	}
	var output bytes.Buffer
	if code := serveLoop(context.Background(), &fakeServeEngine{}, &input, &output); code != 0 {
		t.Fatalf("serveLoop code = %d, want 0", code)
	}
	pong := readServeResponse(t, &output)
	if !pong.OK || pong.Type != "pong" || pong.ID != "p" {
		t.Fatalf("pong response = %+v", pong)
	}
	shutdown := readServeResponse(t, &output)
	if !shutdown.OK || shutdown.Type != "shutdown" || shutdown.ID != "s" {
		t.Fatalf("shutdown response = %+v", shutdown)
	}
}

func testPNGBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png returned error: %v", err)
	}
	return buf.Bytes()
}

func tempPNGFile(t *testing.T) string {
	t.Helper()
	file := t.TempDir() + "/sample.png"
	if err := os.WriteFile(file, testPNGBytes(t), 0o644); err != nil {
		t.Fatalf("write png returned error: %v", err)
	}
	return file
}

func readServeResponse(t *testing.T, reader *bytes.Buffer) serveResponse {
	t.Helper()
	var size [4]byte
	if _, err := reader.Read(size[:]); err != nil {
		t.Fatalf("read size returned error: %v", err)
	}
	data := make([]byte, binary.LittleEndian.Uint32(size[:]))
	if _, err := reader.Read(data); err != nil {
		t.Fatalf("read response returned error: %v", err)
	}
	var resp serveResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("decode response returned error: %v", err)
	}
	return resp
}
