package eval

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"facial-liveness-verification-go/internal/domain"
)

// Engine 是评测所需的引擎能力。
type Engine interface {
	ModelVersion() string
	ExtractLargest(frames [][]byte) ([]float32, error)
	Compare(left, right []float32) (float64, error)
	CheckLiveness(frames [][]byte, threshold, liveRatioThreshold float64) (domain.LivenessVote, error)
}

// Pair 是一对图片的比对分数。
type Pair struct {
	Kind   string  `json:"kind"`
	ID1    string  `json:"id1"`
	Image1 string  `json:"image1"`
	ID2    string  `json:"id2"`
	Image2 string  `json:"image2"`
	Score  float64 `json:"score"`
}

// LivenessSample 是一个活体序列的聚合结果。
type LivenessSample struct {
	ID        string  `json:"id"`
	Label     string  `json:"label,omitempty"`
	Frames    int     `json:"frames"`
	MinScore  float64 `json:"min_score"`
	MeanScore float64 `json:"mean_score"`
	Error     string  `json:"error,omitempty"`
}

// Counts 是评测样本计数。
type Counts struct {
	Identities    int `json:"identities"`
	Images        int `json:"images"`
	Embedded      int `json:"embedded"`
	GenuinePairs  int `json:"genuine_pairs"`
	ImpostorPairs int `json:"impostor_pairs"`
}

// LivenessCounts 是活体样本计数。
type LivenessCounts struct {
	BonaFideSequences int `json:"bona_fide_sequences"`
	AttackSequences   int `json:"attack_sequences"`
	BonaFideScored    int `json:"bona_fide_scored"`
	AttackScored      int `json:"attack_scored"`
}

// LivenessLabel 是单个攻击类型的汇总。
type LivenessLabel struct {
	Label     string  `json:"label"`
	Sequences int     `json:"sequences"`
	APCER     float64 `json:"apcer"`
	MinScore  float64 `json:"min_score"`
	MeanScore float64 `json:"mean_score"`
}

// LivenessReport 是活体评测结果。
type LivenessReport struct {
	Counts           LivenessCounts  `json:"counts"`
	ServiceThreshold float64         `json:"service_threshold"`
	ServicePoint     LivenessPoint   `json:"service_point"`
	Thresholds       []LivenessPoint `json:"thresholds"`
	ByLabel          []LivenessLabel `json:"by_label,omitempty"`
}

// Report 是识别质量评测汇总。
type Report struct {
	GeneratedAt          time.Time        `json:"generated_at"`
	EngineVersion        string           `json:"engine_version"`
	Manifest             string           `json:"manifest"`
	Counts               Counts           `json:"counts"`
	EER                  float64          `json:"eer"`
	EERThreshold         float64          `json:"eer_threshold"`
	AUC                  float64          `json:"auc"`
	TargetFAR            float64          `json:"target_far"`
	RecommendedThreshold float64          `json:"recommended_threshold"`
	FRRAtTarget          float64          `json:"frr_at_target"`
	TargetAchieved       bool             `json:"target_achieved"`
	Thresholds           []ThresholdPoint `json:"thresholds"`
	Liveness             *LivenessReport  `json:"liveness,omitempty"`
	Errors               []string         `json:"errors,omitempty"`
}

// Runner 执行一次评测。
type Runner struct {
	Engine            Engine
	Manifest          *Manifest
	Thresholds        []float64
	TargetFAR         float64
	LivenessThreshold float64
	ImageLoader       func(string) ([]byte, error)
}

// NewRunner 构造评测运行器。
func NewRunner(engine Engine, manifest *Manifest) *Runner {
	return &Runner{
		Engine:            engine,
		Manifest:          manifest,
		Thresholds:        DefaultThresholds(manifest.Options.ThresholdStep),
		TargetFAR:         0.001,
		LivenessThreshold: 0.85,
		ImageLoader:       os.ReadFile,
	}
}

