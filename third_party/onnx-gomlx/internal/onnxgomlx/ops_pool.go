package onnxgomlx

import (
	"fmt"
	"math"

	"github.com/gomlx/compute"
	"github.com/gomlx/compute-onnx/support/protos"
	"github.com/gomlx/compute/dtypes"
	"github.com/gomlx/compute/shapes"
	"github.com/gomlx/exceptions"
	. "github.com/gomlx/gomlx/core/graph"
	timage "github.com/gomlx/gomlx/core/tensors/images"
)

// poolGeometry 保存 ONNX 池化在空间轴上的窗口、步长、膨胀与最终 padding。
// 静态形状使用 padBegin/padEnd；动态形状使用 startNodes/endNodes。
type poolGeometry struct {
	kernel    []int
	strides   []int
	dilations []int

	padBegin []int
	padEnd   []int

	startNodes []*Node
	endNodes   []*Node
	outNodes   []*Node
	outDims    []int
	dynamic    bool
}

// convertAveragePool converts an ONNX AveragePool node to a GoMLX node.
//
// See ONNX documentation in:
// https://onnx.ai/onnx/operators/onnx__AveragePool.html
func convertAveragePool(_ *Model, _ map[string]*Node, node *protos.NodeProto, inputs []*Node) *Node {
	autoPad := GetStringAttrOr(node, "auto_pad", "NOTSET")
	ceilMode := GetIntAttrOr(node, "ceil_mode", 0)
	countIncludePad := GetIntAttrOr(node, "count_include_pad", 0)
	if countIncludePad != 0 && countIncludePad != 1 {
		exceptions.Panicf("AveragePool: count_include_pad must be 0 or 1, got %d", countIncludePad)
	}
	kernelShape := GetIntsAttrOr(node, "kernel_shape", nil)
	if kernelShape == nil {
		exceptions.Panicf("AveragePool: kernel_shape is required")
	}
	strides := GetIntsAttrOr(node, "strides", nil)
	dilations := GetIntsAttrOr(node, "dilations", nil)
	pads := GetIntsAttrOr(node, "pads", nil)

	x := inputs[0]
	g, err := computePoolGeometry(x, autoPad, pads, kernelShape, strides, dilations, ceilMode != 0)
	if err != nil {
		exceptions.Panicf("AveragePool: %v", err)
	}

	kernelSize := 1
	for _, k := range kernelShape {
		kernelSize *= k
	}
	if g.dynamic {
		return dynamicPool(x, g, true, countIncludePad, kernelSize)
	}
	padded := padPoolInput(x, g, Scalar(x.Graph(), x.DType(), 0))
	sum := reducePool(padded, g, compute.ReduceOpSum)
	if countIncludePad == 1 {
		return DivScalar(sum, float64(kernelSize))
	}

	// count_include_pad=0：用有效元素计数做分母，padding 不计入。
	mask := OnesLike(x)
	maskPadded := padPoolInput(mask, g, Scalar(mask.Graph(), mask.DType(), 0))
	count := reducePool(maskPadded, g, compute.ReduceOpSum)
	return Div(sum, count)
}

