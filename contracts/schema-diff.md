# idswyft vs Go 基座 Schema 差异清单（M0 契约冻结输入）

- idswyft 仓库：`/Users/johana/Codes/github/idswyft-community-1.12.39`
- idswyft 提交：`83dd51bdfb6dec608856b41c110f54f5a8bf3efe`（2026-11-17T00:25:31+07:00）
- Go 基座仓库：`/Users/johana/Desktop/facial-liveness-verification-go`
- Go 基座提交：`97c52771462dcf87d09e0f3ace39ab21156838f8`
- 生成时间：2026-11-18T10:08:50.692Z
- 解析方式：静态 SQL 解析（注释/字符串/美元引用块掩码后按语句切分）

## 汇总

| 指标 | idswyft | Go 基座 | 差异 |
| --- | --- | --- | --- |
| 表 | 31 | 9 | 需新增 31，Go 独有 9 |
| 列 | 329 | 84 | - |
| SQL 函数 | 10 | 0 | 全部需实现 |
| 触发器 | 6 | 0 | 全部需实现 |
| RLS 策略 | 53 | 0 | 需决策（保留 SQL 或应用层） |
| 索引 | 77 | - | 随表迁移 |
| 枚举 | 0 | 0 | 需实现 |
| 扩展 | 0 | 0 | - |
| DO $$ 动态块 | 16 | 0 | 需人工复核 |

## 表差异

### idswyft 表（31）

| 表 | 列 | 索引 | 策略 | 触发器 | 来源迁移 |
| --- | --- | --- | --- | --- | --- |
| aml_screenings | 10 | 2 | 2 | 0 | 17_add_aml_screening.sql |
| api_activity_logs | 13 | 2 | 2 | 0 | 28_formalize_api_activity_logs.sql |
| api_keys | 17 | 3 | 2 | 0 | 01_initial_schema.sql |
| batch_items | 8 | 2 | 2 | 0 | 19_add_batch_verification.sql |
| batch_jobs | 9 | 2 | 2 | 0 | 19_add_batch_verification.sql |
| compliance_rules | 6 | 1 | 2 | 0 | 48_add_compliance_engine.sql |
| compliance_rulesets | 8 | 2 | 2 | 0 | 48_add_compliance_engine.sql |
| dedup_fingerprints | 6 | 2 | 2 | 0 | 43_add_duplicate_detection.sql |
| developer_otp_codes | 7 | 1 | 1 | 0 | 25_add_developer_otp_and_github.sql |
| developer_otp_rate_limits | 4 | 0 | 1 | 0 | 25_add_developer_otp_and_github.sql |
| developers | 32 | 3 | 2 | 1 | 01_initial_schema.sql |
| documents | 17 | 5 | 0 | 2 | 01_initial_schema.sql |
| expiry_alerts | 9 | 3 | 2 | 0 | 21_add_monitoring.sql |
| idempotency_keys | 6 | 1 | 2 | 0 | 09_add_idempotency_keys.sql |
| identity_vault | 11 | 4 | 2 | 0 | 46_add_identity_vault.sql |
| mobile_handoff_sessions | 10 | 3 | 2 | 0 | 12_add_mobile_handoff_sessions.sql |
| phone_otp_codes | 8 | 1 | 1 | 0 | 37_add_phone_otp.sql |
| phone_otp_rate_limits | 4 | 0 | 1 | 0 | 37_add_phone_otp.sql |
| provider_metrics | 9 | 2 | 1 | 0 | 11_add_provider_metrics.sql |
| rate_limits | 7 | 3 | 1 | 0 | 07_add_rate_limits_table.sql |
| reverification_schedules | 10 | 4 | 2 | 0 | 21_add_monitoring.sql |
| selfies | 9 | 1 | 0 | 0 | 01_initial_schema.sql |
| users | 7 | 1 | 1 | 1 | 01_initial_schema.sql |
| vault_share_links | 10 | 1 | 2 | 0 | 46_add_identity_vault.sql |
| verifiable_credentials | 10 | 2 | 2 | 0 | 44_add_verifiable_credentials.sql |
| verification_contexts | 3 | 1 | 2 | 1 | 13_add_verification_contexts.sql |
| verification_requests | 45 | 15 | 0 | 1 | 01_initial_schema.sql |
| verification_reviewers | 8 | 3 | 2 | 0 | 33_verification_reviewers.sql |
| verification_risk_scores | 6 | 3 | 2 | 0 | 18_add_risk_scores.sql |
| webhook_deliveries | 11 | 3 | 0 | 0 | 20260319_add_webhook_deliveries.sql |
| webhooks | 9 | 1 | 2 | 0 | 01_initial_schema.sql |

