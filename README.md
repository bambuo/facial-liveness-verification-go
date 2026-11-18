# Go 版身份核验服务

用 Go 1.27 + `onnx-gomlx` + GoMLX 纯 Go 后端实现的人脸核验服务；当前镜像 .NET 版人脸核验语义，目标是以其为内核**契约级 1:1 重写 idswyft**。构建与运行不依赖 cgo，也不加载 `libonnxruntime`。

## 当前状态

| 状态 | 内容 |
| --- | --- |
| 已实现已验证 | 人脸 1:1（YuNet + SFace）、活体（MiniFASNetV2）、动作校验、challenge 一次性/配额/失败锁定、档案加密、合规删除、纯 Go ONNX（PRelu/Dropout/initializer/Clip/MaxPool 等规范补丁）、纯 Go OCR（PP-OCRv5 检测/识别/行分组、中英模型、自动方向、Slice 属性补丁、MRZ/US/UK/CN 字段规则 v1）、GORM 31 模型与启动 AutoMigrate（advisory lock、schema parity、双副本并发）、idswyft 154 接口代码基线（156/158 静态、158/158 运行时）、6/6 引擎端点、handoff/batch/vault/VC/webhooks/admin/developer/compliance/VaaS/monitoring、PDF417/MRZ/篡改/深伪、S3/SMTP/SMS/Webhook 与 9 个周期任务 |
| 已实现未验证 | 语音离线算子（Kaldi fbank、NeMo 归一化、CTC、CAM++/NeMo int8，WER 4–6%）；SMTP/SMS/Webhook 真实外部服务联调；idswyft 全量 154 接口、Docker 构建、并发 P95、CI 集成测试 |
| 未实现 | 条码/MRZ、文档字段规则库、地址验证、批量、移动交接、VaaS、可验证凭证、保险库、合规规则、AML、风险评分、去重、审核工作台、开发者门户、22 页前端、3 套 SDK、9 个定时任务 |
| 基线不做 | 1:N 识别、pgvector、cgo、Node sidecar、cloud `platform/*` 路由 |

契约输入见 [contracts/README.md](contracts/README.md)：154 接口、31 表、schema 差异、7 个硬冲突。

## 已实现范围（12 条 Go 内核路由 + 32 条 idswyft 契约路由 + 6 个引擎端点）

| 方法与路径 | 功能 |
| --- | --- |
| `POST /v1/challenges` | 下发 challenge、动作序列与帧数 |
| `POST /v1/action-check` | 单动作实时预检，不消费 challenge |
| `POST /v1/enrollments` | 活体与动作校验后建立或覆盖档案 |
| `POST /v1/verifications` | 活体与动作校验后做档案 1:1 |
| `GET /v1/subjects/{subject_id}` | 查询档案状态与模型版本 |
| `DELETE /v1/subjects/{subject_id}` | 合规删除档案、索引与对象 |
| `POST /extract/ocr` | OCR 原始文本、行与置信度 |
| `POST /extract/front` | OCR 字段 + 人脸特征；必填字段缺失返回 422 |
| `POST /v1/liveness-checks` | 仅 lab：逐帧证据与 APCER 评测 |
| `POST /dev/token` | 仅 Development 签发 dev JWT |
| `GET /dev/face-test.html` | 仅 Development 的 H5 采集页 |
| `GET /health/live`、`GET /health/ready` | 健康探针 |
| `GET /version` | 构建/模型/配置哈希 |

未实现：`sessions`、`matches`、`documents`、`identifications`、`operations`、结果令牌消费——与 .NET 基线当前注册范围一致。

## 架构

洋葱架构，依赖只由外向内：

```text
cmd/api, cmd/worker                组装根
internal/adapter/*                 HTTP、PostgreSQL、Redis、ONNX、S3、安全、种子
internal/app                       用例编排
internal/port                      端口接口
internal/domain                    实体、值对象、领域规则
```

推理层是唯一替换点：`internal/adapter/onnx` 通过 `third_party/onnx-gomlx`（补齐 PRelu 与 Dropout 等）把模型转换为 GoMLX 计算图，并固定使用纯 Go 后端。详见 [docs/架构全景图.md](docs/架构全景图.md) 与 [docs/功能/01-项目架构.md](docs/功能/01-项目架构.md)。

## 技术栈

- Go 1.27 / `onnx-gomlx v0.5.32` / `gomlx v0.28.35` / `compute v0.1.29`
- PostgreSQL 16（`pgx/v5`）、Redis 7（`go-redis/v9`）、S3（`aws-sdk-go-v2`）
- JWT（`golang-jwt/v5`）、UUID、`golang.org/x/image`
- 模型：YuNet 检测、SFace 识别、MiniFASNetV2 活体

## 环境要求

- Go 1.27 及以上；`CGO_ENABLED=0`
- PostgreSQL 16、Redis 7；S3/RustFS 可选
- 三个 ONNX 模型放在 `models/` 或通过 `FACE_MODEL_DIR` 指定
- 必需环境：`AUTH_DEV_SIGNING_KEY`、`AUTH_RESULT_TOKEN_SIGNING_KEY`、`SECURITY_MASTER_KEY`（生产缺失即启动失败）

## 快速开始

```bash
# 1. 启动本地依赖（PostgreSQL、Redis、RustFS）
cp deploy/dev.env.example deploy/.env
# 修改 deploy/.env 中的全部占位值
docker compose -f deploy/docker-compose.dev.yml up -d

# 2. 构建
CGO_ENABLED=0 go build ./...

# 3. 启动 API
export APP_ENV=development
export AUTH_ENABLE_DEV_TOKEN=true
export AUTH_DEV_SIGNING_KEY="$(openssl rand -base64 32)"
export AUTH_RESULT_TOKEN_SIGNING_KEY="$(openssl rand -base64 32)"
export SECURITY_MASTER_KEY="$(openssl rand -base64 32)"
export DATABASE_URL="host=localhost port=5432 user=faceservice password=<POSTGRES_PASSWORD> dbname=faceservice sslmode=disable"
export REDIS_ADDR=localhost:16379
export REDIS_PASSWORD="<REDIS_PASSWORD>"
go run ./cmd/api
```