// convertMaxPool converts an ONNX MaxPool node to a GoMLX node.
//
// See ONNX documentation in:
// https://onnx.ai/onnx/operators/onnx__MaxPool.html
func convertMaxPool(_ *Model, convertedOutputs map[string]*Node, node *protos.NodeProto, inputs []*Node) *Node {
	autoPad := GetStringAttrOr(node, "auto_pad", "NOTSET")
	ceilMode := GetIntAttrOr(node, "ceil_mode", 0)
	storageOrder := GetIntAttrOr(node, "storage_order", 0)
	if storageOrder != 0 && storageOrder != 1 {
		exceptions.Panicf("MaxPool: storage_order must be 0 or 1, got %d", storageOrder)
	}
	kernelShape := GetIntsAttrOr(node, "kernel_shape", nil)
	if kernelShape == nil {
		exceptions.Panicf("MaxPool: kernel_shape is required")
	}
	strides := GetIntsAttrOr(node, "strides", nil)
	dilations := GetIntsAttrOr(node, "dilations", nil)
	pads := GetIntsAttrOr(node, "pads", nil)

	x := inputs[0]
	g, err := computePoolGeometry(x, autoPad, pads, kernelShape, strides, dilations, ceilMode != 0)
	if err != nil {
		exceptions.Panicf("MaxPool: %v", err)
	}
	if len(node.Output) > 1 && node.Output[1] != "" {
		values, indices := maxPoolWithIndices(x, g, storageOrder)
		convertedOutputs[node.Output[1]] = indices
		return values
	}
	if g.dynamic {
		return dynamicPool(x, g, false, 0, 0)
	}
	padded := padPoolInput(x, g, poolMinValue(x))
	return reducePool(padded, g, compute.ReduceOpMax)
}

// computePoolGeometry 按 ONNX 规范计算 auto_pad/ceil_mode/pads 后的最终 padding。
func computePoolGeometry(x *Node, autoPad string, pads, kernel, strides, dilations []int, ceil bool) (poolGeometry, error) {
	switch autoPad {
	case "NOTSET", "VALID", "SAME_UPPER", "SAME_LOWER":
	default:
		return poolGeometry{}, fmt.Errorf("auto_pad must be NOTSET, VALID, SAME_UPPER or SAME_LOWER, got %q", autoPad)
	}
	numSpatial := x.Rank() - 2
	if numSpatial <= 0 {
		return poolGeometry{}, fmt.Errorf("input must have at least one spatial dimension, got %s", x.Shape())
	}
	if len(kernel) != numSpatial {
		return poolGeometry{}, fmt.Errorf("kernel_shape must have %d values, got %d", numSpatial, len(kernel))
	}
	if autoPad != "NOTSET" && pads != nil {
		return poolGeometry{}, fmt.Errorf("auto_pad and pads cannot be used simultaneously")
	}
	if pads != nil && len(pads) != 2*numSpatial {
		return poolGeometry{}, fmt.Errorf("pads must have %d values, got %d", 2*numSpatial, len(pads))
	}
	if strides == nil {
		strides = make([]int, numSpatial)
		for i := range strides {
			strides[i] = 1
		}
	}
	if len(strides) != numSpatial {
		return poolGeometry{}, fmt.Errorf("strides must have %d values, got %d", numSpatial, len(strides))
	}
	if dilations == nil {
		dilations = make([]int, numSpatial)
		for i := range dilations {
			dilations[i] = 1
		}
	}
	if len(dilations) != numSpatial {
		return poolGeometry{}, fmt.Errorf("dilations must have %d values, got %d", numSpatial, len(dilations))
	}
	for i := range kernel {
		if kernel[i] <= 0 {
			return poolGeometry{}, fmt.Errorf("kernel_shape[%d] must be positive, got %d", i, kernel[i])
		}
		if strides[i] <= 0 {
			return poolGeometry{}, fmt.Errorf("strides[%d] must be positive, got %d", i, strides[i])
		}
		if dilations[i] <= 0 {
			return poolGeometry{}, fmt.Errorf("dilations[%d] must be positive, got %d", i, dilations[i])
		}
	}

	g := poolGeometry{kernel: kernel, strides: strides, dilations: dilations}
	static := !x.Shape().IsDynamic()
	for i := range kernel {
		axis := 2 + i
		effKernel := (kernel[i]-1)*dilations[i] + 1
		pB, pE := 0, 0
		if pads != nil {
			pB, pE = pads[i], pads[i+numSpatial]
			if pB < 0 || pE < 0 {
				return poolGeometry{}, fmt.Errorf("pads must be non-negative, got begin=%d end=%d", pB, pE)
			}
		}
		if static {
			pB, pE, out, err := staticPoolPads(x.Shape().Dim(axis), autoPad, pB, pE, effKernel, strides[i], ceil)
			if err != nil {
				return poolGeometry{}, err
			}
			g.padBegin = append(g.padBegin, pB)
			g.padEnd = append(g.padEnd, pE)
			g.outDims = append(g.outDims, out)
			continue
		}
		startNode, endNode, outNode, err := dynamicPoolPads(x, axis, autoPad, pB, pE, effKernel, strides[i], ceil)
		if err != nil {
			return poolGeometry{}, err
		}
		g.startNodes = append(g.startNodes, startNode)
		g.endNodes = append(g.endNodes, endNode)
		g.outNodes = append(g.outNodes, outNode)
		g.dynamic = true
	}
	return g, nil
}

