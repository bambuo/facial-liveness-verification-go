package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVoiceEngine 用真实 CAM++ 与 NeMo 模型验证声纹维度与数字转写。
func TestVoiceEngine(t *testing.T) {
	dir := os.Getenv("VOICE_MODEL_DIR")
	if dir == "" {
		t.Skip("set VOICE_MODEL_DIR to run the voice engine test")
	}
	engine, err := NewVoiceEngine(VoiceModelPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	wavPath := filepath.Join(dir, "sherpa-onnx-nemo-ctc-en-conformer-small", "test_wavs", "0.wav")
	sr, samples, err := ReadWAV(wavPath)
	if err != nil {
		t.Fatal(err)
	}
	if sr != 16000 {
		t.Fatalf("sample rate=%d, want 16000", sr)
	}
	embedding, err := engine.SpeakerEmbedding(samples)
	if err != nil {
		t.Fatal(err)
	}
	if len(embedding) != 512 {
		t.Fatalf("embedding dimension=%d, want 512", len(embedding))
	}
	text, err := engine.Transcribe(samples)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(text) == "" {
		t.Fatal("transcription is empty")
	}
}
