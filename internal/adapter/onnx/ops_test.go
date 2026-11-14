package onnx

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/gomlx/gomlx/core/tensors"
)

// runSpecModel 加载 testdata 中的规范模型并执行固定输入。
func runSpecModel(t *testing.T, name string) ([]float32, []*tensors.Tensor) {
	t.Helper()
	m, err := loadModel(filepath.Join("testdata", name), 2)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	t.Cleanup(m.close)

	input := make([]float32, 1*3*2*2)
	for i := range input {
		input[i] = float32((i % 97) - 48)
	}
	outs, err := m.run(input)
	if err != nil {
		t.Fatalf("run %s: %v", name, err)
	}
	return input, outs
}

// TestPReluBroadcastSpec 验证 PRelu 的 slope 单向广播与负数分支。
func TestPReluBroadcastSpec(t *testing.T) {
	input, outs := runSpecModel(t, "prelu_broadcast.onnx")
	if len(outs) != 1 {
		t.Fatalf("outputs = %d", len(outs))
	}
	got := flatF32(outs[0])
	slopes := []float32{0.1, 0.2, 0.3}
	for i, v := range input {
		want := v
		if v < 0 {
			want = v * slopes[(i/4)%3]
		}
		if math.Abs(float64(got[i]-want)) > 1e-6 {
			t.Fatalf("index %d: got %v want %v", i, got[i], want)
		}
	}
}

// TestDropoutInferenceSpec 验证推理期输出等于输入且 mask 全为 true。
func TestDropoutInferenceSpec(t *testing.T) {
	for _, name := range []string{"dropout_default.onnx", "dropout_mask.onnx", "dropout_training_false.onnx"} {
		t.Run(name, func(t *testing.T) {
			input, outs := runSpecModel(t, name)
			if len(outs) == 0 {
				t.Fatal("no outputs")
			}
			got := flatF32(outs[0])
			for i := range input {
				if got[i] != input[i] {
					t.Fatalf("output[%d] = %v, want %v", i, got[i], input[i])
				}
			}
			if len(outs) > 1 {
				var mask []bool
				outs[1].MustConstFlatData(func(v any) { mask = v.([]bool) })
				if len(mask) != len(input) {
					t.Fatalf("mask size = %d, want %d", len(mask), len(input))
				}
				for i, v := range mask {
					if !v {
						t.Fatalf("mask[%d] = false, want true", i)
					}
				}
			}
		})
	}
}

// TestDropoutTrainingModeRejected 验证 training_mode=true 不被静默当作推理。
func TestDropoutTrainingModeRejected(t *testing.T) {
	rejected := false
	func() {
		defer func() {
			if recover() != nil {
				rejected = true
			}
		}()
		m, err := loadModel(filepath.Join("testdata", "dropout_training_true.onnx"), 2)
		if err != nil {
			rejected = true
			return
		}
		m.close()
	}()
	if !rejected {
		t.Fatal("training_mode=true should be rejected")
	}
}
