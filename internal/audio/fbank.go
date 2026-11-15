// Package audio 提供纯 Go 的语音前端：Kaldi 兼容 fbank、NeMo 归一化与 CTC 解码。
package audio

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// LogGuard 指定 fbank 对数能量的零值保护方式。
type LogGuard int

const (
	// LogGuardFloor 使用 Kaldi 的 log(max(E, eps))。
	LogGuardFloor LogGuard = iota
	// LogGuardAdd 使用 NeMo 的 log(E + eps)。
	LogGuardAdd
)

// FbankOptions 描述 Kaldi/NeMo fbank 的帧提取与 mel 参数。
type FbankOptions struct {
	SampleRate     int
	NumBins        int
	LowFreq        float64
	HighFreq       float64 // <=0 表示相对 Nyquist 的偏移
	FrameLengthMs  float64
	FrameShiftMs   float64
	Dither         float64 // 标准差；推理默认 0
	DitherSeed     int64   // 0 表示按时间随机；非 0 可复现
	Preemph        float64
	RemoveDCOffset bool
	SnipEdges      bool
	WindowType     string // povey、hann、hamming
	LogGuard       LogGuard
	LogGuardValue  float64
}

// DefaultFbankOptions 返回 Kaldi 兼容默认值；NeMo CTC 使用 NeMoFbankOptions。
func DefaultFbankOptions() FbankOptions {
	return FbankOptions{
		SampleRate:     16000,
		NumBins:        80,
		LowFreq:        20,
		HighFreq:       -400,
		FrameLengthMs:  25,
		FrameShiftMs:   10,
		Preemph:        0.97,
		RemoveDCOffset: true,
		SnipEdges:      false,
		WindowType:     "povey",
		LogGuard:       LogGuardFloor,
		LogGuardValue:  1.1920928955078125e-07, // FLT_EPSILON
	}
}

// NeMoFbankOptions 返回 NeMo per_feature 模型的 fbank 参数：log(E + 2^-24)。
func NeMoFbankOptions() FbankOptions {
	opts := DefaultFbankOptions()
	opts.LogGuard = LogGuardAdd
	opts.LogGuardValue = math.Pow(2, -24)
	return opts
}

// ComputeFbank 按 Kaldi 帧提取、Povey 窗、mel 滤波与对数变换计算 fbank。
// 返回帧数 × NumBins 的二维特征。
func ComputeFbank(samples []float32, opts FbankOptions) ([][]float32, error) {
	if opts.Dither < 0 {
		return nil, errors.New("fbank: dither must be >= 0")
	}
	if opts.NumBins < 3 {
		return nil, fmt.Errorf("fbank: num_bins must be >= 3, got %d", opts.NumBins)
	}
	if opts.SampleRate <= 0 || opts.FrameLengthMs <= 0 || opts.FrameShiftMs <= 0 {
		return nil, errors.New("fbank: sample rate and frame lengths must be positive")
	}
	frameLength := int(math.Round(float64(opts.SampleRate) * opts.FrameLengthMs / 1000))
	frameShift := int(math.Round(float64(opts.SampleRate) * opts.FrameShiftMs / 1000))
	if frameLength <= 0 || frameShift <= 0 {
		return nil, errors.New("fbank: frame length/shift must be positive")
	}
	paddedSize := nextPowerOfTwo(frameLength)
	window, err := makeWindow(opts.WindowType, frameLength)
	if err != nil {
		return nil, err
	}
	mel, err := makeMelBanks(opts, paddedSize)
	if err != nil {
		return nil, err
	}
	numFrames := numFrames(len(samples), frameLength, frameShift, opts.SnipEdges)
	if numFrames == 0 {
		return nil, nil
	}
	var rng *normalRand
	if opts.Dither > 0 {
		rng = newNormalRand(opts.DitherSeed)
	}
	out := make([][]float32, numFrames)
	frame := make([]float32, paddedSize)
	for f := 0; f < numFrames; f++ {
		start := firstSampleOfFrame(f, frameLength, frameShift, opts.SnipEdges)
		extractFrame(samples, start, frameLength, frame)
		for i := frameLength; i < paddedSize; i++ {
			frame[i] = 0
		}
		if rng != nil {
			ditherFrame(frame[:frameLength], opts.Dither, rng)
		}
		removeDC(frame[:frameLength], opts.RemoveDCOffset)
		preemphasize(frame[:frameLength], float32(opts.Preemph))
		for i := 0; i < frameLength; i++ {
			frame[i] *= window[i]
		}
		power := rfftPower(frame, paddedSize)
		out[f] = make([]float32, opts.NumBins)
		for b := range out[f] {
			var energy float32
			for k, w := range mel[b] {
				energy += w * power[k]
			}
			switch opts.LogGuard {
			case LogGuardAdd:
				energy = float32(math.Log(float64(energy) + opts.LogGuardValue))
			case LogGuardFloor:
				energy = float32(math.Log(math.Max(float64(energy), opts.LogGuardValue)))
			default:
				return nil, fmt.Errorf("fbank: unknown log guard %d", opts.LogGuard)
			}
			out[f][b] = energy
		}
	}
	return out, nil
}

