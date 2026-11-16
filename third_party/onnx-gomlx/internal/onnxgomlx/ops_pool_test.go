package onnxgomlx

import (
	"testing"

	"github.com/gomlx/compute-onnx/support/protos"
	"github.com/gomlx/compute/gobackend"
	. "github.com/gomlx/gomlx/core/graph"
	"github.com/stretchr/testify/require"
)

func TestPoolDefaultStrides(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()
	tests := []struct {
		name    string
		convert func(*Model, map[string]*Node, *protos.NodeProto, []*Node) *Node
		stride  []int64
		want    []float32
	}{
		{name: "average omitted", convert: convertAveragePool, want: []float32{1.5, 2.5, 3.5, 4.5}},
		{name: "max omitted", convert: convertMaxPool, want: []float32{2, 3, 4, 5}},
		{name: "average explicit one", convert: convertAveragePool, stride: []int64{1}, want: []float32{1.5, 2.5, 3.5, 4.5}},
		{name: "max explicit one", convert: convertMaxPool, stride: []int64{1}, want: []float32{2, 3, 4, 5}},
		{name: "average explicit two", convert: convertAveragePool, stride: []int64{2}, want: []float32{1.5, 3.5}},
		{name: "max explicit two", convert: convertMaxPool, stride: []int64{2}, want: []float32{2, 4}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exec := MustNewExec(backend, func(x *Node) *Node {
				node := &protos.NodeProto{Attribute: []*protos.AttributeProto{{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2}}}}
				if test.stride != nil {
					node.Attribute = append(node.Attribute, &protos.AttributeProto{Name: "strides", Type: protos.AttributeProto_INTS, Ints: test.stride})
				}
				return test.convert(nil, nil, node, []*Node{x})
			})
			defer exec.Finalize()
			got := exec.MustExec1([][][]float32{{{1, 2, 3, 4, 5}}})
			defer got.FinalizeAll()
			require.Equal(t, [][][]float32{{test.want}}, got.Value())
		})
	}
}

func TestAveragePoolCeilMode(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	run := func(countIncludePad int64, want []float32) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
				{Name: "count_include_pad", Type: protos.AttributeProto_INT, I: countIncludePad},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		})
		defer exec.Finalize()
		got := exec.MustExec1([][][]float32{{{4, 6, 8}}})
		defer got.FinalizeAll()
		require.Equal(t, [][][]float32{{want}}, got.Value())
	}

	// 尾部补零不计入分母：[4,6] -> 5，[8,pad] -> 8。
	run(0, []float32{5, 8})
	// 尾部补零计入分母：[4,6] -> 5，[8,pad] -> 4。
	run(1, []float32{5, 4})
}

func TestAveragePoolDilation(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	exec := MustNewExec(backend, func(x *Node) *Node {
		node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
			{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
			{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1}},
			{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
			{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
		}}
		return convertAveragePool(nil, nil, node, []*Node{x})
	})
	defer exec.Finalize()

	input := [][][][]float32{{{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}}}
	got := exec.MustExec1(input)
	defer got.FinalizeAll()
	require.Equal(t, [][][][]float32{{{{6, 7}, {10, 11}}}}, got.Value())
}

func TestMaxPoolWindowDilation(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	exec := MustNewExec(backend, func(x *Node) *Node {
		node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
			{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
			{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1}},
			{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
		}}
		return convertMaxPool(nil, nil, node, []*Node{x})
	}).WithDynamicAxes([]string{"batch", "", "", ""})
	defer exec.Finalize()

	input := [][][][]float32{{{{
		1, 2, 3, 4,
	}, {
		5, 6, 7, 8,
	}, {
		9, 10, 11, 12,
	}, {
		13, 14, 15, 16,
	}}}}
	got := exec.MustExec1(input)
	defer got.FinalizeAll()
	require.Equal(t, [][][][]float32{{{{11, 12}, {15, 16}}}}, got.Value())
}

