// Package ocr 提供纯 Go PaddleOCR 推理与后处理。
package ocr

import (
	"fmt"
	"sync"

	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/core/tensors"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx"
	"github.com/gomlx/onnx-gomlx/onnx/parser"
)

// maxCachedExecs 限制按输入形状缓存的图数量，避免长跑时图缓存无界增长。
const maxCachedExecs = 4

// runner 是一个 ONNX 模型，按输入形状缓存已编译的 Exec。
type runner struct {
	m     onnx.Model
	store *model.Store
	input string
	mu    sync.Mutex
	execs map[string]*model.Exec
	order []string
}

// loadRunner 加载 ONNX 模型并准备按输入形状缓存的执行器。
func loadRunner(path string) (*runner, error) {
	m, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}
	inNames, _ := m.Inputs()
	if len(inNames) == 0 {
		_ = m.Close()
		return nil, fmt.Errorf("模型没有输入: %s", path)
	}
	store := model.NewStore()
	if err := m.VariablesToScope(store.RootScope()); err != nil {
		_ = m.Close()
		return nil, err
	}
	return &runner{
		m:     m,
		store: store,
		input: inNames[0],
		execs: make(map[string]*model.Exec),
	}, nil
}

// close 释放已编译执行器与模型资源。
func (r *runner) close() {
	for _, e := range r.execs {
		e.Finalize()
	}
	r.execs = nil
	r.order = nil
	_ = r.m.Close()
}

// run 执行一次推理；同一输入形状复用已编译图，超出上限时按 FIFO 释放。
func (r *runner) run(data []float32, dims ...int) ([]float32, []int, error) {
	key := fmt.Sprint(dims)
	r.mu.Lock()
	defer r.mu.Unlock()

	exec := r.execs[key]
	if exec == nil {
		if len(r.order) >= maxCachedExecs {
			oldest := r.order[0]
			r.order = r.order[1:]
			if old := r.execs[oldest]; old != nil {
				old.Finalize()
			}
			delete(r.execs, oldest)
		}
		build := func(scope *model.Scope, input *graph.Node) []*graph.Node {
			return r.m.CallGraph(scope, input.Graph(), map[string]*graph.Node{r.input: input})
		}
		exec = model.MustNewExec(gobackend.GetBackend(), r.store, build)
		r.execs[key] = exec
		r.order = append(r.order, key)
	}
	out, err := exec.Call(tensors.FromFlatDataAndDimensions(data, dims...))
	if err != nil {
		return nil, nil, err
	}
	if len(out) == 0 {
		return nil, nil, fmt.Errorf("模型没有输出")
	}
	t := out[0]
	var flat []float32
	t.MustConstFlatData(func(v any) { flat = v.([]float32) })
	return flat, t.Shape().Dimensions, nil
}
