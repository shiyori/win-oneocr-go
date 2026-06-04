package oneocr

import "testing"

func TestNormalizeResultComputesTextRectsAndConfidence(t *testing.T) {
	result := normalizeResult([]Line{
		{
			Text: " hello ",
			BoundingBox: BoundingBox{
				TopLeft:     Point{X: 1.2, Y: 2.8},
				TopRight:    Point{X: 10.1, Y: 2.2},
				BottomRight: Point{X: 11.4, Y: 7.7},
				BottomLeft:  Point{X: 0.8, Y: 8.1},
			},
			Words: []Word{
				{Text: " hello ", Confidence: 0.5, BoundingBox: BoundingBox{TopLeft: Point{X: 1, Y: 1}, BottomRight: Point{X: 5, Y: 5}}},
				{Text: " world ", Confidence: 0.9, BoundingBox: BoundingBox{TopLeft: Point{X: 6, Y: 1}, BottomRight: Point{X: 10, Y: 5}}},
			},
		},
	})
	if result.Text != "hello" {
		t.Fatalf("Text = %q, want hello", result.Text)
	}
	if len(result.Lines) != 1 {
		t.Fatalf("Lines len = %d, want 1", len(result.Lines))
	}
	line := result.Lines[0]
	if line.Rect.X != 0 || line.Rect.Y != 2 || line.Rect.Width != 12 || line.Rect.Height != 7 {
		t.Fatalf("unexpected line rect: %+v", line.Rect)
	}
	if line.Confidence < 0.69 || line.Confidence > 0.71 {
		t.Fatalf("line confidence = %f, want about 0.7", line.Confidence)
	}
	if line.Words[0].Text != "hello" {
		t.Fatalf("word text = %q, want hello", line.Words[0].Text)
	}
}
