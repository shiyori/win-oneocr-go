package oneocr

import (
	"math"
	"strings"
)

func normalizeResult(lines []Line) *Result {
	out := &Result{Lines: make([]Line, 0, len(lines))}
	texts := make([]string, 0, len(lines))
	for _, line := range lines {
		line.Text = strings.TrimSpace(line.Text)
		line.Rect = rectFromBox(line.BoundingBox)
		if line.Confidence <= 0 {
			line.Confidence = averageWordConfidence(line.Words)
		}
		for i := range line.Words {
			line.Words[i].Text = strings.TrimSpace(line.Words[i].Text)
			line.Words[i].Rect = rectFromBox(line.Words[i].BoundingBox)
		}
		if line.Text != "" {
			texts = append(texts, line.Text)
		}
		out.Lines = append(out.Lines, line)
	}
	out.Text = strings.Join(texts, "\n")
	return out
}

func rectFromBox(box BoundingBox) Rect {
	points := []Point{box.TopLeft, box.TopRight, box.BottomRight, box.BottomLeft}
	minX, minY := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY := float32(-math.MaxFloat32), float32(-math.MaxFloat32)
	for _, point := range points {
		if point.X < minX {
			minX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	if maxX < minX || maxY < minY {
		return Rect{}
	}
	x := int(math.Floor(float64(minX)))
	y := int(math.Floor(float64(minY)))
	right := int(math.Ceil(float64(maxX)))
	bottom := int(math.Ceil(float64(maxY)))
	return Rect{X: x, Y: y, Width: max(0, right-x), Height: max(0, bottom-y)}
}

func averageWordConfidence(words []Word) float32 {
	var sum float32
	var count int
	for _, word := range words {
		if word.Confidence <= 0 {
			continue
		}
		sum += word.Confidence
		count++
	}
	if count == 0 {
		return 1
	}
	return sum / float32(count)
}
