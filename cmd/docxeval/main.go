// 证件字段评测：对 MIDV-500 等带字段标注的图片运行 OCR + 字段规则。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"facial-liveness-verification-go/internal/document/fields"
	"facial-liveness-verification-go/internal/ocr"
)

type gtField struct {
	Value string `json:"value"`
}

type entriesReport struct {
	Image        string  `json:"image"`
	NameOK       bool    `json:"name_ok"`
	NameTokensOK bool    `json:"name_tokens_ok"`
	DOBOK        bool    `json:"dob_ok"`
	IDOK         bool    `json:"id_ok"`
	ExpiryOK     bool    `json:"expiry_ok"`
	Latency      float64 `json:"latency_ms"`
	MRZ          bool    `json:"mrz_detected"`
	Name         string  `json:"name,omitempty"`
	DOB          string  `json:"dob,omitempty"`
	ID           string  `json:"id,omitempty"`
	Expiry       string  `json:"expiry,omitempty"`
}

type report struct {
	GeneratedAt       string          `json:"generated_at"`
	Images            int             `json:"images"`
	MaxSide           int             `json:"max_side"`
	Orientation       string          `json:"orientation"`
	NameAccuracy      float64         `json:"name_accuracy"`
	NameTokenAccuracy float64         `json:"name_token_accuracy"`
	DOBAccuracy       float64         `json:"dob_accuracy"`
	IDAccuracy        float64         `json:"id_accuracy"`
	ExpiryAccuracy    float64         `json:"expiry_accuracy"`
	FieldAccuracy     float64         `json:"field_accuracy"`
	AllFields         int             `json:"all_fields_correct"`
	P50Ms             float64         `json:"p50_ms"`
	P95Ms             float64         `json:"p95_ms"`
	Entries           []entriesReport `json:"entries"`
}

