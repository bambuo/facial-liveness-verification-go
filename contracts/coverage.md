# 契约覆盖报告

由 `go run ./cmd/contractcheck` 生成。

## 冻结契约

| 项 | 数值 |
| --- | --- |
| 接口注册 | 154 |
| 可达路径 | 158 |
| 数据表 / 列 | 31 / 329 |
| Zod schema | 16 |
| 契约 fixtures | 158（含嵌套响应 57） |
| 公开候选签字 | 17 |

## 实现覆盖

后端接口：**156 / 158** 可达路径（154 注册）

引擎端点：**6 / 6**

### 已实现后端接口

| 方法 | 路径 |
| --- | --- |
| DELETE | `/api/admin/user/:userId/data` |
| DELETE | `/api/developer/account` |
| DELETE | `/api/developer/api-key/:keyId` |
| DELETE | `/api/developer/reviewers/:id` |
| DELETE | `/api/developer/webhooks/:webhookId` |
| DELETE | `/api/v2/compliance/rules/:id` |
| DELETE | `/api/v2/compliance/rulesets/:id` |
| DELETE | `/api/v2/monitoring/schedules/:schedule_id` |
| DELETE | `/api/v2/vault/:token` |
| DELETE | `/api/webhooks/:webhookId` |
| GET | `/.well-known/did.json` |
| GET | `/api/admin/analytics` |
| GET | `/api/admin/analytics/fraud-patterns` |
| GET | `/api/admin/analytics/funnel` |
| GET | `/api/admin/analytics/rejections` |
| GET | `/api/admin/analytics/risk-distribution` |
| GET | `/api/admin/audit/export` |
| GET | `/api/admin/dashboard` |
| GET | `/api/admin/developers` |
| GET | `/api/admin/health` |
| GET | `/api/admin/provider-metrics` |
| GET | `/api/admin/thresholds` |
| GET | `/api/admin/verification/:id` |
| GET | `/api/admin/verification/:id/duplicates` |
| GET | `/api/admin/verifications` |
| GET | `/api/auth/csrf-token` |
| GET | `/api/auth/developer/github/url` |
| GET | `/api/developer/activity` |
| GET | `/api/developer/analytics` |
| GET | `/api/developer/api-keys` |
| GET | `/api/developer/profile` |
| GET | `/api/developer/reviewers` |
| GET | `/api/developer/settings/aml` |
| GET | `/api/developer/settings/branding` |
| GET | `/api/developer/settings/duplicate-detection` |
| GET | `/api/developer/settings/llm` |
| GET | `/api/developer/settings/page-builder` |
| GET | `/api/developer/settings/sms` |
| GET | `/api/developer/settings/vault` |
| GET | `/api/developer/settings/vc` |
| GET | `/api/developer/settings/voice-auth` |
| GET | `/api/developer/stats` |
| GET | `/api/developer/verifications/:verificationId` |
| GET | `/api/developer/webhooks` |
| GET | `/api/developer/webhooks/:webhookId/deliveries` |
| GET | `/api/developer/webhooks/:webhookId/secret` |
| GET | `/api/health` |
| GET | `/api/health/api-key-diagnostic` |
| GET | `/api/health/detailed` |
| GET | `/api/health/live` |
| GET | `/api/health/ready` |
| GET | `/api/setup/status` |
| GET | `/api/status` |
| GET | `/api/system/version` |
| GET | `/api/v2/batch` |
| GET | `/api/v2/batch/:id/results` |
| GET | `/api/v2/batch/:id/status` |
| GET | `/api/v2/compliance/rulesets` |
| GET | `/api/v2/compliance/rulesets/:id` |
| GET | `/api/v2/credentials/:jti/status` |
| GET | `/api/v2/monitoring/expiring-documents` |
| GET | `/api/v2/monitoring/schedules` |
| GET | `/api/v2/monitoring/schedules/:schedule_id` |
| GET | `/api/v2/vault` |
| GET | `/api/v2/vault/:token` |
| GET | `/api/v2/vault/:token/attributes/:attr` |
| GET | `/api/v2/vault/share/:shareToken` |
| GET | `/api/v2/verify/:id/credential` |
| GET | `/api/v2/verify/:verification_id/address-status` |
| GET | `/api/v2/verify/:verification_id/selfie` |
| GET | `/api/v2/verify/:verification_id/status` |
| GET | `/api/v2/verify/credentials/:jti/status` |
| GET | `/api/v2/verify/page-config` |
| GET | `/api/v2/verify/page-config/slug/:slug` |
| GET | `/api/v2/verify/session-info` |
| GET | `/api/vaas/developers` |
| GET | `/api/vaas/developers/:id` |
| GET | `/api/vaas/health` |
| GET | `/api/vaas/verify/:verification_id/status` |
| GET | `/api/verify/handoff/:token/session` |
| GET | `/api/verify/handoff/:token/status` |
| GET | `/api/webhooks` |
| GET | `/api/webhooks/:webhookId/deliveries` |
| PATCH | `/api/verify/handoff/:token/complete` |
| PATCH | `/api/verify/handoff/:token/link` |
| POST | `/api/admin/thresholds/preview` |
| POST | `/api/admin/thresholds/reset` |
| POST | `/api/auth/admin/escalate` |
| POST | `/api/auth/admin/login` |
| POST | `/api/auth/developer/github/callback` |
| POST | `/api/auth/developer/otp/complete-registration` |
| POST | `/api/auth/developer/otp/send` |
| POST | `/api/auth/developer/otp/verify` |
| POST | `/api/auth/logout` |
| POST | `/api/auth/reviewer/otp/send` |
| POST | `/api/auth/reviewer/otp/verify` |
| POST | `/api/auth/totp/setup` |
| POST | `/api/auth/totp/verify` |
| POST | `/api/developer/api-key` |
| POST | `/api/developer/api-key/:id/rotate` |
| POST | `/api/developer/avatar` |
| POST | `/api/developer/branding/logo` |
| POST | `/api/developer/register` |
| POST | `/api/developer/reviewers/invite` |
| POST | `/api/developer/webhooks` |
| POST | `/api/developer/webhooks/:webhookId/deliveries/:deliveryId/resend` |
| POST | `/api/developer/webhooks/:webhookId/test` |
| POST | `/api/setup/initialize` |
| POST | `/api/v2/batch/:id/cancel` |
| POST | `/api/v2/batch/upload` |
| POST | `/api/v2/compliance/evaluate` |
| POST | `/api/v2/compliance/rulesets` |
| POST | `/api/v2/compliance/rulesets/:id/rules` |
| POST | `/api/v2/credentials/:jti/revoke` |
| POST | `/api/v2/monitoring/schedules` |
| POST | `/api/v2/vault/:token/share` |
| POST | `/api/v2/vault/store` |
| POST | `/api/v2/verify/:id/credential/send` |
| POST | `/api/v2/verify/:verification_id/address-document` |
| POST | `/api/v2/verify/:verification_id/back-document` |
| POST | `/api/v2/verify/:verification_id/cross-validation` |
| POST | `/api/v2/verify/:verification_id/finalize` |
| POST | `/api/v2/verify/:verification_id/front-document` |
| POST | `/api/v2/verify/:verification_id/internal/session` |
| POST | `/api/v2/verify/:verification_id/live-capture` |
| POST | `/api/v2/verify/:verification_id/phone-otp/send` |
| POST | `/api/v2/verify/:verification_id/phone-otp/verify` |
| POST | `/api/v2/verify/:verification_id/restart` |
| POST | `/api/v2/verify/:verification_id/voice-capture` |
| POST | `/api/v2/verify/:verification_id/voice-challenge` |
| POST | `/api/v2/verify/credentials/:jti/revoke` |
| POST | `/api/v2/verify/initialize` |
| POST | `/api/v2/verify/re-verify` |
| POST | `/api/vaas/developers/:id/suspend` |
| POST | `/api/vaas/developers/:id/unsuspend` |
| POST | `/api/vaas/users` |
| POST | `/api/vaas/verify` |
| POST | `/api/verify/handoff/create` |
| POST | `/api/webhooks/:webhookId/test` |
| POST | `/api/webhooks/register` |
| PUT | `/api/admin/thresholds` |
| PUT | `/api/admin/verification/:id/review` |
| PUT | `/api/developer/profile` |
| PUT | `/api/developer/settings/aml` |
| PUT | `/api/developer/settings/branding` |
| PUT | `/api/developer/settings/duplicate-detection` |
| PUT | `/api/developer/settings/llm` |
| PUT | `/api/developer/settings/page-builder` |
| PUT | `/api/developer/settings/page-builder/slug` |
| PUT | `/api/developer/settings/sms` |
| PUT | `/api/developer/settings/vault` |
| PUT | `/api/developer/settings/vc` |
| PUT | `/api/developer/settings/voice-auth` |
| PUT | `/api/v2/compliance/rules/:id` |
| PUT | `/api/v2/compliance/rulesets/:id` |
| PUT | `/api/webhooks/:webhookId` |

### 已实现引擎端点

| 方法 | 路径 |
| --- | --- |
| POST | `/extract/back` |
| POST | `/extract/front` |
| POST | `/extract/live` |
| POST | `/extract/ocr` |
| POST | `/extract/voice-enroll` |
| POST | `/extract/voice-verify` |

## 待实现模块

| 模块 | 接口数 |
| --- | --- |
| addressVerification | 2 |
| admin | 15 |
| admin-thresholds | 4 |
| auth | 13 |
| batch | 5 |
| compliance | 9 |
| credentials | 8 |
| developer/analytics | 4 |
| developer/apiKeys | 4 |
| developer/profile | 8 |
| developer/register | 1 |
| developer/settings | 19 |
| developer/webhooks | 7 |
| handoff | 5 |
| health | 5 |
| monitoring | 5 |
| newVerification | 15 |
| pageConfig | 3 |
| setup | 2 |
| status | 1 |
| system | 1 |
| vaas | 8 |
| vault | 7 |
| webhooks | 6 |
| well-known | 1 |

## 校验结果

全部通过。
