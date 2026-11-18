# PaddleOCR 纯 Go 移植结论

- 日期：2026-11-18
- 环境：macOS arm64，Go 1.27，`CGO_ENABLED=0`，GoMLX CPU 后端
- 命令：`go run ./cmd/ocrspike -model <dir> -image <demo.jpg>`
- 结论：**GO**；检测/识别/行分组已按 `ppu-paddle-ocr` 逻辑移植，质量接近参考；仍缺黄金集评测与性能优化

## 1. 模型与校验

| 文件 | 大小 | SHA-256 |
| --- | --- | --- |
| `det.onnx`（PP-OCRv5 mobile det） | 4.5 MB | `d7fe3ea74652890722c0f4d02458b7261d9f5ae6c92904d05707c9eb155c7924` |
| `rec.onnx`（en PP-OCRv5 mobile rec） | 7.5 MB | `1081b104a3c44d103511f150763d997a846994431c5775a800c802254c1124bf` |
| `dict.txt`（ppocrv5 en） | 1.4 KB | `c60d46e9e01d500ed6388fe8681051eac9cf6692e0d57238315be171927a0a1b` |
| `demo.jpg` | 191 KB | `69f73db603ace3e67b61c0e30d3bb9c2de54f2a4816aedca3eebbf4e0da6b7af` |

来源：`https://media.githubusercontent.com/media/PT-Perkasa-Pilar-Utama/ppu-paddle-ocr-models/main`（det/rec，Git LFS）与 `https://raw.githubusercontent.com/...`（dict）。模型未入库，生产需放入 `models/ocr/` 并保留许可与校验和。

## 2. 移植内容（与参考实现逐条对齐）

### 2.1 检测

参考 `BaseDetectionService.postprocessDetection`：

1. 概率图转 8 位灰度：`round(p * 255)`；
2. `findContours(RETR_LIST, CHAIN_APPROX_SIMPLE)`，每个 contour 取 `boundingRect`；
3. 过滤 `width * height <= 25`；
4. 外扩：`vertical = round(h * 0.4)`、`horizontal = round(h * 0.6)`（都用高度）；
5. 映射回原图：坐标除以 resize ratio，宽高按原图裁剪；
6. 保留 `width > 5 && height > 5`。

Go 实现：`round(p*255) > 0` 等价于 `p >= 0.5/255`；前景 8 连通域 + 不触边的背景孔洞（近似 RETR_LIST 的外轮廓与内孔）；其余步骤逐条一致。

### 2.2 识别

参考 `BaseRecognitionService`：

1. 逐框裁剪；
2. 等比缩放到高 48，宽 `round(48 * aspect)`（最小 8）；
3. 取 **R 通道**，`(r/255 - 0.5)/0.5`，复制到 3 通道；
4. CTC 贪心：blank=0，跳过重复；字典最后一个索引输出空格；置信度为输出字符最大概率均值。

Go 实现：宽度按 32 分桶并在右侧补中性灰（减少 GoMLX 图编译次数）；其余一致。

### 2.3 行分组

参考 `sortResultsByReadingOrder` + `groupResultsByLine`：

- 排序：`|yA-yB| < (hA+hB)/4` 时按 x，否则按 y；
- 分组：`|y - currentY| < avgHeight/2` 归入当前行，行内按 x 排序；
- 输出：`text`（行内空格、行间换行）+ `lines` + 平均置信度。

Go 实现逐条一致。

## 3. 实测结果

```text
image=2748x1052 dict=438
det input=640x256 output=[1 1 256 640] elapsed=127.7ms
detect: boxes=51 ratio=0.2329 elapsed=153.0ms
recognize: boxes=51 kept=50 elapsed=3.35s
lines=21
  [01] conf=0.990 JA19-789-17015H94 Kasir : MARDIYAN
  [02] conf=0.861 T9.17-01-2024 15:31:44 V.2023.11.0
  [03] conf=0.965 44,900
  [04] conf=0.973 GADING N / 08129
  [05] conf=0.859 0 SMS/WA: 08111064088
  [06] conf=0.747 AFAMART ARTHA ADIN 012
  [07] conf=0.941 T. SUMBER A IJAYA, TBK Kritik&Saran:1500959 Kenbal ian
  [08] conf=0.537 TER  MP
  [09] conf=0.963 Bon Jf
  [10] conf=0.932 PPN 0)
  [11] conf=0.788 A1-7A JA19- 144,900 SUMBER ALFARIA TRIJAYATBK",
  [12] conf=0.791 LBORO LT 2 1 44,900 44,901 A OO L 0 44,900 4900 44,900 Total Iten
  [13] conf=0.759 Item 4,900 ten 1 4509 ALFAMART ARTHA GADING  / 081294665105
  [14] conf=0.845 00 Tunai D
  [15] conf=0.998 44,900
  [16] conf=0.781 .0-02 4 .011.0 T.1011. MARLBORO LGT 20 TOWER LT.12, ALAM SU
  [17] conf=0.724 KMP.RUKOARTHA GADING NIAGA KELAPAGAD eigh3
  [18] conf=0.644 TOWR L.1 STERATAG
  [19] conf=0.814 NPWP01.336.238.9-054.000
  [20] conf=0.886 SUMBER AFARIA TRIJAYA,TBK ex: "NPWP 01.336.238.9-054.000",
  [21] conf=0.999 Original Image Predict Draw Crop Result
```

