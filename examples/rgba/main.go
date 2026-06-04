package main

import (
	"context"
	"fmt"

	oneocr "github.com/shiyori/win-oneocr-go"
	"github.com/shiyori/win-oneocr-go/examples/internal/testimage"
)

func main() {
	ctx := context.Background()
	width, height, rgba, err := testimage.RGBA()
	if err != nil {
		panic(err)
	}

	engine, err := oneocr.New(ctx)
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	result, err := engine.RecognizeRGBA(ctx, width, height, rgba)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Text)
}
