# OCR 黄金集与评测

本目录包含可复现的最小 OCR 黄金集（英文/中文合成证件文本图）与 MIDV-500 公开模拟证件评测报告。

## 文件

| 文件 | 内容 |
| --- | --- |
| `en_doc.png` / `en_doc.txt` / `en_fields.json` | 英文样例与期望字段 |
| `zh_doc.png` / `zh_doc.txt` / `zh_fields.json` | 中文样例与期望字段 |
| `report-en.json` / `report-zh.json` | 最近一次评测结果（含 CER/WER/字段/延迟） |
| `midv-bra.md` / `midv-bra.json` | MIDV-500 巴西护照原始拍摄帧评测（英文 mobile rec） |
| `midv-bra-rect.md` / `midv-bra-rect.json` | MIDV-500 巴西护照 GT 四边形透视校正后评测（英文 mobile rec） |
| `midv-bra-zh.md` / `midv-bra-zh.json` | MIDV-500 巴西护照校正后评测（中文 mobile rec，当前最佳） |

合成样例由 `cmd/ocrimg` 生成（Arial Unicode 字体），参考文本即渲染文本。

## 模型

| 用途 | 文件 | SHA-256 |
| --- | --- | --- |
| 检测 | `det.onnx`（PP-OCRv5 mobile det） | `d7fe3ea74652890722c0f4d02458b7261d9f5ae6c92904d05707c9eb155c7924` |
| 英文识别 | `en_PP-OCRv5_mobile_rec_infer.onnx` | `1081b104a3c44d103511f150763d997a846994431c5775a800c802254c1124bf` |
| 英文词典 | `ppocrv5_en_dict.txt` | `c60d46e9e01d500ed6388fe8681051eac9cf6692e0d57238315be171927a0a1b` |
| 中文识别 | `PP-OCRv5_mobile_rec_infer.onnx` | `d253c3cbee6e507828a5271a30ab0ec8ae7c2a99d0cc8e6f844fe380809d22b3` |
| 中文词典 | `ppocrv5_dict.txt` | `9dfc80c50b6cb07399a47a7cf25d11db475fb4ad0e1fc96b2eff6467c8166ff3` |

模型不入库，放在外部目录（示例：`/tmp/paddle-ocr-models`）。中文模型依赖 `onnx-gomlx` 的 Slice 属性形式补丁。

## 复现

```bash
# 1. 生成样例（已提交时不需要）
go run ./cmd/ocrimg -out eval/ocr/en_doc.png -size 40 \
  -text "NAME: JOHN DOE  DOB: 1990-01-01  ID NO: AB1234567  EXP: 2030-01-01  TOTAL: 44,900"
go run ./cmd/ocrimg -out eval/ocr/zh_doc.png -size 40 \
  -text "姓名：张三 出生日期：1990-01-01 证件号码：1234567890"

# 2. 英文评测
go run ./cmd/ocreval -model /tmp/paddle-ocr-models \
  -image eval/ocr/en_doc.png -ref eval/ocr/en_doc.txt -fields eval/ocr/en_fields.json \
  -width-modes exact,bucket -iters 3 -warmup 256,512,768 -json eval/ocr/report-en.json

# 3. 中文评测
go run ./cmd/ocreval -model /tmp/paddle-ocr-models \
  -rec /tmp/paddle-ocr-models/zh_rec.onnx -dict /tmp/paddle-ocr-models/zh_dict.txt \
  -image eval/ocr/zh_doc.png -ref eval/ocr/zh_doc.txt -fields eval/ocr/zh_fields.json \
  -width-modes exact,bucket -iters 3 -warmup 256,512,768 -json eval/ocr/report-zh.json
```

## 结果（2026-11-18，CPU）

| 样例 | 模式 | CER | CER（去空格） | WER | 字段准确率 | P95 |
| --- | --- | --- | --- | --- | --- | --- |
| 英文 | exact | 0.0000 | 0.0000 | 0.0000 | 1.000 | 463 ms |
| 英文 | bucket | 0.0000 | 0.0000 | 0.0000 | 1.000 | 459 ms |
| 中文 | exact | 0.0541 | 0.0000 | 1.0000 | 1.000 | 371 ms |
| 中文 | bucket | 0.0541 | 0.0000 | 1.0000 | 1.000 | 367 ms |

说明：

- 中文 CER 差异全部来自两个空格；忽略空白后字符完全正确。
- 中文 WER 未分词，不具参考意义；以 CER（去空格）与字段准确率为准。
- 单张合成图仅证明链路与字段抽取；真实证件需要扩充多版式、多国家样本后再谈商用精度。
- exact 与 bucket 在这个样例上无差异；默认使用 exact（与参考实现一致），bucket 作为减少图编译的可选模式。

## MIDV-500 巴西护照

### 数据来源与许可

