// 离线 OCR 评测：CER/WER、字段命中率、延迟 P50/P95、并发正确性。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"facial-liveness-verification-go/internal/ocr"
)

type fieldSpec map[string]string

type modeResult struct {
	Mode          string          `json:"mode"`
	Iterations    int             `json:"iterations"`
	CER           float64         `json:"cer"`
	CERNospace    float64         `json:"cer_nospace"`
	WER           float64         `json:"wer"`
	FieldAccuracy float64         `json:"field_accuracy"`
	Fields        map[string]bool `json:"fields"`
	LatencyMs     []float64       `json:"latency_ms"`
	DetMs         []float64       `json:"det_ms"`
	RecMs         []float64       `json:"rec_ms"`
	P50Ms         float64         `json:"p50_ms"`
	P95Ms         float64         `json:"p95_ms"`
	MeanMs        float64         `json:"mean_ms"`
	Text          string          `json:"text"`
}

type report struct {
	Image            string       `json:"image"`
	Reference        string       `json:"reference"`
	Orientation      string       `json:"orientation"`
	Modes            []modeResult `json:"modes"`
	Concurrency      int          `json:"concurrency"`
	ConcurrentQPS    float64      `json:"concurrent_qps"`
	ConcurrentErrors int          `json:"concurrent_errors"`
}

// main 解析参数，按宽度模式评测并输出 JSON 报告。
func main() {
	modelDir := flag.String("model", "/tmp/paddle-ocr-models", "模型目录")
	imagePath := flag.String("image", "/tmp/paddle-ocr-models/demo.jpg", "输入图片")
	refPath := flag.String("ref", "", "参考文本文件")
	recModel := flag.String("rec", "", "识别模型覆盖路径")
	dictPath := flag.String("dict", "", "词典覆盖路径")
	fieldsPath := flag.String("fields", "", "字段 JSON：{名称: 期望子串}")
	widthModes := flag.String("width-modes", "bucket,exact", "宽度模式列表")
	bucket := flag.Int("bucket", 32, "分桶粒度")
	iters := flag.Int("iters", 3, "每种模式迭代次数")
	warmup := flag.String("warmup", "64,128,192,256,320,384,448,512", "预热宽度")
	orientation := flag.String("orientation", "0", "方向")
	concurrency := flag.Int("concurrency", 0, "并发请求数；0 不压测")
	jsonOut := flag.String("json", "", "JSON 输出路径")
	flag.Parse()

	ref := ""
	if *refPath != "" {
		buf, err := os.ReadFile(*refPath)
		if err != nil {
			fatal(err)
		}
		ref = string(buf)
	}
	var fields fieldSpec
	if *fieldsPath != "" {
		buf, err := os.ReadFile(*fieldsPath)
		if err != nil {
			fatal(err)
		}
		if err := json.Unmarshal(buf, &fields); err != nil {
			fatal(err)
		}
	}
	data, err := os.ReadFile(*imagePath)
	if err != nil {
		fatal(err)
	}

	var modes []modeResult
	for _, mode := range strings.Split(*widthModes, ",") {
		mode = strings.TrimSpace(mode)
		if mode == "" {
			continue
		}
		cfg := ocr.Config{
			DetModel:    filepath.Join(*modelDir, "det.onnx"),
			RecModel:    *recModel,
			DictPath:    *dictPath,
			Orientation: *orientation,
		}
		if cfg.RecModel == "" {
			cfg.RecModel = filepath.Join(*modelDir, "rec.onnx")
		}
		if cfg.DictPath == "" {
			cfg.DictPath = filepath.Join(*modelDir, "dict.txt")
		}
		if mode == "bucket" {
			cfg.WidthBucket = *bucket
		}
		engine, err := ocr.New(cfg)
		if err != nil {
			fatal(err)
		}
		if *warmup != "" {
			if err := engine.Warmup(parseWidths(*warmup)); err != nil {
				fatal(err)
			}
		}
		res := runMode(engine, data, *iters)
		engine.Close()
		res.Mode = mode
		if ref != "" {
			res.CER, res.WER = cerWer(ref, res.Text)
			res.CERNospace = cerNoSpace(ref, res.Text)
		}
		if len(fields) > 0 {
			res.Fields, res.FieldAccuracy = matchFields(fields, res.Text)
		}
		modes = append(modes, res)
	}

	rep := report{Image: *imagePath, Reference: *refPath, Orientation: *orientation, Modes: modes}
	if *concurrency > 0 {
		qps, failures := stress(*modelDir, data, *concurrency, *bucket, *orientation)
		rep.Concurrency = *concurrency
		rep.ConcurrentQPS = qps
		rep.ConcurrentErrors = failures
	}

	out, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(out))
	if *jsonOut != "" {
		if err := os.WriteFile(*jsonOut, append(out, '\n'), 0o644); err != nil {
			fatal(err)
		}
	}
}

