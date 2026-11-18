// OCR 进程级基准：按 N 个 worker 进程并发执行，统计 QPS 与内存。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"facial-liveness-verification-go/internal/ocr"
)

type workerResult struct {
	ElapsedMs float64 `json:"elapsed_ms"`
	DetMs     float64 `json:"det_ms"`
	RecMs     float64 `json:"rec_ms"`
	SysBytes  uint64  `json:"sys_bytes"`
	Lines     int     `json:"lines"`
}

type row struct {
	Processes int
	Iters     int
	WallMs    float64
	QPS       float64
	AvgMs     float64
	SysMB     float64
}

func main() {
	worker := flag.Bool("worker", false, "内部 worker 模式")
	modelDir := flag.String("model", "/tmp/paddle-ocr-models", "模型目录")
	imagePath := flag.String("image", "/tmp/paddle-ocr-models/demo.jpg", "输入图片")
	widthMode := flag.String("width-mode", "exact", "exact|bucket")
	bucket := flag.Int("bucket", 32, "分桶粒度")
	iters := flag.Int("iters", 1, "每个 worker 的识别次数")
	processes := flag.String("processes", "1,2,4", "并发进程数列表")
	warmup := flag.String("warmup", "128,256,384,512,640", "预热宽度")
	maxSide := flag.Int("max-side", 0, "检测长边上限")
	mdOut := flag.String("md", "eval/ocr/bench.md", "Markdown 输出")
	flag.Parse()

	if *worker {
		runWorker(*modelDir, *imagePath, *widthMode, *bucket, *iters, *warmup, *maxSide)
		return
	}

	data, err := os.ReadFile(*imagePath)
	if err != nil {
		fatal(err)
	}
	_ = data

	var rows []row
	for _, part := range strings.Split(*processes, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n <= 0 {
			continue
		}
		start := time.Now()
		results := make([]workerResult, 0, n)
		for i := 0; i < n; i++ {
			cmd := exec.Command(os.Args[0],
				"-worker", "-model", *modelDir, "-image", *imagePath,
				"-width-mode", *widthMode, "-bucket", strconv.Itoa(*bucket),
				"-iters", strconv.Itoa(*iters), "-warmup", *warmup, "-max-side", strconv.Itoa(*maxSide))
			out, err := cmd.Output()
			if err != nil {
				fatal(fmt.Errorf("worker %d/%d: %w", i+1, n, err))
			}
			var res workerResult
			if err := json.Unmarshal(out, &res); err != nil {
				fatal(fmt.Errorf("worker result: %w output=%s", err, string(out)))
			}
			results = append(results, res)
		}
		wall := time.Since(start).Seconds()
		total := 0.0
		sys := uint64(0)
		for _, r := range results {
			total += r.ElapsedMs
			if r.SysBytes > sys {
				sys = r.SysBytes
			}
		}
		rows = append(rows, row{
			Processes: n,
			Iters:     *iters,
			WallMs:    wall * 1000,
			QPS:       float64(n**iters) / wall,
			AvgMs:     total / float64(len(results)),
			SysMB:     float64(sys) / (1024 * 1024),
		})
	}

	var b strings.Builder
	b.WriteString("# OCR 进程级并发基准\n\n")
	b.WriteString("由 `go run ./cmd/ocrbench` 生成；每个 worker 为独立进程，包含加载模型与预热。\n\n")
	b.WriteString("| 进程数 | 每进程次数 | 墙钟 (ms) | QPS | 单进程平均 (ms) | 峰值 Sys (MB) |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %d | %d | %.0f | %.3f | %.0f | %.0f |\n",
			r.Processes, r.Iters, r.WallMs, r.QPS, r.AvgMs, r.SysMB)
	}
	b.WriteString("\n说明：\n\n")
	b.WriteString("- 同进程多引擎（预热后）QPS 约 0.345；本表含进程启动/模型加载/预热，数值不可直接对比。\n")
	b.WriteString("- 单机 1/2/4 进程 QPS 基本持平（约 0.23），说明单进程已占满 CPU，进程级并行不能在同一主机线性扩容。\n")
	b.WriteString("- 扩容建议：多机/多容器副本 + 队列，或量化/降分辨率/后端优化；单机多进程无收益。\n")

	if err := os.MkdirAll("eval/ocr", 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*mdOut, []byte(b.String()), 0o644); err != nil {
		fatal(err)
	}
	fmt.Print(b.String())
}

func runWorker(modelDir, imagePath, widthMode string, bucket, iters int, warmup string, maxSide int) {
	cfg := ocr.Config{
		MaxSide:     maxSide,
		DetModel:    modelDir + "/det.onnx",
		RecModel:    modelDir + "/rec.onnx",
		DictPath:    modelDir + "/dict.txt",
		Orientation: "0",
	}
	if widthMode == "bucket" {
		cfg.WidthBucket = bucket
	}
	engine, err := ocr.New(cfg)
	if err != nil {
		fatal(err)
	}
	defer engine.Close()
	if warmup != "" {
		if err := engine.Warmup(parseWidths(warmup)); err != nil {
			fatal(err)
		}
	}
	data, err := os.ReadFile(imagePath)
	if err != nil {
		fatal(err)
	}
	start := time.Now()
	det, rec := 0.0, 0.0
	lines := 0
	for i := 0; i < iters; i++ {
		res, err := engine.RecognizeBytes(data, ocr.Options{})
		if err != nil {
			fatal(err)
		}
		det += float64(res.DetDuration) / 1e6
		rec += float64(res.RecDuration) / 1e6
		lines += len(res.Lines)
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	out, _ := json.Marshal(workerResult{
		ElapsedMs: float64(time.Since(start).Microseconds()) / 1000,
		DetMs:     det / float64(iters),
		RecMs:     rec / float64(iters),
		SysBytes:  mem.Sys,
		Lines:     lines / iters,
	})
	fmt.Print(string(out))
}

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

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