- 数据集：**MIDV-500**，巴西护照子集 `06_bra_passport`，30 张 `TS` 视频帧。
- 论文：V. V. Arlazarov 等，*MIDV-500: A Dataset for Identity Documents Analysis and Recognition on Mobile Devices in Video Stream*，[DOI 10.18287/2412-6179-2019-43-5-818-824](https://doi.org/10.18287/2412-6179-2019-43-5-818-824)。
- 下载：`https://github.com/fcakyon/midv500` 提供的下载脚本（原始压缩包 `06_bra_passport.zip`）。
- 许可：论文说明数据集中使用的源证件图像为公有领域或公开版权许可；数据集本身用于研究与评测。镜像与原始图片不入库，仅保留评测报告。
- **性质**：MIDV-500 是公开**模拟证件**，不是真实签发证件；指标只能证明链路与相对改进，不能替代真实脱敏样本的商用精度门禁。

### 预处理

1. 解压 `06_bra_passport.zip`，取 `images/TS/TS06_01.png` … `TS06_30.png`；`ground_truth/TS/*.json` 提供文本区域四边形。
2. 评估侧用 Pillow 按 `06_bra_passport/ground_truth/06_bra_passport.json` 的字段四边形做透视变换，输出 2000×1400 的 `rect2/TS/*.png`。校正脚本与图片不入库；透视参数来自数据集 GT，不参与 OCR 推理。
3. 字段期望固定为 `JOANA SILVA COSTA` / `1981-01-10` / `AA000000` / `2025-08-14`，来自同一份 GT。

### 复现命令

```bash
# 英文 mobile rec（低延迟基线）
go run ./cmd/docxeval -model /tmp/paddle-ocr-models \
  -images "$MIDV/bra_ts/06_bra_passport/rect2/TS/*.png" \
  -gt "$MIDV/bra/06_bra_passport/ground_truth/06_bra_passport.json" \
  -max-side 2400 \
  -out-json eval/ocr/midv-bra-rect.json -out-md eval/ocr/midv-bra-rect.md

# 中文 mobile rec（字段优先，当前最佳）
go run ./cmd/docxeval -model /tmp/paddle-ocr-models \
  -rec /tmp/paddle-ocr-models/zh_rec.onnx -dict /tmp/paddle-ocr-models/zh_dict.txt \
  -images "$MIDV/bra_ts/06_bra_passport/rect2/TS/*.png" \
  -gt "$MIDV/bra/06_bra_passport/ground_truth/06_bra_passport.json" \
  -max-side 2000 -width-mode bucket -bucket 64 \
  -out-json eval/ocr/midv-bra-zh.json -out-md eval/ocr/midv-bra-zh.md
```

### 结果

| 模型 | 字段总准确率 | 姓名精确 | 四字段全对 | P50 / P95 |
| --- | --- | --- | --- | --- |
| 英文 mobile（原始帧） | 51.7% | 20.0% | 6 / 30 | 5638 / 7286 ms |
| 英文 mobile（校正） | 71.7% | 13.3% | 4 / 30 | 7735 / 9273 ms |
| **中文 mobile（校正）** | **75.8%** | **53.3%** | **16 / 30** | 7851 / 11206 ms |

中文 mobile 分项：出生日期 76.7%、证件号 93.3%、有效期 80.0%、姓名 token 53.3%。中文词典包含拉丁字母与葡萄牙语重音字符，姓名识别显著优于英文 mobile；但仍有 14/30 张存在至少一个字段错误，主要来自 MRZ 第一行与低质量帧。

`onnx-gomlx` 已补 `Conv auto_pad=SAME_UPPER/SAME_LOWER`（含回归测试），可加载 PP-OCRv5 server 识别模型做离线高精度评测。server rec 前 5 张探测字段 85%、姓名 80%，但 CPU 单图约 40 s，30 张全量因内存/吞吐未完成；仅作离线选项，不作为在线默认或精度门禁。

## 字段规则 v1

`internal/document/fields` 已实现：

- 护照 TD3 MRZ 优先解析（姓名/证件号/出生日期/到期日）。
- 美国/加拿大：驾照标签与 MM/DD/YYYY。
- 英国：16 位驾照号与 DD/MM/YYYY。
- 中国：18 位身份证号（含出生日期）与中文标签。
- 巴西：护照 TD3 MRZ 与葡萄牙语标签/月份（FEV/ABR/MAI/AGO/SET/OUT/DEZ）。
- 通用兜底：标签 + 日期/编号正则。

单测：`go test ./internal/document/fields`。

公开模拟样本已完成 30 张巴西护照评测；**仍缺真实脱敏证件 20–50 张**，因此真实场景的精度门禁尚未完成。

## 进程级并发基准

见 [`bench.md`](bench.md)：单机 1/2/4 进程 QPS 约 0.239/0.235/0.228，单进程已占满 CPU，单机多进程无扩容收益；扩容需多机/多容器副本或模型优化。