func main() {
	modelDir := flag.String("model", "/tmp/paddle-ocr-models", "模型目录")
	recModel := flag.String("rec", "", "识别模型覆盖")
	dictPath := flag.String("dict", "", "词典覆盖")
	imagesGlob := flag.String("images", "", "图片 glob，例如 .../png/TS/*.png")
	gtPath := flag.String("gt", "", "字段标注 JSON")
	country := flag.String("country", "BR", "签发国家")
	limit := flag.Int("limit", 0, "最多评测图片数")
	maxSide := flag.Int("max-side", 1600, "检测长边上限")
	orientation := flag.String("orientation", "0", "方向")
	widthMode := flag.String("width-mode", "exact", "exact|bucket")
	bucket := flag.Int("bucket", 32, "分桶粒度")
	outJSON := flag.String("out-json", "eval/ocr/midv-bra.json", "JSON 输出")
	outMD := flag.String("out-md", "eval/ocr/midv-bra.md", "Markdown 输出")
	flag.Parse()

	if *imagesGlob == "" || *gtPath == "" {
		fatal(fmt.Errorf("必须提供 -images 与 -gt"))
	}
	raw, err := os.ReadFile(*gtPath)
	if err != nil {
		fatal(err)
	}
	var gt map[string]gtField
	if err := json.Unmarshal(raw, &gt); err != nil {
		fatal(err)
	}
	expected := expectedFields(gt)
	if expected.FullName == "" || expected.DateOfBirth == "" || expected.IDNumber == "" || expected.ExpiryDate == "" {
		fatal(fmt.Errorf("标注字段不完整: %+v", expected))
	}

	images, err := filepath.Glob(*imagesGlob)
	if err != nil {
		fatal(err)
	}
	sort.Strings(images)
	if *limit > 0 && len(images) > *limit {
		images = images[:*limit]
	}
	if len(images) == 0 {
		fatal(fmt.Errorf("没有匹配图片: %s", *imagesGlob))
	}

	cfg := ocr.Config{
		DetModel:    filepath.Join(*modelDir, "det.onnx"),
		RecModel:    *recModel,
		DictPath:    *dictPath,
		MaxSide:     *maxSide,
		Orientation: *orientation,
	}
	if cfg.RecModel == "" {
		cfg.RecModel = filepath.Join(*modelDir, "rec.onnx")
	}
	if cfg.DictPath == "" {
		cfg.DictPath = filepath.Join(*modelDir, "dict.txt")
	}
	if *widthMode == "bucket" {
		cfg.WidthBucket = *bucket
	}
	engine, err := ocr.New(cfg)
	if err != nil {
		fatal(err)
	}
	defer engine.Close()
	if err := engine.Warmup([]int{128, 256, 384, 512, 640}); err != nil {
		fatal(err)
	}

	rep := report{GeneratedAt: time.Now().UTC().Format(time.RFC3339), Images: len(images), MaxSide: *maxSide, Orientation: *orientation}
	counts := map[string]int{}
	var latencies []float64
	for _, imgPath := range images {
		data, err := os.ReadFile(imgPath)
		if err != nil {
			fatal(err)
		}
		start := time.Now()
		res, err := engine.RecognizeBytes(data, ocr.Options{Orientation: *orientation})
		if err != nil {
			fatal(fmt.Errorf("%s: %w", imgPath, err))
		}
		latency := float64(time.Since(start).Microseconds()) / 1000
		latencies = append(latencies, latency)
		extracted := fields.Extract(res.Text, *country)
		entry := entriesReport{
			Image:        filepath.Base(imgPath),
			NameOK:       normalizeName(extracted.FullName) == normalizeName(expected.FullName),
			NameTokensOK: sameNameTokens(extracted.FullName, expected.FullName),
			DOBOK:        extracted.DateOfBirth == expected.DateOfBirth,
			IDOK:         normalizeID(extracted.IDNumber) == normalizeID(expected.IDNumber),
			ExpiryOK:     extracted.ExpiryDate == expected.ExpiryDate,
			Latency:      latency,
			MRZ:          strings.Contains(strings.ToUpper(res.Text), "P<BRA"),
			Name:         extracted.FullName,
			DOB:          extracted.DateOfBirth,
			ID:           extracted.IDNumber,
			Expiry:       extracted.ExpiryDate,
		}
		if entry.NameOK {
			counts["name"]++
		}
		if entry.NameTokensOK {
			counts["name_tokens"]++
		}
		if entry.DOBOK {
			counts["dob"]++
		}
		if entry.IDOK {
			counts["id"]++
		}
		if entry.ExpiryOK {
			counts["expiry"]++
		}
		if entry.NameOK && entry.DOBOK && entry.IDOK && entry.ExpiryOK {
			rep.AllFields++
		}
		rep.Entries = append(rep.Entries, entry)
	}

	n := float64(len(images))
	rep.NameAccuracy = float64(counts["name"]) / n
	rep.NameTokenAccuracy = float64(counts["name_tokens"]) / n
	rep.DOBAccuracy = float64(counts["dob"]) / n
	rep.IDAccuracy = float64(counts["id"]) / n
	rep.ExpiryAccuracy = float64(counts["expiry"]) / n
	rep.FieldAccuracy = (rep.NameAccuracy + rep.DOBAccuracy + rep.IDAccuracy + rep.ExpiryAccuracy) / 4
	rep.P50Ms, rep.P95Ms = percentiles(latencies)

	out, _ := json.MarshalIndent(rep, "", "  ")
	if err := os.MkdirAll(filepath.Dir(*outJSON), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outJSON, append(out, '\n'), 0o644); err != nil {
		fatal(err)
	}
	md := renderMarkdown(rep, expected)
	if err := os.WriteFile(*outMD, []byte(md), 0o644); err != nil {
		fatal(err)
	}
	fmt.Print(md)
}

func expectedFields(gt map[string]gtField) fields.Result {
	return fields.Result{
		FullName:    strings.TrimSpace(gt["field05"].Value + " " + gt["field04"].Value),
		DateOfBirth: parseLifeDate(gt["field07"].Value),
		IDNumber:    gt["field03"].Value,
		ExpiryDate:  parseLifeDate(gt["field12"].Value),
	}
}

