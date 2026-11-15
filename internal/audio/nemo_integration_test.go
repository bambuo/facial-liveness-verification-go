package audio

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/core/tensors"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx/parser"
)

// TestNeMoEndToEnd 用真实 NeMo CTC 模型验证 fbank + per-feature normalize + CTC。
// 需要设置 NEMO_MODEL_DIR 指向 sherpa-onnx-nemo-ctc-en-conformer-small 目录。
func TestNeMoEndToEnd(t *testing.T) {
	dir := os.Getenv("NEMO_MODEL_DIR")
	if dir == "" {
		t.Skip("set NEMO_MODEL_DIR to run the NeMo end-to-end test")
	}
	for _, wavName := range []string{"0.wav", "1.wav"} {
		t.Run(wavName, func(t *testing.T) {
			runNeMoEndToEnd(t, dir, wavName)
		})
	}
}

func runNeMoEndToEnd(t *testing.T, dir, wavName string) {
	modelPath := filepath.Join(dir, "model.int8.onnx")
	tokensPath := filepath.Join(dir, "tokens.txt")
	wavPath := filepath.Join(dir, "test_wavs", wavName)
	transPath := filepath.Join(dir, "test_wavs", "trans.txt")
	for _, p := range []string{modelPath, tokensPath, wavPath, transPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing fixture %s: %v", p, err)
		}
	}

	sr, samples, err := ReadWAV(wavPath)
	if err != nil {
		t.Fatal(err)
	}
	if sr != 16000 {
		t.Fatalf("sample rate=%d, want 16000", sr)
	}
	feats, err := ComputeNeMoFeatures(samples)
	if err != nil {
		t.Fatal(err)
	}
	frames := len(feats)
	flat := make([]float32, 80*frames)
	for f := 0; f < frames; f++ {
		for b := 0; b < 80; b++ {
			flat[b*frames+f] = feats[f][b] // [1,80,T]
		}
	}

	mdl, err := parser.ParseFile(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mdl.Close()
	store := model.NewStore()
	if err := mdl.VariablesToScope(store.RootScope()); err != nil {
		t.Fatal(err)
	}
	exec := model.MustNewExec(gobackend.GetBackend(), store, func(scope *model.Scope, a, b *graph.Node) []*graph.Node {
		return mdl.CallGraph(scope, a.Graph(), map[string]*graph.Node{"audio_signal": a, "length": b})
	}).WithDynamicAxes([]string{"batch", "", "time"}, []string{"batch"})
	defer exec.Finalize()

	audioSignal := tensors.FromFlatDataAndDimensions(flat, 1, 80, frames)
	length := tensors.FromFlatDataAndDimensions([]int64{int64(frames)}, 1)
	outs := exec.MustCall(audioSignal, length)
	defer outs[0].FinalizeAll()
	dims := outs[0].Shape().Dimensions
	if len(dims) != 3 || dims[0] != 1 || dims[2] != 1025 {
		t.Fatalf("unexpected logprobs shape %v", dims)
	}
	var logits []float32
	outs[0].MustConstFlatData(func(v any) { logits = v.([]float32) })
	T, V := dims[1], dims[2]
	logprobs := make([][]float32, T)
	for i := 0; i < T; i++ {
		logprobs[i] = logits[i*V : (i+1)*V]
	}

	tokens, err := LoadTokens(tokensPath)
	if err != nil {
		t.Fatal(err)
	}
	text, _ := GreedyDecode(logprobs, tokens)
	want := readTranscript(transPath, wavName)
	gotNorm := normalizeText(text)
	wantNorm := normalizeText(want)
	wer := wordErrorRate(gotNorm, wantNorm)
	t.Logf("transcript: %s", text)
	t.Logf("expected:   %s", want)
	t.Logf("WER:        %.4f", wer)
	if wer > 0.25 {
		t.Fatalf("WER %.4f exceeds 0.25", wer)
	}
}

func readTranscript(path, key string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == key {
			return strings.Join(fields[1:], " ")
		}
	}
	return ""
}

func normalizeText(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func wordErrorRate(got, want string) float64 {
	g := strings.Fields(got)
	w := strings.Fields(want)
	if len(w) == 0 {
		if len(g) == 0 {
			return 0
		}
		return math.Inf(1)
	}
	prev := make([]int, len(w)+1)
	cur := make([]int, len(w)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(g); i++ {
		cur[0] = i
		for j := 1; j <= len(w); j++ {
			cost := 0
			if g[i-1] != w[j-1] {
				cost = 1
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return float64(prev[len(w)]) / float64(len(w))
}

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
