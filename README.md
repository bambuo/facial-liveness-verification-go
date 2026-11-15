# Go 版人脸核验服务

用 Go 1.27 + `onnx-gomlx` + GoMLX 纯 Go 后端复刻 .NET 版人脸核验服务的现有语义。构建与运行不依赖 cgo，也不加载 `libonnxruntime`。

## 已实现范围

| 方法与路径 | 功能 |
| --- | --- |
| `POST /v1/challenges` | 下发 challenge、动作序列与帧数 |
| `POST /v1/action-check` | 单动作实时预检，不消费 challenge |
| `POST /v1/enrollments` | 活体与动作校验后建立或覆盖档案 |
| `POST /v1/verifications` | 活体与动作校验后做档案 1:1 |
| `GET /v1/subjects/{subject_id}` | 查询档案状态与模型版本 |
| `DELETE /v1/subjects/{subject_id}` | 合规删除档案、索引与对象 |
| `POST /dev/token` | 仅 Development 签发 dev JWT |
| `GET /dev/face-test.html` | 仅 Development 的 H5 采集页 |
| `GET /health/live`、`GET /health/ready` | 健康探针 |

未实现：`sessions`、`matches`、`liveness-checks`、`identifications`、`documents`、`operations`、结果令牌消费。与 .NET 基线当前注册范围一致。

## 架构

洋葱架构，依赖只由外向内：

```text
cmd/api, cmd/worker                组装根
internal/adapter/*                 HTTP、PostgreSQL、Redis、ONNX、S3、安全、种子
internal/app                       用例编排
internal/port                      端口接口
internal/domain                    实体、值对象、领域规则
```

推理层是唯一替换点：`internal/adapter/onnx` 通过 `third_party/onnx-gomlx`（补齐 PRelu 与 Dropout）把三个模型转换为 GoMLX 计算图，并固定使用纯 Go 后端。

## 前置条件

- Go 1.27 及以上。
- PostgreSQL 16、Redis 7。
- 三个 ONNX 模型：`face_detection_yunet_2023mar.onnx`、`face_recognition_sface_2021dec.onnx`、`2.7_80x80_MiniFASNetV2.onnx`，放在 `models/` 或通过 `FACE_MODEL_DIR` 指定。

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
export FACE_MODEL_DIR=./models
go run ./cmd/api

# 4. 就绪检查
curl http://localhost:8080/health/ready
# {"engine":"mica-sface-128d-1.0.2","status":"ready"}
```

Worker：

```bash
go run ./cmd/worker
```

Docker：

```bash
docker build -f deploy/Dockerfile -t faceservice-go:local .
docker run --rm -p 8080:8080 --env-file deploy/.env faceservice-go:local
```

配置项与默认值见 [实施方案](docs/需求/0001/0001-Go版人脸核验服务-实施方案.md) 第 3 节。

### 推荐安全档位

高风险场景（KYC、提现、敏感操作）建议显式设置：

```bash
export FACE_LIVENESS_THRESHOLD=0.91
export FACE_ACTIONS_COUNT=3
export FACE_FRAMES_PER_ACTION=5
export FACE_ACTION_VERIFY_ENABLED=true
# 高风险可选：提高 1:1 阈值
export FACE_MATCH_THRESHOLD=0.36
```

本机 AxonData 评测：0.85 阈值下 APCER 24.17%、BPCER 17.24%；0.91 阈值下 APCER 16.67%、BPCER 17.24%。0.91 在不增加真人拒绝率的前提下降低攻击放行率。

说明：代码默认值仍保持与 .NET 基线一致的 0.85，便于语义对账；生产高风险环境通过上述环境变量切到 0.91。该档位仍不构成商业 PAD 合规结论。

### Lab 测试接口

送测/预测试时启用（`LAB_MODE` 在 production 会拒绝启动）：

```bash
export APP_ENV=staging
export LAB_MODE=true
```

- `GET /version`：返回构建版本、模型 SHA-256、配置 SHA-256、阈值与动作参数；
- `POST /v1/liveness-checks`：返回 `pad_score`、`pad_threshold`、`pad_decision`、`frame_scores`、`attack_frames`、`action_checked`、`action_passed`、`model_version`、`config_version`、`device_id` 等冻结证据；
- 可选表单字段 `verify_actions=true`：同时校验 challenge 的动作序列，完整复现生产 enroll/verify 的决策链路。

该接口仅用于测试环境；生产不注册。详细字段见 [识别质量评测](docs/功能/05-识别质量评测.md) §7.3。

## 验证结果（本机，2026-11-15）

| 项 | 结果 |
| --- | --- |
| 单元测试 | `go test ./internal/...` 全部通过 |
| 纯 Go 构建 | `CGO_ENABLED=0 go build ./cmd/api ./cmd/worker` 成功；产物未链接 `libonnxruntime` |
| YuNet vs ONNX Runtime | 12 个输出最大绝对误差 4.8e-06 |
| SFace vs ONNX Runtime | 128 维输出最大绝对误差 6.1e-06 |
| MiniFASNetV2 vs ONNX Runtime | 3 类 logits 最大绝对误差 9.5e-07 |
| ONNX 规范用例 | PRelu 广播、Dropout 推理/mask 与 ONNX Runtime 误差 0；`training_mode=true` 显式拒绝；initializer 默认值与覆盖语义通过 |
| LFW 1:1 公开基线 | 200 人、605 张；EER 2.996%，AUC 0.988190；0.35 阈值 FAR 0.1396%、FRR 4.1721% |
| AxonData PAD 基线 | 29 段真人、120 段攻击；0.91 阈值 APCER 16.67%、BPCER 17.24% |
| HTTP 端到端 | 录入 200、核验 200、score 0.6118、结果令牌签发、主体查询 active、删除后清理 profile 与索引 |
| 与 .NET 基线对账 | 同一组图片 .NET score 0.62328，Go 0.6118，差 0.0115；动作几何 yaw/mouth 差 ≤ 0.004 |

未执行项：Worker 定时任务、S3 实际对象删除、阈值校准、1:N 与 OCR。

## 文档

- [文档索引](docs/README.md)
- [需求清单](docs/需求/0001/0001-Go版人脸核验服务-需求清单.md)
- [方案评估](docs/需求/0001/0001-Go版人脸核验服务-方案评估.md)
- [实施方案](docs/需求/0001/0001-Go版人脸核验服务-实施方案.md)
- [方案验收](docs/需求/0001/0001-Go版人脸核验服务-方案验收.md)
- [前端对接文档](docs/需求/0001/0001-Go版人脸核验服务-前端对接文档.md)

## 安全边界

- 租户身份只从 JWT 获取；所有查询带 `tenant_id`。
- 人脸特征使用租户级 AES-GCM 加密，密钥由主密钥包裹。
- 日志不输出人脸图、特征、证件号、JWT 原文。
- `/dev/*` 仅在 `APP_ENV=development` 且 `AUTH_ENABLE_DEV_TOKEN=true` 时注册；生产启用即启动失败。
