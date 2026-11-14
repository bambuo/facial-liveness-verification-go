// Package eval 提供识别质量评测的指标、清单解析与运行器。
package eval

import (
	"math"
	"sort"
)

// ThresholdPoint 是某一阈值下的识别指标。
type ThresholdPoint struct {
	Threshold float64 `json:"threshold"`
	FAR       float64 `json:"far"`
	FRR       float64 `json:"frr"`
	TAR       float64 `json:"tar"`
}

// DefaultThresholds 生成 0 到 1 的等距阈值，并追加一个高于 1 的阈值。
func DefaultThresholds(step float64) []float64 {
	if step <= 0 || step > 1 {
		step = 0.01
	}
	n := int(math.Round(1 / step))
	out := make([]float64, 0, n+2)
	for i := 0; i <= n; i++ {
		out = append(out, float64(i)*step)
	}
	return append(out, 1.0001)
}

// Sweep 计算同人/异人分数在各阈值下的 FAR、FRR 与 TAR。
func Sweep(genuine, impostor []float64, thresholds []float64) []ThresholdPoint {
	g := append([]float64(nil), genuine...)
	im := append([]float64(nil), impostor...)
	sort.Float64s(g)
	sort.Float64s(im)
	out := make([]ThresholdPoint, 0, len(thresholds))
	for _, t := range thresholds {
		point := ThresholdPoint{Threshold: t}
		if len(im) > 0 {
			accepted := len(im) - sort.SearchFloat64s(im, t)
			point.FAR = float64(accepted) / float64(len(im))
		} else {
			point.FAR = math.NaN()
		}
		if len(g) > 0 {
			rejected := sort.SearchFloat64s(g, t)
			point.FRR = float64(rejected) / float64(len(g))
			point.TAR = 1 - point.FRR
		} else {
			point.FRR = math.NaN()
			point.TAR = math.NaN()
		}
		out = append(out, point)
	}
	return out
}

// EER 返回 FAR 与 FRR 最接近的点和该点阈值。
func EER(points []ThresholdPoint) (eer, threshold float64) {
	bestDiff := math.Inf(1)
	for _, p := range points {
		if math.IsNaN(p.FAR) || math.IsNaN(p.FRR) {
			continue
		}
		diff := math.Abs(p.FAR - p.FRR)
		if diff < bestDiff {
			bestDiff = diff
			eer = (p.FAR + p.FRR) / 2
			threshold = p.Threshold
		}
	}
	return eer, threshold
}

// ROCAUC 用秩和计算 ROC 曲线下面积，时间复杂度 O((n+m) log(n+m))。
func ROCAUC(genuine, impostor []float64) float64 {
	if len(genuine) == 0 || len(impostor) == 0 {
		return math.NaN()
	}
	type item struct {
		score float64
		pos   bool
	}
	items := make([]item, 0, len(genuine)+len(impostor))
	for _, s := range genuine {
		items = append(items, item{score: s, pos: true})
	}
	for _, s := range impostor {
		items = append(items, item{score: s})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].score < items[j].score })

	var positiveRankSum float64
	for i := 0; i < len(items); {
		j := i + 1
		for j < len(items) && items[j].score == items[i].score {
			j++
		}
		avgRank := float64(i+1+j) / 2
		for k := i; k < j; k++ {
			if items[k].pos {
				positiveRankSum += avgRank
			}
		}
		i = j
	}
	n := float64(len(genuine))
	m := float64(len(impostor))
	return (positiveRankSum - n*(n+1)/2) / (n * m)
}

// RecommendThreshold 返回使 FAR 不超过目标的第一个阈值及其 FRR。
func RecommendThreshold(genuine, impostor []float64, thresholds []float64, targetFAR float64) (threshold, frr float64, ok bool) {
	for _, p := range Sweep(genuine, impostor, thresholds) {
		if math.IsNaN(p.FAR) {
			continue
		}
		if p.FAR <= targetFAR {
			return p.Threshold, p.FRR, true
		}
	}
	return 0, math.NaN(), false
}

// LivenessPoint 是某一阈值下的活体指标。
type LivenessPoint struct {
	Threshold float64 `json:"threshold"`
	APCER     float64 `json:"apcer"`
	BPCER     float64 `json:"bpcer"`
}

// LivenessSweep 计算攻击样本通过率与真人拒绝率；分数为序列最低 live 分。
func LivenessSweep(bonaFide, attack []float64, thresholds []float64) []LivenessPoint {
	out := make([]LivenessPoint, 0, len(thresholds))
	for _, t := range thresholds {
		point := LivenessPoint{Threshold: t}
		if len(attack) > 0 {
			pass := 0
			for _, s := range attack {
				if s >= t {
					pass++
				}
			}
			point.APCER = float64(pass) / float64(len(attack))
		} else {
			point.APCER = math.NaN()
		}
		if len(bonaFide) > 0 {
			reject := 0
			for _, s := range bonaFide {
				if s < t {
					reject++
				}
			}
			point.BPCER = float64(reject) / float64(len(bonaFide))
		} else {
			point.BPCER = math.NaN()
		}
		out = append(out, point)
	}
	return out
}

// LivenessAt 返回指定阈值下的活体指标。
func LivenessAt(bonaFide, attack []float64, threshold float64) LivenessPoint {
	points := LivenessSweep(bonaFide, attack, []float64{threshold})
	return points[0]
}

// MinMean 返回一组分数的最小值与均值。
func MinMean(scores []float64) (min, mean float64) {
	if len(scores) == 0 {
		return math.NaN(), math.NaN()
	}
	min = scores[0]
	sum := 0.0
	for _, s := range scores {
		if s < min {
			min = s
		}
		sum += s
	}
	return min, sum / float64(len(scores))
}