// runMode 在指定引擎上重复推理，汇总延迟分位数。
func runMode(engine *ocr.Engine, data []byte, iters int) modeResult {
	res := modeResult{Iterations: iters}
	for i := 0; i < iters; i++ {
		start := time.Now()
		r, err := engine.RecognizeBytes(data, ocr.Options{})
		if err != nil {
			fatal(err)
		}
		res.LatencyMs = append(res.LatencyMs, float64(time.Since(start).Microseconds())/1000)
		res.DetMs = append(res.DetMs, float64(r.DetDuration)/1e6)
		res.RecMs = append(res.RecMs, float64(r.RecDuration)/1e6)
		if i == iters-1 {
			res.Text = r.Text
		}
	}
	sorted := append([]float64(nil), res.LatencyMs...)
	sort.Float64s(sorted)
	res.P50Ms = percentile(sorted, 0.50)
	res.P95Ms = percentile(sorted, 0.95)
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	res.MeanMs = sum / float64(len(sorted))
	return res
}

// stress 多引擎并发推理，返回 QPS 与失败数。
func stress(modelDir string, data []byte, n, bucket int, orientation string) (float64, int) {
	engines := make([]*ocr.Engine, 0, n)
	for i := 0; i < n; i++ {
		cfg := ocr.Config{
			DetModel:    filepath.Join(modelDir, "det.onnx"),
			RecModel:    filepath.Join(modelDir, "rec.onnx"),
			DictPath:    filepath.Join(modelDir, "dict.txt"),
			Orientation: orientation,
			WidthBucket: bucket,
		}
		e, err := ocr.New(cfg)
		if err != nil {
			fatal(err)
		}
		engines = append(engines, e)
	}
	defer func() {
		for _, e := range engines {
			e.Close()
		}
	}()
	// 先预热模型图与推理，避免把首次编译计入吞吐。
	for _, e := range engines {
		_, _ = e.RecognizeBytes(data, ocr.Options{})
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	failures := 0
	start := time.Now()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(e *ocr.Engine) {
			defer wg.Done()
			r, err := e.RecognizeBytes(data, ocr.Options{})
			if err != nil || strings.TrimSpace(r.Text) == "" {
				mu.Lock()
				failures++
				mu.Unlock()
			}
		}(engines[i])
	}
	wg.Wait()
	elapsed := time.Since(start).Seconds()
	qps := 0.0
	if elapsed > 0 {
		qps = float64(n-failures) / elapsed
	}
	return qps, failures
}

// cerWer 计算字符错误率与词错误率。
func cerWer(ref, hyp string) (float64, float64) {
	r := normalize(ref)
	h := normalize(hyp)
	rr := []rune(r)
	hr := []rune(h)
	cer := 0.0
	if len(rr) > 0 {
		cer = float64(levenshteinRunes(rr, hr)) / float64(len(rr))
	}
	rt := strings.Fields(r)
	ht := strings.Fields(h)
	wer := 0.0
	if len(rt) > 0 {
		wer = float64(levenshteinTokens(rt, ht)) / float64(len(rt))
	}
	return cer, wer
}

// cerNoSpace 计算忽略空白后的字符错误率，避免 CJK 空格差异放大 CER。
func cerNoSpace(ref, hyp string) float64 {
	r := []rune(strings.ReplaceAll(normalize(ref), " ", ""))
	h := []rune(strings.ReplaceAll(normalize(hyp), " ", ""))
	if len(r) == 0 {
		return 0
	}
	return float64(levenshteinRunes(r, h)) / float64(len(r))
}

// matchFields 检查期望字段是否出现在 OCR 文本中。
func matchFields(fields fieldSpec, text string) (map[string]bool, float64) {
	compact := compactText(text)
	out := make(map[string]bool, len(fields))
	matched := 0
	for name, want := range fields {
		ok := strings.Contains(compact, compactText(want))
		out[name] = ok
		if ok {
			matched++
		}
	}
	return out, float64(matched) / float64(len(fields))
}

// normalize 统一大小写与空白，供 CER/WER 计算。
func normalize(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}

// compactText 保留字母与数字（含非 ASCII），供字段命中判断。
func compactText(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// levenshteinRunes 计算 rune 级编辑距离。
func levenshteinRunes(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// levenshteinTokens 计算 token 级编辑距离。
func levenshteinTokens(a, b []string) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// percentile 取已排序样本的指定分位数。
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
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

// min3 返回三个整数中的最小值。
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// fatal 打印错误并以非 0 退出。
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
