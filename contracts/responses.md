# idswyft 响应契约（M0 契约冻结输入）

- 来源：`/Users/johana/Codes/github/idswyft-community-1.12.39` @ `83dd51bdfb6dec608856b41c110f54f5a8bf3efe`
- 生成时间：2026-11-18T11:12:04.209Z
- 口径：按源码位置就近归属到路由；同一接口聚合状态码与顶层字段。

## 汇总

| 指标 | 数值 |
| --- | --- |
| 接口 | 158 |
| 有响应证据的接口 | 158 |
| res.json 调用 | 317 |
| 状态码种类 | 12 |

## 响应构建器

`buildVerificationResponse`（`backend/src/verification/statusReader.ts:169`）

- 顶层字段：success, verification_id, verification_mode, is_sandbox, status, current_step, total_steps, front_document_uploaded, back_document_uploaded, live_capture_uploaded, ocr_data, document_authenticity, barcode_data, cross_validation_results, face_match_results, liveness_results, deepfake_check, aml_screening, age_estimation, velocity_analysis, geo_analysis, voice_match_results, risk_score, compliance_flags, duplicate_flags, barcode_extraction_failed, documents_match, face_match_passed, liveness_passed, final_result, rejection_reason, rejection_detail, failure_reason, user_message, manual_review_reason_code, manual_review_reason, created_at, updated_at
- 展开：((state as any).age_verification && { age_verification: (state as any).age_verification })；(mapped.final_result === 'failed' && { retry_available: (verification.retry_count ?? 0) < 3, retry_count: verification.retry_count ?? 0, })

## 接口响应

### POST /api/v2/verify/initialize

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | success, verification_id, session_token, verification_url, verification_mode, status, current_step, total_steps, message | - | `backend/src/routes/newVerification.ts:993` |

### POST /api/v2/verify/re-verify

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | success, verification_id, parent_verification_id, verification_mode, status, current_step, total_steps, message | - | `backend/src/routes/newVerification.ts:1128` |

### POST /api/v2/verify/:verification_id/front-document

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | idempotent_retry | - | `backend/src/routes/newVerification.ts:1223` |
| 409 | success, verification_id, status, current_step, final_result, rejection_reason, rejection_detail, message | - | `backend/src/routes/newVerification.ts:1303` |
| 200 | success, verification_id, verification_mode, status, current_step, total_steps, document_id, ocr_data, detected_document_type, classification_confidence, rejection_reason, rejection_detail, requires_back, message | - | `backend/src/routes/newVerification.ts:1421` |

### POST /api/v2/verify/:verification_id/back-document

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, verification_id, verification_mode, message | - | `backend/src/routes/newVerification.ts:1525` |
| 409 | success, verification_id, status, current_step, final_result, rejection_reason, rejection_detail, message | - | `backend/src/routes/newVerification.ts:1565` |
| 200 | success, verification_id, verification_mode, status, current_step, total_steps, document_id, barcode_data, barcode_extraction_failed, documents_match, cross_validation_results, rejection_reason, rejection_detail, failure_reason, message | - | `backend/src/routes/newVerification.ts:1619` |

### POST /api/v2/verify/:verification_id/cross-validation

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, verification_id, status, current_step, documents_match, cross_validation_results, rejection_reason, rejection_detail, failure_reason, manual_review_reason, message | - | `backend/src/routes/newVerification.ts:1680` |

### POST /api/v2/verify/:verification_id/live-capture

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, verification_id, verification_mode, message | - | `backend/src/routes/newVerification.ts:1735` |
| 409 | success, verification_id, status, current_step, final_result, rejection_reason, rejection_detail, message | - | `backend/src/routes/newVerification.ts:1800` |
| 200 | success, verification_id, verification_mode, status, current_step, total_steps, selfie_id, face_match_results, liveness_results, deepfake_check, age_estimation, velocity_analysis, geo_analysis, final_result, rejection_reason, rejection_detail, failure_reason, manual_review_reason, message | liveness_results{liveness_passed,liveness_score,liveness_mode} | `backend/src/routes/newVerification.ts:1957` |

### POST /api/v2/verify/:verification_id/voice-challenge

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 409 | success, error, current_step | - | `backend/src/routes/newVerification.ts:2030` |
| 200 | success, verification_id, challenge_digits, expires_in_seconds, message | - | `backend/src/routes/newVerification.ts:2052` |

### POST /api/v2/verify/:verification_id/voice-capture

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 409 | success, error, current_step | - | `backend/src/routes/newVerification.ts:2081` |
| 400 | success, error | - | `backend/src/routes/newVerification.ts:2096` |
| 410 | success, error | - | `backend/src/routes/newVerification.ts:2104` |
| 503 | success, error, retryable, current_step | - | `backend/src/routes/newVerification.ts:2122` |
| 503 | success, error | - | `backend/src/routes/newVerification.ts:2131` |
| 200 | success, verification_id, status, current_step, total_steps, voice_match_results, final_result, message | - | `backend/src/routes/newVerification.ts:2248` |