### Go 基座表（9）

| 表 | 列 | 主键/约束 |
| --- | --- | --- |
| face_audit_logs | 9 | id |
| face_consents | 8 | id |
| face_documents | 10 | id |
| face_operations | 11 | id |
| face_profiles | 9 | id |
| face_subjects | 7 | id |
| face_tenant_keys | 6 | id |
| face_tenants | 7 | id |
| face_verification_sessions | 17 | id |

### 需新增到 Go 侧的表

- aml_screenings（10 列）
- api_activity_logs（13 列）
- api_keys（17 列）
- batch_items（8 列）
- batch_jobs（9 列）
- compliance_rules（6 列）
- compliance_rulesets（8 列）
- dedup_fingerprints（6 列）
- developer_otp_codes（7 列）
- developer_otp_rate_limits（4 列）
- developers（32 列）
- documents（17 列）
- expiry_alerts（9 列）
- idempotency_keys（6 列）
- identity_vault（11 列）
- mobile_handoff_sessions（10 列）
- phone_otp_codes（8 列）
- phone_otp_rate_limits（4 列）
- provider_metrics（9 列）
- rate_limits（7 列）
- reverification_schedules（10 列）
- selfies（9 列）
- users（7 列）
- vault_share_links（10 列）
- verifiable_credentials（10 列）
- verification_contexts（3 列）
- verification_requests（45 列）
- verification_reviewers（8 列）
- verification_risk_scores（6 列）
- webhook_deliveries（11 列）
- webhooks（9 列）

### Go 独有表

- face_audit_logs（9 列）
- face_consents（8 列）
- face_documents（10 列）
- face_operations（11 列）
- face_profiles（9 列）
- face_subjects（7 列）
- face_tenant_keys（6 列）
- face_tenants（7 列）
- face_verification_sessions（17 列）

## SQL 函数（需在 Go 侧保留或重写）

| 函数 | 参数 | 返回 | 语言 | 来源 |
| --- | --- | --- | --- | --- |
| check_otp_rate_limit | p_email TEXT, p_max_sends INT, p_window_seconds INT | BOOLEAN | plpgsql | 31_community_rpc_functions.sql |
| check_phone_otp_rate_limit | p_vr_id UUID, p_max_sends INT, p_window_seconds INT | BOOLEAN | plpgsql | 37_add_phone_otp.sql |
| cleanup_old_rate_limits | - | void | plpgsql | 07_add_rate_limits_table.sql |
| link_front_and_back_documents | - | TRIGGER | plpgsql | 04_add_back_of_id_support.sql |
| update_back_of_id_status | - | TRIGGER | plpgsql | 04_add_back_of_id_support.sql |
| update_updated_at_column | - | TRIGGER | - | 01_initial_schema.sql |
| update_verification_contexts_updated_at | - | TRIGGER | plpgsql | 13_add_verification_contexts.sql |
| update_verification_with_state_check | p_verification_id UUID, p_expected_states TEXT[], p_new_status TEXT, p_update_data JSONB | BOOLEAN | plpgsql | 31_community_rpc_functions.sql |
| verify_otp_atomic | p_email TEXT, p_code_hash TEXT, p_max_attempts INT | JSONB | plpgsql | 31_community_rpc_functions.sql |
| verify_phone_otp_atomic | p_vr_id UUID, p_code_hash TEXT, p_max_attempts INT | JSONB | plpgsql | 37_add_phone_otp.sql |

## 触发器

| 触发器 | 时机/事件 | 表 | 来源 |
| --- | --- | --- | --- |
| trg_verification_contexts_updated_at | BEFORE UPDATE | verification_contexts | 13_add_verification_contexts.sql |
| trigger_link_documents | AFTER INSERT | documents | 04_add_back_of_id_support.sql |
| trigger_update_back_of_id_status | AFTER INSERT | documents | 04_add_back_of_id_support.sql |
| update_developers_updated_at | BEFORE UPDATE | developers | 01_initial_schema.sql |
| update_users_updated_at | BEFORE UPDATE | users | 01_initial_schema.sql |
| update_verification_requests_updated_at | BEFORE UPDATE | verification_requests | 01_initial_schema.sql |

