# idswyft 接口清单（M0 契约冻结输入）

- 来源仓库：`/Users/johana/Codes/github/idswyft-community-1.12.39`
- 基线提交：`83dd51bdfb6dec608856b41c110f54f5a8bf3efe`（2026-11-17T00:25:31+07:00）
- 生成时间：2026-11-18T09:21:57.723Z
- 解析器：TypeScript 5.9.3 AST
- 统计口径：`router.<get|post|put|patch|delete>` 注册；同一模块多前缀挂载展开为多条路径；cloud-only 动态挂载不计入

## 汇总

| 指标 | 数值 |
| --- | --- |
| 路由注册数（文档口径 154） | 154 |
| 展开后 method+path | 158 |
| app 级路由 | 7 |
| cloud-only 挂载 | 3 |
| 待人工确认鉴权 | 32 |

### 按模块

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

### 按鉴权

| 鉴权 | 接口数 |
| --- | --- |
| admin-jwt|reviewer-jwt + org-admin|platform-admin | 6 |
| admin-jwt|reviewer-jwt | 2 |
| api-key | 28 |
| api-key|handoff-token | 16 |
| compliance-request | 9 |
| dashboard-jwt | 28 |
| developer-jwt | 15 |
| jwt + admin-jwt | 7 |
| jwt | 2 |
| none-or-inline | 32 |
| review-principal | 4 |
| service-token | 9 |

## 接口明细

> 明细共 158 行：154 条注册 + `credentials` 双挂载（`/api/v2/verify` 与 `/api/v2`）展开的 4 行。

