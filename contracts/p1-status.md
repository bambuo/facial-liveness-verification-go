# P1 状态报告（2026-11-18）

## 结论

P1 **完成**（代码基线与本地验证）：核心 15 接口、6/6 引擎端点、API Key/handoff/session/service-token 鉴权、S3/SMTP/SMS/Webhook 适配器与 9 个周期任务。

边界说明：

- SMTP/SMS/Webhook 需要真实商户/网关凭证才能做外部联调；当前已完成代码路径、配置校验与本地/错误路径测试。
- 核心 15 接口覆盖主流程与状态守卫；审核工作流、完整幂等重放等 1:1 细化归入 P2。
- 其余 126 条契约路由属于 P2/P3，当前覆盖 32/158，不影响 P1 核心链路。

## 交付与证据

| 项 | 结果 | 证据 |
| --- | --- | --- |
| 核心 15 接口 | 已接入 | `/api/v2/verify` initialize/re-verify/front/back/cross/live/voice/finalize/restart/internal/status/selfie/phone-otp |
| 引擎端点 | 6/6 | `back`、`live`、`voice-enroll`、`voice-verify` 新增；`front`、`ocr` 已有 |
| 鉴权矩阵 | 已接入 | `X-API-Key`、`X-Handoff-Token`、`X-Session-Token`、`X-Service-Token`；17 条公开端点已注册 |
| S3 | 已接入 | 正/反面文档与自拍上传，`GET .../selfie` 返回预签名 URL |
| SMTP | 已接入 | 开发者 OTP 邮件；未配置 SMTP 时开发环境回显验证码 |
| SMS | 已接入 | 手机 OTP 通过 HTTP 网关发送；发送失败返回 503 |
| Webhook | 已接入 | finalize 后异步投递，HMAC 签名，落 `webhook_deliveries` |
| 定时任务 | 9 个 | 临时清理、留存、demo、放弃采集、活动日志、幂等键、Webhook 载荷、监控、一致性 |
| 契约覆盖 | 32/158 后端 + 6/6 引擎 | `go run ./cmd/contractcheck` |

## 验证命令

```bash
CGO_ENABLED=0 go build ./...
go test ./... -count=1
VOICE_MODEL_DIR=/path/to/models/voice go test ./internal/audio -run TestVoiceEngine -count=1
go run ./cmd/contractcheck
```

## 非阻塞后续

1. 剩余管理/开发者/合规/保险库/批量和前端属于 P2–P4。
2. 核心 15 接口的审核工作流、完整幂等重放等 1:1 细化归入 P2。
3. SMTP/SMS/Webhook 真实商户凭证联调待提供。
4. 并发 P95、Docker 构建与 CI 门禁属后续验证；CI 继续按用户决定禁用。
