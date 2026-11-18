// OCR CLI 探针：调用 internal/ocr 输出逐行结果与时序。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"facial-liveness-verification-go/internal/ocr"
)

// main 解析参数并执行一次 OCR，打印行结果与时序。
func main() {
	modelDir := flag.String("model", "/tmp/paddle-ocr-models", "模型目录")
	imagePath := flag.String("image", "/tmp/paddle-ocr-models/demo.jpg", "输入图片")
	recModel := flag.String("rec", "", "识别模型覆盖路径")
	dictPath := flag.String("dict", "", "词典覆盖路径")
	widthMode := flag.String("width-mode", "exact", "识别宽度模式: exact|bucket")
	bucket := flag.Int("bucket", 32, "分桶粒度")
	orientation := flag.String("orientation", "0", "方向: 0|auto|90|180|270")
	warmup := flag.String("warmup", "", "预热宽度，逗号分隔")
	flag.Parse()

	if *recModel == "" {
		*recModel = filepath.Join(*modelDir, "rec.onnx")
	}
	if *dictPath == "" {
		*dictPath = filepath.Join(*modelDir, "dict.txt")
	}
	cfg := ocr.Config{
		DetModel:    filepath.Join(*modelDir, "det.onnx"),
		RecModel:    *recModel,
		DictPath:    *dictPath,
		Orientation: *orientation,
	}
	if *widthMode == "bucket" {
		cfg.WidthBucket = *bucket
	}
	engine, err := ocr.New(cfg)
	if err != nil {
		fatal(err)
	}
	defer engine.Close()
	if *warmup != "" {
		widths := parseWidths(*warmup)
		start := time.Now()
		if err := engine.Warmup(widths); err != nil {
			fatal(err)
		}
		fmt.Printf("warmup: widths=%v elapsed=%v\n", widths, time.Since(start))
	}

	data, err := os.ReadFile(*imagePath)
	if err != nil {
		fatal(err)
	}
	start := time.Now()
	res, err := engine.RecognizeBytes(data, ocr.Options{})
	if err != nil {
		fatal(err)
	}
	fmt.Printf("mode=%s orientation=%d total=%v det=%v rec=%v\n",
		*widthMode, res.Orientation, time.Since(start), time.Duration(res.DetDuration), time.Duration(res.RecDuration))
	fmt.Printf("boxes=%d lines=%d confidence=%.3f\n", len(res.Boxes), len(res.Lines), res.Confidence)
	for i, line := range res.Lines {
		parts := make([]string, 0, len(line))
		conf := 0.0
		for _, item := range line {
			parts = append(parts, item.Text)
			conf += item.Confidence
		}
		conf /= float64(len(line))
		fmt.Printf("  [%02d] conf=%.3f %s\n", i+1, conf, strings.Join(parts, " "))
	}
	fmt.Printf("\nTEXT:\n%s\n", res.Text)
}

// parseWidths 解析逗号分隔的预热宽度。
func parseWidths(raw string) []int {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &n); err == nil && n > 0 {
			out = append(out, n)
		}
	}
	return out
}

// fatal 打印错误并以非 0 退出。
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
