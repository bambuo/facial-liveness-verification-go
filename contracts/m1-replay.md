# M1 迁移重放对账

- 日期：2026-11-18
- 环境：`postgres:16`（Docker，临时库 `idswyft`）
- 输入：idswyft `supabase/migrations` 67 个文件
- 目的：验证静态抽取的 31 表/列数与真实重放一致，并定位迁移顺序与 Supabase 专有依赖

## 1. 重放方式

1. 启动临时 Postgres 16；
2. 建立 Supabase 依赖桩：`auth.uid()`、`auth.role()`、`storage.buckets`、`storage.objects`、`storage.foldername/filename/extension`、`storage.protect_delete()`、`service_role` 角色；
3. **按版本序**（`sort -V`）执行全部迁移，不因单条失败中断；
4. 统计 `information_schema` 与 `pg_catalog`，与 `contracts/schema.json` 比对。

命令（复现）：

```bash
docker run -d --name idswyft-m1-pg \
  -e POSTGRES_PASSWORD=... -e POSTGRES_DB=idswyft -p 15432:5432 \
  -v <idswyft>/supabase/migrations:/migrations:ro postgres:16
# 建桩后：
for f in $(ls /migrations/*.sql | sort -V); do psql -v ON_ERROR_STOP=0 -f "$f"; done
```

## 2. 结果

| 对象 | 静态抽取 | 重放结果 | 结论 |
| --- | --- | --- | --- |
| 表 | 31 | 31 | 一致 |
| 列 | 329 | 329 | 一致 |
| SQL 函数 | 10 | 10 | 一致 |
| 触发器 | 6 | 6 | 一致 |
| 显式索引 | 77 | 117（含主键/唯一约束隐式索引） | 口径差异，非缺失 |
| RLS 策略 | 53（静态解析） | 27（重放） | Supabase-only 迁移跳过；C5 已定应用层重写 |

逐表列数比对：**0 差异**。

## 3. 重放中暴露的问题

### 3.1 迁移顺序必须按版本序

`20260629_add_operator_email_to_api_keys.sql` 依赖 `58_add_service_keys.sql` 的 `is_service` 列。字典序（`ls | sort`）会把 `20260629` 排在 `58` 之前，导致迁移失败；`sort -V` 后通过。**M1 迁移器必须使用版本序**。

### 3.2 静态解析两处缺陷（已修复）

1. **多列 `ALTER TABLE ADD COLUMN`**：原解析器只取第一个 `ADD COLUMN`，漏掉同一语句中的后续列。修复后 `api_keys` +3、`api_activity_logs` +1、`developers` +10、`verification_requests` +6。
2. **注释混入列定义**：`expiry_alerts` 的 `webhook_sent` 前有行内注释，原解析器把注释当作列定义前缀而丢弃该列。修复后 `expiry_alerts` +1。

静态列数从 307 修正为 **329**。

### 3.3 预期失败与跳过

| 迁移 | 错误 | 处理 |
| --- | --- | --- |
| `53_create_branding_bucket.sql` | `SUPABASE_ONLY_MIGRATION_SKIPPED` | 预期跳过（Storage bucket） |
| `57_add_rls_policies_all_tables.sql` | `public.admin_users` 不存在 / 角色 `authenticated` 不存在 | Supabase 专有 RLS；C5 应用层重写 |
| `59_enable_rls_migrations.sql` | `public._migrations` 不存在 | Supabase 专有；跳过 |
| `60_create_avatars_bucket.sql` | `SUPABASE_ONLY_MIGRATION_SKIPPED` | 预期跳过（Storage bucket） |
| `34_enable_rls_all_tables.sql` | `webhook_deliveries` 不存在 | RLS 开启顺序依赖；C5 不保留 RLS |

以上失败均不影响表/列结构；在 C4（仅 S3）与 C5（应用层重写）下不需要复刻 Storage bucket 与 RLS。

## 4. 结论

- 表与列契约**已通过重放验证**：31 表 / 329 列，逐表 0 差异。
- 函数与触发器数量一致；实现语义仍需事务级测试（C5）。
- 索引口径需在 M1 DDL 中区分“显式 CREATE INDEX”与“主键/唯一约束隐式索引”。
- RLS 与 Storage bucket 策略**不进入 Go 侧 DDL**，由应用层与 S3 策略替代。

## 5. 后续

1. 用 `schema.json` 生成 M1 基础 DDL（31 表 + 显式索引 + 外键/唯一约束）。
2. 为 10 个 SQL 函数写 Go 等价实现与事务测试（OTP 原子校验、状态检查、证件关联）。
3. 将本报告的命令纳入 CI 的“契约漂移检查”。
