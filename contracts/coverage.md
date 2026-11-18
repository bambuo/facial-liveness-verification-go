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

后端接口：**32 / 158** 可达路径（154 注册）

引擎端点：**6 / 6**

### 已实现后端接口

| 方法 | 路径 |
| --- | --- |
| GET | `/.well-known/did.json` |
| GET | `/api/auth/developer/github/url` |
| GET | `/api/health` |
| GET | `/api/health/detailed` |
| GET | `/api/health/live` |
| GET | `/api/health/ready` |
| GET | `/api/setup/status` |
| GET | `/api/status` |
| GET | `/api/v2/verify/:verification_id/selfie` |
| GET | `/api/v2/verify/:verification_id/status` |
| GET | `/api/v2/verify/page-config/slug/:slug` |
| POST | `/api/auth/admin/escalate` |
| POST | `/api/auth/developer/github/callback` |
| POST | `/api/auth/developer/otp/send` |
| POST | `/api/auth/developer/otp/verify` |
| POST | `/api/auth/reviewer/otp/send` |
| POST | `/api/auth/reviewer/otp/verify` |
| POST | `/api/developer/register` |
| POST | `/api/setup/initialize` |
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
| POST | `/api/v2/verify/initialize` |
| POST | `/api/v2/verify/re-verify` |

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