// staticPoolPads 计算静态空间维的最终 padding；ceil_mode 的尾部额外补零也在此完成。
func staticPoolPads(in int, autoPad string, padBegin, padEnd, effKernel, stride int, ceil bool) (int, int, int, error) {
	if in < 0 {
		return 0, 0, 0, fmt.Errorf("dynamic dimension reached static padding calculation")
	}
	switch autoPad {
	case "VALID":
		padBegin, padEnd = 0, 0
	case "SAME_UPPER", "SAME_LOWER":
		out := ceilDiv(in, stride)
		total := maxInt(0, (out-1)*stride+effKernel-in)
		if autoPad == "SAME_UPPER" {
			padBegin = total / 2
			padEnd = total - padBegin
		} else {
			padEnd = total / 2
			padBegin = total - padEnd
		}
		return padBegin, padEnd, out, nil
	}

	num := in + padBegin + padEnd - effKernel
	if num < 0 {
		return 0, 0, 0, fmt.Errorf("pooling window %d exceeds padded input %d", effKernel, in+padBegin+padEnd)
	}
	out := 0
	if ceil {
		out = ceilDiv(num, stride) + 1
		if (out-1)*stride >= in+padBegin {
			out--
		}
	} else {
		out = num/stride + 1
	}
	extra := maxInt(0, (out-1)*stride+effKernel-(in+padBegin+padEnd))
	return padBegin, padEnd + extra, out, nil
}

// dynamicPoolPads 计算动态空间维的 padding 节点。
func dynamicPoolPads(x *Node, axis int, autoPad string, padBegin, padEnd, effKernel, stride int, ceil bool) (*Node, *Node, *Node, error) {
	g := x.Graph()
	in := DimensionSize(x, axis)
	c := func(v int) *Node { return Scalar(g, dtypes.Int64, int64(v)) }
	if autoPad == "SAME_UPPER" || autoPad == "SAME_LOWER" {
		out := ceilDivNode(in, c(stride))
		total := Max(c(0), Sub(Add(Mul(Sub(out, c(1)), c(stride)), c(effKernel)), in))
		begin := Div(total, c(2))
		end := Sub(total, begin)
		if autoPad == "SAME_LOWER" {
			begin, end = end, begin
		}
		return begin, end, out, nil
	}

	if autoPad == "VALID" {
		padBegin, padEnd = 0, 0
	}
	pB, pE := c(padBegin), c(padEnd)
	num := Sub(Add(in, Add(pB, pE)), c(effKernel))
	numClamped := Max(num, c(0))
	outFloor := Add(Div(numClamped, c(stride)), c(1))
	out := outFloor
	if ceil {
		outCeil := Add(ceilDivNode(numClamped, c(stride)), c(1))
		lastStart := Mul(Sub(outCeil, c(1)), c(stride))
		out = Where(GreaterOrEqual(lastStart, Add(in, pB)), Sub(outCeil, c(1)), outCeil)
	}
	out = Where(LessThan(num, c(0)), c(0), out)
	extra := Max(c(0), Sub(Add(Mul(Sub(out, c(1)), c(stride)), c(effKernel)), Add(Add(in, pB), pE)))
	return pB, Add(pE, extra), out, nil
}