func TestAveragePoolONNXSpec(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	t.Run("ceil", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{3, 3}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		})
		defer exec.Finalize()
		input := [][][][]float32{{{
			{1, 2, 3, 4},
			{5, 6, 7, 8},
			{9, 10, 11, 12},
			{13, 14, 15, 16},
		}}}
		got := exec.MustExec1(input)
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{6, 7.5}, {12, 13.5}}}}, got.Value())
	})

	t.Run("dilations VALID", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{3, 3}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1}},
				{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "auto_pad", Type: protos.AttributeProto_STRING, S: []byte("VALID")},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		})
		defer exec.Finalize()
		input := make([][][][]float32, 1)
		input[0] = make([][][]float32, 1)
		input[0][0] = make([][]float32, 7)
		for i := range input[0][0] {
			input[0][0][i] = make([]float32, 7)
			for j := range input[0][0][i] {
				input[0][0][i][j] = float32(i*7 + j + 1)
			}
		}
		got := exec.MustExec1(input)
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{17, 18, 19}, {24, 25, 26}, {31, 32, 33}}}}, got.Value())
	})

	t.Run("dilations count_include_pad", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "pads", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1, 1, 1}},
				{Name: "count_include_pad", Type: protos.AttributeProto_INT, I: 1},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		})
		defer exec.Finalize()
		input := [][][][]float32{{{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}}
		got := exec.MustExec1(input)
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{1.25, 2.5, 1.25}, {2.5, 5, 2.5}, {1.25, 2.5, 1.25}}}}, got.Value())
	})
	t.Run("same upper", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "auto_pad", Type: protos.AttributeProto_STRING, S: []byte("SAME_UPPER")},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		})
		defer exec.Finalize()
		got := exec.MustExec1([][][][]float32{{{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}})
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{3, 4.5}, {7.5, 9}}}}, got.Value())
	})

	t.Run("ceil last window starts on pad", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{3, 3}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{3, 3}},
				{Name: "pads", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1, 1, 1}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
				{Name: "count_include_pad", Type: protos.AttributeProto_INT, I: 1},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		})
		defer exec.Finalize()
		input := [][][][]float32{{{{0.8580, 0.0786}, {0.2692, 0.1537}}}}
		got := exec.MustExec1(input)
		defer got.FinalizeAll()
		require.InDeltaSlice(t, []float32{0.15105556}, flatten4D(got.Value().([][][][]float32)), 1e-6)
	})
}

func TestPoolDynamicShape(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	t.Run("average ceil", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		}).WithDynamicAxes([]string{"batch", "", "h", "w"})
		defer exec.Finalize()
		got := exec.MustExec1([][][][]float32{{{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}})
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{3, 4.5}, {7.5, 9}}}}, got.Value())
	})

	t.Run("max ceil zero reduction", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
			}}
			return convertMaxPool(nil, nil, node, []*Node{x})
		}).WithDynamicAxes([]string{"batch", "", "h", "w"})
		defer exec.Finalize()
		got := exec.MustExec1([][][][]float32{{{{1, 2}, {3, 4}}}})
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{1}}}}, got.Value())
	})

	t.Run("max same lower negative", func(t *testing.T) {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "auto_pad", Type: protos.AttributeProto_STRING, S: []byte("SAME_LOWER")},
			}}
			return convertMaxPool(nil, nil, node, []*Node{x})
		}).WithDynamicAxes([]string{"batch", "", "h", "w"})
		defer exec.Finalize()
		got := exec.MustExec1([][][][]float32{{{{-5, -4, -3}, {-2, -1, 0}, {1, 2, 3}}}})
		defer got.FinalizeAll()
		require.Equal(t, [][][][]float32{{{{-5, -3}, {1, 3}}}}, got.Value())
	})
}

