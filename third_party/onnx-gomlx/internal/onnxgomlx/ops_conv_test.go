package onnxgomlx

import (
	"testing"

	"github.com/gomlx/compute-onnx/support/protos"
	. "github.com/gomlx/gomlx/core/graph"
)

// TestConvAutoPad 校验 ONNX Conv 的 SAME_UPPER/SAME_LOWER 显式 padding 计算。
func TestConvAutoPad(t *testing.T) {
	m := createTestModelWithDTypePromoConfig(false, false)
	conv := func(g *Graph, autoPad string, kernel int, strides int64) *Node {
		x := Const(g, [][][][]float32{{{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}})
		w := Const(g, ones4D(1, 1, kernel, kernel))
		node := &protos.NodeProto{
			OpType: "Conv",
			Input:  []string{"x", "w"},
			Output: []string{"y"},
			Attribute: []*protos.AttributeProto{
				stringAttr("auto_pad", autoPad),
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{int64(kernel), int64(kernel)}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{strides, strides}},
			},
		}
		return convertTestNode(m, g, node, map[string]*Node{"x": x, "w": w})[0]
	}

	runGoBackendTest(t, "Conv SAME_UPPER", func(g *Graph) (inputs, outputs []*Node) {
		outputs = []*Node{conv(g, "SAME_UPPER", 3, 2)}
		return
	}, []any{
		[][][][]float32{{{{12, 16}, {24, 28}}}},
	}, -1)

	runGoBackendTest(t, "Conv SAME_LOWER", func(g *Graph) (inputs, outputs []*Node) {
		outputs = []*Node{conv(g, "SAME_LOWER", 2, 2)}
		return
	}, []any{
		[][][][]float32{{{{1, 5}, {11, 28}}}},
	}, -1)
}

// ones4D 生成指定形状、元素全为 1 的四维 float32 张量。
func ones4D(n, c, h, w int) [][][][]float32 {
	out := make([][][][]float32, n)
	for i := range out {
		out[i] = make([][][]float32, c)
		for j := range out[i] {
			out[i][j] = make([][]float32, h)
			for k := range out[i][j] {
				out[i][j][k] = make([]float32, w)
				for l := range out[i][j][k] {
					out[i][j][k][l] = 1
				}
			}
		}
	}
	return out
}