### POST /api/v2/verify/:verification_id/finalize

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, verification_id, status, current_step, final_result, rejection_reason, rejection_detail, message | - | `backend/src/routes/newVerification.ts:2294` |

### POST /api/v2/verify/:verification_id/restart

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, message | - | `backend/src/routes/newVerification.ts:2326` |
| 400 | success, message, retry_count | - | `backend/src/routes/newVerification.ts:2335` |
| 503 | success, message | - | `backend/src/routes/newVerification.ts:2348` |
| 409 | success, message | - | `backend/src/routes/newVerification.ts:2374` |
| 200 | success, verification_id, retry_count, message | - | `backend/src/routes/newVerification.ts:2406` |

### POST /api/v2/verify/:verification_id/internal/session

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | success, message | - | `backend/src/routes/newVerification.ts:2472` |
| 200 | success, verification_id, session_token, session_token_expires_at, verification_url | - | `backend/src/routes/newVerification.ts:2497` |

### GET /api/v2/verify/:verification_id/status

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → response | - | `backend/src/routes/newVerification.ts:2564` |

### GET /api/v2/verify/:verification_id/selfie

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/newVerification.ts:2588` |
| 404 | error | - | `backend/src/routes/newVerification.ts:2604` |
| 404 | selfie_url | - | `backend/src/routes/newVerification.ts:2613` |
| 200 | verification_id, selfie_url, expires_in | - | `backend/src/routes/newVerification.ts:2618` |

### POST /api/v2/verify/:verification_id/phone-otp/send

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 429 | success, message | - | `backend/src/routes/newVerification.ts:2675` |
| 200 | → response | - | `backend/src/routes/newVerification.ts:2691` |

### POST /api/v2/verify/:verification_id/phone-otp/verify

来源：`backend/src/routes/newVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, message | - | `backend/src/routes/newVerification.ts:2711` |
| 200 | success, message, phone_verified | - | `backend/src/routes/newVerification.ts:2717` |

### POST /api/verify/handoff/create

来源：`backend/src/routes/handoff.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | token, expires_at | - | `backend/src/routes/handoff.ts:112` |

### GET /api/verify/handoff/:token/session

来源：`backend/src/routes/handoff.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/handoff.ts:120` |
| 404 | error | - | `backend/src/routes/handoff.ts:133` |
| 410 | error | - | `backend/src/routes/handoff.ts:144` |
| 409 | error | - | `backend/src/routes/handoff.ts:148` |
| 200 | user_id, source, branding, page_builder_config | - | `backend/src/routes/handoff.ts:179` |

### PATCH /api/verify/handoff/:token/link

来源：`backend/src/routes/handoff.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/handoff.ts:194` |
| 404 | error | - | `backend/src/routes/handoff.ts:220` |
| 200 | success | - | `backend/src/routes/handoff.ts:223` |

### PATCH /api/verify/handoff/:token/complete

来源：`backend/src/routes/handoff.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/handoff.ts:231` |
| 404 | error | - | `backend/src/routes/handoff.ts:260` |
| 410 | error | - | `backend/src/routes/handoff.ts:264` |
| 409 | error | - | `backend/src/routes/handoff.ts:281` |
| 200 | success | - | `backend/src/routes/handoff.ts:284` |

### GET /api/verify/handoff/:token/status

来源：`backend/src/routes/handoff.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/handoff.ts:292` |
| 404 | error | - | `backend/src/routes/handoff.ts:304` |
| 410 | error | - | `backend/src/routes/handoff.ts:315` |
| 200 | status | - | `backend/src/routes/handoff.ts:318` |

### POST /api/developer/register

来源：`backend/src/routes/developer/register.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | developer, message | developer{id,email,name,company,is_verified,created_at} | `backend/src/routes/developer/register.ts:98` |

### POST /api/developer/api-key

来源：`backend/src/routes/developer/apiKeys.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | api_key, key_id, name, is_sandbox, created_at, expires_at, key_prefix, security_info, message | security_info{store_securely,environment_variable,revocation} | `backend/src/routes/developer/apiKeys.ts:154` |

### GET /api/developer/api-keys

来源：`backend/src/routes/developer/apiKeys.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | api_keys, total_keys, active_keys, security_recommendations | security_recommendations{rotate_unused_keys,monitor_usage,use_environment_variables} | `backend/src/routes/developer/apiKeys.ts:197` |

### DELETE /api/developer/api-key/:keyId

来源：`backend/src/routes/developer/apiKeys.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message, revoked_key, security_info | revoked_key{id,name,key_preview}<br>security_info{immediate_effect,cleanup_recommendation,regeneration} | `backend/src/routes/developer/apiKeys.ts:283` |

### POST /api/developer/api-key/:id/rotate

来源：`backend/src/routes/developer/apiKeys.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | new_key, new_key_id, old_key_expires_at, grace_period_hours, message | - | `backend/src/routes/developer/apiKeys.ts:462` |