## RLS 策略

| 策略 | 表 | 命令 | USING | WITH CHECK | 来源 |
| --- | --- | --- | --- | --- | --- |
| developers_own_aml_screenings | public.aml_screenings | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_api_activity_logs | public.api_activity_logs | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_api_keys | public.api_keys | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| developers_own_batch_items | public.batch_items | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_batch_jobs | public.batch_jobs | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_compliance_rules | public.compliance_rules | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_compliance_rulesets | public.compliance_rulesets | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_dedup_fingerprints | public.dedup_fingerprints | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_expiry_alerts | public.expiry_alerts | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_idempotency_keys | public.idempotency_keys | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_identity_vault | public.identity_vault | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_mobile_handoff_sessions | public.mobile_handoff_sessions | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_profile | public.developers | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_reverification_schedules | public.reverification_schedules | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_risk_scores | public.verification_risk_scores | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_vault_share_links | public.vault_share_links | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_verifiable_credentials | public.verifiable_credentials | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_verification_contexts | public.verification_contexts | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_verification_reviewers | public.verification_reviewers | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_webhook_deliveries | public.webhook_deliveries | SELECT | 是 | 否 | 57_add_rls_policies_all_tables.sql |
| developers_own_webhooks | public.webhooks | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_admin_users | public.admin_users | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_aml_screenings | public.aml_screenings | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_api_activity_logs | public.api_activity_logs | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_api_keys | public.api_keys | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_batch_items | public.batch_items | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_batch_jobs | public.batch_jobs | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_compliance_rules | public.compliance_rules | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_compliance_rulesets | public.compliance_rulesets | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_dedup_fingerprints | public.dedup_fingerprints | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_developer_otp_codes | public.developer_otp_codes | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_developer_otp_rate_limits | public.developer_otp_rate_limits | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_developers | public.developers | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_expiry_alerts | public.expiry_alerts | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_idempotency_keys | public.idempotency_keys | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_identity_vault | public.identity_vault | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_migrations | public._migrations | ALL | 是 | 是 | 59_enable_rls_migrations.sql |
| service_role_all_mobile_handoff_sessions | public.mobile_handoff_sessions | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_org_threshold_settings | public.organization_threshold_settings | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_organizations | public.organizations | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_phone_otp_codes | public.phone_otp_codes | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_phone_otp_rate_limits | public.phone_otp_rate_limits | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_provider_metrics | public.provider_metrics | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_rate_limits | public.rate_limits | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_reverification_schedules | public.reverification_schedules | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_users | public.users | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_vault_share_links | public.vault_share_links | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_verifiable_credentials | public.verifiable_credentials | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_verification_contexts | public.verification_contexts | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_verification_reviewers | public.verification_reviewers | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_verification_risk_scores | public.verification_risk_scores | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_webhook_deliveries | public.webhook_deliveries | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |
| service_role_all_webhooks | public.webhooks | ALL | 是 | 是 | 57_add_rls_policies_all_tables.sql |

## 枚举

无 `CREATE TYPE ... AS ENUM`（状态值在应用层常量中定义）。

## 待复核与决策

- 存在 16 个 DO $$ 动态 SQL 块，静态解析无法展开，M1 需用 pg_dump 重放对账。
- 策略/触发器引用 `storage.objects`、`auth.uid()` 等 Supabase 内置对象；Go 侧需决定保留 Supabase Storage 语义还是替换为 S3 + 应用层策略。
- 迁移存在重复数字前缀（如两个 04），当前按文件名自然序应用；M1 重放时需与 idswyft 迁移器顺序一致。
- 同名表为 0：Go 侧 9 表与 idswyft 31 表是两套主体模型，需在 M0 决策“同库并存”或“新库独立 + /v1 降级为内部引擎”。
- 本清单为静态解析结果；M1 门禁为：空库重放 67 个迁移后 `pg_dump --schema-only` 归一化 diff = 0。

## 复算命令

```bash
cd contracts/tools && node extract-schema.mjs
```
