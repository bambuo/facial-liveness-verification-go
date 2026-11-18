# idswyft 字段契约（M0 契约冻结输入）

- 来源：`/Users/johana/Codes/github/idswyft-community-1.12.39` @ `83dd51bdfb6dec608856b41c110f54f5a8bf3efe`
- 生成时间：2026-11-18T09:21:57.940Z

## 汇总

| 指标 | 数值 |
| --- | --- |
| Zod schema | 16 |
| Zod 字段 | 86 |
| 有请求字段的接口 | 73 / 158 |
| 请求字段总数 | 158 |
| 状态词表 | 5 |

## 状态词表与冲突

| 名称 | 值数量 | 值 | 来源 |
| --- | --- | --- | --- |
| RejectionReason | 16 | FRONT_OCR_FAILED, FRONT_LOW_CONFIDENCE, BACK_BARCODE_NOT_FOUND, BACK_MRZ_CHECKSUM_FAILED, BACK_MRZ_BARCODE_MISMATCH, CROSS_VALIDATION_FAILED, DOCUMENT_EXPIRED, LIVENESS_FAILED, FACE_NOT_DETECTED, FACE_MATCH_FAILED, AML_MATCH_FOUND, AML_POTENTIAL_MATCH, DOCUMENT_TAMPERED, DEEPFAKE_DETECTED, VOICE_MATCH_FAILED, VOICE_CHALLENGE_FAILED | shared/src/verification/models/schemas.ts |
| VerificationStatus | 12 | AWAITING_FRONT, FRONT_PROCESSING, AWAITING_BACK, BACK_PROCESSING, CROSS_VALIDATING, AWAITING_LIVE, LIVE_PROCESSING, FACE_MATCHING, AWAITING_VOICE, VOICE_MATCHING, COMPLETE, HARD_REJECTED | shared/src/verification/models/schemas.ts |
| VerificationState | 12 | pending, document_uploaded, ocr_processing, ocr_completed, back_id_processing, cross_validation_completed, live_capture_processing, face_matching, liveness_checking, verified, failed, manual_review | backend/src/services/verificationStateMachine.ts |
| VerificationEvent | 12 | document_upload, ocr_success, ocr_failure, back_id_upload, cross_validation_success, cross_validation_failure, live_capture_upload, face_match_success, face_match_failure, liveness_success, liveness_failure, manual_review_required | backend/src/services/verificationStateMachine.ts |
| VerificationStatus | 5 | pending, processing, verified, failed, manual_review | backend/src/types/index.ts |

- **冲突**：VerificationStatus 有 2 套不同取值：shared/src/verification/models/schemas.ts(12)；backend/src/types/index.ts(5)
- **冲突**：对外 VerificationStatus 与内部 VerificationState 是两套状态；API 响应 status 使用前者，内部状态机使用后者

## Zod schema

### AnalysisFrameSchema

来源：`shared/src/verification/models/headTurnLivenessSchema.ts:5`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| frame_base64 | `z.string().max(200_000)` |  |  |
| timestamp | `z.number()` |  |  |
| phase | `z.enum([ 'turn1_start', 'turn1_peak', 'turn1_return', 'turn_start', 'turn_peak', 'turn_return', ])` |  |  |
| color_rgb | `z.tuple([z.number(), z.number(), z.number()]).optional()` | 是 |  |

### HeadTurnLivenessMetadataSchema

来源：`shared/src/verification/models/headTurnLivenessSchema.ts:23`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| challenge_type | `z.literal('head_turn')` |  |  |
| challenge_direction | `z.enum(['left', 'right'])` |  |  |
| frames | `z.array(AnalysisFrameSchema).min(5).max(12)` |  |  |
| color_sequence | `z.array(z.tuple([z.number(), z.number(), z.number()])).optional().default([])` | 是 |  |
| start_timestamp | `z.number()` |  |  |
| end_timestamp | `z.number()` |  |  |
| screen_width | `z.number().optional()` | 是 |  |
| screen_height | `z.number().optional()` | 是 |  |
| virtual_camera_check | `z.object({ label: z.string(), suspected_virtual: z.boolean(), }).optional()` | 是 |  |

### IDCardOCRSchema

