package testimage

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

const sampleText = "ONEOCR 123"

var glyphs = map[rune][]string{
	' ': {
		".....",
		".....",
		".....",
		".....",
		".....",
		".....",
		".....",
	},
	'1': {
		"..#..",
		".##..",
		"..#..",
		"..#..",
		"..#..",
		"..#..",
		".###.",
	},
	'2': {
		".###.",
		"#...#",
		"....#",
		"...#.",
		"..#..",
		".#...",
		"#####",
	},
	'3': {
		"####.",
		"....#",
		"....#",
		".###.",
		"....#",
		"....#",
		"####.",
	},
	'C': {
		".####",
		"#....",
		"#....",
		"#....",
		"#....",
		"#....",
		".####",
	},
	'E': {
		"#####",
		"#....",
		"#....",
		"####.",
		"#....",
		"#....",
		"#####",
	},
	'N': {
		"#...#",
		"##..#",
		"#.#.#",
		"#..##",
		"#...#",
		"#...#",
		"#...#",
	},
	'O': {
		".###.",
		"#...#",
		"#...#",
		"#...#",
		"#...#",
		"#...#",
		".###.",
	},
	'R': {
		"####.",
		"#...#",
		"#...#",
		"####.",
		"#.#..",
		"#..#.",
		"#...#",
	},
}

func PNGBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, Image()); err != nil {
		return nil, fmt.Errorf("encode sample png: %w", err)
	}
	return buf.Bytes(), nil
}

func Image() image.Image {
	const (
		scale  = 18
		margin = 36
		gap    = 2
	)
	width := margin*2 + len([]rune(sampleText))*(5+gap)*scale
	height := margin*2 + 7*scale
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	x := margin
	for _, r := range sampleText {
		pattern, ok := glyphs[r]
		if !ok {
			pattern = glyphs[' ']
		}
		drawGlyph(img, x, margin, scale, pattern)
		x += (5 + gap) * scale
	}
	return img
}

func RGBA() (width, height int, rgba []byte, err error) {
	img := Image()
	bounds := img.Bounds()
	width, height = bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), img, bounds.Min, draw.Src)
	return width, height, dst.Pix, nil
}

func WritePNG(path string) error {
	data, err := PNGBytes()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write sample png: %w", err)
	}
	return nil
}

func drawGlyph(img *image.RGBA, x, y, scale int, pattern []string) {
	black := color.RGBA{A: 255}
	for row, line := range pattern {
		for col, cell := range line {
			if cell != '#' {
				continue
			}
			rect := image.Rect(
				x+col*scale,
				y+row*scale,
				x+(col+1)*scale,
				y+(row+1)*scale,
			)
			draw.Draw(img, rect, &image.Uniform{C: black}, image.Point{}, draw.Src)
		}
	}
}