// maxPoolWithIndices 同时计算 MaxPool 的 Y 与 Indices。
// 窗口按 kernel 轴升序遍历，严格大于才更新，因此并列最大值取最先遇到的位置。
func maxPoolWithIndices(x *Node, g poolGeometry, storageOrder int) (values, indices *Node) {
	flatIndex := flatIndexTensor(x, g, storageOrder)
	fill := poolMinValue(x)
	zero := Scalar(x.Graph(), x.DType(), 0)
	minusOne := Scalar(x.Graph(), dtypes.Int64, -1)

	var visit func(dim int, offsets []int)
	visit = func(dim int, offsets []int) {
		if dim < len(g.kernel) {
			for k := 0; k < g.kernel[dim]; k++ {
				visit(dim+1, append(append([]int(nil), offsets...), k))
			}
			return
		}
		gathered := x
		gatheredIdx := flatIndex
		var mask *Node
		for i, off := range offsets {
			axis := 2 + i
			outVec, padBegin := poolAxisIota(x, g, i)
			in := DimensionSize(x, axis)
			stride := Scalar(x.Graph(), dtypes.Int64, int64(g.strides[i]))
			shift := Scalar(x.Graph(), dtypes.Int64, int64(off*g.dilations[i]))
			raw := Sub(Add(Mul(outVec, stride), shift), padBegin)
			idx := Min(Max(raw, Scalar(x.Graph(), dtypes.Int64, 0)), Sub(in, Scalar(x.Graph(), dtypes.Int64, 1)))
			gathered = onnxGather(gathered, idx, axis)
			gatheredIdx = onnxGather(gatheredIdx, idx, axis)
			valid := LogicalAnd(GreaterOrEqual(raw, Scalar(x.Graph(), dtypes.Int64, 0)), LessThan(raw, in))
			specs := make([]DimensionSpec, x.Rank())
			for a := range specs {
				specs[a] = StaticDim(1)
			}
			specs[axis] = poolOutputSpec(g, i)
			valid = DynamicReshape(valid, specs...)
			if mask == nil {
				mask = valid
			} else {
				mask = LogicalAnd(mask, valid)
			}
		}
		mask = BroadcastLike(mask, gathered)
		if values == nil {
			values = Where(mask, gathered, fill)
			indices = Where(mask, gatheredIdx, minusOne)
			return
		}
		better := LogicalAnd(mask, GreaterThan(gathered, values))
		values = Where(better, gathered, values)
		indices = Where(better, gatheredIdx, indices)
	}
	visit(0, nil)
	// 全 padding 的窗口按 ONNX 参考实现保持 Y=0、Indices=-1。
	values = Where(Equal(indices, minusOne), zero, values)
	return values, indices
}

// poolOutputSpec 返回当前空间轴输出长度的 DimensionSpec。
func poolOutputSpec(g poolGeometry, i int) DimensionSpec {
	if g.dynamic {
		return DynamicDim(g.outNodes[i])
	}
	return StaticDim(g.outDims[i])
}

// poolAxisIota 返回当前空间轴的输出位置向量与左 padding 标量。
func poolAxisIota(x *Node, g poolGeometry, i int) (*Node, *Node) {
	if g.dynamic {
		return DynamicIota(x.Graph(), dtypes.Int64, 0, DynamicDim(g.outNodes[i])), g.startNodes[i]
	}
	out := g.outDims[i]
	iota := Iota(x.Graph(), shapes.Make(dtypes.Int64, out), 0)
	return iota, Scalar(x.Graph(), dtypes.Int64, int64(g.padBegin[i]))
}