来源：`shared/src/verification/models/schemas.ts:66`（passthrough）

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| full_name | `z.string().min(1)` |  |  |
| date_of_birth | `z.string().min(1)` |  |  |
| id_number | `z.string().min(1)` |  |  |
| expiry_date | `z.string().min(1)` |  |  |
| nationality | `z.string().optional()` | 是 |  |
| issuing_country | `z.string().length(2).optional()` | 是 |  |

### DocumentAuthenticitySchema

来源：`shared/src/verification/models/schemas.ts:76`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| score | `z.number().min(0).max(1)` |  |  |
| flags | `z.array(z.string())` |  |  |
| isAuthentic | `z.boolean()` |  |  |
| ganScore | `z.number().min(0).max(1).optional()` | 是 |  |
| zoneScore | `z.number().min(0).max(1).optional()` | 是 |  |

### FrontExtractionResultSchema

来源：`shared/src/verification/models/schemas.ts:85`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| ocr | `IDCardOCRSchema` |  |  |
| face_embedding | `z.array(z.number()).nullable()` |  | 是 |
| face_confidence | `confidence` |  |  |
| ocr_confidence | `confidence` |  |  |
| mrz_from_front | `z.array(z.string()).nullable()` |  | 是 |
| authenticity | `DocumentAuthenticitySchema` |  |  |
| face_age | `z.number().optional()` | 是 |  |
| face_gender | `z.string().optional()` | 是 |  |

### QRPayloadSchema

来源：`shared/src/verification/models/schemas.ts:99`（passthrough）

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| first_name | `z.string().optional()` | 是 |  |
| last_name | `z.string().optional()` | 是 |  |
| full_name | `z.string().optional()` | 是 |  |
| date_of_birth | `z.string().optional()` | 是 |  |
| id_number | `z.string().optional()` | 是 |  |
| expiry_date | `z.string().optional()` | 是 |  |
| nationality | `z.string().optional()` | 是 |  |

### MRZResultSchema

来源：`shared/src/verification/models/schemas.ts:110`（passthrough）

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| raw_lines | `z.array(z.string())` |  |  |
| fields | `z.record(z.string()).optional()` | 是 |  |
| checksums_valid | `z.boolean()` |  |  |

### BackExtractionResultSchema

来源：`shared/src/verification/models/schemas.ts:117`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| qr_payload | `QRPayloadSchema.nullable()` |  | 是 |
| mrz_result | `MRZResultSchema.nullable()` |  | 是 |
| barcode_format | `z.enum(['PDF417', 'QR_CODE', 'DATA_MATRIX', 'CODE_128', 'MRZ_TD1', 'MRZ_TD2', 'MRZ_TD3']).nullable()` |  | 是 |
| raw_barcode_data | `z.string().nullable()` |  | 是 |

### FieldScoreSchema

来源：`shared/src/verification/models/schemas.ts:127`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| score | `confidence` |  |  |
| passed | `z.boolean()` |  |  |
| weight | `confidence` |  |  |

### DlFormatValidationSchema

来源：`shared/src/verification/models/schemas.ts:134`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| valid | `z.boolean()` |  |  |
| verdict | `z.enum(['PASS', 'FAIL', 'REVIEW', 'SKIP'])` |  |  |
| matched_pattern | `z.string().nullable()` |  | 是 |
| issuing_state | `z.string().nullable()` |  | 是 |
| detail | `z.string()` |  |  |

### AddressValidationSchema

来源：`shared/src/verification/models/schemas.ts:143`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| score | `z.number().min(0).max(1)` |  |  |
| verdict | `z.enum(['PASS', 'REVIEW', 'FAIL'])` |  |  |
| front_address | `z.string()` |  |  |
| back_address | `z.string()` |  |  |

### CrossValidationResultSchema

来源：`shared/src/verification/models/schemas.ts:151`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| overall_score | `confidence` |  |  |
| field_scores | `z.record(FieldScoreSchema)` |  |  |
| has_critical_failure | `z.boolean()` |  |  |
| document_expired | `z.boolean()` |  |  |
| verdict | `z.enum(['PASS', 'REVIEW', 'REJECT'])` |  |  |
| dl_format_validation | `DlFormatValidationSchema.optional()` | 是 |  |
| address_validation | `AddressValidationSchema.optional()` | 是 |  |

