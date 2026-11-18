# idswyft 鉴权证据（M0 契约冻结输入）

- 来源：`/Users/johana/Codes/github/idswyft-community-1.12.39` @ `83dd51bdfb6dec608856b41c110f54f5a8bf3efe`
- 生成时间：2026-11-18T09:21:58.349Z
- 口径：中间件鉴权来自 endpoints.json；内联鉴权来自 handler 内调用/成员/请求头；无证据标记为 public-candidate，仍需人工签字。

## 汇总

| 分类 | 数量 |
| --- | --- |
| middleware | 126 |
| inline | 15 |
| public-candidate | 17 |

## 内联鉴权证据

| 方法 | 路径 | 机制 | 证据 |
| --- | --- | --- | --- |
| POST | `/api/verify/handoff/create` | session-token-header, handoff-token | req.headers['x-session-token']@18<br>hashHandoffToken()@34<br>from('api_keys')@66<br>hashHandoffToken()@87 |
| GET | `/api/verify/handoff/:token/session` | path-token, handoff-token | const { token } = req.params@117<br>hashHandoffToken()@123<br>from('api_keys')@155<br>from('developers')@162 |
| PATCH | `/api/verify/handoff/:token/link` | path-token, handoff-token | const { token } = req.params@190<br>hashHandoffToken()@205 |
| PATCH | `/api/verify/handoff/:token/complete` | path-token, handoff-token | const { token } = req.params@228<br>hashHandoffToken()@250 |
| GET | `/api/verify/handoff/:token/status` | path-token, handoff-token | const { token } = req.params@289<br>hashHandoffToken()@295 |
| POST | `/api/auth/logout` | cookie | res.clearCookie()@84 |
| GET | `/api/auth/csrf-token` | csrf | generateToken()@90 |
| POST | `/api/auth/admin/login` | jwt | verifyToken()@148 |
| POST | `/api/auth/developer/otp/complete-registration` | registration-token | verifyRegistrationToken()@283<br>from('developers')@289<br>from('developers')@300<br>from('api_keys')@318 |
| GET | `/api/health/api-key-diagnostic` | api-key-query | req.query.key@110<br>from('api_keys')@124 |
| GET | `/api/v2/verify/page-config` | api-key-query | req.query.api_key@18<br>from('api_keys')@30<br>from('developers')@45 |
| GET | `/api/v2/verify/session-info` | session-token-header, handoff-token | req.headers['x-session-token']@114<br>hashHandoffToken()@125<br>from('developers')@144 |
| GET | `/api/v2/verify/credentials/:jti/status` | path-jti | const { jti } = req.params@211 |
| GET | `/api/v2/credentials/:jti/status` | path-jti | const { jti } = req.params@211 |
| GET | `/api/v2/vault/share/:shareToken` | share-token | const { shareToken } = req.params@106 |

## public-candidate（需签字）

| 方法 | 路径 | 模块 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/developer/register` | developer/register | 未发现中间件或内联鉴权证据 |
| POST | `/api/auth/admin/escalate` | auth | 未发现中间件或内联鉴权证据 |
| POST | `/api/auth/developer/otp/send` | auth | 未发现中间件或内联鉴权证据 |
| POST | `/api/auth/developer/otp/verify` | auth | 未发现中间件或内联鉴权证据 |
| POST | `/api/auth/reviewer/otp/send` | auth | 未发现中间件或内联鉴权证据 |
| POST | `/api/auth/reviewer/otp/verify` | auth | 未发现中间件或内联鉴权证据 |
| POST | `/api/auth/developer/github/callback` | auth | 未发现中间件或内联鉴权证据 |
| GET | `/api/auth/developer/github/url` | auth | 未发现中间件或内联鉴权证据 |
| GET | `/api/health` | health | 未发现中间件或内联鉴权证据 |
| GET | `/api/health/detailed` | health | 未发现中间件或内联鉴权证据 |
| GET | `/api/health/ready` | health | 未发现中间件或内联鉴权证据 |
| GET | `/api/health/live` | health | 未发现中间件或内联鉴权证据 |
| GET | `/api/v2/verify/page-config/slug/:slug` | pageConfig | 未发现中间件或内联鉴权证据 |
| GET | `/api/status` | status | 未发现中间件或内联鉴权证据 |
| GET | `/api/setup/status` | setup | 未发现中间件或内联鉴权证据 |
| POST | `/api/setup/initialize` | setup | 未发现中间件或内联鉴权证据 |
| GET | `/.well-known/did.json` | well-known | 未发现中间件或内联鉴权证据 |

## 说明

- 内联鉴权按调用名/成员访问/请求头匹配，不判断完整分支逻辑；public-candidate 必须人工签字
- authenticateComplianceRequest 的内部实现可能是 API Key 或 Reviewer JWT，需在 M1 打开实现确认
- C3 已决定完全采用 idswyft 认证模型；本文件作为 M1 鉴权矩阵的实现依据
