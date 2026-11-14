# ONNX 规范用例

这些小模型用于验证 `third_party/onnx-gomlx` 中 PRelu 与 Dropout 的补丁是否符合 ONNX 算子规范。

| 文件 | 语义 | 参考 |
| --- | --- | --- |
| `prelu_broadcast.onnx` | PRelu，slope 形状 `[3,1,1]` 单向广播到 `[1,3,2,2]` | ONNX PRelu-16，`Less + Where + Mul` 函数体 |
| `dropout_default.onnx` | Dropout 推理，`training_mode` 缺省 | 输出等于输入 |
| `dropout_mask.onnx` | Dropout 推理并请求 mask | 输出等于输入，mask 全为 true |
| `dropout_training_false.onnx` | `training_mode=false` 常量 | 输出等于输入，mask 全为 true |
| `dropout_training_true.onnx` | `training_mode=true` | 推理转换器应拒绝 |
| `initializer_default.onnx` | initializer `w` 同时作为 graph input | 未传用默认值 `[2,3]`，传入 `[10,20]` 时覆盖 |

生成方式：使用 Python `onnx` 包按官方示例构造，IR version 10，PRelu opset 16，Dropout opset 13。输入固定为 `[1,3,2,2]` float32。

参考文档：

- https://onnx.ai/onnx/operators/onnx__PRelu.html
- https://onnx.ai/onnx/operators/onnx__Dropout.html