### GET /api/developer/webhooks

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | webhooks | - | `backend/src/routes/developer/webhooks.ts:78` |

### POST /api/developer/webhooks

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | webhook, message | webhook{secret_key} | `backend/src/routes/developer/webhooks.ts:210` |

### GET /api/developer/webhooks/:webhookId/secret

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | secret_key | - | `backend/src/routes/developer/webhooks.ts:258` |

### GET /api/developer/webhooks/:webhookId/deliveries

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | deliveries, total | - | `backend/src/routes/developer/webhooks.ts:295` |

### POST /api/developer/webhooks/:webhookId/test

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, status_code, error | - | `backend/src/routes/developer/webhooks.ts:373` |
| 200 | success, status_code | - | `backend/src/routes/developer/webhooks.ts:397` |
| 200 | success, status_code, error | - | `backend/src/routes/developer/webhooks.ts:408` |

### POST /api/developer/webhooks/:webhookId/deliveries/:deliveryId/resend

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, delivery | delivery{id,status,created_at} | `backend/src/routes/developer/webhooks.ts:475` |

### DELETE /api/developer/webhooks/:webhookId

来源：`backend/src/routes/developer/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/developer/webhooks.ts:535` |

### GET /api/developer/stats

来源：`backend/src/routes/developer/analytics.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | period, period_start, total_requests, successful_requests, failed_requests, pending_requests, manual_review_requests, success_rate, monthly_limit, monthly_usage, remaining_quota, quota_reset_date | - | `backend/src/routes/developer/analytics.ts:70` |

### GET /api/developer/activity

来源：`backend/src/routes/developer/analytics.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | statistics, recent_activities, total_activities, session_outcomes | - | `backend/src/routes/developer/analytics.ts:186` |

### GET /api/developer/verifications/:verificationId

来源：`backend/src/routes/developer/analytics.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, verification_id, status, current_step, total_steps, message | - | `backend/src/routes/developer/analytics.ts:232` |
| 200 | → response | - | `backend/src/routes/developer/analytics.ts:252` |

### GET /api/developer/analytics

来源：`backend/src/routes/developer/analytics.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | daily_volume, rejection_breakdown, daily_latency, quota, funnel, daily_webhooks | quota{used,limit} | `backend/src/routes/developer/analytics.ts:291` |

### GET /api/developer/settings/llm

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | configured, provider, api_key_preview, endpoint_url, model | - | `backend/src/routes/developer/settings.ts:29` |
| 200 | configured, provider, api_key_preview, endpoint_url, model | - | `backend/src/routes/developer/settings.ts:42` |

### PUT /api/developer/settings/llm

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message | - | `backend/src/routes/developer/settings.ts:72` |
| 200 | success, message, provider, api_key_preview, endpoint_url, model | - | `backend/src/routes/developer/settings.ts:101` |

### GET /api/developer/settings/sms

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | configured, provider, api_key_preview, phone_number | - | `backend/src/routes/developer/settings.ts:128` |
| 200 | configured, provider, api_key_preview, phone_number | - | `backend/src/routes/developer/settings.ts:141` |

### PUT /api/developer/settings/sms

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message | - | `backend/src/routes/developer/settings.ts:170` |
| 200 | success, message, provider, api_key_preview, phone_number | - | `backend/src/routes/developer/settings.ts:190` |

### GET /api/developer/settings/branding

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | configured, logo_url, accent_color, company_name | - | `backend/src/routes/developer/settings.ts:215` |

### PUT /api/developer/settings/branding

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | error | - | `backend/src/routes/developer/settings.ts:248` |
| 200 | success, logo_url, accent_color, company_name | - | `backend/src/routes/developer/settings.ts:251` |

### GET /api/developer/settings/page-builder

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | configured, config, slug, logo_url | - | `backend/src/routes/developer/settings.ts:307` |