### LiveCaptureResultSchema

来源：`shared/src/verification/models/schemas.ts:164`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| face_embedding | `z.array(z.number()).nullable()` |  | 是 |
| face_confidence | `confidence` |  |  |
| liveness_passed | `z.boolean()` |  |  |
| liveness_score | `confidence` |  |  |
| deepfake_check | `z.object({ isReal: z.boolean(), realProbability: z.number().min(0).max(1), fakeProbability: z.number().min(0).max(1), }).optional()` | 是 |  |
| face_age | `z.number().optional()` | 是 |  |
| face_gender | `z.string().optional()` | 是 |  |

### FaceMatchResultSchema

来源：`shared/src/verification/models/schemas.ts:181`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| similarity_score | `confidence` |  |  |
| passed | `z.boolean()` |  |  |
| threshold_used | `confidence` |  |  |
| skipped_reason | `z.string().optional()` | 是 |  |

### VoiceMatchResultSchema

来源：`shared/src/verification/models/schemas.ts:192`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| similarity_score | `confidence` |  |  |
| passed | `z.boolean()` |  |  |
| threshold_used | `confidence` |  |  |
| challenge_verified | `z.boolean()` |  |  |
| challenge_digits | `z.string()` |  |  |
| skipped_reason | `z.string().optional()` | 是 |  |

### GateResultSchema

来源：`shared/src/verification/models/schemas.ts:205`

| 字段 | 类型 | 可选 | 可空 |
| --- | --- | --- | --- |
| passed | `z.boolean()` |  |  |
| rejection_reason | `RejectionReasonEnum.nullable()` |  | 是 |
| rejection_detail | `z.string().nullable()` |  | 是 |
| user_message | `z.string().nullable()` |  | 是 |

## 接口请求字段

