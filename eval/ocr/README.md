# OCR 黄金集与评测

本目录是可复现的最小 OCR 黄金集：英文/中文各一张合成证件文本图，附带参考文本与字段期望。

## 文件

| 文件 | 内容 |
| --- | --- |
| `en_doc.png` / `en_doc.txt` / `en_fields.json` | 英文样例与期望字段 |
| `zh_doc.png` / `zh_doc.txt` / `zh_fields.json` | 中文样例与期望字段 |
| `report-en.json` / `report-zh.json` | 最近一次评测结果（含 CER/WER/字段/延迟） |

图片由 `cmd/ocrimg` 生成（Arial Unicode 字体），参考文本即渲染文本。

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