| # | 模块 | 方法 | 路径 | 挂载 | 鉴权 | 中间件 | 源码位置 | 待复核 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | newVerification | POST | `/api/v2/verify/initialize` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, checkSandboxMode, idempotencyMiddleware, verificationRateLimit, validate | `backend/src/routes/newVerification.ts:818` |  |
| 2 | newVerification | POST | `/api/v2/verify/re-verify` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, checkSandboxMode, verificationRateLimit, validate | `backend/src/routes/newVerification.ts:1025` |  |
| 3 | newVerification | POST | `/api/v2/verify/:verification_id/front-document` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, idempotencyMiddleware, verificationRateLimit, upload, validate | `backend/src/routes/newVerification.ts:1150` |  |
| 4 | newVerification | POST | `/api/v2/verify/:verification_id/back-document` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, idempotencyMiddleware, verificationRateLimit, upload, validate | `backend/src/routes/newVerification.ts:1475` |  |
| 5 | newVerification | POST | `/api/v2/verify/:verification_id/cross-validation` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, verificationRateLimit, validate | `backend/src/routes/newVerification.ts:1663` |  |
| 6 | newVerification | POST | `/api/v2/verify/:verification_id/live-capture` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, idempotencyMiddleware, verificationRateLimit, upload, validate | `backend/src/routes/newVerification.ts:1698` |  |
| 7 | newVerification | POST | `/api/v2/verify/:verification_id/voice-challenge` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, validate | `backend/src/routes/newVerification.ts:2017` |  |
| 8 | newVerification | POST | `/api/v2/verify/:verification_id/voice-capture` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, upload, validate | `backend/src/routes/newVerification.ts:2063` |  |
| 9 | newVerification | POST | `/api/v2/verify/:verification_id/finalize` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff | `backend/src/routes/newVerification.ts:2279` |  |
| 10 | newVerification | POST | `/api/v2/verify/:verification_id/restart` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, verificationRateLimit, validate | `backend/src/routes/newVerification.ts:2308` |  |
| 11 | newVerification | POST | `/api/v2/verify/:verification_id/internal/session` | `/api/v2/verify` | service-token | authenticateServiceToken, validate | `backend/src/routes/newVerification.ts:2425` |  |
| 12 | newVerification | GET | `/api/v2/verify/:verification_id/status` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, validate | `backend/src/routes/newVerification.ts:2513` |  |
| 13 | newVerification | GET | `/api/v2/verify/:verification_id/selfie` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, validate | `backend/src/routes/newVerification.ts:2577` |  |
| 14 | newVerification | POST | `/api/v2/verify/:verification_id/phone-otp/send` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, validate | `backend/src/routes/newVerification.ts:2658` |  |
| 15 | newVerification | POST | `/api/v2/verify/:verification_id/phone-otp/verify` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, validate | `backend/src/routes/newVerification.ts:2695` |  |
| 16 | handoff | POST | `/api/verify/handoff/create` | `/api/verify/handoff` | none-or-inline | basicRateLimit | `backend/src/routes/handoff.ts:16` | 是 |
| 17 | handoff | GET | `/api/verify/handoff/:token/session` | `/api/verify/handoff` | none-or-inline | - | `backend/src/routes/handoff.ts:116` | 是 |
| 18 | handoff | PATCH | `/api/verify/handoff/:token/link` | `/api/verify/handoff` | none-or-inline | - | `backend/src/routes/handoff.ts:189` | 是 |
| 19 | handoff | PATCH | `/api/verify/handoff/:token/complete` | `/api/verify/handoff` | none-or-inline | - | `backend/src/routes/handoff.ts:227` | 是 |
| 20 | handoff | GET | `/api/verify/handoff/:token/status` | `/api/verify/handoff` | none-or-inline | - | `backend/src/routes/handoff.ts:288` | 是 |
| 21 | developer/register | POST | `/api/developer/register` | `/api/developer` | none-or-inline | registrationRateLimit, validate | `backend/src/routes/developer/register.ts:26` | 是 |
| 22 | developer/apiKeys | POST | `/api/developer/api-key` | `/api/developer` | developer-jwt | apiKeyRateLimit, authenticateDeveloperJWT, validate | `backend/src/routes/developer/apiKeys.ts:25` |  |
| 23 | developer/apiKeys | GET | `/api/developer/api-keys` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard | `backend/src/routes/developer/apiKeys.ts:173` |  |
| 24 | developer/apiKeys | DELETE | `/api/developer/api-key/:keyId` | `/api/developer` | developer-jwt | apiKeyRateLimit, authenticateDeveloperJWT, validate | `backend/src/routes/developer/apiKeys.ts:233` |  |
| 25 | developer/apiKeys | POST | `/api/developer/api-key/:id/rotate` | `/api/developer` | developer-jwt | apiKeyRateLimit, authenticateDeveloperJWT, validate | `backend/src/routes/developer/apiKeys.ts:308` |  |
| 26 | developer/webhooks | GET | `/api/developer/webhooks` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard | `backend/src/routes/developer/webhooks.ts:43` |  |
| 27 | developer/webhooks | POST | `/api/developer/webhooks` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/webhooks.ts:83` |  |
| 28 | developer/webhooks | GET | `/api/developer/webhooks/:webhookId/secret` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/webhooks.ts:221` |  |
| 29 | developer/webhooks | GET | `/api/developer/webhooks/:webhookId/deliveries` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/webhooks.ts:263` |  |
| 30 | developer/webhooks | POST | `/api/developer/webhooks/:webhookId/test` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/webhooks.ts:313` |  |
| 31 | developer/webhooks | POST | `/api/developer/webhooks/:webhookId/deliveries/:deliveryId/resend` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/webhooks.ts:420` |  |
| 32 | developer/webhooks | DELETE | `/api/developer/webhooks/:webhookId` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/webhooks.ts:487` |  |
| 33 | developer/analytics | GET | `/api/developer/stats` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard | `backend/src/routes/developer/analytics.ts:37` |  |
| 34 | developer/analytics | GET | `/api/developer/activity` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard | `backend/src/routes/developer/analytics.ts:88` |  |
| 35 | developer/analytics | GET | `/api/developer/verifications/:verificationId` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard, validate | `backend/src/routes/developer/analytics.ts:196` |  |
| 36 | developer/analytics | GET | `/api/developer/analytics` | `/api/developer` | dashboard-jwt | apiKeyRateLimit, authenticateDashboard | `backend/src/routes/developer/analytics.ts:258` |  |
| 37 | developer/settings | GET | `/api/developer/settings/llm` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/settings.ts:17` |  |
| 38 | developer/settings | PUT | `/api/developer/settings/llm` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/settings.ts:52` |  |
| 39 | developer/settings | GET | `/api/developer/settings/sms` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/settings.ts:116` |  |
| 40 | developer/settings | PUT | `/api/developer/settings/sms` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/settings.ts:150` |  |
| 41 | developer/settings | GET | `/api/developer/settings/branding` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/settings.ts:202` |  |
| 42 | developer/settings | PUT | `/api/developer/settings/branding` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/settings.ts:226` |  |
| 43 | developer/settings | GET | `/api/developer/settings/page-builder` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/settings.ts:297` |  |
| 44 | developer/settings | PUT | `/api/developer/settings/page-builder` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/settings.ts:316` |  |
| 45 | developer/settings | PUT | `/api/developer/settings/page-builder/slug` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/settings.ts:335` |  |
| 46 | developer/settings | GET | `/api/developer/settings/aml` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/settings.ts:370` |  |
| 47 | developer/settings | PUT | `/api/developer/settings/aml` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/settings.ts:387` |  |
| 48 | developer/settings | GET | `/api/developer/settings/voice-auth` | `/api/developer` | developer-jwt | authenticateDeveloperJWT | `backend/src/routes/developer/settings.ts:412` |  |
| 49 | developer/settings | PUT | `/api/developer/settings/voice-auth` | `/api/developer` | developer-jwt | authenticateDeveloperJWT, validate | `backend/src/routes/developer/settings.ts:429` |  |
| 50 | developer/settings | GET | `/api/developer/settings/vc` | `/api/developer` | developer-jwt | authenticateDeveloperJWT | `backend/src/routes/developer/settings.ts:454` |  |
| 51 | developer/settings | PUT | `/api/developer/settings/vc` | `/api/developer` | developer-jwt | authenticateDeveloperJWT, validate | `backend/src/routes/developer/settings.ts:471` |  |
| 52 | developer/settings | GET | `/api/developer/settings/duplicate-detection` | `/api/developer` | developer-jwt | authenticateDeveloperJWT | `backend/src/routes/developer/settings.ts:498` |  |
| 53 | developer/settings | PUT | `/api/developer/settings/duplicate-detection` | `/api/developer` | developer-jwt | authenticateDeveloperJWT, validate | `backend/src/routes/developer/settings.ts:516` |  |
| 54 | developer/settings | GET | `/api/developer/settings/vault` | `/api/developer` | developer-jwt | authenticateDeveloperJWT | `backend/src/routes/developer/settings.ts:545` |  |
| 55 | developer/settings | PUT | `/api/developer/settings/vault` | `/api/developer` | developer-jwt | authenticateDeveloperJWT, validate | `backend/src/routes/developer/settings.ts:563` |  |
| 56 | developer/profile | DELETE | `/api/developer/account` | `/api/developer` | developer-jwt | apiKeyRateLimit, authenticateDeveloperJWT, validate | `backend/src/routes/developer/profile.ts:29` |  |
| 57 | developer/profile | GET | `/api/developer/profile` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/profile.ts:91` |  |
| 58 | developer/profile | PUT | `/api/developer/profile` | `/api/developer` | developer-jwt | authenticateDeveloperJWT, validate | `backend/src/routes/developer/profile.ts:128` |  |
| 59 | developer/profile | POST | `/api/developer/avatar` | `/api/developer` | developer-jwt | authenticateDeveloperJWT, avatarUpload | `backend/src/routes/developer/profile.ts:177` |  |
| 60 | developer/profile | POST | `/api/developer/branding/logo` | `/api/developer` | dashboard-jwt | authenticateDashboard, avatarUpload | `backend/src/routes/developer/profile.ts:216` |  |
| 61 | developer/profile | POST | `/api/developer/reviewers/invite` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/profile.ts:255` |  |
| 62 | developer/profile | GET | `/api/developer/reviewers` | `/api/developer` | dashboard-jwt | authenticateDashboard | `backend/src/routes/developer/profile.ts:329` |  |
| 63 | developer/profile | DELETE | `/api/developer/reviewers/:id` | `/api/developer` | dashboard-jwt | authenticateDashboard, validate | `backend/src/routes/developer/profile.ts:348` |  |
| 64 | admin | GET | `/api/admin/dashboard` | `/api/admin` | review-principal | authenticateReviewPrincipal | `backend/src/routes/admin.ts:27` |  |
| 65 | admin | GET | `/api/admin/verifications` | `/api/admin` | review-principal | authenticateReviewPrincipal, validate | `backend/src/routes/admin.ts:61` |  |
| 66 | admin | GET | `/api/admin/verification/:id` | `/api/admin` | review-principal | authenticateReviewPrincipal, validate | `backend/src/routes/admin.ts:147` |  |
| 67 | admin | GET | `/api/admin/verification/:id/duplicates` | `/api/admin` | admin-jwt|reviewer-jwt | authenticateAdminOrReviewer, validate | `backend/src/routes/admin.ts:308` |  |
| 68 | admin | PUT | `/api/admin/verification/:id/review` | `/api/admin` | review-principal | authenticateReviewPrincipal, validate | `backend/src/routes/admin.ts:378` |  |
| 69 | admin | GET | `/api/admin/developers` | `/api/admin` | jwt + admin-jwt | authenticateJWT, requireAdminRole, validate | `backend/src/routes/admin.ts:516` |  |
| 70 | admin | GET | `/api/admin/analytics` | `/api/admin` | admin-jwt|reviewer-jwt + org-admin|platform-admin | authenticateAdminOrReviewer, requireOrgAdminOrPlatformAdmin, validate | `backend/src/routes/admin.ts:576` |  |
| 71 | admin | GET | `/api/admin/health` | `/api/admin` | jwt + admin-jwt | authenticateJWT, requireAdminRole | `backend/src/routes/admin.ts:602` |  |
| 72 | admin | GET | `/api/admin/analytics/funnel` | `/api/admin` | admin-jwt|reviewer-jwt + org-admin|platform-admin | authenticateAdminOrReviewer, requireOrgAdminOrPlatformAdmin, validate | `backend/src/routes/admin.ts:755` |  |
| 73 | admin | GET | `/api/admin/analytics/rejections` | `/api/admin` | admin-jwt|reviewer-jwt + org-admin|platform-admin | authenticateAdminOrReviewer, requireOrgAdminOrPlatformAdmin, validate | `backend/src/routes/admin.ts:772` |  |
| 74 | admin | GET | `/api/admin/analytics/fraud-patterns` | `/api/admin` | admin-jwt|reviewer-jwt + org-admin|platform-admin | authenticateAdminOrReviewer, requireOrgAdminOrPlatformAdmin, validate | `backend/src/routes/admin.ts:787` |  |
| 75 | admin | GET | `/api/admin/analytics/risk-distribution` | `/api/admin` | admin-jwt|reviewer-jwt + org-admin|platform-admin | authenticateAdminOrReviewer, requireOrgAdminOrPlatformAdmin, validate | `backend/src/routes/admin.ts:802` |  |
| 76 | admin | GET | `/api/admin/provider-metrics` | `/api/admin` | jwt + admin-jwt | authenticateJWT, requireAdminRole | `backend/src/routes/admin.ts:817` |  |
| 77 | admin | GET | `/api/admin/audit/export` | `/api/admin` | admin-jwt|reviewer-jwt | authenticateAdminOrReviewer, validate | `backend/src/routes/admin.ts:834` |  |
| 78 | admin | DELETE | `/api/admin/user/:userId/data` | `/api/admin` | admin-jwt|reviewer-jwt + org-admin|platform-admin | authenticateAdminOrReviewer, requireOrgAdminOrPlatformAdmin | `backend/src/routes/admin.ts:1050` |  |
| 79 | admin-thresholds | GET | `/api/admin/thresholds` | `/api/admin/thresholds` | jwt + admin-jwt | authenticateJWT, requireAdminRole | `backend/src/routes/admin-thresholds.ts:20` |  |
| 80 | admin-thresholds | PUT | `/api/admin/thresholds` | `/api/admin/thresholds` | jwt + admin-jwt | authenticateJWT, requireAdminRole | `backend/src/routes/admin-thresholds.ts:82` |  |
| 81 | admin-thresholds | POST | `/api/admin/thresholds/preview` | `/api/admin/thresholds` | jwt + admin-jwt | authenticateJWT, requireAdminRole | `backend/src/routes/admin-thresholds.ts:194` |  |
| 82 | admin-thresholds | POST | `/api/admin/thresholds/reset` | `/api/admin/thresholds` | jwt + admin-jwt | authenticateJWT, requireAdminRole | `backend/src/routes/admin-thresholds.ts:249` |  |
| 83 | auth | POST | `/api/auth/logout` | `/api/auth` | none-or-inline | - | `backend/src/routes/auth.ts:83` | 是 |
| 84 | auth | GET | `/api/auth/csrf-token` | `/api/auth` | none-or-inline | - | `backend/src/routes/auth.ts:89` | 是 |
| 85 | auth | POST | `/api/auth/admin/login` | `/api/auth` | none-or-inline | adminLoginLimiter, validate | `backend/src/routes/auth.ts:95` | 是 |
| 86 | auth | POST | `/api/auth/admin/escalate` | `/api/auth` | none-or-inline | - | `backend/src/routes/auth.ts:175` | 是 |
| 87 | auth | POST | `/api/auth/developer/otp/send` | `/api/auth` | none-or-inline | otpSendLimiter, emailValidation, validate | `backend/src/routes/auth.ts:192` | 是 |
| 88 | auth | POST | `/api/auth/developer/otp/verify` | `/api/auth` | none-or-inline | otpVerifyLimiter, validate | `backend/src/routes/auth.ts:211` | 是 |
| 89 | auth | POST | `/api/auth/developer/otp/complete-registration` | `/api/auth` | none-or-inline | validate | `backend/src/routes/auth.ts:270` | 是 |
| 90 | auth | POST | `/api/auth/reviewer/otp/send` | `/api/auth` | none-or-inline | otpSendLimiter, emailValidation, validate | `backend/src/routes/auth.ts:365` | 是 |
| 91 | auth | POST | `/api/auth/reviewer/otp/verify` | `/api/auth` | none-or-inline | otpVerifyLimiter, validate | `backend/src/routes/auth.ts:398` | 是 |
| 92 | auth | POST | `/api/auth/developer/github/callback` | `/api/auth` | none-or-inline | validate | `backend/src/routes/auth.ts:490` | 是 |
| 93 | auth | GET | `/api/auth/developer/github/url` | `/api/auth` | none-or-inline | - | `backend/src/routes/auth.ts:660` | 是 |
| 94 | auth | POST | `/api/auth/totp/setup` | `/api/auth` | jwt | authenticateJWT | `backend/src/routes/auth.ts:675` |  |
| 95 | auth | POST | `/api/auth/totp/verify` | `/api/auth` | jwt | authenticateJWT | `backend/src/routes/auth.ts:693` |  |
| 96 | health | GET | `/api/health` | `/api/health` | none-or-inline | - | `backend/src/routes/health.ts:11` | 是 |
| 97 | health | GET | `/api/health/detailed` | `/api/health` | none-or-inline | - | `backend/src/routes/health.ts:24` | 是 |
| 98 | health | GET | `/api/health/ready` | `/api/health` | none-or-inline | - | `backend/src/routes/health.ts:89` | 是 |
| 99 | health | GET | `/api/health/live` | `/api/health` | none-or-inline | - | `backend/src/routes/health.ts:104` | 是 |
| 100 | health | GET | `/api/health/api-key-diagnostic` | `/api/health` | none-or-inline | - | `backend/src/routes/health.ts:109` | 是 |
| 101 | webhooks | POST | `/api/webhooks/register` | `/api/webhooks` | api-key | authenticateAPIKey, validate | `backend/src/routes/webhooks.ts:16` |  |
| 102 | webhooks | GET | `/api/webhooks` | `/api/webhooks` | api-key | authenticateAPIKey | `backend/src/routes/webhooks.ts:86` |  |
| 103 | webhooks | PUT | `/api/webhooks/:webhookId` | `/api/webhooks` | api-key | authenticateAPIKey, validate | `backend/src/routes/webhooks.ts:105` |  |
| 104 | webhooks | DELETE | `/api/webhooks/:webhookId` | `/api/webhooks` | api-key | authenticateAPIKey | `backend/src/routes/webhooks.ts:167` |  |
| 105 | webhooks | GET | `/api/webhooks/:webhookId/deliveries` | `/api/webhooks` | api-key | authenticateAPIKey | `backend/src/routes/webhooks.ts:194` |  |
| 106 | webhooks | POST | `/api/webhooks/:webhookId/test` | `/api/webhooks` | api-key | authenticateAPIKey | `backend/src/routes/webhooks.ts:231` |  |
| 107 | vaas | POST | `/api/vaas/users` | `/api/vaas` | service-token | authenticateServiceToken | `backend/src/routes/vaas.ts:13` |  |
| 108 | vaas | POST | `/api/vaas/verify` | `/api/vaas` | service-token | authenticateServiceToken | `backend/src/routes/vaas.ts:63` |  |
| 109 | vaas | GET | `/api/vaas/verify/:verification_id/status` | `/api/vaas` | service-token | authenticateServiceToken | `backend/src/routes/vaas.ts:114` |  |
| 110 | vaas | GET | `/api/vaas/developers` | `/api/vaas` | service-token | authenticateServiceToken | `backend/src/routes/vaas.ts:145` |  |
| 111 | vaas | GET | `/api/vaas/developers/:id` | `/api/vaas` | service-token | authenticateServiceToken, validate | `backend/src/routes/vaas.ts:213` |  |
| 112 | vaas | POST | `/api/vaas/developers/:id/suspend` | `/api/vaas` | service-token | authenticateServiceToken, validate | `backend/src/routes/vaas.ts:244` |  |
| 113 | vaas | POST | `/api/vaas/developers/:id/unsuspend` | `/api/vaas` | service-token | authenticateServiceToken, validate | `backend/src/routes/vaas.ts:264` |  |
| 114 | vaas | GET | `/api/vaas/health` | `/api/vaas` | service-token | authenticateServiceToken | `backend/src/routes/vaas.ts:284` |  |
| 115 | batch | POST | `/api/v2/batch/upload` | `/api/v2/batch` | api-key | authenticateAPIKey, checkSandboxMode, validate | `backend/src/routes/batch.ts:206` |  |
| 116 | batch | GET | `/api/v2/batch/:id/status` | `/api/v2/batch` | api-key | authenticateAPIKey, validate | `backend/src/routes/batch.ts:260` |  |
| 117 | batch | GET | `/api/v2/batch/:id/results` | `/api/v2/batch` | api-key | authenticateAPIKey, validate | `backend/src/routes/batch.ts:294` |  |
| 118 | batch | POST | `/api/v2/batch/:id/cancel` | `/api/v2/batch` | api-key | authenticateAPIKey, validate | `backend/src/routes/batch.ts:318` |  |
| 119 | batch | GET | `/api/v2/batch` | `/api/v2/batch` | api-key | authenticateAPIKey, validate | `backend/src/routes/batch.ts:348` |  |
| 120 | addressVerification | POST | `/api/v2/verify/:verification_id/address-document` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, verificationRateLimit, upload, validate | `backend/src/routes/addressVerification.ts:41` |  |
| 121 | addressVerification | GET | `/api/v2/verify/:verification_id/address-status` | `/api/v2/verify` | api-key|handoff-token | authenticateAPIKeyOrHandoff, validate | `backend/src/routes/addressVerification.ts:165` |  |
| 122 | pageConfig | GET | `/api/v2/verify/page-config` | `/api/v2/verify` | none-or-inline | basicRateLimit | `backend/src/routes/pageConfig.ts:15` | 是 |
| 123 | pageConfig | GET | `/api/v2/verify/page-config/slug/:slug` | `/api/v2/verify` | none-or-inline | basicRateLimit | `backend/src/routes/pageConfig.ts:76` | 是 |
| 124 | pageConfig | GET | `/api/v2/verify/session-info` | `/api/v2/verify` | none-or-inline | basicRateLimit | `backend/src/routes/pageConfig.ts:111` | 是 |
| 125 | monitoring | POST | `/api/v2/monitoring/schedules` | `/api/v2/monitoring` | api-key | authenticateAPIKey, validate | `backend/src/routes/monitoring.ts:32` |  |
| 126 | monitoring | GET | `/api/v2/monitoring/schedules` | `/api/v2/monitoring` | api-key | authenticateAPIKey, validate | `backend/src/routes/monitoring.ts:78` |  |
| 127 | monitoring | GET | `/api/v2/monitoring/schedules/:schedule_id` | `/api/v2/monitoring` | api-key | authenticateAPIKey, validate | `backend/src/routes/monitoring.ts:116` |  |
| 128 | monitoring | DELETE | `/api/v2/monitoring/schedules/:schedule_id` | `/api/v2/monitoring` | api-key | authenticateAPIKey, validate | `backend/src/routes/monitoring.ts:147` |  |
| 129 | monitoring | GET | `/api/v2/monitoring/expiring-documents` | `/api/v2/monitoring` | api-key | authenticateAPIKey, validate | `backend/src/routes/monitoring.ts:180` |  |
| 130 | status | GET | `/api/status` | `/api/status` | none-or-inline | - | `backend/src/routes/status.ts:23` | 是 |
| 131 | setup | GET | `/api/setup/status` | `/api/setup` | none-or-inline | - | `backend/src/routes/setup.ts:31` | 是 |
| 132 | setup | POST | `/api/setup/initialize` | `/api/setup` | none-or-inline | setupRateLimit, validate | `backend/src/routes/setup.ts:54` | 是 |
| 133 | credentials | GET | `/api/v2/verify/:id/credential` | `/api/v2/verify` | api-key | authenticateAPIKey | `backend/src/routes/credentials.ts:18` |  |
| 134 | credentials | POST | `/api/v2/verify/:id/credential/send` | `/api/v2/verify` | api-key | authenticateAPIKey | `backend/src/routes/credentials.ts:91` |  |
| 135 | credentials | POST | `/api/v2/verify/credentials/:jti/revoke` | `/api/v2/verify` | api-key | authenticateAPIKey | `backend/src/routes/credentials.ts:188` |  |
| 136 | credentials | GET | `/api/v2/verify/credentials/:jti/status` | `/api/v2/verify` | none-or-inline | - | `backend/src/routes/credentials.ts:209` | 是 |
| 137 | credentials | GET | `/api/v2/:id/credential` | `/api/v2` | api-key | authenticateAPIKey | `backend/src/routes/credentials.ts:18` |  |
| 138 | credentials | POST | `/api/v2/:id/credential/send` | `/api/v2` | api-key | authenticateAPIKey | `backend/src/routes/credentials.ts:91` |  |
| 139 | credentials | POST | `/api/v2/credentials/:jti/revoke` | `/api/v2` | api-key | authenticateAPIKey | `backend/src/routes/credentials.ts:188` |  |
| 140 | credentials | GET | `/api/v2/credentials/:jti/status` | `/api/v2` | none-or-inline | - | `backend/src/routes/credentials.ts:209` | 是 |
| 141 | compliance | POST | `/api/v2/compliance/rulesets` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:74` |  |
| 142 | compliance | GET | `/api/v2/compliance/rulesets` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:103` |  |
| 143 | compliance | GET | `/api/v2/compliance/rulesets/:id` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:139` |  |
| 144 | compliance | PUT | `/api/v2/compliance/rulesets/:id` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:167` |  |
| 145 | compliance | DELETE | `/api/v2/compliance/rulesets/:id` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:212` |  |
| 146 | compliance | POST | `/api/v2/compliance/rulesets/:id/rules` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:237` |  |
| 147 | compliance | PUT | `/api/v2/compliance/rules/:id` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:280` |  |
| 148 | compliance | DELETE | `/api/v2/compliance/rules/:id` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:339` |  |
| 149 | compliance | POST | `/api/v2/compliance/evaluate` | `/api/v2/compliance` | compliance-request | authenticateComplianceRequest | `backend/src/routes/compliance.ts:379` |  |
| 150 | vault | POST | `/api/v2/vault/store` | `/api/v2/vault` | api-key | authenticateAPIKey, requireVaultEnabled | `backend/src/routes/vault.ts:39` |  |
| 151 | vault | GET | `/api/v2/vault/share/:shareToken` | `/api/v2/vault` | none-or-inline | - | `backend/src/routes/vault.ts:104` | 是 |
| 152 | vault | GET | `/api/v2/vault/:token` | `/api/v2/vault` | api-key | authenticateAPIKey, requireVaultEnabled | `backend/src/routes/vault.ts:174` |  |
| 153 | vault | GET | `/api/v2/vault/:token/attributes/:attr` | `/api/v2/vault` | api-key | authenticateAPIKey, requireVaultEnabled | `backend/src/routes/vault.ts:202` |  |
| 154 | vault | DELETE | `/api/v2/vault/:token` | `/api/v2/vault` | api-key | authenticateAPIKey, requireVaultEnabled | `backend/src/routes/vault.ts:232` |  |
| 155 | vault | GET | `/api/v2/vault` | `/api/v2/vault` | api-key | authenticateAPIKey, requireVaultEnabled | `backend/src/routes/vault.ts:261` |  |
| 156 | vault | POST | `/api/v2/vault/:token/share` | `/api/v2/vault` | api-key | authenticateAPIKey, requireVaultEnabled | `backend/src/routes/vault.ts:301` |  |
| 157 | well-known | GET | `/.well-known/did.json` | `/.well-known` | none-or-inline | - | `backend/src/routes/well-known.ts:12` | 是 |
| 158 | system | GET | `/api/system/version` | `/api/system` | developer-jwt | authenticateDeveloperJWT | `backend/src/routes/system.ts:52` |  |

## app 级路由

| 方法 | 路径 | 处理函数 | 条件 | 源码位置 |
| --- | --- | --- | --- | --- |
| GET | `/api/public/assets/*` | servePublicAsset | storage.provider=local|s3 | `backend/src/server.ts:218` |
| GET | `/api/files/*` | serveLocalFile | - | `backend/src/server.ts:226` |
| GET | `/health` | inline | - | `backend/src/server.ts:238` |
| GET | `/api/health` | inline | - | `backend/src/server.ts:246` |
| GET | `/` | inline | - | `backend/src/server.ts:256` |
| GET | `/api/docs/markdown` | inline | - | `backend/src/server.ts:268` |
| GET | `/api/docs` | inline | - | `backend/src/server.ts:283` |

## 说明与待确认

- 静态解析；鉴权仅按中间件名推断，handler 内联鉴权标记为 none-or-inline，需要逐条人工确认
- 同一路由模块挂载到多个前缀时展开为多条路径（credentials: /api/v2/verify 与 /api/v2）
- cloud-only 动态挂载不计入 154
- 请求/响应字段未在本清单展开，见 M0 后续契约 schema 抽取
