# 人脸模型

本目录存放人脸核验平台使用的三个 ONNX 模型（`Face:ModelDir` 指向此处）。模型随仓库分发，便于克隆后直接运行。

## 模型清单

| 文件 | 用途 | 大小 | SHA-256 | 许可证 |
| --- | --- | --- | --- | --- |
| `face_detection_yunet_2023mar.onnx` | 人脸检测（YuNet）、5 点关键点 | 233 KB | `8f2383e4dd3cfbb4553ea8718107fc0423210dc964f9f4280604804ed2552fa4` | MIT |
| `face_recognition_sface_2021dec.onnx` | 128 维人脸特征（SFace） | 37 MB | `0ba9fbfa01b5270c96627c4ef784da859931e02f04419c829e83484087c34e79` | Apache-2.0 |
| `2.7_80x80_MiniFASNetV2.onnx` | 静默活体（MiniFASNetV2） | 1.7 MB | `0cbe5caec95c31de9d2ef845cb85407d76aecd1b6a2c0e343f7d35306bfbccb8` | Apache-2.0 |

## 来源

- YuNet、SFace：OpenCV Model Zoo（`https://github.com/opencv/opencv_zoo`）的
  `models/face_detection_yunet` 与 `models/face_recognition_sface`。
- MiniFASNetV2：Silent-Face-Anti-Spoofing（`https://github.com/minivision-ai/Silent-Face-Anti-Spoofing`）
  的 `2.7_80x80_MiniFASNetV2` 权重导出的 ONNX。

## 许可与合规

- 三个模型均可免费商用（MIT / Apache-2.0）。分发时须保留原始版权与许可声明，
  不得移除模型内的许可信息。
- 生产引入前应复核上游仓库的许可变更；许可证以各上游仓库当前声明为准。
- 模型文件不含人脸数据，不涉及生物特征合规。

## 完整性校验

启动时如需强校验，可按上表 SHA-256 比对，例如：

```bash
shasum -a 256 models/*.onnx
```

## 更换或升级模型

1. 替换文件并保证文件名与表一致（引擎按固定文件名加载）。
2. 同步更新本文件的 SHA-256、大小与许可证信息。
3. 模型变更会改变特征分布：`IFaceEngine.ModelVersion` 必须同步升版，
   旧档案会因版本不一致被拒绝比对（需重新录入或重建索引）。
4. 重跑引擎自检与集成测试，确认同人/异人区分度仍达标。
