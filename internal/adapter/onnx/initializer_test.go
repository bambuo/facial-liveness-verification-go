package onnx

import (
	"path/filepath"
	"testing"

	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/core/tensors"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx/parser"
)

// TestInitializerInputDefaults 验证 initializer 作为 graph input 时的默认值与覆盖语义。
func TestInitializerInputDefaults(t *testing.T) {
	mdl, err := parser.ParseFile(filepath.Join("testdata", "initializer_default.onnx"))
	if err != nil {
		t.Fatal(err)
	}
	defer mdl.Close()

	names, _ := mdl.Inputs()
	if len(names) != 1 || names[0] != "x" {
		t.Fatalf("required inputs = %v, want [x]", names)
	}

	store := model.NewStore()
	if err := mdl.VariablesToScope(store.RootScope()); err != nil {
		t.Fatal(err)
	}

	execDefault := model.MustNewExec(gobackend.GetBackend(), store, func(scope *model.Scope, x *graph.Node) []*graph.Node {
		return mdl.CallGraph(scope, x.Graph(), map[string]*graph.Node{"x": x})
	})
	defer execDefault.Finalize()

	execOverride := model.MustNewExec(gobackend.GetBackend(), store, func(scope *model.Scope, x, w *graph.Node) []*graph.Node {
		return mdl.CallGraph(scope, x.Graph(), map[string]*graph.Node{"x": x, "w": w})
	})
	defer execOverride.Finalize()

	x := tensors.FromFlatDataAndDimensions([]float32{1, 1}, 1, 2)
	got := flatF32(execDefault.MustCall(x)[0])
	if got[0] != 2 || got[1] != 3 {
		t.Fatalf("default output = %v, want [2 3]", got)
	}

	w := tensors.FromFlatDataAndDimensions([]float32{10, 20}, 2)
	got = flatF32(execOverride.MustCall(x, w)[0])
	if got[0] != 10 || got[1] != 20 {
		t.Fatalf("override output = %v, want [10 20]", got)
	}
}
