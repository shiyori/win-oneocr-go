package main

import (
	"flag"
	"fmt"

	"github.com/shiyori/win-oneocr-go/examples/internal/testimage"
)

func main() {
	output := flag.String("out", "oneocr-sample.png", "output png path")
	flag.Parse()
	if err := testimage.WritePNG(*output); err != nil {
		panic(err)
	}
	fmt.Println(*output)
}
