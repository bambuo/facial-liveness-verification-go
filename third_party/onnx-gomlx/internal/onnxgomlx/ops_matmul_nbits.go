// Copyright 2023-2026 The GoMLX Authors. SPDX-License-Identifier: Apache-2.0

package onnxgomlx

import (
	"github.com/gomlx/compute-onnx/support/protos"
	"github.com/gomlx/compute/dtypes"
	"github.com/gomlx/compute/shapes"
	"github.com/gomlx/exceptions"
	. "github.com/gomlx/gomlx/core/graph"
)

// convertMatMulNBits implements Microsoft's block-quantized MatMulNBits operator
// via nn.QuantizedDense so backends with OpTypeFusedQuantizedDense execute fused
// sub-byte/8-bit kernels and other backends fall back cleanly.
func (m *Model) convertMatMulNBits(node *protos.NodeProto, inputs []*Node) *Node {
	if len(inputs) < 3 {
		exceptions.Panicf("MatMulNBits requires at least 3 inputs (A, B, scales), got %d in %s", len(inputs), NodeToString(node))
	}
	a, packed, scales := inputs[0], inputs[1], inputs[2]
	k, n := GetIntAttrOr(node, "K", 0), GetIntAttrOr(node, "N", 0)
	bits, block := GetIntAttrOr(node, "bits", 4), GetIntAttrOr(node, "block_size", 128)
	if k <= 0 || n <= 0 || block < 16 || block&(block-1) != 0 || (bits != 2 && bits != 4 && bits != 8) {
		exceptions.Panicf("MatMulNBits: invalid K=%d, N=%d, bits=%d, block_size=%d", k, n, bits, block)
	}
	if len(inputs) > 4 && inputs[4] != nil {
		exceptions.Panicf("MatMulNBits: g_idx is not supported")
	}
	blocks := (k + block - 1) / block
	blobBytes := block * bits / 8
	if packed.DType() != dtypes.Uint8 || packed.Shape().Size() != n*blocks*blobBytes || scales.Shape().Size() != n*blocks {
		exceptions.Panicf("MatMulNBits: invalid packed weights %s or scales %s", packed.Shape(), scales.Shape())
	}

	// Constrain A's trailing contracting dimension to K when dynamic.
	if a.Rank() == 0 {
		exceptions.Panicf("MatMulNBits: input A must have rank >= 1, got %s", a.Shape())
	}
	if a.Shape().Dim(a.Rank()-1) == shapes.DynamicDim {
		specs := DimensionSpecsFor(a)
		specs[len(specs)-1] = StaticDim(k)
		a = DynamicReshape(a, specs...)
	} else if a.Shape().Dim(a.Rank()-1) != k {
		exceptions.Panicf("MatMulNBits: input A trailing dim (%d) must match K (%d)", a.Shape().Dim(a.Rank()-1), k)
	}

	var bias *Node
	if len(inputs) > 5 && inputs[5] != nil {
		bias = inputs[5]
	}

	// Float64 activations are not supported by FusedQuantizedDense (which targets Float32/Float16/BFloat16);
	// dequantize directly in Float64 to preserve full precision.
	if a.DType() == dtypes.Float64 {
		weights := ConvertDType(unpackNBits(Reshape(packed, n, blocks, blobBytes), bits), dtypes.Float64)
		zp := matMulNBitsZeroPoint(a.Graph(), inputs, n, blocks, bits)
		zp = ConvertDType(zp, dtypes.Float64)
		if !zp.IsScalar() {
			zp = Reshape(zp, n, blocks, 1)
		}
		weights = Mul(Sub(weights, zp), Reshape(ConvertDType(scales, dtypes.Float64), n, blocks, 1))
		weights = Reshape(weights, n, blocks*block)
		if k != blocks*block {
			weights = Slice(weights, AxisRange(), AxisRange(0, k))
		}
		out := m.convertMatMul(a, Transpose(weights, 0, 1))
		if bias != nil {
			out = m.convertBinaryOp(Add, out, bias)
		}
		return out
	}

	// Dequantize weights and use a regular MatMul. This preserves dynamic
	// contracting dimensions, which FusedQuantizedDense does not support.
	g := a.Graph()
	unpacked := unpackNBits(Reshape(packed, n, blocks, blobBytes), bits)
	weights := ConvertDType(unpacked, a.DType())
	zp := matMulNBitsZeroPoint(g, inputs, n, blocks, bits)
	zp = ConvertDType(zp, a.DType())
	if !zp.IsScalar() {
		zp = Reshape(zp, n, blocks, 1)
	}
	scalesTyped := Reshape(ConvertDType(scales, a.DType()), n, blocks, 1)
	weights = Mul(Sub(weights, zp), scalesTyped)
	weights = Reshape(weights, n, blocks*block)
	if k != blocks*block {
		weights = Slice(weights, AxisRange(), AxisRange(0, k))
	}
	out := m.convertMatMul(a, Transpose(weights, 0, 1))
	if bias != nil {
		if bias.DType() != a.DType() {
			bias = ConvertDType(bias, a.DType())
		}
		out = m.convertBinaryOp(Add, out, bias)
	}
	return out

}

func matMulNBitsZeroPoint(g *Graph, inputs []*Node, n, blocks, bits int) *Node {
	if len(inputs) > 3 && inputs[3] != nil {
		zero := inputs[3]
		if bits < 8 && zero.DType() == dtypes.Uint8 {
			expectedPackedCols := (blocks*bits + 7) / 8
			if zero.Shape().Size() == n*expectedPackedCols {
				zero = unpackNBits(Reshape(zero, n, expectedPackedCols), bits)
				zero = Slice(zero, AxisRange(), AxisRange(0, blocks))
				return ConvertDType(zero, dtypes.Uint8)
			}
		}
		return Reshape(ConvertDType(zero, dtypes.Float32), n, blocks)
	}
	return Scalar(g, dtypes.Uint8, uint8(1<<(bits-1)))
}

// unpackNBits expands the trailing byte axis into 8/bits integer elements (low bits first).
func unpackNBits(packed *Node, bits int) *Node {
	if bits == 8 {
		return packed
	}
	x := ConvertDType(packed, dtypes.Int32)
	parts := make([]*Node, 8/bits)
	mask := Scalar(x.Graph(), dtypes.Int32, (1<<bits)-1)
	for i := range parts {
		part := x
		if i > 0 {
			part = BitwiseShiftRightLogical(x, Scalar(x.Graph(), dtypes.Int32, i*bits))
		}
		parts[i] = InsertAxes(BitwiseAnd(part, mask), -1)
	}
	dims := append([]int(nil), packed.Shape().Dimensions...)
	dims[len(dims)-1] *= len(parts)
	return Reshape(Concatenate(parts, -1), dims...)
}