// flatIndexTensor 构造与输入同形状的展平索引张量，支持 storage_order 0/1。
func flatIndexTensor(x *Node, _ poolGeometry, storageOrder int) *Node {
	rank := x.Rank()
	g := x.Graph()
	var acc *Node
	for axis := 0; axis < rank; axis++ {
		var iota *Node
		if x.Shape().IsDynamic() {
			iota = DynamicIota(g, dtypes.Int64, 0, DimensionSpecFor(x, axis))
			specs := make([]DimensionSpec, rank)
			for a := range specs {
				specs[a] = StaticDim(1)
			}
			specs[axis] = DimensionSpecFor(x, axis)
			iota = DynamicReshape(iota, specs...)
		} else {
			dim := x.Shape().Dim(axis)
			iota = Iota(g, shapes.Make(dtypes.Int64, dim), 0)
			dims := make([]int, rank)
			for a := range dims {
				dims[a] = 1
			}
			dims[axis] = dim
			iota = Reshape(iota, dims...)
		}
		iota = BroadcastLike(iota, x)
		coeff := indexCoefficient(x, axis, storageOrder)
		term := Mul(iota, coeff)
		if acc == nil {
			acc = term
		} else {
			acc = Add(acc, term)
		}
	}
	return acc
}

// indexCoefficient 返回某一轴的展平系数；storage_order=1 时空间轴使用列优先。
func indexCoefficient(x *Node, axis, storageOrder int) *Node {
	g := x.Graph()
	coeff := Scalar(g, dtypes.Int64, 1)
	if storageOrder == 0 || axis <= 1 {
		for j := axis + 1; j < x.Rank(); j++ {
			coeff = Mul(coeff, DimensionSize(x, j))
		}
		return coeff
	}
	for j := 2; j < axis; j++ {
		coeff = Mul(coeff, DimensionSize(x, j))
	}
	return coeff
}

// dynamicPool 在动态空间维上用 Gather 展开窗口实现 ONNX 池化。
// 每条 gather 的索引按 stride/dilation/pad 计算并做边界裁剪，padding 由有效掩码排除。
func dynamicPool(x *Node, g poolGeometry, average bool, countIncludePad, kernelSize int) *Node {
	zero := Scalar(x.Graph(), x.DType(), 0)
	one := Scalar(x.Graph(), x.DType(), 1)
	fill := zero
	if !average {
		fill = poolMinValue(x)
	}
	var acc, count *Node

	var visit func(dim int, offsets []int)
	visit = func(dim int, offsets []int) {
		if dim < len(g.kernel) {
			for k := range g.kernel[dim] {
				visit(dim+1, append(append([]int(nil), offsets...), k))
			}
			return
		}
		gathered := x
		var mask *Node
		for i, off := range offsets {
			axis := 2 + i
			in := DimensionSize(x, axis)
			out := g.outNodes[i]
			iota := DynamicIota(x.Graph(), dtypes.Int64, 0, DynamicDim(out))
			stride := Scalar(x.Graph(), dtypes.Int64, int64(g.strides[i]))
			shift := Scalar(x.Graph(), dtypes.Int64, int64(off*g.dilations[i]))
			raw := Sub(Add(Mul(iota, stride), shift), g.startNodes[i])
			idx := Min(Max(raw, Scalar(x.Graph(), dtypes.Int64, 0)), Sub(in, Scalar(x.Graph(), dtypes.Int64, 1)))
			gathered = onnxGather(gathered, idx, axis)
			valid := LogicalAnd(GreaterOrEqual(raw, Scalar(x.Graph(), dtypes.Int64, 0)), LessThan(raw, in))
			specs := make([]DimensionSpec, x.Rank())
			for a := range specs {
				specs[a] = StaticDim(1)
			}
			specs[axis] = DynamicDim(out)
			valid = DynamicReshape(valid, specs...)
			if mask == nil {
				mask = valid
			} else {
				mask = LogicalAnd(mask, valid)
			}
		}
		mask = BroadcastLike(mask, gathered)
		var val *Node
		if average {
			val = Where(mask, gathered, zero)
		} else {
			val = Where(mask, gathered, fill)
		}
		if acc == nil {
			acc = val
		} else if average {
			acc = Add(acc, val)
		} else {
			acc = Max(acc, val)
		}
		if average && countIncludePad == 0 {
			cnt := Where(mask, one, zero)
			if count == nil {
				count = cnt
			} else {
				count = Add(count, cnt)
			}
		}
	}
	visit(0, nil)
	if average {
		if countIncludePad == 1 {
			return DivScalar(acc, float64(kernelSize))
		}
		return Div(acc, count)
	}
	return acc
}

