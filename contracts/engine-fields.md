# 引擎 6 端点字段契约

直接从 idswyft `engine/src/routes/extract.ts` 抽取。所有端点：

- 方法：`POST`
- Content-Type：`multipart/form-data`
- 文件字段：`file`（单文件）
- 上限：10 MB（`multer` 配置）
- 成功：`{ success: true, result: ... }`
- 失败：`{ success: false, error: string, ... }`（HTTP 4xx/5xx）

## 1. 端点总表

| 端点 | 请求附加字段 | 结果结构 | Zod / 类型 |
| --- | --- | --- | --- |
| `/extract/front` | `document_type`、`issuing_country`、`document_id`、`verification_id`、`llm_config`（JSON 字符串或对象） | OCR + 人脸 + 篡改结果 | `FrontExtractionResultSchema` |
| `/extract/back` | 无 | 条码/QR + MRZ | `BackExtractionResultSchema` |
| `/extract/live` | `is_sandbox`（`'true'`/布尔）、`head_turn_metadata`（JSON） | 人脸检测 + 活体 + 深伪 | `LiveCaptureResultSchema` |
| `/extract/ocr` | `document_type` | OCR 文本与字段置信度 | `OCRData`（实现类型） |
| `/extract/voice-enroll` | 无（`mimetype` 决定解码） | 声纹向量 | `{ speaker_embedding: number[], embedding_dimension: number }` |
| `/extract/voice-verify` | 无 | 声纹向量 + 数字转写 | `{ speaker_embedding, embedding_dimension, transcription }` |

## 2. 结果字段（来自 Zod）

### FrontExtractionResult

`ocr`（`IDCardOCRSchema`：`full_name`、`date_of_birth`、`id_number`、`expiry_date`、`nationality?`、`issuing_country?`，passthrough）、`face_embedding`（可空）、`face_confidence`、`ocr_confidence`、`mrz_from_front`（可空）、`authenticity`（`score`、`flags`、`isAuthentic`、`ganScore?`、`zoneScore?`）、`face_age?`、`face_gender?`。

### BackExtractionResult

`qr_payload`（`QRPayloadSchema`，可空）、`mrz_result`（`MRZResultSchema`，可空）、`barcode_format`（`PDF417`/`QR_CODE`/`DATA_MATRIX`/`CODE_128`/`MRZ_TD1/TD2/TD3`，可空）、`raw_barcode_data`（可空字符串）。

### LiveCaptureResult

`face_embedding`（可空数组）、`face_confidence`、`liveness_passed`、`liveness_score`、`deepfake_check?`（`isReal`、`realProbability`、`fakeProbability`）、`face_age?`、`face_gender?`。

### OCRData

`raw_text`、`confidence_scores`（字段名→0–1），以及按证件类型抽取的字段（`name`、`date_of_birth`、`document_number`、`expiration_date` 等，随 provider 变化）。

## 3. 与 Go 侧的关系

| 现状 | 目标 |
| --- | --- |
| Go 无人脸以外的引擎端点 | 新增 6 端点：OCR/条码/MRZ/活体/深伪/声纹 |
| OCR 纯 Go spike 已通（`ocr-spike.md`） | 生产化后接入 `/extract/front`、`/extract/ocr` |
| 语音算子已通（fbank/CTC/CAM++） | 接入 `/extract/voice-enroll`、`/extract/voice-verify` |
| 人脸已通 | `/extract/live` 复用活体/人脸，补深伪 |

## 4. 证据

- 路由与请求字段：`engine/src/routes/extract.ts:171-700`
- 结果 schema：`shared/src/verification/models/schemas.ts`
- 活体元数据：`shared/src/verification/models/headTurnLivenessSchema.ts`
- 错误封装：`engine/src/routes/extract.ts` 各端点 `catch` 分支