func TestAveragePoolStrideAndDilation(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()
	exec := MustNewExec(backend, func(x *Node) *Node {
		node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
			{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
			{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
			{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
		}}
		return convertAveragePool(nil, nil, node, []*Node{x})
	})
	defer exec.Finalize()
	got := exec.MustExec1([][][]float32{{{1, 2, 3, 4, 5}}})
	defer got.FinalizeAll()
	require.Equal(t, [][][]float32{{{2, 4}}}, got.Value())
}

func TestPoolDynamicCountIncludePad(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()
	for _, countIncludePad := range []int64{0, 1} {
		exec := MustNewExec(backend, func(x *Node) *Node {
			node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
				{Name: "count_include_pad", Type: protos.AttributeProto_INT, I: countIncludePad},
			}}
			return convertAveragePool(nil, nil, node, []*Node{x})
		}).WithDynamicAxes([]string{"batch", "", "w"})
		got := exec.MustExec1([][][]float32{{{1, 2, 3, 4, 5}}})
		want := []float32{1.5, 3.5, 5}
		if countIncludePad == 1 {
			want = []float32{1.5, 3.5, 2.5}
		}
		require.Equal(t, [][][]float32{{want}}, got.Value())
		got.FinalizeAll()
		exec.Finalize()
	}
}

