# 鉴权签字记录（2026-11-18）

依据：idswyft 当前实现、C3（完全采用 idswyft 认证模型）与 P0 评审结论。

## 签字结论

17 条公开候选全部确认为**公开端点**，M1 鉴权矩阵按下表实现。

| # | 模块 | 方法 | 路径 | 保护方式 |
| --- | --- | --- | --- | --- |
| 1 | 开发者 | POST | `/api/developer/register` | 注册限流、邮箱校验 |
| 2 | 认证 | POST | `/api/auth/admin/escalate` | 固定返回 410（已下线） |
| 3 | 认证 | POST | `/api/auth/developer/otp/send` | OTP 发送限流 |
| 4 | 认证 | POST | `/api/auth/developer/otp/verify` | OTP 校验与尝试次数限制 |
| 5 | 认证 | POST | `/api/auth/reviewer/otp/send` | OTP 发送限流 |
| 6 | 认证 | POST | `/api/auth/reviewer/otp/verify` | OTP 校验与尝试次数限制 |
| 7 | 认证 | POST | `/api/auth/developer/github/callback` | OAuth state 校验 |
| 8 | 认证 | GET | `/api/auth/developer/github/url` | 仅返回配置状态与跳转 URL |
| 9 | 健康 | GET | `/api/health` | 公开 |
| 10 | 健康 | GET | `/api/health/detailed` | 公开；不返回敏感配置 |
| 11 | 健康 | GET | `/api/health/ready` | 公开 |
| 12 | 健康 | GET | `/api/health/live` | 公开 |
| 13 | 页面配置 | GET | `/api/v2/verify/page-config/slug/:slug` | 按公开 slug 查询 |
| 14 | 状态 | GET | `/api/status` | 公开 |
| 15 | 初始化 | GET | `/api/setup/status` | 公开；只返回是否已初始化 |
| 16 | 初始化 | POST | `/api/setup/initialize` | 首跑保护：零开发者时才可初始化 |
| 17 | `.well-known` | GET | `/.well-known/did.json` | 公开 |

## 相关但非公开候选

以下端点已有内联证据，M1 按证据实现：

- `GET /api/health/api-key-diagnostic`：query `key` 做 HMAC 诊断，不返回密钥。
- `POST /api/auth/logout`、`GET /api/auth/csrf-token`：Cookie/CSRF 语义。
- `GET /api/v2/vault/share/:shareToken`：分享令牌（过期/次数限制）。
- `GET /api/v2/verify/credentials/:jti/status`：按 jti 查询凭证状态（两个挂载前缀）。

## 影响

- `contracts/auth.json` 中这 17 条由 `public-candidate` 转为签名公开；M1 不再要求鉴权。
- 契约测试与文档按“公开”断言，不引入 401/403。
- 若后续改为需鉴权，必须走新的 M0 决策并更新本文件与鉴权矩阵。