| 方法 | 路径 | body | query | param | files |
| --- | --- | --- | --- | --- | --- |
| POST | `/api/v2/verify/initialize` | user_id(body.isUUID.withMessage)<br>document_type(body.optional.isIn.withMessage)<br>issuing_country(body.optional.isLength.isAlpha.withMessage)<br>sandbox(body.optional.isBoolean.withMessage)<br>source(body.optional.isIn.withMessage)<br>addons(body.optional.isObject.withMessage)<br>addons.aml_screening(body.optional.isBoolean.withMessage)<br>addons.address_verification(body.optional.isBoolean.withMessage)<br>verification_mode(body.optional.isIn.withMessage)<br>age_threshold(body.optional.isInt.withMessage) | - | - | - |
| POST | `/api/v2/verify/re-verify` | user_id(body.isUUID.withMessage)<br>previous_verification_id(body.isUUID.withMessage)<br>source(body.optional.isIn.withMessage) | - | - | - |
| POST | `/api/v2/verify/:verification_id/front-document` | document_type(body.optional.isIn.withMessage)<br>issuing_country(body.optional.isLength.isAlpha.withMessage) | - | verification_id(param.isUUID.withMessage) | document(upload.single) |
| POST | `/api/v2/verify/:verification_id/back-document` | document_type(body.optional.isIn.withMessage)<br>issuing_country(body.optional.isLength.isAlpha.withMessage) | - | verification_id(param.isUUID.withMessage) | document(upload.single) |
| POST | `/api/v2/verify/:verification_id/cross-validation` | - | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/v2/verify/:verification_id/live-capture` | - | - | verification_id(param.isUUID.withMessage) | selfie(upload.single) |
| POST | `/api/v2/verify/:verification_id/voice-challenge` | - | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/v2/verify/:verification_id/voice-capture` | - | - | verification_id(param.isUUID.withMessage) | file(upload.single) |
| POST | `/api/v2/verify/:verification_id/finalize` | - | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/v2/verify/:verification_id/restart` | - | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/v2/verify/:verification_id/internal/session` | - | - | verification_id(param.isUUID.withMessage) | - |
| GET | `/api/v2/verify/:verification_id/status` | - | - | verification_id(param.isUUID.withMessage) | - |
| GET | `/api/v2/verify/:verification_id/selfie` | - | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/v2/verify/:verification_id/phone-otp/send` | phone_number(body.matches.withMessage) | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/v2/verify/:verification_id/phone-otp/verify` | code(body.matches.withMessage) | - | verification_id(param.isUUID.withMessage) | - |
| POST | `/api/verify/handoff/create` | - | - | - | - |
| GET | `/api/verify/handoff/:token/session` | - | - | - | - |
| PATCH | `/api/verify/handoff/:token/link` | - | - | - | - |
| PATCH | `/api/verify/handoff/:token/complete` | - | - | - | - |
| GET | `/api/verify/handoff/:token/status` | - | - | - | - |
| POST | `/api/developer/register` | email(body.isEmail.normalizeEmail.withMessage)<br>name(body.trim.escape.isLength.withMessage)<br>company(body.optional.trim.escape.isLength.withMessage)<br>webhook_url(body.optional.isURL.withMessage) | - | - | - |
| POST | `/api/developer/api-key` | name(body.trim.escape.isLength.withMessage)<br>is_sandbox(body.optional.isBoolean.withMessage)<br>expires_in_days(body.optional.isInt.withMessage) | - | - | - |
| GET | `/api/developer/api-keys` | - | - | - | - |
| DELETE | `/api/developer/api-key/:keyId` | - | - | keyId(param.isUUID.withMessage) | - |
| POST | `/api/developer/api-key/:id/rotate` | gracePeriodHours(body.optional.isInt.withMessage) | - | id(param.isUUID.withMessage) | - |
| GET | `/api/developer/webhooks` | - | - | - | - |
| POST | `/api/developer/webhooks` | url(body.isURL.withMessage)<br>is_sandbox(body.optional.isBoolean.withMessage)<br>events(body.optional.isArray.withMessage)<br>secret(body.optional.isString.withMessage)<br>api_key_id(body.optional.isUUID.withMessage) | - | - | - |
| GET | `/api/developer/webhooks/:webhookId/secret` | - | - | webhookId(param.isUUID.withMessage) | - |
| GET | `/api/developer/webhooks/:webhookId/deliveries` | - | - | webhookId(param.isUUID.withMessage) | - |
| POST | `/api/developer/webhooks/:webhookId/test` | - | - | webhookId(param.isUUID.withMessage) | - |
| POST | `/api/developer/webhooks/:webhookId/deliveries/:deliveryId/resend` | - | - | webhookId(param.isUUID.withMessage)<br>deliveryId(param.isUUID.withMessage) | - |
| DELETE | `/api/developer/webhooks/:webhookId` | - | - | webhookId(param.isUUID.withMessage) | - |
| GET | `/api/developer/stats` | - | - | - | - |
| GET | `/api/developer/activity` | - | - | - | - |
| GET | `/api/developer/verifications/:verificationId` | - | - | verificationId(param.isUUID.withMessage) | - |
| GET | `/api/developer/analytics` | - | - | - | - |
| GET | `/api/developer/settings/llm` | - | - | - | - |
| PUT | `/api/developer/settings/llm` | provider(body.isIn.withMessage)<br>api_key(body.optional.isString.isLength.withMessage)<br>endpoint_url(body.optional.isURL.withMessage)<br>model(body.optional.isString.isLength.withMessage) | - | - | - |
| GET | `/api/developer/settings/sms` | - | - | - | - |
| PUT | `/api/developer/settings/sms` | provider(body.isIn.withMessage)<br>api_key(body.optional.isString.isLength.withMessage)<br>api_secret(body.optional.isString.isLength.withMessage)<br>phone_number(body.optional.matches.withMessage) | - | - | - |
| GET | `/api/developer/settings/branding` | - | - | - | - |
| PUT | `/api/developer/settings/branding` | logo_url(body.optional.isURL.withMessage)<br>accent_color(body.optional.matches.withMessage)<br>company_name(body.optional.isString.trim.isLength.withMessage) | - | - | - |
| GET | `/api/developer/settings/page-builder` | - | - | - | - |
| PUT | `/api/developer/settings/page-builder` | - | - | - | - |
| PUT | `/api/developer/settings/page-builder/slug` | slug(body.optional.isString.matches.withMessage) | - | - | - |
| GET | `/api/developer/settings/aml` | - | - | - | - |
| PUT | `/api/developer/settings/aml` | enabled(body.isBoolean.withMessage) | - | - | - |
| GET | `/api/developer/settings/voice-auth` | - | - | - | - |
| PUT | `/api/developer/settings/voice-auth` | enabled(body.isBoolean.withMessage) | - | - | - |
| GET | `/api/developer/settings/vc` | - | - | - | - |
| PUT | `/api/developer/settings/vc` | enabled(body.isBoolean.withMessage) | - | - | - |
| GET | `/api/developer/settings/duplicate-detection` | - | - | - | - |
| PUT | `/api/developer/settings/duplicate-detection` | enabled(body.isBoolean.withMessage)<br>action(body.isIn.withMessage) | - | - | - |
| GET | `/api/developer/settings/vault` | - | - | - | - |
| PUT | `/api/developer/settings/vault` | enabled(body.optional.isBoolean.withMessage)<br>auto_store(body.optional.isBoolean.withMessage) | - | - | - |
| DELETE | `/api/developer/account` | confirm_email(body.isEmail.normalizeEmail.withMessage) | - | - | - |
| GET | `/api/developer/profile` | - | - | - | - |
| PUT | `/api/developer/profile` | name(body.trim.escape.isLength.withMessage)<br>company(body.optional.trim.escape.isLength.withMessage)<br>locale(body.optional.isIn.withMessage) | - | - | - |
| POST | `/api/developer/avatar` | - | - | - | - |
| POST | `/api/developer/branding/logo` | - | - | - | - |
| POST | `/api/developer/reviewers/invite` | email(body.isEmail.normalizeEmail.withMessage)<br>name(body.optional.trim.escape.isLength.withMessage)<br>role(body.optional.isIn.withMessage) | - | - | - |
| GET | `/api/developer/reviewers` | - | - | - | - |
| DELETE | `/api/developer/reviewers/:id` | - | - | id(param.isUUID.withMessage) | - |
| GET | `/api/admin/dashboard` | - | - | - | - |
| GET | `/api/admin/verifications` | - | status(query.optional.isIn.withMessage)<br>developer_id(query.optional.isUUID.withMessage)<br>page(query.optional.isInt.withMessage)<br>limit(query.optional.isInt.withMessage) | - | - |
| GET | `/api/admin/verification/:id` | - | - | id(param.isUUID.withMessage) | - |
| GET | `/api/admin/verification/:id/duplicates` | - | - | id(param.isUUID.withMessage) | - |
| PUT | `/api/admin/verification/:id/review` | decision(body.isIn.withMessage)<br>reason(body.optional.trim.isLength.withMessage)<br>new_status(body.optional.isIn.withMessage) | - | id(param.isUUID.withMessage) | - |
| GET | `/api/admin/developers` | - | page(query.optional.isInt.withMessage)<br>limit(query.optional.isInt.withMessage) | - | - |
| GET | `/api/admin/analytics` | - | period(query.optional.isIn.withMessage)<br>developer_id(query.optional.isUUID.withMessage) | - | - |
| GET | `/api/admin/health` | - | - | - | - |
| GET | `/api/admin/analytics/funnel` | - | period(query.optional.isIn.withMessage)<br>developer_id(query.optional.isUUID.withMessage) | - | - |
| GET | `/api/admin/analytics/rejections` | - | period(query.optional.isIn.withMessage) | - | - |
| GET | `/api/admin/analytics/fraud-patterns` | - | period(query.optional.isIn.withMessage) | - | - |
| GET | `/api/admin/analytics/risk-distribution` | - | period(query.optional.isIn.withMessage) | - | - |
| GET | `/api/admin/provider-metrics` | - | - | - | - |
| GET | `/api/admin/audit/export` | - | from(query.optional.isISO8601.withMessage)<br>to(query.optional.isISO8601.withMessage)<br>status(query.optional.isIn.withMessage)<br>developer_id(query.optional.isUUID.withMessage)<br>format(query.optional.isIn.withMessage)<br>limit(query.optional.isInt.withMessage)<br>offset(query.optional.isInt.withMessage)<br>include_sandbox(query.optional.isBoolean.withMessage) | - | - |
| DELETE | `/api/admin/user/:userId/data` | - | - | - | - |
| GET | `/api/admin/thresholds` | - | - | - | - |
| PUT | `/api/admin/thresholds` | - | - | - | - |
| POST | `/api/admin/thresholds/preview` | - | - | - | - |
| POST | `/api/admin/thresholds/reset` | - | - | - | - |
| POST | `/api/auth/logout` | - | - | - | - |
| GET | `/api/auth/csrf-token` | - | - | - | - |
| POST | `/api/auth/admin/login` | email(body.isEmail.withMessage)<br>password(body.notEmpty.withMessage.isLength.withMessage) | - | - | - |
| POST | `/api/auth/admin/escalate` | - | - | - | - |
| POST | `/api/auth/developer/otp/send` | - | - | - | - |
| POST | `/api/auth/developer/otp/verify` | code(body.isString.isLength.isNumeric.withMessage) | - | - | - |
| POST | `/api/auth/developer/otp/complete-registration` | registration_token(body.isString.notEmpty.withMessage)<br>name(body.isString.trim.isLength.withMessage)<br>company(body.optional.isString.trim) | - | - | - |
| POST | `/api/auth/reviewer/otp/send` | - | - | - | - |
| POST | `/api/auth/reviewer/otp/verify` | code(body.isString.isLength.isNumeric.withMessage) | - | - | - |
| POST | `/api/auth/developer/github/callback` | code(body.isString.matches.withMessage)<br>state(body.isString.notEmpty.withMessage) | - | - | - |
| GET | `/api/auth/developer/github/url` | - | - | - | - |
| POST | `/api/auth/totp/setup` | - | - | - | - |
| POST | `/api/auth/totp/verify` | - | - | - | - |
| GET | `/api/health` | - | - | - | - |
| GET | `/api/health/detailed` | - | - | - | - |
| GET | `/api/health/ready` | - | - | - | - |
| GET | `/api/health/live` | - | - | - | - |
| GET | `/api/health/api-key-diagnostic` | - | - | - | - |
| POST | `/api/webhooks/register` | url(body.isURL.withMessage)<br>is_sandbox(body.optional.isBoolean.withMessage)<br>secret_token(body.optional.isLength.withMessage) | - | - | - |
| GET | `/api/webhooks` | - | - | - | - |
| PUT | `/api/webhooks/:webhookId` | url(body.optional.isURL.withMessage)<br>is_active(body.optional.isBoolean.withMessage)<br>secret_token(body.optional.isLength.withMessage) | - | - | - |
| DELETE | `/api/webhooks/:webhookId` | - | - | - | - |
| GET | `/api/webhooks/:webhookId/deliveries` | - | - | - | - |
| POST | `/api/webhooks/:webhookId/test` | - | - | - | - |
| POST | `/api/vaas/users` | - | - | - | - |
| POST | `/api/vaas/verify` | - | - | - | - |
| GET | `/api/vaas/verify/:verification_id/status` | - | - | - | - |
| GET | `/api/vaas/developers` | - | - | - | - |
| GET | `/api/vaas/developers/:id` | - | - | id(param.isUUID) | - |
| POST | `/api/vaas/developers/:id/suspend` | - | - | id(param.isUUID) | - |
| POST | `/api/vaas/developers/:id/unsuspend` | - | - | id(param.isUUID) | - |
| GET | `/api/vaas/health` | - | - | - | - |
| POST | `/api/v2/batch/upload` | items(body.isArray.withMessage)<br>items.*.user_id(body.isUUID.withMessage)<br>items.*.document_type(body.optional.isIn.withMessage)<br>items.*.front_document_url(body.optional.isURL.withMessage)<br>items.*.back_document_url(body.optional.isURL.withMessage) | - | - | - |
| GET | `/api/v2/batch/:id/status` | - | - | id(param.isUUID.withMessage) | - |
| GET | `/api/v2/batch/:id/results` | - | - | id(param.isUUID.withMessage) | - |
| POST | `/api/v2/batch/:id/cancel` | - | - | id(param.isUUID.withMessage) | - |
| GET | `/api/v2/batch` | - | page(query.optional.isInt.withMessage)<br>limit(query.optional.isInt.withMessage) | - | - |
| POST | `/api/v2/verify/:verification_id/address-document` | document_type(body.isIn.withMessage) | - | verification_id(param.isUUID) | - |
| GET | `/api/v2/verify/:verification_id/address-status` | - | - | verification_id(param.isUUID) | - |
| GET | `/api/v2/verify/page-config` | - | - | - | - |
| GET | `/api/v2/verify/page-config/slug/:slug` | - | - | - | - |
| GET | `/api/v2/verify/session-info` | - | - | - | - |
| POST | `/api/v2/monitoring/schedules` | user_id(body.isUUID.withMessage)<br>interval_days(body.isInt.withMessage)<br>verification_request_id(body.optional.isUUID) | - | - | - |
| GET | `/api/v2/monitoring/schedules` | - | status(query.optional.isIn)<br>page(query.optional.isInt)<br>limit(query.optional.isInt) | - | - |
| GET | `/api/v2/monitoring/schedules/:schedule_id` | - | - | schedule_id(param.isUUID) | - |
| DELETE | `/api/v2/monitoring/schedules/:schedule_id` | - | - | schedule_id(param.isUUID) | - |
| GET | `/api/v2/monitoring/expiring-documents` | - | days_ahead(query.optional.isInt)<br>page(query.optional.isInt)<br>limit(query.optional.isInt) | - | - |
| GET | `/api/status` | - | - | - | - |
| GET | `/api/setup/status` | - | - | - | - |
| POST | `/api/setup/initialize` | name(body.isString.trim.escape.isLength.withMessage)<br>email(body.isEmail.normalizeEmail.withMessage)<br>company(body.optional.isString.trim.escape.isLength.withMessage)<br>locale(body.optional.isIn.withMessage) | - | - | - |
| GET | `/api/v2/verify/:id/credential` | - | - | - | - |
| POST | `/api/v2/verify/:id/credential/send` | - | - | - | - |
| POST | `/api/v2/verify/credentials/:jti/revoke` | - | - | - | - |
| GET | `/api/v2/verify/credentials/:jti/status` | - | - | - | - |
| GET | `/api/v2/:id/credential` | - | - | - | - |
| POST | `/api/v2/:id/credential/send` | - | - | - | - |
| POST | `/api/v2/credentials/:jti/revoke` | - | - | - | - |
| GET | `/api/v2/credentials/:jti/status` | - | - | - | - |
| POST | `/api/v2/compliance/rulesets` | - | - | - | - |
| GET | `/api/v2/compliance/rulesets` | - | - | - | - |
| GET | `/api/v2/compliance/rulesets/:id` | - | - | - | - |
| PUT | `/api/v2/compliance/rulesets/:id` | - | - | - | - |
| DELETE | `/api/v2/compliance/rulesets/:id` | - | - | - | - |
| POST | `/api/v2/compliance/rulesets/:id/rules` | - | - | - | - |
| PUT | `/api/v2/compliance/rules/:id` | - | - | - | - |
| DELETE | `/api/v2/compliance/rules/:id` | - | - | - | - |
| POST | `/api/v2/compliance/evaluate` | - | - | - | - |
| POST | `/api/v2/vault/store` | - | - | - | - |
| GET | `/api/v2/vault/share/:shareToken` | - | - | - | - |
| GET | `/api/v2/vault/:token` | - | - | - | - |
| GET | `/api/v2/vault/:token/attributes/:attr` | - | - | - | - |
| DELETE | `/api/v2/vault/:token` | - | - | - | - |
| GET | `/api/v2/vault` | - | - | - | - |
| POST | `/api/v2/vault/:token/share` | - | - | - | - |
| GET | `/.well-known/did.json` | - | - | - | - |
| GET | `/api/system/version` | - | - | - | - |

## 说明

- 请求字段来自 express-validator 链与 multer 字段；未覆盖 handler 内直接读 `req.body.x` 的分支。
- 响应字段尚未逐接口展开；M0 下一步从 shared/backend 类型与 `res.json` 结构生成。
- 状态词表存在多套口径，必须在 C8 决策后统一。