打开 `http://localhost:8080/dev/face-test.html` 做 H5 联调。

## 契约冻结（M0）

```bash
cd contracts/tools
npm ci
node extract-endpoints.mjs --out ..   # 154 注册 / 158 路径
node extract-schema.mjs --out ..      # 31 表 / 329 列 / 10 函数 / 6 触发器 / 53 策略
```

产物：`contracts/endpoints.json`、`contracts/schema.json`、`contracts/schema-diff.md`、`contracts/conflicts.md`。数量校验失败时脚本非 0 退出。7 个硬冲突（双主体模型、幂等/jti、认证映射、存储 provider、SQL 函数/RLS、Remotion、ML 分数）未决策前不进入 M1。

## 接口约定

- 统一响应封装与错误结构；用户可见错误带 `messageKey`（en/zh-CN）。
- `X-Request-Id` 统一为 `req_` + 20 hex；缺失时服务端生成。
- 鉴权：租户 JWT（`tenant` 非空且 active；`sub` 为业务用户 ID）；结果令牌 `aud` 等于租户 code。
- 请求体上限：challenge/action-check/enroll/verify 统一；单帧超限 422 `LOW_QUALITY`。
- 风控拦截在 challenge 消费之前；活体 `real` 清零，`spoof`/`fuzzy` 累计。

## 配置

只列字段名，真实值通过环境变量/`deploy/.env` 注入：

`APP_ENV`、`DATABASE_URL`、`DB_AUTO_MIGRATE`、`DB_MIGRATION_LOCK_SECONDS`、`DB_MIGRATION_STATEMENT_SECONDS`、`REDIS_ADDR`、`REDIS_PASSWORD`、`FACE_MODEL_DIR`、`AUTH_DEV_SIGNING_KEY`、`AUTH_RESULT_TOKEN_SIGNING_KEY`、`SECURITY_MASTER_KEY`、`LIVENESS_GUARD_ENABLED`、`LIVENESS_FAILURE_LIMIT`、`LIVENESS_WINDOW_SECONDS`、`LIVENESS_LOCK_SECONDS`、`LIVENESS_LAB_ENABLED`、`OCR_ENABLED`、`OCR_MODEL_DIR`、`OCR_DET_MODEL`、`OCR_REC_MODEL`、`OCR_DICT`、`OCR_WIDTH_MODE`、`OCR_ORIENTATION`、`VOICE_ENABLED`、`VOICE_MODEL_DIR`、`VOICE_SPEAKER_MODEL`、`VOICE_ASR_MODEL`、`VOICE_TOKENS`、`S3_*`、`SMTP_ENABLED`、`SMTP_HOST`、`SMTP_PORT`、`SMTP_USERNAME`、`SMTP_PASSWORD`、`SMTP_FROM`、`SMS_ENABLED`、`SMS_ENDPOINT`、`SMS_API_KEY`、`SMS_FROM`、`API_KEY_SECRET`、`SERVICE_TOKEN`、`DATA_RETENTION_DAYS`。

## 测试

```bash
CGO_ENABLED=0 go build ./...
go test ./...
```

全量 `go test ./...` 通过；OCR 端点单测与真实模型集成测试已加入（未设置 `OCR_TEST_MODEL_DIR` 时跳过）；语音引擎真实模型测试通过（`VOICE_MODEL_DIR`）；无 CI 集成测试、无并发 P95 基线。模型相关测试在缺少模型时跳过。OCR 黄金集与指标见 [eval/ocr/README.md](eval/ocr/README.md)，评测见 [docs/功能/05-识别质量评测.md](docs/功能/05-识别质量评测.md)。

## 部署

```bash
docker compose -f deploy/docker-compose.dev.yml up -d
# 生产由 Caddy/Nginx 终止 TLS 并反代 api:8080
```

`/health/live`、`/health/ready` 用于探针；`/version` 返回构建/模型/配置哈希与冻结参数。

## 开发规范

见 [AGENTS.md](AGENTS.md)：洋葱依赖、纯 Go、契约优先、状态机严格有序、Webhook 不阻塞、测试与交付标准。

## 安全

- 人脸向量加密落库；日志不输出证件号、完整姓名、令牌。
- 保留期与删除路径遵循 `dataRetention`；删除档案时清理对象。
- 不削弱 API Key HMAC、JWT、CSRF、限流、SSRF 守卫、幂等与路径校验。
- 与 .NET 共用库/主密钥的兼容性变更先确认。

## 已知限制

- PAD 商用可信度不足：0.91 阈值 APCER 16.67% / BPCER 17.24%；高仿真面具仍可漏过。
- 单图识别约 1.1s（AuraFace 评测口径）；并发 P95 未验证。
- 无 HA、监控、告警与 7×24 SLA 依据。
- Redis 配额键格式与 .NET 不一致，计数不合并。
- `scene` 与 JWT 一致性要求未落到代码（待决策）。

## 文档

- [文档索引](docs/README.md)
- [架构全景图](docs/架构全景图.md)
- [功能清单](docs/功能/02-功能清单.md)
- [领域术语表](docs/功能/03-领域术语表.md)
- [idswyft 重写与契约冻结](docs/流程/03-idswyft重写与契约冻结-v1.md)
- [契约冻结输入](contracts/README.md)
- [OCR 黄金集与评测](eval/ocr/README.md)
