package audio

import "math"

// NormalizePerFeature 按 NeMo per_feature 方式对每个 mel 维做时间轴归一化。
// 使用参考值中心化，减少低能量帧的浮点抵消。
func NormalizePerFeature(features [][]float32) [][]float32 {
	if len(features) == 0 {
		return nil
	}
	frames := len(features)
	bins := len(features[0])
	out := make([][]float32, frames)
	for i := range out {
		out[i] = make([]float32, bins)
	}
	for b := 0; b < bins; b++ {
		ref := float64(features[0][b])
		var sum float64
		for f := 0; f < frames; f++ {
			sum += float64(features[f][b]) - ref
		}
		mean := ref + sum/float64(frames)
		var variance float64
		if frames > 1 {
			for f := 0; f < frames; f++ {
				d := float64(features[f][b]) - mean
				variance += d * d
			}
			variance /= float64(frames - 1)
		}
		std := math.Sqrt(math.Max(variance, 0)) + 1e-5
		for f := 0; f < frames; f++ {
			out[f][b] = float32((float64(features[f][b]) - mean) / std)
		}
	}
	return out
}

// ComputeNeMoFeatures 执行 NeMo fbank 与 per_feature 归一化，返回模型输入特征。
func ComputeNeMoFeatures(samples []float32) ([][]float32, error) {
	feats, err := ComputeFbank(samples, NeMoFbankOptions())
	if err != nil {
		return nil, err
	}
	return NormalizePerFeature(feats), nil
}