### PUT /api/developer/settings/page-builder

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/developer/settings.ts:323` |
| 500 | error | - | `backend/src/routes/developer/settings.ts:330` |
| 200 | success | - | `backend/src/routes/developer/settings.ts:331` |

### PUT /api/developer/settings/page-builder/slug

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 409 | error | - | `backend/src/routes/developer/settings.ts:355` |
| 500 | error | - | `backend/src/routes/developer/settings.ts:363` |
| 200 | success, slug | - | `backend/src/routes/developer/settings.ts:364` |

### GET /api/developer/settings/aml

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | enabled | - | `backend/src/routes/developer/settings.ts:381` |

### PUT /api/developer/settings/aml

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | error | - | `backend/src/routes/developer/settings.ts:403` |
| 200 | success, enabled | - | `backend/src/routes/developer/settings.ts:406` |

### GET /api/developer/settings/voice-auth

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | enabled | - | `backend/src/routes/developer/settings.ts:423` |

### PUT /api/developer/settings/voice-auth

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | error | - | `backend/src/routes/developer/settings.ts:445` |
| 200 | success, enabled | - | `backend/src/routes/developer/settings.ts:448` |

### GET /api/developer/settings/vc

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | enabled | - | `backend/src/routes/developer/settings.ts:465` |

### PUT /api/developer/settings/vc

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | error | - | `backend/src/routes/developer/settings.ts:487` |
| 200 | success, enabled | - | `backend/src/routes/developer/settings.ts:490` |

### GET /api/developer/settings/duplicate-detection

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | enabled, action | - | `backend/src/routes/developer/settings.ts:509` |

### PUT /api/developer/settings/duplicate-detection

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | error | - | `backend/src/routes/developer/settings.ts:536` |
| 200 | success, enabled, action | - | `backend/src/routes/developer/settings.ts:539` |

### GET /api/developer/settings/vault

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | enabled, auto_store | - | `backend/src/routes/developer/settings.ts:556` |

### PUT /api/developer/settings/vault

来源：`backend/src/routes/developer/settings.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/developer/settings.ts:578` |
| 500 | error | - | `backend/src/routes/developer/settings.ts:587` |
| 200 | success | - | `backend/src/routes/developer/settings.ts:590` |

### DELETE /api/developer/account

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/developer/profile.ts:72` |

### GET /api/developer/profile

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, scope, operator | operator{email,api_key_id,key_prefix,service_label,service_product,service_environment} | `backend/src/routes/developer/profile.ts:96` |
| 200 | success, data | data{id,email,name,company,avatar_url,locale,created_at} | `backend/src/routes/developer/profile.ts:112` |

### PUT /api/developer/profile

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, data | data{avatar_url} | `backend/src/routes/developer/profile.ts:166` |

### POST /api/developer/avatar

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, data | data{avatar_url} | `backend/src/routes/developer/profile.ts:209` |

### POST /api/developer/branding/logo

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, data | data{logo_url} | `backend/src/routes/developer/profile.ts:248` |

### POST /api/developer/reviewers/invite

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | reviewer | - | `backend/src/routes/developer/profile.ts:324` |

### GET /api/developer/reviewers

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | reviewers | - | `backend/src/routes/developer/profile.ts:343` |

### DELETE /api/developer/reviewers/:id

来源：`backend/src/routes/developer/profile.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message, reviewer | - | `backend/src/routes/developer/profile.ts:368` |

### GET /api/admin/dashboard

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | stats, recent_verifications, developer_count, system_health | - | `backend/src/routes/admin.ts:51` |

### GET /api/admin/verifications

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | verifications, pagination | pagination{page,limit,total,pages} | `backend/src/routes/admin.ts:134` |

### GET /api/admin/verification/:id

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | verification | verification{document_url,selfie_url,documents,debug} | `backend/src/routes/admin.ts:295` |

### GET /api/admin/verification/:id/duplicates

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | verification_id, linked_verifications | - | `backend/src/routes/admin.ts:334` |
| 200 | verification_id, linked_verifications | - | `backend/src/routes/admin.ts:356` |

### PUT /api/admin/verification/:id/review

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | verification, message | - | `backend/src/routes/admin.ts:464` |

### GET /api/admin/developers

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | developers, pagination | pagination{page,limit,total,pages} | `backend/src/routes/admin.ts:563` |

### GET /api/admin/analytics

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → analytics | - | `backend/src/routes/admin.ts:597` |

### GET /api/admin/health

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → health | - | `backend/src/routes/admin.ts:607` |

### GET /api/admin/analytics/funnel

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | funnel | - | `backend/src/routes/admin.ts:767` |

### GET /api/admin/analytics/rejections

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | rejections | - | `backend/src/routes/admin.ts:782` |

### GET /api/admin/analytics/fraud-patterns

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | patterns | - | `backend/src/routes/admin.ts:797` |

### GET /api/admin/analytics/risk-distribution

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | distribution | - | `backend/src/routes/admin.ts:812` |

### GET /api/admin/provider-metrics

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/admin.ts:823` |
| 200 | → summary | - | `backend/src/routes/admin.ts:827` |

### GET /api/admin/audit/export

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | export_date, record_count, filters, records | filters{from,to,status,developer_id,include_sandbox} | `backend/src/routes/admin.ts:1034` |

### DELETE /api/admin/user/:userId/data

来源：`backend/src/routes/admin.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message | - | `backend/src/routes/admin.ts:1072` |

### GET /api/admin/thresholds

来源：`backend/src/routes/admin-thresholds.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:24` |
| 200 | success, data | data{production,sandbox,meta} | `backend/src/routes/admin-thresholds.ts:36` |
| 500 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:71` |

### PUT /api/admin/thresholds

来源：`backend/src/routes/admin-thresholds.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:88` |
| 400 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:106` |
| 200 | success, data | data{updated_settings,production,sandbox,meta} | `backend/src/routes/admin-thresholds.ts:151` |
| 500 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:180` |

### POST /api/admin/thresholds/preview