// Run 计算配对分数、阈值指标与活体指标。
func (r *Runner) Run(ctx context.Context) (*Report, []Pair, []LivenessSample, error) {
	report := &Report{
		GeneratedAt:   time.Now().UTC(),
		EngineVersion: r.Engine.ModelVersion(),
		Manifest:      r.Manifest.Path,
		TargetFAR:     r.TargetFAR,
	}
	cache := map[string][]byte{}
	embeddings := map[string][]float32{}

	type group struct {
		id    string
		paths []string
	}
	var groups []group
	for _, identity := range r.Manifest.Identities {
		g := group{id: identity.ID}
		for _, rel := range identity.Images {
			if err := ctx.Err(); err != nil {
				return nil, nil, nil, err
			}
			report.Counts.Images++
			path := r.Manifest.Resolve(rel)
			data, err := r.load(cache, path)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", rel, err))
				continue
			}
			emb, err := r.Engine.ExtractLargest([][]byte{data})
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", rel, err))
				continue
			}
			embeddings[path] = emb
			g.paths = append(g.paths, path)
			report.Counts.Embedded++
		}
		if len(g.paths) > 0 {
			groups = append(groups, g)
		}
	}
	report.Counts.Identities = len(groups)
	if len(groups) < 2 {
		return nil, nil, nil, fmt.Errorf("至少需要两个可检测的 identity 才能计算同人与异人指标")
	}

	var pairs []Pair
	genuine := make([]float64, 0, 64)
	for _, g := range groups {
		for i := 0; i < len(g.paths); i++ {
			for j := i + 1; j < len(g.paths); j++ {
				score, err := r.Engine.Compare(embeddings[g.paths[i]], embeddings[g.paths[j]])
				if err != nil {
					report.Errors = append(report.Errors, fmt.Sprintf("compare %s/%s: %v", g.paths[i], g.paths[j], err))
					continue
				}
				genuine = append(genuine, score)
				pairs = append(pairs, Pair{Kind: "genuine", ID1: g.id, Image1: g.paths[i], ID2: g.id, Image2: g.paths[j], Score: score})
			}
		}
	}
	if len(genuine) == 0 {
		return nil, nil, nil, fmt.Errorf("没有可用的同人配对，请给至少一个 identity 提供两张可检测图片")
	}

	totalImpostor := 0
	for a := 0; a < len(groups); a++ {
		for b := a + 1; b < len(groups); b++ {
			totalImpostor += len(groups[a].paths) * len(groups[b].paths)
		}
	}
	stride := 1
	if max := r.Manifest.Options.MaxImpostorPairs; max > 0 && totalImpostor > max {
		stride = (totalImpostor + max - 1) / max
	}
	impostor := make([]float64, 0, totalImpostor/stride+1)
	counter := 0
	for a := 0; a < len(groups); a++ {
		for b := a + 1; b < len(groups); b++ {
			for _, p1 := range groups[a].paths {
				for _, p2 := range groups[b].paths {
					take := counter%stride == 0
					counter++
					if !take {
						continue
					}
					if err := ctx.Err(); err != nil {
						return nil, nil, nil, err
					}
					score, err := r.Engine.Compare(embeddings[p1], embeddings[p2])
					if err != nil {
						report.Errors = append(report.Errors, fmt.Sprintf("compare %s/%s: %v", p1, p2, err))
						continue
					}
					impostor = append(impostor, score)
					pairs = append(pairs, Pair{Kind: "impostor", ID1: groups[a].id, Image1: p1, ID2: groups[b].id, Image2: p2, Score: score})
				}
			}
		}
	}
	if len(impostor) == 0 {
		return nil, nil, nil, fmt.Errorf("没有可用的异人配对，请至少提供两个 identity")
	}
	report.Counts.GenuinePairs = len(genuine)
	report.Counts.ImpostorPairs = len(impostor)

	report.Thresholds = Sweep(genuine, impostor, r.Thresholds)
	report.EER, report.EERThreshold = EER(report.Thresholds)
	report.AUC = ROCAUC(genuine, impostor)
	threshold, frr, ok := RecommendThreshold(genuine, impostor, r.Thresholds, r.TargetFAR)
	report.RecommendedThreshold = threshold
	report.FRRAtTarget = frr
	report.TargetAchieved = ok

	var samples []LivenessSample
	if r.Manifest.Liveness != nil {
		var bonaScores, attackScores []float64
		byLabel := make(map[string][]float64)
		for _, seq := range r.Manifest.Liveness.BonaFide {
			sample, min := r.evalSequence(ctx, cache, seq)
			samples = append(samples, sample)
			bonaScores = append(bonaScores, min)
		}
		for _, seq := range r.Manifest.Liveness.Attacks {
			sample, min := r.evalSequence(ctx, cache, seq)
			samples = append(samples, sample)
			attackScores = append(attackScores, min)
			byLabel[seq.Label] = append(byLabel[seq.Label], min)
		}
		labels := make([]string, 0, len(byLabel))
		for label := range byLabel {
			labels = append(labels, label)
		}
		sort.Strings(labels)
		byLabelReport := make([]LivenessLabel, 0, len(labels))
		for _, label := range labels {
			scores := byLabel[label]
			passed := 0
			for _, score := range scores {
				if score >= r.LivenessThreshold {
					passed++
				}
			}
			min, mean := MinMean(scores)
			byLabelReport = append(byLabelReport, LivenessLabel{
				Label:     label,
				Sequences: len(scores),
				APCER:     float64(passed) / float64(len(scores)),
				MinScore:  min,
				MeanScore: mean,
			})
		}
		report.Liveness = &LivenessReport{
			Counts: LivenessCounts{
				BonaFideSequences: len(r.Manifest.Liveness.BonaFide),
				AttackSequences:   len(r.Manifest.Liveness.Attacks),
				BonaFideScored:    len(bonaScores),
				AttackScored:      len(attackScores),
			},
			ServiceThreshold: r.LivenessThreshold,
			ServicePoint:     LivenessAt(bonaScores, attackScores, r.LivenessThreshold),
			Thresholds:       LivenessSweep(bonaScores, attackScores, r.Thresholds),
			ByLabel:          byLabelReport,
		}
	}
	return report, pairs, samples, nil
}

func (r *Runner) load(cache map[string][]byte, path string) ([]byte, error) {
	if data, ok := cache[path]; ok {
		return data, nil
	}
	data, err := r.ImageLoader(path)
	if err != nil {
		return nil, err
	}
	cache[path] = data
	return data, nil
}

func (r *Runner) evalSequence(ctx context.Context, cache map[string][]byte, seq Sequence) (LivenessSample, float64) {
	sample := LivenessSample{ID: seq.ID, Label: seq.Label, Frames: len(seq.Images)}
	scores := make([]float64, 0, len(seq.Images))
	var errs []string
	for _, rel := range seq.Images {
		if err := ctx.Err(); err != nil {
			sample.Error = err.Error()
			break
		}
		path := r.Manifest.Resolve(rel)
		data, err := r.load(cache, path)
		if err != nil {
			errs = append(errs, rel+": "+err.Error())
			scores = append(scores, 0)
			continue
		}
		vote, err := r.Engine.CheckLiveness([][]byte{data}, 0, 1)
		if err != nil {
			errs = append(errs, rel+": "+err.Error())
			scores = append(scores, 0)
			continue
		}
		scores = append(scores, vote.MinScore)
	}
	if len(errs) > 0 {
		sample.Error = strings.Join(errs, "; ")
	}
	min, mean := MinMean(scores)
	sample.MinScore = min
	sample.MeanScore = mean
	return sample, min
}