var monthNames = map[string]int{
	"JAN": 1, "FEV": 2, "FEB": 2, "MAR": 3, "ABR": 4, "APR": 4, "MAI": 5, "MAY": 5,
	"JUN": 6, "JUL": 7, "AGO": 8, "AUG": 8, "SET": 9, "SEP": 9, "OUT": 10, "OCT": 10,
	"NOV": 11, "DEZ": 12, "DEC": 12,
}

func parseLifeDate(raw string) string {
	m := regexp.MustCompile(`(\d{1,2})\s+([A-Za-z]{3})[^0-9]*(\d{4})`).FindStringSubmatch(raw)
	if len(m) != 4 {
		return ""
	}
	day, _ := strconv.Atoi(m[1])
	month := monthNames[strings.ToUpper(m[2])]
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func normalizeName(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToUpper(s) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r > 127 {
			b.WriteRune(r)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func normalizeID(s string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
}

func sameNameTokens(a, b string) bool {
	ta := nameTokens(a)
	tb := nameTokens(b)
	if len(ta) != len(tb) || len(tb) == 0 {
		return false
	}
	for i := range ta {
		if ta[i] != tb[i] {
			return false
		}
	}
	return true
}

func nameTokens(s string) []string {
	out := []string{}
	for _, token := range strings.Fields(normalizeName(s)) {
		if len([]rune(token)) >= 2 {
			out = append(out, token)
		}
	}
	sort.Strings(out)
	return out
}

func percentiles(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	p := func(q float64) float64 {
		idx := int(math.Ceil(q*float64(len(sorted)))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		return sorted[idx]
	}
	return p(0.5), p(0.95)
}

func renderMarkdown(rep report, expected fields.Result) string {
	var b strings.Builder
	b.WriteString("# MIDV 巴西护照字段评测\n\n")
	fmt.Fprintf(&b, "- 图片：%d；MaxSide：%d；方向：%s\n", rep.Images, rep.MaxSide, rep.Orientation)
	fmt.Fprintf(&b, "- 期望字段：name=%q dob=%s id=%s expiry=%s\n\n", expected.FullName, expected.DateOfBirth, expected.IDNumber, expected.ExpiryDate)
	b.WriteString("| 指标 | 值 |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| 姓名准确率（精确） | %.1f%% |\n", rep.NameAccuracy*100)
	fmt.Fprintf(&b, "| 姓名准确率（token） | %.1f%% |\n", rep.NameTokenAccuracy*100)
	fmt.Fprintf(&b, "| 出生日期准确率 | %.1f%% |\n", rep.DOBAccuracy*100)
	fmt.Fprintf(&b, "| 证件号准确率 | %.1f%% |\n", rep.IDAccuracy*100)
	fmt.Fprintf(&b, "| 有效期准确率 | %.1f%% |\n", rep.ExpiryAccuracy*100)
	fmt.Fprintf(&b, "| 字段总准确率 | %.1f%% |\n", rep.FieldAccuracy*100)
	fmt.Fprintf(&b, "| 四字段全对 | %d / %d |\n", rep.AllFields, rep.Images)
	fmt.Fprintf(&b, "| P50 / P95 | %.0f ms / %.0f ms |\n\n", rep.P50Ms, rep.P95Ms)
	b.WriteString("| 图片 | MRZ | 姓名 | 出生 | 证件号 | 有效期 | 耗时(ms) |\n| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, e := range rep.Entries {
		mark := func(ok bool) string {
			if ok {
				return "✓"
			}
			return "✗"
		}
		fmt.Fprintf(&b, "| %s | %v | %s | %s | %s | %s | %.0f |\n",
			e.Image, e.MRZ, mark(e.NameOK), mark(e.DOBOK), mark(e.IDOK), mark(e.ExpiryOK), e.Latency)
	}
	b.WriteString("\n> MIDV-500 为公开模拟证件（非真实证件）；许可与来源见 `eval/ocr` 说明。\n")
	return b.String()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
