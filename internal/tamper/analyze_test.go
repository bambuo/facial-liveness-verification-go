package tamper

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// TestAnalyzeSynthetic 验证参考检测器返回有界分数与明细。
func TestAnalyzeSynthetic(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: uint8((x + y) / 2), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	result, err := Analyze(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if result.Score < 0 || result.Score > 1 {
		t.Fatalf("score out of range: %f", result.Score)
	}
	if result.Details == nil || result.Details["ela"] == nil {
		t.Fatalf("missing details: %+v", result.Details)
	}
}
