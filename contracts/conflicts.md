# 契约冻结硬冲突（M0 决策单）

本文件记录“契约级 1:1 全量复刻 idswyft”与 Go 基座现状之间必须先拍板的 7 个冲突。每条包含：证据、1:1 要求、候选方案、推荐默认值、影响面、状态。**未决策项不得进入 M1 实现。**

## 已锁定口径

| 决策 | 结论 | 影响 |
| --- | --- | --- |
| 复刻口径 | 契约级 1:1 | 154 接口、31 表、错误码、鉴权矩阵、前端路由逐条对账 |
| 交付范围 | 全量 154 接口 | 不做“核心先行”裁剪 |
| 商用目标 | 先试点、后商用 | 试点门禁与商用门禁（PAD/精度/合规）分开 |
| 前端基线 | 现有 React 页面逐页对齐 | 24 路由 / 22 页截图 diff + 交互脚本 |
| 运行时 | 纯 Go、无 cgo、无 Node sidecar | ML 走纯 Go ONNX；前端构建期可用 Node |

---

## 冻结记录（2026-11-18）

| 编号 | 决策 | 验收方式 |
| --- | --- | --- |
| C1 | 新库独立；31 表落新库；/v1 人脸接口降级为内部引擎 | 空库重放 31 表；/v1 不再对外注册 |
| C2 | 采用 idswyft 幂等/jti/结果令牌语义 | 幂等键落库、重复请求返回首次响应、TTL 清理测试 |
| C3 | 完全采用 idswyft 认证模型（Developer/User/Reviewer/Service Token）；tenant 不作为对外概念 | 鉴权矩阵测试；结果令牌 aud 按 idswyft 规则 |
| C4 | 仅 S3（批准偏差 D2） | S3 provider 冒烟；非 s3 配置 fail-fast；文档标注 local/supabase 不支持 |
| C5 | 全部应用层重写（函数/触发器/RLS 均不保留） | 事务级集成测试：OTP 原子校验、状态检查、证件关联 |
| C6 | 构建期预渲染 MP4，运行时不引入 Node | 前端构建产出视频资产；页面原生 video 播放 |
| C7 | 字段/范围 1:1；SFace 与 face-api 各自独立校准；分数不可跨版本比较 | 校准报告 + FAR/FRR/ROC；字段级契约测试 |
| C8 | status 与 final_result 统一为 12 值（批准偏差 D1） | statusReader/webhook/SDK/前端映射与测试全部改为 12 值 |

### 批准偏差

| 编号 | 偏差 | 影响 | 处理 |
| --- | --- | --- | --- |
| D1 | idswyft 的 `final_result` 为 `verified`/`failed`/`manual_review`；本重写统一使用 12 值状态词表 | SDK、前端、Webhook 消费方需同步改造；与 idswyft 基线响应不一致 | 在契约测试中登记为例外；M1 同步 statusReader、Webhook 载荷与前端映射 |
| D2 | idswyft 支持 `supabase`/`local`/`s3`；本重写只保留 S3 | 本地开发与 Supabase 部署需改用 S3/RustFS；文档需标注 | 非 s3 provider 启动 fail-fast；dev compose 使用 RustFS |

## C1. 双主体模型：31 表 vs 9 表

**证据**：`schema-diff.md` 显示 idswyft 31 表、Go 基座 9 表（`face_tenants`、`face_subjects`、`face_profiles`、`face_documents`、`face_verification_sessions`、`face_audit_logs`、`face_operations`、`face_consents`、`face_tenant_keys`），同名表为 0。

**1:1 要求**：idswyft 的 `users`、`developers`、`api_keys`、`verification_requests`、`verification_contexts`、`documents`、`selfies` 等 31 表成为主契约。

**候选方案**：

| 方案 | 说明 | 代价 |
| --- | --- | --- |
| A. 新库独立（推荐） | 31 表落在新库；/v1 人脸接口降级为内部引擎，不再对外 | 需要一次性数据迁移/重新建档；.NET 兼容只在内部引擎层 |
| B. 同库并存 | 31 表与 9 表同库，`subject` 与 `user` 双写 | 双写一致性、删除/保留语义冲突、审计口径分裂 |
| C. 视图映射 | 9 表用视图映射到 31 表 | 写入语义无法完全映射；DDL 复杂度高 |

