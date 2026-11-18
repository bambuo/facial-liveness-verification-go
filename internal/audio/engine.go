package audio

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/core/tensors"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx"
	"github.com/gomlx/onnx-gomlx/onnx/parser"
)

const maxSpeakerExecs = 2

// VoiceConfig 是声纹提取与数字转写模型路径。
type VoiceConfig struct {
	SpeakerModel string
	ASRModel     string
	TokensPath   string
}

// VoiceEngine 提供 CAM++ 声纹向量与 NeMo CTC 数字转写。
type VoiceEngine struct {
	cfg VoiceConfig

	mu           sync.Mutex
	speakerModel onnx.Model
	speakerStore *model.Store
	speakerExecs map[int]*model.Exec
	speakerOrder []int
	asrModel     onnx.Model
	asrStore     *model.Store
	asrExec      *model.Exec
	tokens       []string
}

// NewVoiceEngine 加载声纹与 ASR 模型；模型缺失时返回错误，由调用方决定是否启用。
func NewVoiceEngine(cfg VoiceConfig) (*VoiceEngine, error) {
	for _, path := range []string{cfg.SpeakerModel, cfg.ASRModel, cfg.TokensPath} {
		if path == "" {
			return nil, fmt.Errorf("语音模型路径不能为空")
		}
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("语音模型不存在 %s: %w", path, err)
		}
	}
	speakerModel, speakerStore, err := loadVoiceModel(cfg.SpeakerModel)
	if err != nil {
		return nil, fmt.Errorf("加载声纹模型: %w", err)
	}
	asrModel, asrStore, err := loadVoiceModel(cfg.ASRModel)
	if err != nil {
		_ = speakerModel.Close()
		return nil, fmt.Errorf("加载 ASR 模型: %w", err)
	}
	tokens, err := LoadTokens(cfg.TokensPath)
	if err != nil {
		_ = speakerModel.Close()
		_ = asrModel.Close()
		return nil, fmt.Errorf("加载 ASR 词表: %w", err)
	}
	asrExec := model.MustNewExec(gobackend.GetBackend(), asrStore, func(scope *model.Scope, signal, length *graph.Node) []*graph.Node {
		return asrModel.CallGraph(scope, signal.Graph(), map[string]*graph.Node{
			"audio_signal": signal, "length": length,
		})
	}).WithDynamicAxes([]string{"batch", "", "time"}, []string{"batch"})
	return &VoiceEngine{
		cfg: cfg, speakerModel: speakerModel, speakerStore: speakerStore,
		speakerExecs: make(map[int]*model.Exec), asrModel: asrModel,
		asrStore: asrStore, asrExec: asrExec, tokens: tokens,
	}, nil
}

// loadVoiceModel 读取 ONNX 模型与权重作用域。
func loadVoiceModel(path string) (onnx.Model, *model.Store, error) {
	m, err := parser.ParseFile(path)
	if err != nil {
		return nil, nil, err
	}
	store := model.NewStore()
	if err := m.VariablesToScope(store.RootScope()); err != nil {
		_ = m.Close()
		return nil, nil, err
	}
	return m, store, nil
}

// Close 释放编译图与模型资源。
func (v *VoiceEngine) Close() {
	v.mu.Lock()
	defer v.mu.Unlock()
	for _, exec := range v.speakerExecs {
		exec.Finalize()
	}
	v.speakerExecs = nil
	v.speakerOrder = nil
	if v.asrExec != nil {
		v.asrExec.Finalize()
		v.asrExec = nil
	}
	if v.speakerModel != nil {
		_ = v.speakerModel.Close()
	}
	if v.asrModel != nil {
		_ = v.asrModel.Close()
	}
}

