package deepfake

import (
	"image"
	"image/color"
	"math"
	"os"
	"testing"
)

// TestDeepfakeInference 用真实模型验证输出概率归一化。
func TestDeepfakeInference(t *testing.T) {
	path := os.Getenv("DEEPFAKE_MODEL")
	if path == "" {
		t.Skip("set DEEPFAKE_MODEL to run the deepfake inference test")
	}
	engine, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	img := image.NewRGBA(image.Rect(0, 0, 224, 224))
	for y := 0; y < 224; y++ {
		for x := 0; x < 224; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	result, err := engine.Detect(img)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(result.RealProbability+result.FakeProbability-1) > 1e-6 {
		t.Fatalf("probabilities do not sum to 1: %+v", result)
	}
}