// padPoolInput 在池化前按 ONNX 语义补齐输入；无 padding 时直接返回原张量。
func padPoolInput(x *Node, g poolGeometry, fill *Node) *Node {
	if !g.dynamic && allZeroPads(g.padBegin) && allZeroPads(g.padEnd) {
		return x
	}
	specs := make([]DynamicPadAxisSpec, x.Rank())
	for i := range specs {
		specs[i] = DynamicPadAxis(0, 0, 0)
	}
	for i := range g.kernel {
		axis := 2 + i
		if g.dynamic {
			specs[axis] = DynamicPadAxisSpec{StartNode: g.startNodes[i], EndNode: g.endNodes[i]}
		} else {
			specs[axis] = DynamicPadAxis(g.padBegin[i], g.padEnd[i], 0)
		}
	}
	return DynamicPad(x, fill, specs...)
}

// reducePool 在已补齐的输入上做无 padding 的 reduce-window，支持 stride 与 dilation 同时出现。
func reducePool(x *Node, g poolGeometry, op compute.ReduceOpType) *Node {
	rank := x.Rank()
	window := make([]int, rank)
	strides := make([]int, rank)
	dilations := make([]int, rank)
	for i := range window {
		window[i], strides[i], dilations[i] = 1, 1, 1
	}
	for i := range g.kernel {
		axis := 2 + i
		window[axis] = g.kernel[i]
		strides[axis] = g.strides[i]
		dilations[axis] = g.dilations[i]
	}
	return BackendReduceWindow(x, op, window, strides, nil, dilations, nil)
}

// poolMinValue 返回 MaxPool 的 padding 填充值，保证 padding 不影响最大值。
func poolMinValue(x *Node) *Node {
	g, dt := x.Graph(), x.DType()
	switch dt {
	case dtypes.Float16, dtypes.BFloat16, dtypes.Float32, dtypes.Float64:
		return Scalar(g, dt, math.Inf(-1))
	case dtypes.Int8:
		return Scalar(g, dt, int8(math.MinInt8))
	case dtypes.Uint8:
		return Scalar(g, dt, uint8(0))
	default:
		exceptions.Panicf("MaxPool: unsupported dtype %s", dt)
		return nil
	}
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}

func ceilDivNode(a, b *Node) *Node {
	return Div(Add(a, Sub(b, Scalar(a.Graph(), b.DType(), 1))), b)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func allZeroPads(pads []int) bool {
	for _, p := range pads {
		if p != 0 {
			return false
		}
	}
	return true
}

// convertGlobalAveragePool converts an ONNX GlobalAveragePool node to a GoMLX node.
//
// See ONNX documentation in:
// https://onnx.ai/onnx/operators/onnx__GlobalAveragePool.html
func convertGlobalAveragePool(_ *Model, _ map[string]*Node, _ *protos.NodeProto, inputs []*Node) *Node {
	x := inputs[0]
	spatialDims := x.Rank() - 2
	if x.Shape().IsDynamic() {
		spatialAxes := make([]int, spatialDims)
		for i := range spatialAxes {
			spatialAxes[i] = i + 2
		}
		return ReduceAndKeep(x, ReduceMean, spatialAxes...)
	}
	window := make([]int, spatialDims)
	for i := range window {
		window[i] = x.Shape().Dim(i + 2)
	}
	pool := MeanPool(x).ChannelsAxis(timage.ChannelsFirst).Strides(1).WindowPerAxis(window...)
	return pool.Done()
}
