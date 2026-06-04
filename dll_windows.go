//go:build windows

package oneocr

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

const (
	modelKey           = `kj)TGtrK>f]b[Piow.gU+nC@s""""""4`
	maxOneOCRTextBytes = 1 << 20
)

var errNoText = errors.New("no text detected")

type oneOCRImage struct {
	Type     int32
	Width    int32
	Height   int32
	Reserved int32
	Step     int64
	DataPtr  *byte
}

type inProcessEngine struct {
	mu          sync.Mutex
	dll         *oneOCRDLL
	initOpts    uintptr
	processOpts uintptr
	pipeline    uintptr
	closed      bool
}

type oneOCRDLL struct {
	dll                      *syscall.LazyDLL
	createOcrInitOptions     *syscall.LazyProc
	createOcrPipeline        *syscall.LazyProc
	createOcrProcessOptions  *syscall.LazyProc
	runOcrPipeline           *syscall.LazyProc
	getOcrLineCount          *syscall.LazyProc
	getOcrLine               *syscall.LazyProc
	getOcrLineBoundingBox    *syscall.LazyProc
	getOcrLineContent        *syscall.LazyProc
	getOcrLineWordCount      *syscall.LazyProc
	getOcrWord               *syscall.LazyProc
	getOcrWordBoundingBox    *syscall.LazyProc
	getOcrWordContent        *syscall.LazyProc
	getOcrWordConfidence     *syscall.LazyProc
	releaseOcrInitOptions    *syscall.LazyProc
	releaseOcrPipeline       *syscall.LazyProc
	releaseOcrProcessOptions *syscall.LazyProc
	releaseOcrResult         *syscall.LazyProc
}

func newInProcessEngine(runtimeDir string) (*inProcessEngine, error) {
	if err := setDLLDirectory(runtimeDir); err != nil {
		return nil, fmt.Errorf("set OneOCR DLL directory: %w", err)
	}
	dll, err := loadOneOCRDLL(runtimeDir)
	if err != nil {
		return nil, err
	}
	engine := &inProcessEngine{dll: dll}
	if err := engine.init(runtimeDir); err != nil {
		_ = engine.Close()
		return nil, err
	}
	return engine, nil
}

func loadOneOCRDLL(runtimeDir string) (*oneOCRDLL, error) {
	dll := syscall.NewLazyDLL(filepath.Join(runtimeDir, "oneocr.dll"))
	if err := dll.Load(); err != nil {
		return nil, fmt.Errorf("load oneocr.dll: %w", err)
	}
	return &oneOCRDLL{
		dll:                      dll,
		createOcrInitOptions:     dll.NewProc("CreateOcrInitOptions"),
		createOcrPipeline:        dll.NewProc("CreateOcrPipeline"),
		createOcrProcessOptions:  dll.NewProc("CreateOcrProcessOptions"),
		runOcrPipeline:           dll.NewProc("RunOcrPipeline"),
		getOcrLineCount:          dll.NewProc("GetOcrLineCount"),
		getOcrLine:               dll.NewProc("GetOcrLine"),
		getOcrLineBoundingBox:    dll.NewProc("GetOcrLineBoundingBox"),
		getOcrLineContent:        dll.NewProc("GetOcrLineContent"),
		getOcrLineWordCount:      dll.NewProc("GetOcrLineWordCount"),
		getOcrWord:               dll.NewProc("GetOcrWord"),
		getOcrWordBoundingBox:    dll.NewProc("GetOcrWordBoundingBox"),
		getOcrWordContent:        dll.NewProc("GetOcrWordContent"),
		getOcrWordConfidence:     dll.NewProc("GetOcrWordConfidence"),
		releaseOcrInitOptions:    dll.NewProc("ReleaseOcrInitOptions"),
		releaseOcrPipeline:       dll.NewProc("ReleaseOcrPipeline"),
		releaseOcrProcessOptions: dll.NewProc("ReleaseOcrProcessOptions"),
		releaseOcrResult:         dll.NewProc("ReleaseOcrResult"),
	}, nil
}