来源：`backend/src/routes/admin-thresholds.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, data | data{preview,explanation} | `backend/src/routes/admin-thresholds.ts:213` |
| 500 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:238` |

### POST /api/admin/thresholds/reset

来源：`backend/src/routes/admin-thresholds.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:255` |
| 200 | success, data | data{message,default_settings} | `backend/src/routes/admin-thresholds.ts:300` |
| 500 | success, error | error{code,message} | `backend/src/routes/admin-thresholds.ts:309` |

### POST /api/auth/logout

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/auth.ts:85` |

### GET /api/auth/csrf-token

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | csrfToken | - | `backend/src/routes/auth.ts:91` |

### POST /api/auth/admin/login

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 401 | requires_totp | - | `backend/src/routes/auth.ts:145` |
| 200 | token, user | user{id,email,role} | `backend/src/routes/auth.ts:163` |

### POST /api/auth/admin/escalate

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 410 | message | - | `backend/src/routes/auth.ts:176` |

### POST /api/auth/developer/otp/send

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/auth.ts:203` |

### POST /api/auth/developer/otp/verify

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | token, developer, is_new | developer{id,email,name,company,is_verified,created_at} | `backend/src/routes/auth.ts:244` |
| 200 | registration_token, is_new | - | `backend/src/routes/auth.ts:262` |

### POST /api/auth/developer/otp/complete-registration

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | token, developer, api_key, is_new | developer{id,email,name,company,is_verified,created_at} | `backend/src/routes/auth.ts:340` |

### POST /api/auth/reviewer/otp/send

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/auth.ts:383` |
| 200 | message | - | `backend/src/routes/auth.ts:390` |

### POST /api/auth/reviewer/otp/verify

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | token, reviewer | reviewer{id,email,name,developer_id,role} | `backend/src/routes/auth.ts:446` |

### POST /api/auth/developer/github/callback

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | token, developer, api_key, is_new | developer{id,email,name,company,is_verified,avatar_url,created_at} | `backend/src/routes/auth.ts:617` |
| 200 | token, developer, is_new | developer{id,email,name,company,is_verified,avatar_url,created_at} | `backend/src/routes/auth.ts:643` |

### GET /api/auth/developer/github/url

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | url, configured | - | `backend/src/routes/auth.ts:663` |
| 200 | url, state, configured | - | `backend/src/routes/auth.ts:670` |

### POST /api/auth/totp/setup

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | qrCode | - | `backend/src/routes/auth.ts:688` |

### POST /api/auth/totp/verify

来源：`backend/src/routes/auth.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/auth.ts:717` |

### GET /api/health

来源：`backend/src/routes/health.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → healthcheck | - | `backend/src/routes/health.ts:20` |

### GET /api/health/detailed

来源：`backend/src/routes/health.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| dynamic | → healthcheck | - | `backend/src/routes/health.ts:85` |

### GET /api/health/ready

来源：`backend/src/routes/health.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | status | - | `backend/src/routes/health.ts:97` |
| 503 | status, error | - | `backend/src/routes/health.ts:99` |

### GET /api/health/live

来源：`backend/src/routes/health.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | status | - | `backend/src/routes/health.ts:105` |

### GET /api/health/api-key-diagnostic

来源：`backend/src/routes/health.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/health.ts:113` |
| dynamic | → diagnostics | - | `backend/src/routes/health.ts:164` |
| 500 | timestamp, error, details, environment | environment{NODE_ENV,API_KEY_SECRET} | `backend/src/routes/health.ts:167` |

### POST /api/webhooks/register

来源：`backend/src/routes/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | webhook, message | webhook{id,url,is_sandbox,is_active,created_at} | `backend/src/routes/webhooks.ts:72` |

### GET /api/webhooks

来源：`backend/src/routes/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | webhooks | - | `backend/src/routes/webhooks.ts:92` |

### PUT /api/webhooks/:webhookId

来源：`backend/src/routes/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | webhook, message | webhook{id,url,is_sandbox,is_active,created_at} | `backend/src/routes/webhooks.ts:153` |

### DELETE /api/webhooks/:webhookId

来源：`backend/src/routes/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message | - | `backend/src/routes/webhooks.ts:187` |

### GET /api/webhooks/:webhookId/deliveries

来源：`backend/src/routes/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | webhook_id, page, limit, total, deliveries | - | `backend/src/routes/webhooks.ts:211` |

### POST /api/webhooks/:webhookId/test