// SpeakerEmbedding 从 16kHz 单声道样本提取 512 维声纹向量。
func (v *VoiceEngine) SpeakerEmbedding(samples []float32) ([]float32, error) {
	if len(samples) == 0 {
		return nil, fmt.Errorf("音频没有采样")
	}
	feats, err := ComputeFbank(samples, DefaultFbankOptions())
	if err != nil {
		return nil, err
	}
	if len(feats) == 0 {
		return nil, fmt.Errorf("音频过短，无法提取声纹")
	}
	flat := make([]float32, 0, len(feats)*80)
	for _, row := range feats {
		flat = append(flat, row...)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	exec := v.speakerExecs[len(feats)]
	if exec == nil {
		if len(v.speakerOrder) >= maxSpeakerExecs {
			oldest := v.speakerOrder[0]
			v.speakerOrder = v.speakerOrder[1:]
			if old := v.speakerExecs[oldest]; old != nil {
				old.Finalize()
			}
			delete(v.speakerExecs, oldest)
		}
		exec = model.MustNewExec(gobackend.GetBackend(), v.speakerStore, func(scope *model.Scope, input *graph.Node) []*graph.Node {
			return v.speakerModel.CallGraph(scope, input.Graph(), map[string]*graph.Node{"feats": input})
		})
		v.speakerExecs[len(feats)] = exec
		v.speakerOrder = append(v.speakerOrder, len(feats))
	}
	out, err := exec.Call(tensors.FromFlatDataAndDimensions(flat, 1, len(feats), 80))
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("声纹模型没有输出")
	}
	defer out[0].FinalizeAll()
	var embedding []float32
	out[0].MustConstFlatData(func(value any) { embedding = value.([]float32) })
	if len(embedding) == 0 {
		return nil, fmt.Errorf("声纹模型输出为空")
	}
	return embedding, nil
}

// Transcribe 从 16kHz 单声道样本做 NeMo CTC 数字转写。
func (v *VoiceEngine) Transcribe(samples []float32) (string, error) {
	if len(samples) == 0 {
		return "", fmt.Errorf("音频没有采样")
	}
	feats, err := ComputeNeMoFeatures(samples)
	if err != nil {
		return "", err
	}
	if len(feats) == 0 {
		return "", fmt.Errorf("音频过短，无法转写")
	}
	flat := make([]float32, 80*len(feats))
	for frame := 0; frame < len(feats); frame++ {
		for bin := 0; bin < 80; bin++ {
			flat[bin*len(feats)+frame] = feats[frame][bin]
		}
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	signal := tensors.FromFlatDataAndDimensions(flat, 1, 80, len(feats))
	length := tensors.FromFlatDataAndDimensions([]int64{int64(len(feats))}, 1)
	out, err := v.asrExec.Call(signal, length)
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", fmt.Errorf("ASR 模型没有输出")
	}
	defer out[0].FinalizeAll()
	dims := out[0].Shape().Dimensions
	if len(dims) != 3 {
		return "", fmt.Errorf("ASR 输出维度异常: %v", dims)
	}
	var logits []float32
	out[0].MustConstFlatData(func(value any) { logits = value.([]float32) })
	timeSteps, vocab := dims[1], dims[2]
	logprobs := make([][]float32, timeSteps)
	for i := 0; i < timeSteps; i++ {
		logprobs[i] = logits[i*vocab : (i+1)*vocab]
	}
	text, _ := GreedyDecode(logprobs, v.tokens)
	return text, nil
}

// DecodeUploadAudio 将 WAV 字节转成 16kHz 单声道样本；仅接受 WAV。
func DecodeUploadAudio(data []byte) ([]float32, error) {
	sr, samples, err := ReadWAVBytes(data)
	if err != nil {
		return nil, err
	}
	if sr != 16000 {
		samples = resampleLinear(samples, sr, 16000)
	}
	return samples, nil
}

// resampleLinear 用线性插值把采样率转到 16kHz，语音场景足够。
func resampleLinear(input []float32, fromRate, toRate int) []float32 {
	if fromRate == toRate || len(input) == 0 {
		return input
	}
	ratio := float64(fromRate) / float64(toRate)
	outLen := int(math.Floor(float64(len(input)) / ratio))
	out := make([]float32, outLen)
	for i := 0; i < outLen; i++ {
		src := float64(i) * ratio
		left := int(math.Floor(src))
		frac := float32(src - float64(left))
		a := input[left]
		b := a
		if left+1 < len(input) {
			b = input[left+1]
		}
		out[i] = a + frac*(b-a)
	}
	return out
}

// VoiceModelPaths 返回默认语音模型路径，便于配置与测试复用。
func VoiceModelPaths(modelDir string) VoiceConfig {
	return VoiceConfig{
		SpeakerModel: filepath.Join(modelDir, "wespeaker_en_voxceleb_CAM++_LM.onnx"),
		ASRModel:     filepath.Join(modelDir, "sherpa-onnx-nemo-ctc-en-conformer-small", "model.int8.onnx"),
		TokensPath:   filepath.Join(modelDir, "sherpa-onnx-nemo-ctc-en-conformer-small", "tokens.txt"),
	}
}
