package main

import (
	"context"
	"fmt"

	oneocr "github.com/shiyori/win-oneocr-go"
	"github.com/shiyori/win-oneocr-go/examples/internal/testimage"
)

func main() {
	ctx := context.Background()
	img := testimage.Image()

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
