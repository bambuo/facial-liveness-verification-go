// 引擎可行性探针：加载 ONNX、转换为 GoMLX 图、用纯 Go 后端跑一次推理。
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gomlx/compute/dtypes"
	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/core/tensors"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: spike <model.onnx>")
		os.Exit(2)
	}
	path := os.Args[1]
	m, err := parser.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read model: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	inNames, inShapes := m.Inputs()
	outNames, outShapes := m.Outputs()
	fmt.Printf("model=%s\n", path)
	fmt.Printf("inputs=%v shapes=%v\n", inNames, inShapes)
	fmt.Printf("outputs=%v shapes=%v\n", outNames, outShapes)

	store := model.NewStore()
	if err := m.VariablesToScope(store.RootScope()); err != nil {
		fmt.Fprintf(os.Stderr, "variables: %v\n", err)
		os.Exit(1)
	}

	exec := model.MustNewExec(gobackend.GetBackend(), store, func(scope *model.Scope, input *graph.Node) []*graph.Node {
		return m.CallGraph(scope, input.Graph(), map[string]*graph.Node{inNames[0]: input})
	})
	defer exec.Finalize()

	shape := append([]int(nil), inShapes[0].Dimensions...)
	n := 1
	for i, d := range shape {
		if d < 0 {
			shape[i] = 1
			d = 1
		}
		n *= d
	}
	start := time.Now()
	data := make([]float32, n)
	divisor := float32(1)
	zero := false
	if len(os.Args) > 3 {
		switch os.Args[3] {
		case "norm":
			divisor = 100
		case "zero":
			zero = true
		}
	}
	for i := range data {
		if zero {
			data[i] = 0
			continue
		}
		data[i] = float32((i%97)-48) / divisor
	}
	input := tensors.FromFlatDataAndDimensions(data, shape...)

	outs := exec.MustCall(input)
	fmt.Printf("inference=%v outputs=%d\n", time.Since(start), len(outs))
	dumpDir := ""
	if len(os.Args) > 2 {
		dumpDir = os.Args[2]
	}
	for i, t := range outs {
		if t.DType() == dtypes.Bool {
			var flat []bool
			t.MustConstFlatData(func(v any) { flat = v.([]bool) })
			trueCount := 0
			for _, v := range flat {
				if v {
					trueCount++
				}
			}
			fmt.Printf("  [%d] shape=%v dtype=Bool true=%d/%d\n", i, t.Shape().Dimensions, trueCount, len(flat))
			continue
		}
		var flat []float32
		t.MustConstFlatData(func(v any) { flat = v.([]float32) })
		head := flat
		if len(head) > 4 {
			head = head[:4]
		}
		fmt.Printf("  [%d] shape=%v dtype=%s head=%v\n", i, t.Shape().Dimensions, t.DType(), head)
		if dumpDir != "" {
			raw := make([]byte, len(flat)*4)
			for j, v := range flat {
				bits := math.Float32bits(v)
				raw[j*4] = byte(bits)
				raw[j*4+1] = byte(bits >> 8)
				raw[j*4+2] = byte(bits >> 16)
				raw[j*4+3] = byte(bits >> 24)
			}
			name := fmt.Sprintf("%s.%d.bin", dumpDir, i)
			if err := os.WriteFile(name, raw, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", name, err)
				os.Exit(1)
			}
		}
	}
}