来源：`backend/src/routes/webhooks.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | message, result, payload | result{delivered,response_status} | `backend/src/routes/webhooks.ts:269` |

### POST /api/vaas/users

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | success, error | - | `backend/src/routes/vaas.ts:37` |
| 201 | success, data | data{id,email,phone,first_name,last_name,external_id,metadata} | `backend/src/routes/vaas.ts:48` |

### POST /api/vaas/verify

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | success, error | - | `backend/src/routes/vaas.ts:93` |
| 200 | success, verification_id, status, message | - | `backend/src/routes/vaas.ts:105` |

### GET /api/vaas/verify/:verification_id/status

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | success, error | - | `backend/src/routes/vaas.ts:124` |
| 200 | success, verification | verification{id,status,confidence_score,failure_reason,updated_at} | `backend/src/routes/vaas.ts:130` |

### GET /api/vaas/developers

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 500 | success, error | - | `backend/src/routes/vaas.ts:172` |
| 200 | success, data, meta | meta{total,page,per_page} | `backend/src/routes/vaas.ts:201` |

### GET /api/vaas/developers/:id

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | success, error | - | `backend/src/routes/vaas.ts:223` |
| 200 | success, data | data{avatar_url,api_key_count,verification_count} | `backend/src/routes/vaas.ts:232` |

### POST /api/vaas/developers/:id/suspend

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | success, error | - | `backend/src/routes/vaas.ts:256` |
| 200 | success, data | - | `backend/src/routes/vaas.ts:260` |

### POST /api/vaas/developers/:id/unsuspend

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | success, error | - | `backend/src/routes/vaas.ts:276` |
| 200 | success, data | - | `backend/src/routes/vaas.ts:280` |

### GET /api/vaas/health

来源：`backend/src/routes/vaas.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message, timestamp | - | `backend/src/routes/vaas.ts:285` |

### POST /api/v2/batch/upload

来源：`backend/src/routes/batch.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | success, batch_id, status, total_items, message | - | `backend/src/routes/batch.ts:246` |

### GET /api/v2/batch/:id/status

来源：`backend/src/routes/batch.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | batch_id, status, total_items, processed_items, succeeded_items, failed_items, progress_percentage, created_at, completed_at | - | `backend/src/routes/batch.ts:274` |

### GET /api/v2/batch/:id/results

来源：`backend/src/routes/batch.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | results | - | `backend/src/routes/batch.ts:310` |

### POST /api/v2/batch/:id/cancel

来源：`backend/src/routes/batch.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error, message | - | `backend/src/routes/batch.ts:331` |
| 200 | success, message | - | `backend/src/routes/batch.ts:337` |

### GET /api/v2/batch

来源：`backend/src/routes/batch.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | jobs, pagination | pagination{page,limit,total,pages} | `backend/src/routes/batch.ts:362` |

### POST /api/v2/verify/:verification_id/address-document

来源：`backend/src/routes/addressVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error, message | - | `backend/src/routes/addressVerification.ts:77` |
| 200 | success, verification_id, address_verification | address_verification{status,score,name_match_score,address,document_fresh,reasons} | `backend/src/routes/addressVerification.ts:146` |

### GET /api/v2/verify/:verification_id/address-status

来源：`backend/src/routes/addressVerification.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error, message | - | `backend/src/routes/addressVerification.ts:181` |
| 200 | verification_id, address_verification, message | - | `backend/src/routes/addressVerification.ts:188` |
| 200 | verification_id, address_verification | address_verification{status,score,name_match_score,address,document_type,document_fresh,reasons} | `backend/src/routes/addressVerification.ts:197` |

### GET /api/v2/verify/page-config

来源：`backend/src/routes/pageConfig.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/pageConfig.ts:21` |
| 200 | branding | branding{logo_url,accent_color,company_name} | `backend/src/routes/pageConfig.ts:40` |
| 200 | branding | branding{logo_url,accent_color,company_name} | `backend/src/routes/pageConfig.ts:54` |
| 200 | branding, page_builder_config | branding{logo_url,accent_color,company_name} | `backend/src/routes/pageConfig.ts:63` |

### GET /api/v2/verify/page-config/slug/:slug

来源：`backend/src/routes/pageConfig.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | branding, page_builder_config | branding{logo_url,accent_color,company_name} | `backend/src/routes/pageConfig.ts:90` |
| 200 | branding, page_builder_config | branding{logo_url,accent_color,company_name} | `backend/src/routes/pageConfig.ts:97` |

### GET /api/v2/verify/session-info

来源：`backend/src/routes/pageConfig.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/pageConfig.ts:117` |
| 400 | error | - | `backend/src/routes/pageConfig.ts:122` |
| 404 | error | - | `backend/src/routes/pageConfig.ts:135` |
| 410 | error | - | `backend/src/routes/pageConfig.ts:140` |
| 200 | verification_id, user_id, status, verification_mode, branding, page_builder_config | branding{logo_url,accent_color,company_name} | `backend/src/routes/pageConfig.ts:153` |

### POST /api/v2/monitoring/schedules

来源：`backend/src/routes/monitoring.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | success, schedule | schedule{id,user_id,interval_days,next_verification_at,status,created_at} | `backend/src/routes/monitoring.ts:60` |

### GET /api/v2/monitoring/schedules

来源：`backend/src/routes/monitoring.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | schedules, total, page, limit | - | `backend/src/routes/monitoring.ts:95` |

