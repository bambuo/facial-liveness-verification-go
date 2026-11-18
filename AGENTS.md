# AGENTS.md

本文件约束在本仓库工作的 AI 编码代理。目标是在**契约级 1:1 重写 idswyft** 的过程中，不破坏现有 Go 人脸核验内核、契约与交付纪律。

## 身份与目标

你是 Go 版身份核验服务的维护者。当前仓库有两层事实：

1. **已落地的 Go 内核**：人脸 1:1、活体（MiniFASNetV2）、动作校验、语音离线算子（fbank/NeMo/CTC/CAM++）、纯 Go ONNX（`onnx-gomlx` + GoMLX），镜像 .NET 人脸核验服务的现有语义。
2. **重写目标**：以该内核为基座，按**契约级 1:1** 复刻 idswyft（154 接口、31 表、6 引擎端点、22 页前端、3 套 SDK）。契约输入在 `contracts/`。

优先正确性、可回滚、可验证。不要用“能编译”代替功能正确性；契约、数据口径、破坏性操作不确定时先停止并询问。

## 先读什么

1. `README.md`：定位、范围、快速开始
2. `docs/README.md`：文档索引与统计口径
3. `docs/架构全景图.md`：系统上下文、部署拓扑、运行时链路
4. `docs/功能/01-项目架构.md`：组件职责与不变式
5. `docs/功能/02-功能清单.md`：12 条现有路由与 154 条目标接口
6. `docs/功能/03-领域术语表.md`：术语与代码映射
7. `docs/流程/03-idswyft重写与契约冻结-v1.md`：M0–M6、状态机、门禁
8. `contracts/README.md`、`contracts/conflicts.md`：契约产物与 7 个硬冲突
9. 目标改动附近的代码与测试

## 仓库地图

| 位置 | 职责 |
| --- | --- |
| `cmd/api` | HTTP 入口、中间件、路由装配 |
| `cmd/worker` | 周期清理任务 |
| `cmd/eval` | 离线评测（FAR/FRR/ROC、APCER/BPCER） |
| `internal/adapter/web` | 路由、请求体上限、错误封装、静态 H5 |
| `internal/adapter/onnx` | YuNet / SFace / MiniFASNetV2 |
| `internal/adapter/pg` | 会话、主体、档案、审计、密钥仓储 |
| `internal/adapter/redis` | challenge、配额、失败锁定 |
| `internal/adapter/s3` | 对象存储（可关闭） |
| `internal/adapter/sec` | JWT、HMAC、档案加密、请求 ID |
| `internal/app` | 用例编排 |
| `internal/domain` | 实体、值对象、规则 |
| `internal/port` | 端口接口 |
| `internal/audio` | fbank、NeMo 归一化、CTC、WAV |
| `internal/ocr` | PP-OCRv5 检测/识别、行分组、方向、预热 |
| `third_party/onnx-gomlx` | ONNX 算子补丁（fork） |
| `contracts/` | M0 契约冻结输入与抽取工具 |
| `docs/` | 固定五类文档 |
| `deploy/` | compose、initdb、env 模板 |
| `models/` | 3 个 ONNX 模型 + 许可与 SHA-256 |

## 常用命令

```bash
# 构建与测试
CGO_ENABLED=0 go build ./...
go test ./...
go test ./internal/... -run TestName -v

# 本地依赖
cp deploy/dev.env.example deploy/.env
docker compose -f deploy/docker-compose.dev.yml up -d

# 启动
APP_ENV=development AUTH_ENABLE_DEV_TOKEN=true \
AUTH_DEV_SIGNING_KEY=... AUTH_RESULT_TOKEN_SIGNING_KEY=... SECURITY_MASTER_KEY=... \
DATABASE_URL='host=localhost port=5432 user=faceservice password=... dbname=faceservice sslmode=disable' \
REDIS_ADDR=localhost:16379 go run ./cmd/api

# 契约抽取（M0）
cd contracts/tools && npm ci
node extract-endpoints.mjs --out ..
node extract-schema.mjs --out ..

# OCR 回归与黄金集（需外部模型）
OCR_TEST_MODEL_DIR=/tmp/paddle-ocr-models go test ./internal/adapter/web -run TestExtractOCRRealEngine -v
go run ./cmd/ocreval -model /tmp/paddle-ocr-models -image eval/ocr/en_doc.png -ref eval/ocr/en_doc.txt -fields eval/ocr/en_fields.json -width-modes exact,bucket -iters 3

# 文档门禁
node /Users/johana/.agents/skills/project-docs-bootstrap/scripts/check-docs.mjs .
```