func TestPoolDynamicStrideAndDilation(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()
	exec := MustNewExec(backend, func(x *Node) *Node {
		node := &protos.NodeProto{Attribute: []*protos.AttributeProto{
			{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
			{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
			{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
		}}
		return convertAveragePool(nil, nil, node, []*Node{x})
	}).WithDynamicAxes([]string{"batch", "", "w"})
	defer exec.Finalize()
	got := exec.MustExec1([][][]float32{{{1, 2, 3, 4, 5}}})
	defer got.FinalizeAll()
	require.Equal(t, [][][]float32{{{2, 4}}}, got.Value())
}

func TestMaxPoolOptionalIndices(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	node := &protos.NodeProto{
		Output: []string{"Y", ""},
		Attribute: []*protos.AttributeProto{
			{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2}},
		},
	}
	exec := MustNewExec(backend, func(x *Node) *Node {
		return convertMaxPool(nil, nil, node, []*Node{x})
	})
	defer exec.Finalize()
	got := exec.MustExec1([][][]float32{{{1, 2, 3, 4}}})
	defer got.FinalizeAll()
	require.Equal(t, [][][]float32{{{2, 3, 4}}}, got.Value())
}

func TestMaxPoolIndices(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	run := func(dynamic bool, storageOrder int64) (any, any) {
		node := &protos.NodeProto{
			Output: []string{"Y", "I"},
			Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
				{Name: "storage_order", Type: protos.AttributeProto_INT, I: storageOrder},
			},
		}
		exec := MustNewExec(backend, func(x *Node) []*Node {
			outputs := map[string]*Node{}
			y := convertMaxPool(nil, outputs, node, []*Node{x})
			require.NotNil(t, outputs["I"])
			return []*Node{y, outputs["I"]}
		})
		if dynamic {
			exec = exec.WithDynamicAxes([]string{"batch", "", "h", "w"})
		}
		defer exec.Finalize()
		outs := exec.MustExec([][][][]float32{{{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}})
		defer outs[0].FinalizeAll()
		defer outs[1].FinalizeAll()
		return outs[0].Value(), outs[1].Value()
	}

	for _, tc := range []struct {
		name    string
		dynamic bool
		storage int64
		wantIdx [][][][]int64
	}{
		{name: "static row major", storage: 0, wantIdx: [][][][]int64{{{{4, 5}, {7, 8}}}}},
		{name: "static column major", storage: 1, wantIdx: [][][][]int64{{{{4, 7}, {5, 8}}}}},
		{name: "dynamic row major", dynamic: true, storage: 0, wantIdx: [][][][]int64{{{{4, 5}, {7, 8}}}}},
		{name: "dynamic column major", dynamic: true, storage: 1, wantIdx: [][][][]int64{{{{4, 7}, {5, 8}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values, indices := run(tc.dynamic, tc.storage)
			require.Equal(t, [][][][]float32{{{{5, 6}, {8, 9}}}}, values)
			require.Equal(t, tc.wantIdx, indices)
		})
	}
}

func TestMaxPoolIndicesTieBreak(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	node := &protos.NodeProto{
		Output: []string{"Y", "I"},
		Attribute: []*protos.AttributeProto{
			{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
			{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
		},
	}
	exec := MustNewExec(backend, func(x *Node) []*Node {
		outputs := map[string]*Node{}
		return []*Node{convertMaxPool(nil, outputs, node, []*Node{x}), outputs["I"]}
	})
	defer exec.Finalize()
	outs := exec.MustExec([][][][]float32{{{{1, 1}, {1, 1}}}})
	defer outs[0].FinalizeAll()
	defer outs[1].FinalizeAll()
	require.Equal(t, [][][][]float32{{{{1}}}}, outs[0].Value())
	require.Equal(t, [][][][]int64{{{{0}}}}, outs[1].Value())
}

func TestGlobalAveragePoolDynamic(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	exec := MustNewExec(backend, func(x *Node) *Node {
		return convertGlobalAveragePool(nil, nil, &protos.NodeProto{OpType: "GlobalAveragePool"}, []*Node{x})
	}).WithDynamicAxes([]string{"batch", "", "h", "w"})
	defer exec.Finalize()

	got1 := exec.MustExec1([][][][]float32{{{{1, 3}, {5, 7}}}})
	defer got1.FinalizeAll()
	require.InDeltaSlice(t, []float32{4}, got1.Value().([][][][]float32)[0][0][0], 1e-6)

	got2 := exec.MustExec1([][][][]float32{{{{2, 4, 6}}}, {{{10, 20, 30}}}})
	defer got2.FinalizeAll()
	vals := got2.Value().([][][][]float32)
	require.InDelta(t, 4.0, vals[0][0][0][0], 1e-6)
	require.InDelta(t, 20.0, vals[1][0][0][0], 1e-6)
}

func flatten4D(v [][][][]float32) []float32 {
	var out []float32
	for _, a := range v {
		for _, b := range a {
			for _, c := range b {
				out = append(out, c...)
			}
		}
	}
	return out
}

func TestMaxPoolIndicesDilationAndPadding(t *testing.T) {
	backend, err := gobackend.New("")
	require.NoError(t, err)
	defer backend.Finalize()

	run := func(node *protos.NodeProto, input [][][][]float32) (any, any) {
		exec := MustNewExec(backend, func(x *Node) []*Node {
			outputs := map[string]*Node{}
			y := convertMaxPool(nil, outputs, node, []*Node{x})
			return []*Node{y, outputs["I"]}
		})
		defer exec.Finalize()
		outs := exec.MustExec(input)
		defer outs[0].FinalizeAll()
		defer outs[1].FinalizeAll()
		return outs[0].Value(), outs[1].Value()
	}

	t.Run("dilation ceil", func(t *testing.T) {
		node := &protos.NodeProto{
			Output: []string{"Y", "I"},
			Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1}},
				{Name: "dilations", Type: protos.AttributeProto_INTS, Ints: []int64{2, 2}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
			},
		}
		input := [][][][]float32{{{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}, {13, 14, 15, 16}}}}
		values, indices := run(node, input)
		require.Equal(t, [][][][]float32{{{{11, 12}, {15, 16}}}}, values)
		require.Equal(t, [][][][]int64{{{{10, 11}, {14, 15}}}}, indices)
	})

	t.Run("padding excluded", func(t *testing.T) {
		node := &protos.NodeProto{
			Output: []string{"Y", "I"},
			Attribute: []*protos.AttributeProto{
				{Name: "kernel_shape", Type: protos.AttributeProto_INTS, Ints: []int64{3, 3}},
				{Name: "strides", Type: protos.AttributeProto_INTS, Ints: []int64{3, 3}},
				{Name: "pads", Type: protos.AttributeProto_INTS, Ints: []int64{1, 1, 1, 1}},
				{Name: "ceil_mode", Type: protos.AttributeProto_INT, I: 1},
			},
		}
		input := [][][][]float32{{{{1, 2}, {3, 4}}}}
		values, indices := run(node, input)
		require.Equal(t, [][][][]float32{{{{4}}}}, values)
		require.Equal(t, [][][][]int64{{{{3}}}}, indices)
	})
}