**推荐**：A。已定“新库重建 + 人脸重新建档”，与 C1 推荐一致。

**影响面**：数据层、删除/保留、审计、密钥、SDK、前端账户模型。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C2. 幂等 / jti / 结果令牌语义

**证据**：Go 基座镜像 .NET 现状——`jti` 不做一次性、`Idempotency-Key` 接受但不生效、结果令牌返回原始 JWT；idswyft 有 `idempotency_keys` 表（`09_add_idempotency_keys.sql`）与 `idempotencyMiddleware`。

**1:1 要求**：采用 idswyft 语义——幂等键落库、重复请求返回首次响应、TTL 清理任务（`backend/src/server.ts` 幂等清理 cron）。

**候选方案**：

| 方案 | 说明 |
| --- | --- |
| A. 采用 idswyft 语义（推荐） | 1:1 契约；.NET 客户端需改造 |
| B. 保留 .NET 休眠语义 | 与 154 接口契约冲突，需在契约中显式标注差异 |

**推荐**：A。差异必须在 `endpoints.json` 的 `idempotency` 字段与契约测试中体现。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C3. 认证模型映射

**证据**：`endpoints.md` 鉴权分布——`api-key` 16、`api-key|handoff-token` 16、`service-token` 9、`developer-jwt` 14、`dashboard-jwt` 28、`review-principal` 4、`admin-jwt|reviewer-jwt` 2、`admin-jwt|reviewer-jwt + org-admin|platform-admin` 6、`jwt + admin-jwt` 6、`jwt` 2、`none-or-inline` 55（待逐条确认）。

**1:1 要求**：API Key HMAC（`x-api-key`）、Handoff Token、Service Token、Developer JWT、Reviewer JWT、Admin JWT、Session Token、公开端点行为一致；CSRF 仅对 Cookie 认证的变更请求生效。

**待决策**：

| 决策项 | 选项 |
| --- | --- |
| `tenant`（.NET）与 `developer/user`（idswyft）映射 | 单系统常量租户 / 一开发者一租户 / 取消 tenant |
| 55 条 `none-or-inline` | 逐条确认公开或内联校验（M0 评审） |
| 结果令牌 | 原始 JWT / 一次性消费（与 C2 联动） |

**推荐**：保留 idswyft 的 Developer/User 模型；`tenant` 仅在内部引擎层保留常量；55 条在 M0 评审会上逐条勾选。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C4. 存储 provider

**证据**：idswyft 支持 `supabase | local | s3` 三种 provider（`backend/src/config/index.ts`），含信封加密与签名 URL；Go 基座只实现 S3。

**1:1 要求**：三 provider 行为一致（写入、读取、删除、签名 URL TTL、加密元数据、公共资源代理 `/api/public/assets/*`）。

**候选方案**：

| 方案 | 说明 |
| --- | --- |
| A. 全量实现三 provider（推荐） | 契约 1:1；local 用文件系统，s3 用 SDK，supabase 用 Storage API |
| B. 只保留 s3 | 契约降级，需在契约中标注 local/supabase 不支持 |

**推荐**：A，但 `supabase` provider 仅实现 Storage REST 兼容层，不引入 supabase-js。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C5. SQL 函数 / 触发器 / RLS

**证据**：`schema.json` 抽取到 10 个 SQL 函数、6 个触发器、53 条 RLS 策略、16 个 `DO $$` 动态块（`34_enable_rls_all_tables.sql`、`39_admin_restructure.sql`、`57_add_rls_policies_all_tables.sql` 等）。

**1:1 要求**：OTP 原子校验（`verify_otp_atomic`）、速率限制（`check_otp_rate_limit`）、前后证件关联（`link_front_and_back_documents`）、状态检查（`update_verification_with_state_check`）等语义不能漂移。

**候选方案**：

| 方案 | 说明 |
| --- | --- |
| A. 保留 SQL 函数/触发器（推荐） | pgx 直接调用；RLS 仅在“同库并存”时保留 |
| B. 应用层重写 | 需事务级测试证明原子性等价 |

