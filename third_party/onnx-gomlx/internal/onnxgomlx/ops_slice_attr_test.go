package onnxgomlx

import (
	"math"
	"testing"

	"github.com/gomlx/compute-onnx/support/protos"
	. "github.com/gomlx/gomlx/core/graph"
)

// TestSliceAttributes 覆盖 opset<10 的 Slice：starts/ends/axes 为属性而非输入。
func TestSliceAttributes(t *testing.T) {
	m := createTestModelWithDTypePromoConfig(false, false)
	sliceAttrs := func(g *Graph, x *Node, starts, ends, axes []int64) *Node {
		attrs := []*protos.AttributeProto{
			{Name: "starts", Type: protos.AttributeProto_INTS, Ints: starts},
			{Name: "ends", Type: protos.AttributeProto_INTS, Ints: ends},
		}
		if axes != nil {
			attrs = append(attrs, &protos.AttributeProto{Name: "axes", Type: protos.AttributeProto_INTS, Ints: axes})
		}
		node := &protos.NodeProto{
			OpType:    "Slice",
			Input:     []string{"x"},
			Output:    []string{"y"},
			Attribute: attrs,
		}
		return convertTestNode(m, g, node, map[string]*Node{"x": x})[0]
	}
	runGoBackendTest(t, "Slice attributes", func(g *Graph) (inputs, outputs []*Node) {
		x := Const(g, []float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
		x2 := Const(g, [][]int32{{1, 2, 3}, {4, 5, 6}})
		outputs = []*Node{
			sliceAttrs(g, x, []int64{2}, []int64{6}, []int64{0}),
			sliceAttrs(g, x, []int64{-3}, []int64{math.MaxInt64}, nil),
			sliceAttrs(g, x2, []int64{0}, []int64{2}, []int64{1}),
		}
		return
	}, []any{
		[]float32{2, 3, 4, 5},
		[]float32{7, 8, 9},
		[][]int32{{1, 2}, {4, 5}},
	}, -1)
}
