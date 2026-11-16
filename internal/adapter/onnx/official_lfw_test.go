package onnx

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"facial-liveness-verification-go/internal/eval"
)

// officialPair 是 LFW 官方 6000 对中的一条。
type officialPair struct {
	Fold  int
	Kind  string
	Path1 string
	Path2 string
	Score float64
}

// loadOfficialPairs 解析 LFW pairs.txt；前 3000 行同人，后 3000 行异人，各 10 折。
func loadOfficialPairs(path, imageRoot string) ([]officialPair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var matched, mismatched []officialPair
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		switch len(fields) {
		case 3:
			i, err1 := strconv.Atoi(fields[1])
			j, err2 := strconv.Atoi(fields[2])
			if err1 != nil || err2 != nil {
				continue
			}
			matched = append(matched, officialPair{
				Kind:  "genuine",
				Path1: filepath.Join(imageRoot, fields[0], fmt.Sprintf("%s_%04d.jpg", fields[0], i)),
				Path2: filepath.Join(imageRoot, fields[0], fmt.Sprintf("%s_%04d.jpg", fields[0], j)),
			})
		case 4:
			i, err1 := strconv.Atoi(fields[1])
			j, err2 := strconv.Atoi(fields[3])
			if err1 != nil || err2 != nil {
				continue
			}
			mismatched = append(mismatched, officialPair{
				Kind:  "impostor",
				Path1: filepath.Join(imageRoot, fields[0], fmt.Sprintf("%s_%04d.jpg", fields[0], i)),
				Path2: filepath.Join(imageRoot, fields[2], fmt.Sprintf("%s_%04d.jpg", fields[2], j)),
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(matched) != 3000 || len(mismatched) != 3000 {
		return nil, fmt.Errorf("官方配对数量异常: matched=%d mismatched=%d", len(matched), len(mismatched))
	}
	for i := 0; i < 3000; i++ {
		matched[i].Fold = i/300 + 1
		mismatched[i].Fold = i/300 + 1
	}
	return append(matched, mismatched...), nil
}

// loadEmbeddingCache 读取特征缓存；文件不存在时返回空缓存。
func loadEmbeddingCache(path string) (map[string][]float32, error) {
	cache := map[string][]float32{}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return cache, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = gob.NewDecoder(f).Decode(&cache)
	return cache, err
}

// saveEmbeddingCache 原子写入特征缓存。
func saveEmbeddingCache(path string, cache map[string][]float32) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := gob.NewEncoder(f).Encode(cache); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// foldAccuracy 返回 10 折协议下训练折选阈值、测试折评估的平均准确率与标准差。
func foldAccuracy(pairs []officialPair) (float64, float64, []float64) {
	thresholds := make([]float64, 1001)
	for i := range thresholds {
		thresholds[i] = float64(i) / 1000
	}
	accs := make([]float64, 0, 10)
	for fold := 1; fold <= 10; fold++ {
		bestThreshold, bestAcc := 0.0, -1.0
		for _, th := range thresholds {
			correct, total := 0, 0
			for _, p := range pairs {
				if p.Fold == fold {
					continue
				}
				total++
				if (p.Kind == "genuine" && p.Score >= th) || (p.Kind == "impostor" && p.Score < th) {
					correct++
				}
			}
			if acc := float64(correct) / float64(total); acc > bestAcc {
				bestAcc, bestThreshold = acc, th
			}
		}
		correct, total := 0, 0
		for _, p := range pairs {
			if p.Fold != fold {
				continue
			}
			total++
			if (p.Kind == "genuine" && p.Score >= bestThreshold) || (p.Kind == "impostor" && p.Score < bestThreshold) {
				correct++
			}
		}
		accs = append(accs, float64(correct)/float64(total))
	}
	mean, variance := 0.0, 0.0
	for _, acc := range accs {
		mean += acc
	}
	mean /= float64(len(accs))
	for _, acc := range accs {
		variance += (acc - mean) * (acc - mean)
	}
	return mean, math.Sqrt(variance / float64(len(accs))), accs
}

// TestOfficialLFW 用官方 6000 对评测，特征缓存支持分块续跑；默认跳过。
func TestOfficialLFW(t *testing.T) {
	if os.Getenv("OFFICIAL") != "1" {
		t.Skip("设置 OFFICIAL=1 运行 LFW 官方 6000 对")
	}
	modelDir := os.Getenv("MODEL_DIR")
	pairsPath := os.Getenv("PAIRS_FILE")
	imageRoot := os.Getenv("IMAGE_ROOT")
	cachePath := os.Getenv("CACHE_FILE")
	outDir := os.Getenv("OUT_DIR")
	mode := os.Getenv("REC_MODE")
	detMode := os.Getenv("DET_MODE")
	if modelDir == "" || pairsPath == "" || imageRoot == "" || cachePath == "" || outDir == "" || mode == "" {
		t.Fatal("需要 MODEL_DIR、PAIRS_FILE、IMAGE_ROOT、CACHE_FILE、OUT_DIR、REC_MODE")
	}
	if detMode == "" {
		detMode = "yunet"
	}
	maxImages, _ := strconv.Atoi(os.Getenv("MAX_IMAGES"))

	pairs, err := loadOfficialPairs(pairsPath, imageRoot)
	if err != nil {
		t.Fatalf("加载官方配对失败: %v", err)
	}
	cache, err := loadEmbeddingCache(cachePath)
	if err != nil {
		t.Fatalf("读取特征缓存失败: %v", err)
	}
	required := map[string]bool{}
	for _, p := range pairs {
		required[p.Path1] = true
		required[p.Path2] = true
	}
	pending := make([]string, 0, len(required))
	for path := range required {
		if _, ok := cache[path]; !ok {
			pending = append(pending, path)
		}
	}
	sort.Strings(pending)
	if maxImages > 0 && len(pending) > maxImages {
		pending = pending[:maxImages]
	}

	if len(pending) > 0 {
		var err error
		var det *execModel
		var scrfd *scrfdDetector
		switch detMode {
		case "yunet":
			det, err = loadModel(modelDir+"/face_detection_yunet_2023mar.onnx", yunetInputSize)
			if err != nil {
				t.Fatalf("加载检测模型失败: %v", err)
			}
			defer det.close()
		case "scrfd":
			scrfd, err = newSCRFD(os.Getenv("SCRFD_MODEL"))
			if err != nil {
				t.Fatalf("加载 SCRFD 失败: %v", err)
			}
			defer scrfd.close()
		default:
			t.Fatalf("未知 DET_MODE: %s", detMode)
		}

		var recPath string
		switch mode {
		case "aura":
			recPath = os.Getenv("AURAFACE_MODEL")
			if recPath == "" {
				t.Fatal("aura 模式需要 AURAFACE_MODEL")
			}
		case "sface":
			recPath = modelDir + "/face_recognition_sface_2021dec.onnx"
		default:
			t.Fatalf("未知 REC_MODE: %s", mode)
		}
		rec, err := loadModel(recPath, 112)
		if err != nil {
			t.Fatalf("加载识别模型失败: %v", err)
		}
		defer rec.close()

		engine := &recSpikeEngine{
			det: &Engine{
				cfg: Config{DetectionThreshold: 0.9, NMSThreshold: 0.3, AnalysisThreshold: 0.6},
				det: det,
			},
			scrfd: scrfd,
			rec:   rec,
			aura:  mode == "aura",
		}
		start := time.Now()
		for i, path := range pending {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("读取 %s: %v", path, err)
			}
			emb, err := engine.ExtractLargest([][]byte{data})
			if err != nil {
				t.Fatalf("提取 %s: %v", path, err)
			}
			cache[path] = emb
			if (i+1)%200 == 0 {
				if err := saveEmbeddingCache(cachePath, cache); err != nil {
					t.Fatalf("写缓存失败: %v", err)
				}
			}
		}
		if err := saveEmbeddingCache(cachePath, cache); err != nil {
			t.Fatalf("写缓存失败: %v", err)
		}
		fmt.Printf("processed=%d remaining=%d elapsed=%.1fs det_p50=%.1fms rec_p50=%.1fms\n",
			len(pending), len(required)-len(cache), time.Since(start).Seconds(),
			percentile(engine.detMS, 0.5), percentile(engine.recMS, 0.5))
	}

	missing := len(required) - len(cache)
	if missing > 0 {
		fmt.Printf("embedding coverage %d/%d; rerun same command to continue\n", len(cache), len(required))
		return
	}

	var genuine, impostor []float64
	for i := range pairs {
		p := &pairs[i]
		p.Score = float64(0)
		for j := range cache[p.Path1] {
			p.Score += float64(cache[p.Path1][j]) * float64(cache[p.Path2][j])
		}
		if p.Kind == "genuine" {
			genuine = append(genuine, p.Score)
		} else {
			impostor = append(impostor, p.Score)
		}
	}
	thresholds := eval.DefaultThresholds(0.001)
	sweep := eval.Sweep(genuine, impostor, thresholds)
	eer, eerThreshold := eval.EER(sweep)
	auc := eval.ROCAUC(genuine, impostor)
	recommended, frr, achieved := eval.RecommendThreshold(genuine, impostor, thresholds, 0.001)
	foldMean, foldStd, foldAccs := foldAccuracy(pairs)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("创建输出目录失败: %v", err)
	}
	report := map[string]any{
		"engine":             mode,
		"detector":           detMode,
		"pairs":              len(pairs),
		"unique_images":      len(required),
		"eer":                eer,
		"eer_threshold":      eerThreshold,
		"auc":                auc,
		"target_far":         0.001,
		"threshold":          recommended,
		"frr_at_target":      frr,
		"target_achieved":    achieved,
		"fold_accuracy_mean": foldMean,
		"fold_accuracy_std":  foldStd,
		"fold_accuracies":    foldAccs,
	}
	raw, _ := json.MarshalIndent(report, "", "  ")
	if err := os.WriteFile(filepath.Join(outDir, "official_report.json"), raw, 0o644); err != nil {
		t.Fatalf("写报告失败: %v", err)
	}
	f, err := os.Create(filepath.Join(outDir, "official_scores.csv"))
	if err != nil {
		t.Fatalf("写分数失败: %v", err)
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{"fold", "kind", "image1", "image2", "score"})
	for _, p := range pairs {
		_ = w.Write([]string{strconv.Itoa(p.Fold), p.Kind, p.Path1, p.Path2, strconv.FormatFloat(p.Score, 'f', 6, 64)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		f.Close()
		t.Fatalf("写分数失败: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("关闭分数文件失败: %v", err)
	}

	fmt.Printf("pairs=%d EER=%.4f%% AUC=%.6f threshold=%.3f FRR=%.3f%%\n",
		len(pairs), eer*100, auc, recommended, frr*100)
	fmt.Printf("LFW 10-fold accuracy=%.4f%% ± %.4f%%\n", foldMean*100, foldStd*100)
	fmt.Printf("report=%s/official_report.json\n", outDir)
}