**推荐**：A。RLS 按 C1 决策：新库独立时改为应用层租户过滤 + 契约测试；同库并存时保留。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C6. Remotion 演示视频

**证据**：`frontend` 使用 `remotion` + `@remotion/player` 渲染开发者集成演示视频。

**1:1 要求**：页面行为一致，但运行时禁止 Node。

**推荐**：构建期预渲染 MP4，作为静态资产随前端发布；页面用原生 `<video>` 播放。不把 Remotion 运行时打进产物。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C7. ML 分数分布与阈值

**证据**：idswyft 人脸识别为 `@vladmandic/face-api`（SSD MobileNet + 128 维描述子，TFJS WASM）；Go 基座为 YuNet + SFace（512 维）+ MiniFASNetV2，纯 Go ONNX。PAD 0.91 阈值 APCER 16.67% / BPCER 17.24%。OCR spike 已证明 PP-OCRv5 mobile det/rec 可在纯 Go 下运行（`ocr-spike.md`），但检测后处理与识别宽度仍需生产化。

**1:1 要求**：字段名、范围、状态枚举一致；**分数值不可能一致**。

**决策**：

| 决策项 | 选项 |
| --- | --- |
| 阈值 | 重新校准（推荐）/ 沿用 idswyft 阈值 |
| 旧档案 | 不可复用，重新建档（已定） |
| deepfake/篡改模型 | 移植 idswyft 的 `deepfake-detector.onnx` / Sharp 算法，或声明差异 |
| OCR 质量门禁 | 字段级准确率 + CER/WER 黄金集（推荐）/ 仅要求链路跑通 |

**推荐**：字段级 1:1 + 独立校准集；OCR 纯 Go 路线 GO，但生产前必须补后处理、多语言与评测；在契约中标注“分数不可跨版本比较”；商用门禁另立（FAR/FRR/ROC + APCER/BPCER + ISO/IEC 30107-3 如需要）。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## C8. 状态词表与响应字段口径

**证据**：`fields.json` 检测到三套状态口径：

| 来源 | 值 |
| --- | --- |
| `shared/src/verification/models/schemas.ts` `VerificationStatus` | 12 值：`AWAITING_FRONT` … `COMPLETE`、`HARD_REJECTED` |
| `backend/src/services/verificationStateMachine.ts` `VerificationState` | 12 值：`pending`、`document_uploaded` … `verified`、`failed`、`manual_review` |
| `backend/src/types/index.ts` `VerificationStatus` | 5 值：`pending`、`processing`、`verified`、`failed`、`manual_review` |

`statusReader.ts` 对外返回 `status`（shared 12 值）与 `final_result`（`verified`/`failed`/`manual_review`）。

**1:1 要求**：每个响应字段只能绑定一套词表，SDK/前端/测试按同一口径。

**候选方案**：

| 方案 | 说明 |
| --- | --- |
| A. 对外 12 值 + `final_result` 3 值（推荐） | 与 idswyft 当前 API 行为一致；内部状态机独立 |
| B. 对外 5 值 | 简化前端，但与 1:1 契约冲突 |
| C. 统一 12 值 | 需要改造 `final_result` 与前端映射 |

**推荐**：A。文档与 SDK 明确 `status` 与 `final_result` 的区别。

**影响面**：前端状态展示、SDK 类型、Webhook 载荷、契约测试、文档。

**状态**：已冻结（2026-11-18，最终决策见冻结记录）。

---

## 决策记录模板

```text
## Cx. 标题
- 决策：
- 依据：
- 影响面：
- 验收方式：
- 决策人 / 日期：
```

## 冻结门禁

- [x] C1–C8 全部给出决策与验收方式（见冻结记录）
- [ ] 32 条待复核鉴权：`auth-review.md` 已给 15 条内联证据；17 条 public-candidate 待签字
- [x] 请求字段契约（Zod 16 schema / 86 字段；73 条接口）完成
- [x] 响应字段契约（158/158 接口、317 个 `res.json`）完成
- [x] PaddleOCR 纯 Go 移植完成（det/rec/行分组；评测待 M1）
- [x] M1 空库重放对账通过（`m1-replay.md`：31 表 / 329 列 / 10 函数 / 6 触发器一致）
