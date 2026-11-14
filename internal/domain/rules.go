package domain

import (
	"fmt"
	"math"
)

// Satisfies 判断单帧是否满足指定动作；baseline 为张嘴判定的基线段 mouthRatio。
func Satisfies(action string, m FrameMetric, baseline *float64, p Policy) bool {
	if !m.FaceFound {
		return false
	}
	switch action {
	case "straight":
		return math.Abs(m.YawRatio) <= p.StraightYaw
	case "turn_left":
		return m.YawRatio <= -p.YawThreshold
	case "turn_right":
		return m.YawRatio >= p.YawThreshold
	case "mouth_open":
		return baseline != nil && m.MouthRatio >= *baseline*p.MouthOpenRatio
	default:
		return true
	}
}

// RequiredPasses 返回某段通过所需的最少帧数。
func RequiredPasses(total int, ratio float64) int {
	n := int(math.Ceil(float64(total) * ratio))
	if n < 1 {
		return 1
	}
	return n
}

// ActionStats 输出动作诊断用的几何范围。
func ActionStats(metrics []FrameMetric) string {
	minY, maxY := math.Inf(1), math.Inf(-1)
	minM, maxM := math.Inf(1), math.Inf(-1)
	for _, m := range metrics {
		if !m.FaceFound {
			continue
		}
		minY = math.Min(minY, m.YawRatio)
		maxY = math.Max(maxY, m.YawRatio)
		minM = math.Min(minM, m.MouthRatio)
		maxM = math.Max(maxM, m.MouthRatio)
	}
	return fmt.Sprintf("yaw %.3f~%.3f, mouth %.3f~%.3f", minY, maxY, minM, maxM)
}

// AggregateLiveness 将逐帧计数聚合为活体结论；judged=0 时 ok=false。
func AggregateLiveness(live, attack int, minScore float64, p Policy) (LivenessVote, bool) {
	judged := live + attack
	if judged == 0 {
		return LivenessVote{}, false
	}
	ratio := float64(live) / float64(judged)
	status := LivenessReal
	if attack > 0 {
		status = LivenessSpoof
	} else if ratio < p.LiveRatioThreshold {
		status = LivenessFuzzy
	}
	return LivenessVote{Status: status, MinScore: minScore, LiveRatio: ratio}, true
}

// Compare 计算两个特征的余弦相似度；两侧均应已 L2 归一化。
func Compare(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, New(CodeServiceUnavailable, fmt.Sprintf("特征维度不一致：%d vs %d", len(a), len(b)))
	}
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot, nil
}

// RoundAway 按远离零方向四舍五入到指定小数位。
func RoundAway(v float64, digits int) float64 {
	p := math.Pow10(digits)
	return math.Round(v*p) / p
}
