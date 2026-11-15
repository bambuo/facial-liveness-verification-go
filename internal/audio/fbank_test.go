package audio

import (
	"encoding/binary"
	"math"
	"os"
	"testing"
)

func TestFbankGolden(t *testing.T) {
	sr, samples, err := ReadWAV("testdata/sine_16k.wav")
	if err != nil {
		t.Fatal(err)
	}
	if sr != 16000 {
		t.Fatalf("sample rate=%d, want 16000", sr)
	}

	got, err := ComputeFbank(samples, NeMoFbankOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 || len(got[0]) != 80 {
		t.Fatalf("shape=%dx%d, want 100x80", len(got), len(got[0]))
	}

	raw, err := os.ReadFile("testdata/fbank_golden.f32")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 100*80*4 {
		t.Fatalf("golden size=%d, want %d", len(raw), 100*80*4)
	}
	maxErr := 0.0
	sumAbs := 0.0
	worstFrame, worstBin := 0, 0
	for f := 0; f < 100; f++ {
		for b := 0; b < 80; b++ {
			off := (f*80 + b) * 4
			want := math.Float32frombits(binary.LittleEndian.Uint32(raw[off : off+4]))
			diff := math.Abs(float64(got[f][b] - want))
			sumAbs += diff
			if diff > maxErr {
				maxErr = diff
				worstFrame, worstBin = f, b
			}
		}
	}
	meanAbs := sumAbs / float64(100*80)
	t.Logf("fbank maxErr=%.8g meanAbsErr=%.8g", maxErr, meanAbs)
	if meanAbs > 1e-4 {
		t.Fatalf("fbank mean abs error %.8g exceeds 1e-4", meanAbs)
	}
	if maxErr > 2e-3 {
		off := (worstFrame*80 + worstBin) * 4
		want := math.Float32frombits(binary.LittleEndian.Uint32(raw[off : off+4]))
		t.Fatalf("fbank mismatch: max abs error %.8g at frame=%d bin=%d; got=%.8f want=%.8f",
			maxErr, worstFrame, worstBin, got[worstFrame][worstBin], want)
	}
}

func TestNormalizePerFeature(t *testing.T) {
	in := [][]float32{
		{1, 10},
		{3, 20},
		{5, 30},
	}
	got := NormalizePerFeature(in)
	if len(got) != 3 || len(got[0]) != 2 {
		t.Fatalf("shape=%dx%d, want 3x2", len(got), len(got[0]))
	}
	want := [][]float32{{-1, -1}, {0, 0}, {1, 1}}
	for f := range want {
		for b := range want[f] {
			if math.Abs(float64(got[f][b]-want[f][b])) > 1e-4 {
				t.Fatalf("got[%d][%d]=%f, want %f", f, b, got[f][b], want[f][b])
			}
		}
	}
}

func TestGreedyDecode(t *testing.T) {
	tokens := []string{"<unk>", "▁hello", "▁world", "s"}
	logprobs := [][]float32{
		{9, 1, 0, 0},
		{0, 9, 1, 0},
		{0, 9, 1, 0},
		{9, 0, 0, 0},
		{0, 0, 9, 0},
		{0, 0, 0, 9},
		{0, 0, 0, 9},
	}
	text, ids := GreedyDecode(logprobs, tokens)
	if text != "hello worlds" {
		t.Fatalf("text=%q, want %q", text, "hello worlds")
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf("ids=%v, want [1 2 3]", ids)
	}
}

func TestLoadTokens(t *testing.T) {
	path := t.TempDir() + "/tokens.txt"
	if err := os.WriteFile(path, []byte("<unk> 0\n▁hello 1\ns 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tokens, err := LoadTokens(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 3 || tokens[0] != "<unk>" || tokens[1] != "▁hello" || tokens[2] != "s" {
		t.Fatalf("tokens=%v", tokens)
	}
}

func TestDigitsOnly(t *testing.T) {
	cases := map[string]string{
		"one two three four five six": "123456",
		"Zero, nine, 3!":              "093",
		"after early nightfall":       "",
	}
	for in, want := range cases {
		if got := DigitsOnly(in); got != want {
			t.Fatalf("DigitsOnly(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestFbankDitherDeterministic(t *testing.T) {
	sr, samples, err := ReadWAV("testdata/sine_16k.wav")
	if err != nil || sr != 16000 {
		t.Fatalf("read wav: %v sr=%d", err, sr)
	}
	opts := NeMoFbankOptions()
	opts.Dither = 0.003
	opts.DitherSeed = 42
	a, err := ComputeFbank(samples, opts)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ComputeFbank(samples, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) || len(a[0]) != len(b[0]) {
		t.Fatal("dither changed feature shape")
	}
	for f := range a {
		for k := range a[f] {
			if a[f][k] != b[f][k] {
				t.Fatalf("dither not reproducible at frame=%d bin=%d: %f != %f", f, k, a[f][k], b[f][k])
			}
		}
	}
	base, err := ComputeFbank(samples, NeMoFbankOptions())
	if err != nil {
		t.Fatal(err)
	}
	maxDiff := 0.0
	for f := range a {
		for k := range a[f] {
			if d := math.Abs(float64(a[f][k] - base[f][k])); d > maxDiff {
				maxDiff = d
			}
		}
	}
	if maxDiff < 1e-4 {
		t.Fatalf("dither had no effect, max diff=%g", maxDiff)
	}
}

func TestNormalRandStats(t *testing.T) {
	rng := newNormalRand(7)
	const n = 50000
	var sum, sumSq float64
	for i := 0; i < n; i++ {
		v := rng.gauss()
		sum += v
		sumSq += v * v
	}
	mean := sum / n
	std := math.Sqrt(sumSq/n - mean*mean)
	if math.Abs(mean) > 0.02 {
		t.Fatalf("gaussian mean=%f, want ~0", mean)
	}
	if math.Abs(std-1) > 0.03 {
		t.Fatalf("gaussian std=%f, want ~1", std)
	}
}
