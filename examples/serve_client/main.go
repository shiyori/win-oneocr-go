package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/shiyori/win-oneocr-go/examples/internal/testimage"
)

type request struct {
	ID            string `json:"id,omitempty"`
	Type          string `json:"type"`
	Input         string `json:"input,omitempty"`
	PayloadLength int    `json:"payloadLength,omitempty"`
}

type response struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	OK    bool   `json:"ok"`
	Text  string `json:"text"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	oneocrPath := flag.String("oneocr", "oneocr", "path to oneocr executable")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, *oneocrPath, "serve")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	reader := bufio.NewReader(stdout)

	if err := writeFrame(stdin, request{ID: "p", Type: "ping"}, nil); err != nil {
		panic(err)
	}
	if resp, err := readFrame(reader); err != nil || !resp.OK || resp.Type != "pong" {
		panic(fmt.Sprintf("ping failed: response=%+v error=%v", resp, err))
	}

	pngBytes, err := testimage.PNGBytes()
	if err != nil {
		panic(err)
	}
	if err := writeFrame(stdin, request{
		ID:            "ocr",
		Type:          "recognize",
		Input:         "image_bytes",
		PayloadLength: len(pngBytes),
	}, pngBytes); err != nil {
		panic(err)
	}
	resp, err := readFrame(reader)
	if err != nil {
		panic(err)
	}
	if !resp.OK {
		panic(fmt.Sprintf("ocr failed: %+v", resp.Error))
	}
	fmt.Println(resp.Text)

	if err := writeFrame(stdin, request{ID: "s", Type: "shutdown"}, nil); err != nil {
		panic(err)
	}
	if _, err := readFrame(reader); err != nil {
		panic(err)
	}
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}

func writeFrame(writer io.Writer, header request, payload []byte) error {
	data, err := json.Marshal(header)
	if err != nil {
		return err
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

func readFrame(reader *bufio.Reader) (response, error) {
	var size [4]byte
	if _, err := io.ReadFull(reader, size[:]); err != nil {
		return response{}, err
	}
	data := make([]byte, binary.LittleEndian.Uint32(size[:]))
	if _, err := io.ReadFull(reader, data); err != nil {
		return response{}, err
	}
	var resp response
	return resp, json.Unmarshal(data, &resp)
}