func (e *inProcessEngine) init(runtimeDir string) error {
	initOpts := e.dll.createInitOptions()
	if initOpts == 0 {
		return errors.New("CreateOcrInitOptions returned empty handle")
	}
	processOpts := e.dll.createProcessOptions()
	if processOpts == 0 {
		e.dll.releaseInitOptions(initOpts)
		return errors.New("CreateOcrProcessOptions returned empty handle")
	}
	pipeline := e.dll.createPipeline(initOpts, filepath.Join(runtimeDir, "oneocr.onemodel"), modelKey)
	if pipeline == 0 {
		e.dll.releaseProcessOptions(processOpts)
		e.dll.releaseInitOptions(initOpts)
		return errors.New("CreateOcrPipeline returned empty handle")
	}
	e.initOpts = initOpts
	e.processOpts = processOpts
	e.pipeline = pipeline
	return nil
}

func (e *inProcessEngine) RecognizeRGBA(ctx context.Context, width, height, stride int, rgba []byte) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.pipeline == 0 {
		return nil, ErrUnavailable
	}
	if stride <= 0 {
		stride = width * 4
	}
	img := &oneOCRImage{
		Type:     3,
		Width:    int32(width),
		Height:   int32(height),
		Reserved: 0,
		Step:     int64(stride),
		DataPtr:  &rgba[0],
	}
	handle, err := e.dll.runPipeline(e.pipeline, e.processOpts, img)
	if errors.Is(err, errNoText) {
		return &Result{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer e.dll.releaseResult(handle)
	lines := e.dll.extractLines(handle)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return normalizeResult(lines), nil
}

func (e *inProcessEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	if e.pipeline != 0 {
		e.dll.releasePipeline(e.pipeline)
		e.pipeline = 0
	}
	if e.processOpts != 0 {
		e.dll.releaseProcessOptions(e.processOpts)
		e.processOpts = 0
	}
	if e.initOpts != 0 {
		e.dll.releaseInitOptions(e.initOpts)
		e.initOpts = 0
	}
	e.closed = true
	return nil
}

func (d *oneOCRDLL) createInitOptions() uintptr {
	var handle uintptr
	d.createOcrInitOptions.Call(uintptr(unsafe.Pointer(&handle)))
	return handle
}

func (d *oneOCRDLL) createPipeline(initOpts uintptr, modelPath, key string) uintptr {
	var handle uintptr
	modelPathPtr, _ := syscall.BytePtrFromString(modelPath)
	modelKeyPtr, _ := syscall.BytePtrFromString(key)
	d.createOcrPipeline.Call(
		uintptr(unsafe.Pointer(modelPathPtr)),
		uintptr(unsafe.Pointer(modelKeyPtr)),
		initOpts,
		uintptr(unsafe.Pointer(&handle)),
	)
	return handle
}

func (d *oneOCRDLL) createProcessOptions() uintptr {
	var handle uintptr
	d.createOcrProcessOptions.Call(uintptr(unsafe.Pointer(&handle)))
	return handle
}

func (d *oneOCRDLL) runPipeline(pipeline, processOpts uintptr, img *oneOCRImage) (uintptr, error) {
	var result uintptr
	ret, _, _ := d.runOcrPipeline.Call(
		pipeline,
		uintptr(unsafe.Pointer(img)),
		processOpts,
		uintptr(unsafe.Pointer(&result)),
	)
	if ret != 0 {
		if ret == 3 {
			return 0, errNoText
		}
		return 0, fmt.Errorf("RunOcrPipeline returned error code %d", ret)
	}
	return result, nil
}

func (d *oneOCRDLL) extractLines(result uintptr) []Line {
	lineCount := d.lineCount(result)
	lines := make([]Line, 0, lineCount)
	for i := 0; i < lineCount; i++ {
		lineHandle := d.line(result, i)
		wordCount := d.wordCount(lineHandle)
		words := make([]Word, 0, wordCount)
		for j := 0; j < wordCount; j++ {
			wordHandle := d.word(lineHandle, j)
			words = append(words, Word{
				Text:        d.wordContent(wordHandle),
				BoundingBox: d.wordBox(wordHandle),
				Confidence:  d.wordConfidence(wordHandle),
			})
		}
		lines = append(lines, Line{
			Text:        d.lineContent(lineHandle),
			BoundingBox: d.lineBox(lineHandle),
			Words:       words,
		})
	}
	return lines
}

func (d *oneOCRDLL) lineCount(result uintptr) int {
	var count int
	d.getOcrLineCount.Call(result, uintptr(unsafe.Pointer(&count)))
	return count
}

func (d *oneOCRDLL) line(result uintptr, index int) uintptr {
	var handle uintptr
	d.getOcrLine.Call(result, uintptr(index), uintptr(unsafe.Pointer(&handle)))
	return handle
}

func (d *oneOCRDLL) lineBox(line uintptr) BoundingBox {
	var box *BoundingBox
	d.getOcrLineBoundingBox.Call(line, uintptr(unsafe.Pointer(&box)))
	if box == nil {
		return BoundingBox{}
	}
	return *box
}

func (d *oneOCRDLL) lineContent(line uintptr) string {
	var ptr *byte
	d.getOcrLineContent.Call(line, uintptr(unsafe.Pointer(&ptr)))
	return safeCString(ptr)
}

func (d *oneOCRDLL) wordCount(line uintptr) int {
	var count int
	d.getOcrLineWordCount.Call(line, uintptr(unsafe.Pointer(&count)))
	return count
}

func (d *oneOCRDLL) word(line uintptr, index int) uintptr {
	var handle uintptr
	d.getOcrWord.Call(line, uintptr(index), uintptr(unsafe.Pointer(&handle)))
	return handle
}

func (d *oneOCRDLL) wordBox(word uintptr) BoundingBox {
	var box *BoundingBox
	d.getOcrWordBoundingBox.Call(word, uintptr(unsafe.Pointer(&box)))
	if box == nil {
		return BoundingBox{}
	}
	return *box
}

func (d *oneOCRDLL) wordContent(word uintptr) string {
	var ptr *byte
	d.getOcrWordContent.Call(word, uintptr(unsafe.Pointer(&ptr)))
	return safeCString(ptr)
}

func (d *oneOCRDLL) wordConfidence(word uintptr) float32 {
	var confidence float32
	d.getOcrWordConfidence.Call(word, uintptr(unsafe.Pointer(&confidence)))
	return confidence
}

func (d *oneOCRDLL) releaseInitOptions(handle uintptr) {
	d.releaseOcrInitOptions.Call(handle)
}

func (d *oneOCRDLL) releasePipeline(handle uintptr) {
	d.releaseOcrPipeline.Call(handle)
}

func (d *oneOCRDLL) releaseProcessOptions(handle uintptr) {
	d.releaseOcrProcessOptions.Call(handle)
}

func (d *oneOCRDLL) releaseResult(handle uintptr) {
	d.releaseOcrResult.Call(handle)
}

func safeCString(ptr *byte) string {
	if ptr == nil {
		return ""
	}
	buf := make([]byte, 0, 64)
	for offset := 0; offset < maxOneOCRTextBytes; offset++ {
		b := *(*byte)(unsafe.Add(unsafe.Pointer(ptr), offset))
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf)
}

func setDLLDirectory(dir string) error {
	ptr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return err
	}
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("SetDllDirectoryW")
	ret, _, callErr := proc.Call(uintptr(unsafe.Pointer(ptr)))
	if ret == 0 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return errors.New("SetDllDirectoryW returned false")
	}
	return nil
}