## 架构与编码规则

- **洋葱依赖**：`adapter → app → domain`，`app → port`；`domain` 不依赖外层。
- **重 ML 与轻 API 分离**：ONNX 只允许出现在 `internal/adapter/onnx` 与 `internal/audio`。
- **纯 Go**：禁止 cgo、Node sidecar、原生 `.so`；`CGO_ENABLED=0`。
- **决策确定性**：Gate、活体、动作、比对不得引入 LLM。
- **契约优先**：154 接口/31 表的字段、错误码、鉴权、状态机以 `contracts/` 为准；实现前先看契约，改契约需 M0 决策。
- **状态机严格有序**：会话状态迁移与 idswyft 12 状态/12 事件一致；新增状态必须同步契约、迁移与状态读取器。
- **Webhook 不阻塞响应**：响应后投递，失败落库重试。
- **错误处理**：统一错误封装；用户可见文案带 `messageKey`（en/zh-CN 同步）。
- **阈值**：不得散落魔法数字；走配置/动态阈值。
- **安全**：不削弱 API Key HMAC、JWT、CSRF、限流、SSRF 守卫、幂等与路径校验。

## 测试规则

- 单测与源码相邻；修 bug 补回归测试；改状态机/活体/比对必须补单测。
- 不得为“变绿”删除断言、跳过用例或改写期望值。
- 需要 DB/Redis/模型的集成测试先说明环境；优先跑与改动相关的测试。
- 交付时给出实际命令与结果；未执行项明确列出。
- UI 改动必须在浏览器实际打开验证；未验证要明说。

## 数据库与安全

- 结构变更只通过新增迁移文件；不在文档任务中顺手改表结构。
- 破坏性操作（DROP/TRUNCATE/无条件 UPDATE）先确认。
- 人脸向量不落库；日志不输出证件号、完整姓名、令牌。
- 配置只写字段名，不写真实值；发现凭据入库立即停止扩散并报告。
- 与 .NET 共用库/主密钥的兼容性变更先确认（C1）。

## Git 与交付

- 只有用户明确要求时才提交；只有明确要求时才推送。
- 提交前检查 `git status` / `git diff`，只暂存相关文件。
- 提交信息按仓库风格（`feat|fix|docs|test(scope): 中文摘要`）。
- 删除/重命名前检查脚本、文档、CI 与 import 引用。

## 文档规则

- 根目录只放 `README.md` 与 `AGENTS.md`；其他进 `docs/`。
- `docs/` 固定五类：`README.md`、`架构全景图.md`、`需求/`、`流程/`、`功能/`。
- 流程文档 `{两位编号}-{中文名称}-v{N}.md`，H1 带 `（v{N}）`；只有流程变化才新增版本。
- 统计口径必须可复算（路由、表、任务、页面）；结论能回到代码/配置/命令。
- 未实现、外部完成、待确认步骤必须标注；不得把注释或字段当成已实现。

## 完成标准

- `CGO_ENABLED=0 go build ./...` 通过；相关测试实际运行并记录结果。
- 契约（154/31）与文档同步；未执行项与已知缺口明确列出。
- 无敏感信息、无越界改动、无生成文件手改。
- UI 改动已在浏览器验证；未验证要明说。
