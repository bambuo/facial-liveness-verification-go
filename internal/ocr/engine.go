package ocr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"
	"time"

	_ "golang.org/x/image/tiff"
)

const (
	// DefaultMaxSide 是检测阶段长边上限。
	DefaultMaxSide = 640
	// DefaultMinBoxArea 是检测框最小面积（缩放后坐标）。
	DefaultMinBoxArea = 25
	recHeight         = 48
	recMinWidth       = 8
)

// Config 是 OCR 引擎配置。
type Config struct {
	DetModel          string
	RecModel          string
	DictPath          string
	MaxSide           int
	MinBoxArea        int
	PaddingVertical   float64
	PaddingHorizontal float64
	// WidthBucket > 0 时按该粒度分桶识别宽度；<= 0 使用精确宽度。
	WidthBucket int
	// Orientation 默认 "0"；可选 "auto"、"90"、"180"、"270"。
	Orientation string
}

// Options 是单次调用选项。
type Options struct {
	Orientation string
	MaxBoxes    int
}

// Engine 是纯 Go OCR 引擎。
type Engine struct {
	cfg  Config
	det  *runner
	rec  *runner
	dict []rune
}

// New 加载检测/识别模型与词典。
func New(cfg Config) (*Engine, error) {
	if cfg.DetModel == "" || cfg.RecModel == "" || cfg.DictPath == "" {
		return nil, fmt.Errorf("OCR 模型路径不完整")
	}
	if cfg.MaxSide <= 0 {
		cfg.MaxSide = DefaultMaxSide
	}
	if cfg.MinBoxArea <= 0 {
		cfg.MinBoxArea = DefaultMinBoxArea
	}
	if cfg.PaddingVertical == 0 {
		cfg.PaddingVertical = 0.4
	}
	if cfg.PaddingHorizontal == 0 {
		cfg.PaddingHorizontal = 0.6
	}
	if cfg.Orientation == "" {
		cfg.Orientation = "0"
	}
	dict, err := LoadDict(cfg.DictPath)
	if err != nil {
		return nil, err
	}
	det, err := loadRunner(cfg.DetModel)
	if err != nil {
		return nil, fmt.Errorf("加载检测模型: %w", err)
	}
	rec, err := loadRunner(cfg.RecModel)
	if err != nil {
		det.close()
		return nil, fmt.Errorf("加载识别模型: %w", err)
	}
	return &Engine{cfg: cfg, det: det, rec: rec, dict: dict}, nil
}

// Close 释放模型资源。
func (e *Engine) Close() {
	if e.det != nil {
		e.det.close()
	}
	if e.rec != nil {
		e.rec.close()
	}
}

// DictionarySize 返回词典长度。
func (e *Engine) DictionarySize() int { return len(e.dict) }

// RecognizeBytes 解码图片并识别。
func (e *Engine) RecognizeBytes(data []byte, opts Options) (Result, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Result{}, fmt.Errorf("解码图片: %w", err)
	}
	return e.RecognizeImage(img, opts)
}

// RecognizeImage 执行检测与识别；auto 方向会尝试 0/90/180/270 并取最佳。
func (e *Engine) RecognizeImage(img image.Image, opts Options) (Result, error) {
	start := time.Now()
	angleOpt := opts.Orientation
	if angleOpt == "" {
		angleOpt = e.cfg.Orientation
	}
	angles := orientationAngles(angleOpt)
	var best Result
	bestScore := math.Inf(-1)
	bestAngle := 0
	var lastErr error
	for _, angle := range angles {
		r, err := e.runOnce(rotate(img, angle))
		if err != nil {
			lastErr = err
			continue
		}
		score := resultScore(r)
		if score > bestScore {
			bestScore = score
			best = r
			bestAngle = angle
		}
	}
	if bestScore == math.Inf(-1) {
		if lastErr != nil {
			return Result{}, lastErr
		}
		return Result{}, fmt.Errorf("OCR 无结果")
	}
	best.Orientation = bestAngle
	best.Total = time.Since(start).Nanoseconds()
	return best, nil
}

// Warmup 预编译常用识别宽度。
func (e *Engine) Warmup(widths []int) error {
	for _, w := range widths {
		if w < recMinWidth {
			w = recMinWidth
		}
		if e.cfg.WidthBucket > 0 {
			w = int(math.Ceil(float64(w)/float64(e.cfg.WidthBucket))) * e.cfg.WidthBucket
		}
		zeros := make([]float32, 3*recHeight*w)
		if _, _, err := e.rec.run(zeros, 1, 3, recHeight, w); err != nil {
			return err
		}
	}
	return nil
}

// runOnce 对单一方向执行检测、识别、阅读排序与分行。
func (e *Engine) runOnce(img image.Image) (Result, error) {
	detStart := time.Now()
	boxes, _, err := e.detect(img)
	if err != nil {
		return Result{}, err
	}
	detDur := time.Since(detStart).Nanoseconds()

	recStart := time.Now()
	candidates := make([]candidate, 0, len(boxes))
	for _, b := range boxes {
		text, conf, err := e.recognize(img, b)
		if err != nil {
			continue
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		candidates = append(candidates, candidate{text: text, confidence: conf, box: b})
	}
	recDur := time.Since(recStart).Nanoseconds()

	sortReadingOrder(candidates)
	lines := groupResultsByLine(candidates)
	text, conf := joinLines(lines)
	return Result{
		Text:        text,
		Lines:       lines,
		Confidence:  conf,
		Boxes:       boxes,
		DetDuration: detDur,
		RecDuration: recDur,
	}, nil
}

// resultScore 用字符数加权置信度，作为自动方向选择分数。
func resultScore(r Result) float64 {
	if len(r.Lines) == 0 {
		return math.Inf(-1)
	}
	score := 0.0
	for _, line := range r.Lines {
		for _, item := range line {
			score += float64(len([]rune(item.Text))) * item.Confidence
		}
	}
	return score
}

// orientationAngles 把方向配置展开为待尝试的角度列表。
func orientationAngles(opt string) []int {
	switch opt {
	case "auto":
		return []int{0, 90, 180, 270}
	case "90":
		return []int{90}
	case "180":
		return []int{180}
	case "270":
		return []int{270}
	default:
		return []int{0}
	}
}

// LoadDict 读取 PaddleOCR 词典；首行空串对应 blank。
func LoadDict(path string) ([]rune, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := strings.TrimPrefix(string(buf), "\ufeff")
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	dict := make([]rune, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			dict = append(dict, 0)
			continue
		}
		dict = append(dict, []rune(line)[0])
	}
	return dict, nil
}