func nextPowerOfTwo(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func makeWindow(name string, n int) ([]float32, error) {
	w := make([]float32, n)
	a := 2 * math.Pi / float64(n-1)
	if name == "hann" || name == "hanning" {
		a = 2 * math.Pi / float64(n)
	}
	for i := 0; i < n; i++ {
		x := a * float64(i)
		switch name {
		case "povey":
			w[i] = float32(math.Pow(0.5-0.5*math.Cos(x), 0.85))
		case "hann", "hanning":
			w[i] = float32(0.5 - 0.5*math.Cos(x))
		case "hamming":
			w[i] = float32(0.54 - 0.46*math.Cos(x))
		case "rectangular":
			w[i] = 1
		default:
			return nil, fmt.Errorf("fbank: unsupported window %q", name)
		}
	}
	return w, nil
}

func makeMelBanks(opts FbankOptions, paddedSize int) ([][]float32, error) {
	nyquist := 0.5 * float64(opts.SampleRate)
	high := opts.HighFreq
	if high <= 0 {
		high = nyquist + high
	}
	if opts.LowFreq < 0 || opts.LowFreq >= nyquist || high <= opts.LowFreq || high > nyquist {
		return nil, fmt.Errorf("fbank: invalid low/high frequency %.1f/%.1f (nyquist %.1f)", opts.LowFreq, high, nyquist)
	}
	numFFTBins := paddedSize / 2
	fftBinWidth := float32(opts.SampleRate) / float32(paddedSize)
	melLow := melScale(float32(opts.LowFreq))
	melHigh := melScale(float32(high))
	delta := (melHigh - melLow) / float32(opts.NumBins+1)
	banks := make([][]float32, opts.NumBins)
	for b := range banks {
		left := melLow + float32(b)*delta
		center := melLow + float32(b+1)*delta
		right := melLow + float32(b+2)*delta
		banks[b] = make([]float32, numFFTBins)
		for i := 0; i < numFFTBins; i++ {
			mel := melScale(fftBinWidth * float32(i))
			if mel <= left || mel >= right {
				continue
			}
			if mel <= center {
				banks[b][i] = (mel - left) / (center - left)
			} else {
				banks[b][i] = (right - mel) / (right - center)
			}
		}
	}
	return banks, nil
}

func melScale(f float32) float32 {
	return float32(1127.0 * math.Log1p(float64(f)/700.0))
}

func numFrames(numSamples, frameLength, frameShift int, snipEdges bool) int {
	if numSamples <= 0 {
		return 0
	}
	if snipEdges {
		if numSamples < frameLength {
			return 0
		}
		return 1 + (numSamples-frameLength)/frameShift
	}
	return (numSamples + frameShift/2) / frameShift
}

func firstSampleOfFrame(frame, frameLength, frameShift int, snipEdges bool) int {
	if snipEdges {
		return frame * frameShift
	}
	return frame*frameShift + frameShift/2 - frameLength/2
}

func extractFrame(samples []float32, start, frameLength int, dst []float32) {
	for i := 0; i < frameLength; i++ {
		idx := start + i
		for idx < 0 || idx >= len(samples) {
			if idx < 0 {
				idx = -idx - 1
			} else {
				idx = 2*len(samples) - 1 - idx
			}
		}
		dst[i] = samples[idx]
	}
}

func removeDC(frame []float32, enabled bool) {
	if !enabled || len(frame) == 0 {
		return
	}
	var sum float32
	for _, v := range frame {
		sum += v
	}
	mean := sum / float32(len(frame))
	for i := range frame {
		frame[i] -= mean
	}
}

func preemphasize(frame []float32, coeff float32) {
	if coeff == 0 || len(frame) == 0 {
		return
	}
	for i := len(frame) - 1; i > 0; i-- {
		frame[i] -= coeff * frame[i-1]
	}
	frame[0] -= coeff * frame[0]
}

// normalRand 是 Box-Muller 高斯采样器，状态仅属于单次 ComputeFbank 调用。
type normalRand struct {
	state    uint64
	spare    float64
	hasSpare bool
}

func newNormalRand(seed int64) *normalRand {
	if seed == 0 {
		seed = time.Now().UnixNano() | 1
	}
	return &normalRand{state: uint64(seed)}
}

func (r *normalRand) nextUint64() uint64 {
	x := r.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	r.state = x
	return x * 2685821657736338717
}

func (r *normalRand) uniform() float64 {
	return (float64(r.nextUint64()>>11) + 1.0) / (float64(uint64(1)<<53) + 2.0)
}

func (r *normalRand) gauss() float64 {
	if r.hasSpare {
		r.hasSpare = false
		return r.spare
	}
	u1 := r.uniform()
	u2 := r.uniform()
	mag := math.Sqrt(-2 * math.Log(u1))
	r.spare = mag * math.Sin(2*math.Pi*u2)
	r.hasSpare = true
	return mag * math.Cos(2*math.Pi*u2)
}

func ditherFrame(frame []float32, amount float64, rng *normalRand) {
	for i := range frame {
		frame[i] += float32(rng.gauss() * amount)
	}
}