## 4. 与参考实现对照

参考实现（`ppu-paddle-ocr@4.1.1` + `onnxruntime-node`，同图同模型）：

| 指标 | 参考实现 | Go 移植 |
| --- | --- | --- |
| 检测框 | 未打印 | 51 |
| 识别保留 | 24 行 | 50 框 / 21 行 |
| 平均置信度 | 0.729 | ≈0.848（仅非空行） |
| 耗时 | 352 ms（模型已初始化） | 3.35 s（含按宽度编译图） |
| 文本 token（≥4 字符） | 62 | 59 |
| token 重合 | — | 30 |

重合的关键 token：`JA19-789-17015H94`、`KASIR`、`MARDIYAN`、`ARTHA`、`V.2023.11.0`、`44,900`、`ALFAMART`、`GADING`、`SMS/WA`、`KRITIK`、`SARAN`、`1500959`、`144,900`、`MARLBORO`、`4,900`、`TUNAI`、`NIAGA`、`SUMBER`、`TRIJAYA,TBK`、`NPWP`、`01.336.238.9-054.000`。

结论：模型推理正确；移植后行结构与关键字段已接近参考，剩余差异主要来自宽度分桶、孔洞近似、英文词典与图片本身的水印噪声。

## 5. M1 已完成项

| 项 | 结果 |
| --- | --- |
| 精确宽度 | 默认 `exact`；与 32 分桶对比：英文 CER/WER 均为 0，中文去空格 CER 均为 0 |
| 多语言 | 中文 PP-OCRv5 + 18384 词典跑通；需 `onnx-gomlx` 的 Slice 属性形式补丁（已加回归测试） |
| 自动方向 | 90° 旋转图选中 270，180° 旋转图选中 180，文本正常识别 |
| 预热 | 识别宽度预热 280–310 ms，首次请求不再承担全部编译 |
| 并发 | 4 个引擎并发 0 错误；QPS ≈ 0.345，说明 gobackend 推理全局串行，需进程级并行才能扩容 |
| 黄金集 | 英文 CER 0 / WER 0 / 字段 1.0 / P95 463 ms；中文去空格 CER 0 / 字段 1.0 / P95 371 ms |
| API | `/extract/ocr`、`/extract/front` 已接入，真实模型集成测试通过 |

黄金集与复现命令见 [`eval/ocr/README.md`](../eval/ocr/README.md)。

## 6. 仍待完善

| 项 | 现状 | 生产要求 |
| --- | --- | --- |
| 真实证件 | 合成样例 + 演示图 | 多版式、多国家、多拍摄条件黄金集 |
| 字段规则 | 通用正则/标签抽取 | 按国家证件移植 idswyft 字段规则库 |
| 条码/MRZ | 未实现 | PDF417/QR/MRZ 与交叉校验 |
| 篡改/深伪 | 未实现 | Sharp 算法移植、deepfake ONNX |
| 性能 | 单图 3–4 s，4 并发 QPS 0.345 | 预编译/多进程/后端优化，目标 P95 与 QPS 门禁 |
| 引擎端点 | 2/6 已接入 | front/back/live/ocr/voice-enroll/voice-verify 全量 |

## 7. 下一步（M2）

1. 扩充真实证件黄金集，输出字段级准确率、CER/WER 与 P95 门禁。
2. 移植 idswyft 的国家证件字段规则与交叉校验。
3. 接入条码/MRZ、篡改/深伪与其余 4 个引擎端点。
4. 优化并发：进程级并行或后端改造，建立 QPS/内存基线。