### GET /api/v2/monitoring/schedules/:schedule_id

来源：`backend/src/routes/monitoring.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | schedule | schedule{id,user_id,interval_days,next_verification_at,last_verification_at,status,created_at} | `backend/src/routes/monitoring.ts:129` |

### DELETE /api/v2/monitoring/schedules/:schedule_id

来源：`backend/src/routes/monitoring.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message | - | `backend/src/routes/monitoring.ts:167` |

### GET /api/v2/monitoring/expiring-documents

来源：`backend/src/routes/monitoring.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | alerts, total, page, limit | - | `backend/src/routes/monitoring.ts:197` |

### GET /api/status

来源：`backend/src/routes/status.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | overall, services, checked_at | - | `backend/src/routes/status.ts:62` |

### GET /api/setup/status

来源：`backend/src/routes/setup.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | needs_setup | - | `backend/src/routes/setup.ts:34` |
| 200 | needs_setup | - | `backend/src/routes/setup.ts:50` |

### POST /api/setup/initialize

来源：`backend/src/routes/setup.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 201 | token, developer, api_key | developer{id,email,name,company,is_verified,locale,created_at} | `backend/src/routes/setup.ts:133` |

### GET /api/v2/verify/:id/credential

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 503 | error | - | `backend/src/routes/credentials.ts:25` |
| 403 | error | - | `backend/src/routes/credentials.ts:36` |
| 404 | error | - | `backend/src/routes/credentials.ts:48` |
| 200 | success, credential, jti, expires_at | - | `backend/src/routes/credentials.ts:53` |
| 200 | success, already_issued, jti, issued_at, expires_at, message | - | `backend/src/routes/credentials.ts:71` |
| 400 | error | - | `backend/src/routes/credentials.ts:81` |

### POST /api/v2/verify/:id/credential/send

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/credentials.ts:100` |
| 503 | error | - | `backend/src/routes/credentials.ts:104` |
| 403 | error | - | `backend/src/routes/credentials.ts:115` |
| 404 | error | - | `backend/src/routes/credentials.ts:127` |
| 200 | success, jti, expires_at, email_sent | - | `backend/src/routes/credentials.ts:175` |

### POST /api/v2/verify/credentials/:jti/revoke

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message | - | `backend/src/routes/credentials.ts:197` |
| 400 | error | - | `backend/src/routes/credentials.ts:199` |

### GET /api/v2/verify/credentials/:jti/status

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → status | - | `backend/src/routes/credentials.ts:213` |

### GET /api/v2/:id/credential

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 503 | error | - | `backend/src/routes/credentials.ts:25` |
| 403 | error | - | `backend/src/routes/credentials.ts:36` |
| 404 | error | - | `backend/src/routes/credentials.ts:48` |
| 200 | success, credential, jti, expires_at | - | `backend/src/routes/credentials.ts:53` |
| 200 | success, already_issued, jti, issued_at, expires_at, message | - | `backend/src/routes/credentials.ts:71` |
| 400 | error | - | `backend/src/routes/credentials.ts:81` |

### POST /api/v2/:id/credential/send

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/credentials.ts:100` |
| 503 | error | - | `backend/src/routes/credentials.ts:104` |
| 403 | error | - | `backend/src/routes/credentials.ts:115` |
| 404 | error | - | `backend/src/routes/credentials.ts:127` |
| 200 | success, jti, expires_at, email_sent | - | `backend/src/routes/credentials.ts:175` |

### POST /api/v2/credentials/:jti/revoke

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, message | - | `backend/src/routes/credentials.ts:197` |
| 400 | error | - | `backend/src/routes/credentials.ts:199` |

### GET /api/v2/credentials/:jti/status

来源：`backend/src/routes/credentials.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → status | - | `backend/src/routes/credentials.ts:213` |

### POST /api/v2/compliance/rulesets

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:81` |
| 201 | success, ruleset | - | `backend/src/routes/compliance.ts:98` |

### GET /api/v2/compliance/rulesets

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, rulesets | - | `backend/src/routes/compliance.ts:128` |

### GET /api/v2/compliance/rulesets/:id

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:145` |
| 404 | error | - | `backend/src/routes/compliance.ts:154` |
| 200 | success, ruleset | ruleset{rules} | `backend/src/routes/compliance.ts:162` |

### PUT /api/v2/compliance/rulesets/:id

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:174` |
| 400 | error | - | `backend/src/routes/compliance.ts:179` |
| 400 | error | - | `backend/src/routes/compliance.ts:186` |
| 400 | error | - | `backend/src/routes/compliance.ts:192` |
| 404 | error | - | `backend/src/routes/compliance.ts:205` |
| 200 | success, ruleset | - | `backend/src/routes/compliance.ts:207` |

