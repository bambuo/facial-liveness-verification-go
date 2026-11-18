# 契约冻结输入（M0）

本目录是 idswyft 重写的 **M0 契约抽取产物**，不是运行时配置。目标是把“契约级 1:1”固化为可复算、可评审、可进 CI 的清单。

## 产物

| 文件 | 内容 | 数量校验 |
| --- | --- | --- |
| `endpoints.json` / `endpoints.md` | 154 条路由注册、挂载前缀、中间件、鉴权推断、源码位置 | 注册 154 / 可达路径 158 |
| `fields.json` / `fields.md` | 16 个 Zod schema（86 字段）、73 条接口的请求字段、状态/拒绝原因词表 | 16 schema / 2 词表冲突 |
| `responses.json` / `responses.md` | 158/158 接口的响应证据、317 个 `res.json`、12 类状态码、嵌套字段、`buildVerificationResponse` 字段 | 158 覆盖 / 44 含嵌套 |
| `fixtures.json` | 158 条可达路径的请求/响应夹具（鉴权、请求字段、状态码、嵌套响应、实现标记） | 158 |
| `errors-i18n.md` | en / zh-CN 142 键与源码 `messageKey` 引用对账 | 142/142，1 个未定义 |
| `auth.json` / `auth-review.md` / `auth-decisions.md` | 126 中间件 / 15 内联 / 17 公开候选取证与签字 | 158 覆盖，17 签字 |
| `coverage.md` | 契约与实现覆盖（fixtures、后端/引擎实现率） | 后端 0/158，引擎 2/6 |
| `schema.json` | idswyft 31 表、329 列、10 个 SQL 函数、6 个触发器、53 条 RLS 策略、77 个索引 | 31 表通过 |
| `go-schema.json` | Go 基座 9 表、84 列、17 个索引 | 9 表 |
| `schema-diff.md` | idswyft vs Go 基座差异、需新增表、函数/触发器/RLS 清单 | 需新增 22 表 |
| `ocr-spike.md` | PP-OCRv5 纯 Go 移植、中英黄金集、方向/并发/API 接入 | 已移植并评测 |
| `engine-fields.md` | 6 个引擎端点的请求/结果字段与 Zod 映射 | 6 端点 |
| `m1-replay.md` | 67 个迁移在 Postgres 16 的版本序重放与逐表列数对账 | 31 表 / 329 列一致 |
| `p1-status.md` | P1 核心 15 接口、引擎 6/6、鉴权与集成基线 | 32/158 契约路由 |
| `conflicts.md` | 8 个硬冲突（C1–C8），2026-11-18 已全部冻结，含 2 项批准偏差（D1 状态词表、D2 仅 S3） | 已冻结 |

## 来源与口径

- idswyft：`83dd51bdfb6dec608856b41c110f54f5a8bf3efe`（2026-11-17）
- Go 基座：`badcc5c`
- 路由：TypeScript `5.9.3` AST，只统计 `router.<get|post|put|patch|delete>` 注册；`credentials` 模块挂载两次，展开为 8 条路径；cloud-only 动态挂载 3 个不计入 154。
- 数据库：静态解析 `supabase/migrations`（67 个文件）与 `docs/需求/0001/sql`；注释、单引号字符串、美元引用块掩码后按语句切分。
- 鉴权：仅按中间件名推断；55 条 `none-or-inline` 需在 M0 评审中逐条确认（见 `endpoints.md` 的“待复核”列）。

## 复算

```bash
cd contracts/tools
npm ci
node extract-endpoints.mjs --out ..
node extract-fields.mjs --out ..
node extract-responses.mjs --out ..
node extract-auth.mjs --out ..
node extract-schema.mjs --out ..
```

两个数量校验脚本失败时以非 0 退出：接口注册数必须等于 154，表数必须等于 31。

## 已知缺口（M1 门禁）

1. **M1 重放已完成**：31 表 / 329 列 / 10 函数 / 6 触发器一致；RLS 与 Storage 策略按 C4/C5 不复刻（`m1-replay.md`）。
2. **响应字段**：已从 `res.json` 与 `buildVerificationResponse` 抽取（158/158 接口）；嵌套结构与条件展开仍需 M1 细化。
3. **鉴权**：`endpoints.md` 仍有 32 条待复核；`auth-review.md` 已给出 15 条内联证据、17 条 public-candidate（需签字）。
4. **状态词表冲突**：对外 `VerificationStatus`（12 值）、内部 `VerificationState`（12 值）、`backend/src/types/index.ts`（5 值）并存，C8 已决定统一 12 值。
5. **storage/auth 内置对象**：策略/触发器引用 `storage.objects`、`auth.uid()`，C4 已决定仅 S3。
6. **迁移顺序**：存在重复数字前缀（两个 04），当前按文件名自然序应用；M1 需与 idswyft 迁移器顺序一致。
7. **同名表为 0**：Go 9 表与 idswyft 31 表是两套主体模型，C1 已决定新库独立。
8. **OCR**：中英模型、精确/分桶、自动方向、字段规则 v1（MRZ/US/UK/CN）、黄金集与 `/extract/ocr`、`/extract/front` 已完成；真实 20–50 张证件样本、条码/篡改/深伪与单机并发扩容仍待（`ocr-spike.md`、`eval/ocr/`）。

## 冻结门禁

- [ ] 154 条接口逐条确认鉴权与错误码（`endpoints.md`）
- [ ] 31 表 + 10 函数 + 6 触发器 + 53 策略按 C5 应用层重写方案逐条落地（`schema-diff.md`）
- [x] C1–C8 全部给出决策（`conflicts.md` 冻结记录）
- [x] Zod + 请求字段契约完成（`fields.json`）
- [x] PaddleOCR 纯 Go 移植、中英黄金集与 API 接入完成（`ocr-spike.md`、`eval/ocr/`）
- [ ] 响应字段契约
- [x] M1 空库重放对账通过（表/列/函数/触发器；RLS 按 C5 不复刻）