### DELETE /api/v2/compliance/rulesets/:id

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:218` |
| 404 | error | - | `backend/src/routes/compliance.ts:228` |
| 200 | success, message | - | `backend/src/routes/compliance.ts:230` |

### POST /api/v2/compliance/rulesets/:id/rules

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:244` |
| 404 | error | - | `backend/src/routes/compliance.ts:254` |
| 400 | error | - | `backend/src/routes/compliance.ts:257` |
| 400 | error | - | `backend/src/routes/compliance.ts:260` |
| 201 | success, rule | - | `backend/src/routes/compliance.ts:275` |

### PUT /api/v2/compliance/rules/:id

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:287` |
| 404 | error | - | `backend/src/routes/compliance.ts:296` |
| 404 | error | - | `backend/src/routes/compliance.ts:305` |
| 400 | error | - | `backend/src/routes/compliance.ts:311` |
| 400 | error | - | `backend/src/routes/compliance.ts:316` |
| 400 | error | - | `backend/src/routes/compliance.ts:322` |
| 200 | success, rule | - | `backend/src/routes/compliance.ts:334` |

### DELETE /api/v2/compliance/rules/:id

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:345` |
| 404 | error | - | `backend/src/routes/compliance.ts:354` |
| 404 | error | - | `backend/src/routes/compliance.ts:363` |
| 200 | success, message | - | `backend/src/routes/compliance.ts:372` |

### POST /api/v2/compliance/evaluate

来源：`backend/src/routes/compliance.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/compliance.ts:386` |
| 200 | success, matched_rules, matches, resolved_action | - | `backend/src/routes/compliance.ts:402` |

### POST /api/v2/vault/store

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/vault.ts:47` |
| 404 | error | - | `backend/src/routes/vault.ts:58` |
| 400 | error | - | `backend/src/routes/vault.ts:60` |
| 200 | success, already_stored, vault_token, message | - | `backend/src/routes/vault.ts:73` |
| 404 | error | - | `backend/src/routes/vault.ts:83` |
| 400 | error | - | `backend/src/routes/vault.ts:87` |
| 201 | success, vault_token | - | `backend/src/routes/vault.ts:92` |

### GET /api/v2/vault/share/:shareToken

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/vault.ts:114` |
| 410 | error | - | `backend/src/routes/vault.ts:117` |
| 410 | error | - | `backend/src/routes/vault.ts:121` |
| 410 | error | - | `backend/src/routes/vault.ts:134` |
| 404 | error | - | `backend/src/routes/vault.ts:145` |
| 404 | error | - | `backend/src/routes/vault.ts:150` |
| 200 | success, attributes, expires_at | - | `backend/src/routes/vault.ts:162` |

### GET /api/v2/vault/:token

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/vault.ts:182` |
| 200 | success, vault_token, status, data, created_at, expires_at | - | `backend/src/routes/vault.ts:187` |

### GET /api/v2/vault/:token/attributes/:attr

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/vault.ts:210` |
| 404 | error | - | `backend/src/routes/vault.ts:217` |
| 200 | success, attribute, value | - | `backend/src/routes/vault.ts:220` |

### DELETE /api/v2/vault/:token

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/vault.ts:240` |
| 404 | error | - | `backend/src/routes/vault.ts:248` |
| 200 | success, message | - | `backend/src/routes/vault.ts:253` |

### GET /api/v2/vault

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | success, entries, pagination | pagination{page,limit,total,total_pages} | `backend/src/routes/vault.ts:284` |

### POST /api/v2/vault/:token/share

来源：`backend/src/routes/vault.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 400 | error | - | `backend/src/routes/vault.ts:310` |
| 400 | error | - | `backend/src/routes/vault.ts:314` |
| 400 | error | - | `backend/src/routes/vault.ts:319` |
| 404 | error | - | `backend/src/routes/vault.ts:330` |
| 400 | error | - | `backend/src/routes/vault.ts:331` |
| 201 | success, share_url, share_token, expires_at, allowed_attributes | - | `backend/src/routes/vault.ts:351` |

### GET /.well-known/did.json

来源：`backend/src/routes/well-known.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 404 | error | - | `backend/src/routes/well-known.ts:14` |
| 200 | @context, id, verificationMethod, authentication, assertionMethod | - | `backend/src/routes/well-known.ts:23` |

### GET /api/system/version

来源：`backend/src/routes/system.ts`

| 状态码 | 顶层字段 / 引用 | 嵌套字段 | 源码位置 |
| --- | --- | --- | --- |
| 200 | → response | - | `backend/src/routes/system.ts:97` |

## 未归属的 res.json

- `backend/src/routes/vault.ts:28` status=403 keys=error
- `backend/src/routes/system.ts:36` status=200 keys=-

## 说明

- 按源码位置就近归属；helper 函数内产生的 res.json 可能归属到相邻路由，需人工复核
- 对象字面量只取顶层字段；嵌套结构与类型约束需结合 TypeScript 类型继续展开
- buildVerificationResponse 的展开表达式（条件对象）未逐字段展开
- C8 已决定 status 与 final_result 统一 12 值；本文件反映 idswyft 当前实现，落地时按 C8 改造
