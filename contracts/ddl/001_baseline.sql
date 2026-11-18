-- 由 contracts/tools/gen-ddl.mjs 从 contracts/schema.json 生成。
-- 依据：C1 新库独立；C5 SQL 函数/触发器/RLS 由应用层重写。
-- 注意：外键 ON DELETE 语义来自列定义；未捕获的约束需在 M1 对账。

CREATE TABLE "aml_screenings" (
  "date_of_birth" TEXT,
  "full_name" TEXT NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "lists_checked" TEXT[] DEFAULT '{}',
  "match_found" BOOLEAN NOT NULL DEFAULT false,
  "matches" JSONB DEFAULT '[]',
  "nationality" TEXT,
  "risk_level" VARCHAR(20) NOT NULL DEFAULT 'clear',
  "screened_at" TIMESTAMPTZ DEFAULT NOW(),
  "verification_request_id" UUID
);

CREATE TABLE "api_activity_logs" (
  "api_key_id" UUID,
  "developer_id" UUID,
  "endpoint" TEXT NOT NULL,
  "error_message" TEXT,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "ip_address" TEXT,
  "is_service" BOOLEAN NOT NULL DEFAULT FALSE,
  "method" VARCHAR(10) NOT NULL,
  "response_time_ms" INTEGER NOT NULL DEFAULT 0,
  "service_product" TEXT,
  "status_code" INTEGER NOT NULL,
  "timestamp" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "user_agent" TEXT
);

CREATE TABLE "api_keys" (
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "developer_id" UUID,
  "expires_at" TIMESTAMP WITH TIME ZONE,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "is_active" BOOLEAN DEFAULT TRUE,
  "is_sandbox" BOOLEAN DEFAULT FALSE,
  "is_service" BOOLEAN NOT NULL DEFAULT FALSE,
  "key_hash" TEXT NOT NULL,
  "key_prefix" VARCHAR(20) NOT NULL,
  "last_used_at" TIMESTAMP WITH TIME ZONE,
  "name" VARCHAR(255) NOT NULL,
  "operator_email" TEXT,
  "revoked_at" TIMESTAMP,
  "revoked_reason" TEXT,
  "service_environment" TEXT,
  "service_label" TEXT,
  "service_product" TEXT,
  CONSTRAINT "api_keys_service_fields_consistent" CHECK ( (is_service = TRUE AND service_product IS NOT NULL AND service_environment IS NOT NULL) OR (is_service = FALSE AND service_product IS NULL AND service_environment IS NULL) ),
  CONSTRAINT "api_keys_service_environment_valid" CHECK ( service_environment IS NULL OR service_environment IN ('production', 'staging', 'development') ),
  CONSTRAINT "api_keys_service_product_valid" CHECK ( service_product IS NULL OR service_product IN ('gatepass', 'idswyft-internal', 'kazivio') ),
  CONSTRAINT "api_keys_operator_email_service_only" CHECK (operator_email IS NULL OR is_service = TRUE)
);

CREATE TABLE "batch_items" (
  "batch_id" UUID,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "error" TEXT,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "input_data" JSONB NOT NULL,
  "status" VARCHAR(20) NOT NULL DEFAULT 'pending',
  "user_id" UUID,
  "verification_id" UUID
);

CREATE TABLE "batch_jobs" (
  "completed_at" TIMESTAMPTZ,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "developer_id" UUID,
  "failed_items" INTEGER NOT NULL DEFAULT 0,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "processed_items" INTEGER NOT NULL DEFAULT 0,
  "status" VARCHAR(20) NOT NULL DEFAULT 'pending',
  "succeeded_items" INTEGER NOT NULL DEFAULT 0,
  "total_items" INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "compliance_rules" (
  "action" JSONB NOT NULL,
  "condition" JSONB NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "description" TEXT,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "ruleset_id" UUID NOT NULL
);

CREATE TABLE "compliance_rulesets" (
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "description" TEXT,
  "developer_id" UUID NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "is_active" BOOLEAN NOT NULL DEFAULT true,
  "name" TEXT NOT NULL,
  "priority" INTEGER NOT NULL DEFAULT 100,
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "dedup_fingerprints" (
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "developer_id" UUID NOT NULL,
  "fingerprint_type" TEXT NOT NULL,
  "hash_value" TEXT NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "verification_request_id" UUID NOT NULL
);

CREATE TABLE "developer_otp_codes" (
  "attempts" INT NOT NULL DEFAULT 0,
  "code_hash" TEXT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "email" TEXT NOT NULL,
  "expires_at" TIMESTAMPTZ NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "used_at" TIMESTAMPTZ
);

CREATE TABLE "developer_otp_rate_limits" (
  "email" TEXT PRIMARY KEY,
  "send_count" INT NOT NULL DEFAULT 1,
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "window_start" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "developers" (
  "aml_enabled" BOOLEAN DEFAULT true,
  "avatar_url" TEXT,
  "branding_accent_color" TEXT,
  "branding_company_name" TEXT,
  "branding_logo_url" TEXT,
  "company" VARCHAR(255),
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "duplicate_detection_action" TEXT,
  "duplicate_detection_enabled" BOOLEAN DEFAULT false,
  "email" VARCHAR(255) NOT NULL UNIQUE,
  "github_id" BIGINT UNIQUE,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "is_verified" BOOLEAN DEFAULT FALSE,
  "llm_api_key_encrypted" TEXT,
  "llm_endpoint_url" TEXT,
  "llm_model" TEXT,
  "llm_provider" TEXT,
  "locale" TEXT NOT NULL DEFAULT 'en',
  "name" VARCHAR(255) NOT NULL,
  "page_builder_config" JSONB DEFAULT NULL,
  "sms_api_key_encrypted" TEXT,
  "sms_api_secret_encrypted" TEXT,
  "sms_phone_number" TEXT,
  "sms_provider" TEXT,
  "status" TEXT NOT NULL,
  "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "vault_auto_store" BOOLEAN DEFAULT false,
  "vault_enabled" BOOLEAN DEFAULT false,
  "vc_enabled" BOOLEAN DEFAULT false,
  "verification_slug" TEXT DEFAULT NULL,
  "voice_auth_enabled" BOOLEAN DEFAULT false,
  "webhook_url" TEXT
);

CREATE TABLE "documents" (
  "back_of_id_document_id" UUID,
  "barcode_data" JSONB,
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "cross_validation_results" JSONB,
  "document_type" VARCHAR(50) NOT NULL,
  "file_name" VARCHAR(255) NOT NULL,
  "file_path" TEXT NOT NULL,
  "file_size" INTEGER,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "is_back_of_id" BOOLEAN DEFAULT FALSE,
  "issuing_country" VARCHAR(2),
  "mime_type" VARCHAR(100),
  "ocr_data" JSONB,
  "ocr_extracted" BOOLEAN DEFAULT FALSE,
  "quality_analysis" JSONB,
  "quality_score" DECIMAL(3,2),
  "verification_request_id" UUID
);

CREATE TABLE "expiry_alerts" (
  "alert_type" VARCHAR(20) NOT NULL,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "developer_id" UUID NOT NULL,
  "document_id" UUID,
  "expiry_date" DATE NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "user_id" UUID,
  "verification_request_id" UUID NOT NULL,
  "webhook_sent" BOOLEAN DEFAULT false,
  CONSTRAINT chk_alert_type CHECK (alert_type IN ('90_day', '60_day', '30_day', 'expired'))
);

CREATE TABLE "idempotency_keys" (
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "developer_id" UUID NOT NULL,
  "expires_at" TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '24 hours'),
  "key" TEXT NOT NULL,
  "response_body" JSONB NOT NULL,
  "response_status" INTEGER NOT NULL,
  PRIMARY KEY (key, developer_id)
);

CREATE TABLE "identity_vault" (
  "access_count" INTEGER NOT NULL DEFAULT 0,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "developer_id" UUID NOT NULL,
  "encrypted_data" TEXT NOT NULL,
  "encryption_key_id" TEXT NOT NULL DEFAULT 'v1',
  "expires_at" TIMESTAMPTZ,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "last_accessed_at" TIMESTAMPTZ,
  "status" TEXT NOT NULL DEFAULT 'active',
  "vault_token" TEXT NOT NULL UNIQUE,
  "verification_id" UUID NOT NULL
);

CREATE TABLE "mobile_handoff_sessions" (
  "api_key_id" UUID NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "expires_at" TIMESTAMPTZ NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "result" JSONB,
  "source" TEXT NOT NULL DEFAULT 'api',
  "status" TEXT NOT NULL,
  "token" CHAR(64) NOT NULL UNIQUE,
  "user_id" TEXT NOT NULL,
  "verification_id" TEXT
);

CREATE TABLE "phone_otp_codes" (
  "attempts" INT NOT NULL DEFAULT 0,
  "code_hash" TEXT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "expires_at" TIMESTAMPTZ NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "phone_number" TEXT NOT NULL,
  "used_at" TIMESTAMPTZ,
  "verification_request_id" UUID NOT NULL
);

CREATE TABLE "phone_otp_rate_limits" (
  "send_count" INT NOT NULL DEFAULT 1,
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "verification_request_id" UUID PRIMARY KEY,
  "window_start" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "provider_metrics" (
  "confidence_score" DECIMAL(4,3),
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "error_type" TEXT,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "latency_ms" INTEGER,
  "provider_name" TEXT NOT NULL,
  "provider_type" TEXT NOT NULL,
  "success" BOOLEAN NOT NULL,
  "verification_id" UUID
);

CREATE TABLE "rate_limits" (
  "blocked_until" TIMESTAMP WITH TIME ZONE,
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "identifier" TEXT NOT NULL,
  "identifier_type" TEXT NOT NULL,
  "request_count" INTEGER NOT NULL DEFAULT 0,
  "window_start" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE "reverification_schedules" (
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "developer_id" UUID NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "interval_days" INTEGER NOT NULL DEFAULT 365,
  "last_verification_at" TIMESTAMPTZ,
  "next_verification_at" TIMESTAMPTZ NOT NULL,
  "status" VARCHAR(20) NOT NULL DEFAULT 'active',
  "updated_at" TIMESTAMPTZ DEFAULT NOW(),
  "user_id" UUID NOT NULL,
  "verification_request_id" UUID,
  CONSTRAINT chk_interval_days CHECK (interval_days >= 30 AND interval_days <= 730)
);

CREATE TABLE "selfies" (
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "enrollment_audio_path" TEXT,
  "face_detected" BOOLEAN DEFAULT FALSE,
  "file_name" VARCHAR(255) NOT NULL,
  "file_path" TEXT NOT NULL,
  "file_size" INTEGER,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "liveness_score" DECIMAL(3,2),
  "verification_request_id" UUID
);

CREATE TABLE "users" (
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "email" VARCHAR(255),
  "first_name" VARCHAR(100),
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "last_name" VARCHAR(100),
  "phone" VARCHAR(20),
  "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE "vault_share_links" (
  "access_count" INTEGER NOT NULL DEFAULT 0,
  "allowed_attributes" TEXT[] NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "developer_id" UUID NOT NULL,
  "expires_at" TIMESTAMPTZ NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "max_accesses" INTEGER DEFAULT 10,
  "recipient_label" TEXT,
  "share_token" TEXT NOT NULL UNIQUE,
  "vault_id" UUID NOT NULL
);

CREATE TABLE "verifiable_credentials" (
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "credential_jti" TEXT NOT NULL UNIQUE,
  "credential_jwt" TEXT,
  "developer_id" UUID NOT NULL,
  "expires_at" TIMESTAMPTZ NOT NULL,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "issued_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "revocation_reason" TEXT,
  "revoked_at" TIMESTAMPTZ,
  "verification_request_id" UUID NOT NULL
);

CREATE TABLE "verification_contexts" (
  "context" JSONB NOT NULL DEFAULT '{}',
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "verification_id" UUID PRIMARY KEY
);

CREATE TABLE "verification_requests" (
  "addons" JSONB DEFAULT NULL,
  "address_data" JSONB,
  "address_match_score" DECIMAL(3,2),
  "address_verification_status" VARCHAR(20),
  "age_threshold" SMALLINT,
  "api_key_id" UUID,
  "back_of_id_uploaded" BOOLEAN DEFAULT FALSE,
  "client_ip" TEXT,
  "confidence_score" DECIMAL(3,2),
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "cross_validation_score" DECIMAL(3,2),
  "developer_id" UUID,
  "document_id" UUID,
  "duplicate_flags" JSONB DEFAULT NULL,
  "enhanced_verification_completed" BOOLEAN DEFAULT FALSE,
  "external_verification_id" VARCHAR(255),
  "face_match_score" DECIMAL(3,2),
  "failure_reason" TEXT,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "is_sandbox" BOOLEAN DEFAULT FALSE,
  "issuing_country" VARCHAR(2),
  "live_capture_completed" BOOLEAN DEFAULT FALSE,
  "liveness_score" DECIMAL(3,2),
  "manual_review_reason" TEXT,
  "parent_verification_id" UUID,
  "photo_consistency_score" DECIMAL(3,2),
  "processing_completed_at" TIMESTAMPTZ,
  "retry_count" INTEGER NOT NULL DEFAULT 0,
  "reviewed_at" TIMESTAMPTZ,
  "reviewed_by" TEXT,
  "selfie_id" UUID,
  "session_api_key_id" UUID,
  "session_started_at" TIMESTAMPTZ,
  "session_token_expires_at" TIMESTAMPTZ,
  "session_token_hash" CHAR(64) UNIQUE,
  "source" TEXT NOT NULL DEFAULT 'api',
  "status" VARCHAR(20) NOT NULL DEFAULT 'pending',
  "step_timestamps" JSONB DEFAULT '{}',
  "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "user_id" UUID,
  "verification_mode" VARCHAR(20) DEFAULT 'full',
  "verification_type" VARCHAR(20) DEFAULT 'document',
  "voice_challenge" TEXT,
  "voice_challenge_created_at" TIMESTAMPTZ,
  "voice_match_score" REAL,
  CONSTRAINT "verification_requests_status_check" CHECK (status IN ('pending', 'processing', 'verified', 'failed', 'manual_review')),
  CONSTRAINT "verification_requests_verification_mode_check" CHECK (verification_mode IN ('full', 'document_only', 'identity', 'liveness_only', 'document_refresh', 'age_only')),
  CONSTRAINT "chk_session_token_expiry" CHECK (session_token_hash IS NULL OR session_token_expires_at IS NOT NULL)
);

CREATE TABLE "verification_reviewers" (
  "developer_id" UUID NOT NULL,
  "email" TEXT NOT NULL UNIQUE,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "invited_at" TIMESTAMPTZ DEFAULT NOW(),
  "last_login_at" TIMESTAMPTZ,
  "name" TEXT,
  "role" TEXT NOT NULL DEFAULT 'reviewer',
  "status" TEXT NOT NULL DEFAULT 'invited'
);

CREATE TABLE "verification_risk_scores" (
  "computed_at" TIMESTAMPTZ DEFAULT NOW(),
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "overall_score" INTEGER NOT NULL,
  "risk_factors" JSONB DEFAULT '[]',
  "risk_level" VARCHAR(20) NOT NULL DEFAULT 'low',
  "verification_request_id" UUID UNIQUE
);

CREATE TABLE "webhook_deliveries" (
  "attempts" INTEGER DEFAULT 0,
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "delivered_at" TIMESTAMP WITH TIME ZONE,
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "next_retry_at" TIMESTAMP WITH TIME ZONE,
  "payload" JSONB NOT NULL,
  "response_body" TEXT,
  "response_status" INTEGER,
  "status" VARCHAR(20) NOT NULL DEFAULT 'pending',
  "verification_request_id" UUID NOT NULL,
  "webhook_id" UUID NOT NULL
);

CREATE TABLE "webhooks" (
  "api_key_id" UUID,
  "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  "developer_id" UUID,
  "events" TEXT[],
  "id" UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  "is_active" BOOLEAN DEFAULT TRUE,
  "is_sandbox" BOOLEAN DEFAULT FALSE,
  "secret_key" VARCHAR(255),
  "url" TEXT NOT NULL
);

-- 外键在所有表创建后添加，避免表顺序依赖。
ALTER TABLE "aml_screenings" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "api_activity_logs" ADD FOREIGN KEY ("api_key_id") REFERENCES "api_keys" ON DELETE SET NULL;
ALTER TABLE "api_activity_logs" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "api_keys" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "batch_items" ADD FOREIGN KEY ("batch_id") REFERENCES "batch_jobs" ON DELETE CASCADE;
ALTER TABLE "batch_jobs" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "compliance_rules" ADD FOREIGN KEY ("ruleset_id") REFERENCES "compliance_rulesets" ON DELETE CASCADE;
ALTER TABLE "compliance_rulesets" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "dedup_fingerprints" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "dedup_fingerprints" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "documents" ADD FOREIGN KEY ("back_of_id_document_id") REFERENCES "documents";
ALTER TABLE "documents" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "expiry_alerts" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "expiry_alerts" ADD FOREIGN KEY ("document_id") REFERENCES "documents" ON DELETE SET NULL;
ALTER TABLE "expiry_alerts" ADD FOREIGN KEY ("user_id") REFERENCES "users" ON DELETE SET NULL;
ALTER TABLE "expiry_alerts" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "identity_vault" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "identity_vault" ADD FOREIGN KEY ("verification_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "mobile_handoff_sessions" ADD FOREIGN KEY ("api_key_id") REFERENCES "api_keys";
ALTER TABLE "phone_otp_codes" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "phone_otp_rate_limits" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "provider_metrics" ADD FOREIGN KEY ("verification_id") REFERENCES "verification_requests" ON DELETE SET NULL;
ALTER TABLE "reverification_schedules" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "reverification_schedules" ADD FOREIGN KEY ("user_id") REFERENCES "users" ON DELETE CASCADE;
ALTER TABLE "reverification_schedules" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE SET NULL;
ALTER TABLE "selfies" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "vault_share_links" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "vault_share_links" ADD FOREIGN KEY ("vault_id") REFERENCES "identity_vault" ON DELETE CASCADE;
ALTER TABLE "verifiable_credentials" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "verifiable_credentials" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("api_key_id") REFERENCES "api_keys" ON DELETE SET NULL;
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("document_id") REFERENCES "documents" ON DELETE SET NULL;
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("parent_verification_id") REFERENCES "verification_requests";
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("selfie_id") REFERENCES "selfies" ON DELETE SET NULL;
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("session_api_key_id") REFERENCES "api_keys";
ALTER TABLE "verification_requests" ADD FOREIGN KEY ("user_id") REFERENCES "users" ON DELETE CASCADE;
ALTER TABLE "verification_reviewers" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;
ALTER TABLE "verification_risk_scores" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "webhook_deliveries" ADD FOREIGN KEY ("verification_request_id") REFERENCES "verification_requests" ON DELETE CASCADE;
ALTER TABLE "webhook_deliveries" ADD FOREIGN KEY ("webhook_id") REFERENCES "webhooks" ON DELETE CASCADE;
ALTER TABLE "webhooks" ADD FOREIGN KEY ("api_key_id") REFERENCES "api_keys" ON DELETE SET NULL;
ALTER TABLE "webhooks" ADD FOREIGN KEY ("developer_id") REFERENCES "developers" ON DELETE CASCADE;

CREATE INDEX "api_activity_logs_service_product_idx" ON "api_activity_logs" (service_product, timestamp DESC) WHERE is_service = TRUE;
CREATE INDEX "api_keys_operator_email_idx" ON "api_keys" (operator_email) WHERE operator_email IS NOT NULL;
CREATE INDEX "api_keys_service_lookup_idx" ON "api_keys" (key_hash) WHERE is_service = TRUE AND is_active = TRUE;
CREATE INDEX "idx_aml_screenings_risk" ON "aml_screenings" (risk_level);
CREATE INDEX "idx_aml_screenings_verification" ON "aml_screenings" (verification_request_id);
CREATE INDEX "idx_api_activity_developer_ts" ON "api_activity_logs" (developer_id, timestamp DESC);
CREATE INDEX "idx_api_keys_developer_id" ON "api_keys" (developer_id);
CREATE INDEX "idx_batch_items_batch" ON "batch_items" (batch_id);
CREATE INDEX "idx_batch_items_status" ON "batch_items" (status);
CREATE INDEX "idx_batch_jobs_dev" ON "batch_jobs" (developer_id);
CREATE INDEX "idx_batch_jobs_status" ON "batch_jobs" (status);
CREATE INDEX "idx_compliance_rules_ruleset" ON "compliance_rules" (ruleset_id);
CREATE INDEX "idx_compliance_rulesets_active" ON "compliance_rulesets" (developer_id, is_active) WHERE is_active = true;
CREATE INDEX "idx_compliance_rulesets_dev" ON "compliance_rulesets" (developer_id);
CREATE INDEX "idx_dedup_lookup" ON "dedup_fingerprints" (developer_id, fingerprint_type, hash_value);
CREATE INDEX "idx_dedup_vr_id" ON "dedup_fingerprints" (verification_request_id);
CREATE INDEX "idx_developer_otp_codes_email_expires" ON "developer_otp_codes" (email, expires_at);
CREATE INDEX "idx_developers_email" ON "developers" (email);
CREATE INDEX "idx_developers_github_id" ON "developers" (github_id) WHERE github_id IS NOT NULL;
CREATE UNIQUE INDEX "idx_developers_verification_slug" ON "developers" (verification_slug) WHERE verification_slug IS NOT NULL;
CREATE INDEX "idx_documents_back_of_id_document_id" ON "documents" (back_of_id_document_id);
CREATE INDEX "idx_documents_is_back_of_id" ON "documents" (is_back_of_id);
CREATE INDEX "idx_documents_issuing_country" ON "documents" (issuing_country);
CREATE INDEX "idx_documents_ocr_extracted" ON "documents" (ocr_extracted);
CREATE INDEX "idx_documents_verification_request_id" ON "documents" (verification_request_id);
CREATE INDEX "idx_expiry_alerts_developer" ON "expiry_alerts" (developer_id);
CREATE UNIQUE INDEX "idx_expiry_alerts_unique" ON "expiry_alerts" (verification_request_id, alert_type);
CREATE INDEX "idx_expiry_alerts_verification" ON "expiry_alerts" (verification_request_id);
CREATE INDEX "idx_idempotency_expires" ON "idempotency_keys" (expires_at);
CREATE INDEX "idx_phone_otp_codes_vr_expires" ON "phone_otp_codes" (verification_request_id, expires_at);
CREATE INDEX "idx_provider_metrics_created" ON "provider_metrics" (created_at);
CREATE INDEX "idx_provider_metrics_name" ON "provider_metrics" (provider_name);
CREATE INDEX "idx_rate_limits_blocked_until" ON "rate_limits" (blocked_until) WHERE blocked_until IS NOT NULL;
CREATE INDEX "idx_rate_limits_lookup" ON "rate_limits" (identifier, identifier_type);
CREATE INDEX "idx_rate_limits_window_start" ON "rate_limits" (window_start);
CREATE UNIQUE INDEX "idx_reverification_active_unique" ON "reverification_schedules" (developer_id, user_id) WHERE status = 'active';
CREATE INDEX "idx_reverification_developer" ON "reverification_schedules" (developer_id);
CREATE INDEX "idx_reverification_next" ON "reverification_schedules" (next_verification_at) WHERE status = 'active';
CREATE INDEX "idx_reverification_user" ON "reverification_schedules" (user_id);
CREATE INDEX "idx_risk_scores_computed_at" ON "verification_risk_scores" (computed_at);
CREATE INDEX "idx_risk_scores_level" ON "verification_risk_scores" (risk_level);
CREATE INDEX "idx_risk_scores_verification" ON "verification_risk_scores" (verification_request_id);
CREATE INDEX "idx_selfies_verification_request_id" ON "selfies" (verification_request_id);
CREATE INDEX "idx_share_token" ON "vault_share_links" (share_token);
CREATE INDEX "idx_users_email" ON "users" (email);
CREATE INDEX "idx_vault_developer" ON "identity_vault" (developer_id);
CREATE INDEX "idx_vault_expires" ON "identity_vault" (expires_at) WHERE status = 'active';
CREATE INDEX "idx_vault_token" ON "identity_vault" (vault_token);
CREATE INDEX "idx_vault_verification" ON "identity_vault" (verification_id);
CREATE INDEX "idx_vc_jti" ON "verifiable_credentials" (credential_jti);
CREATE INDEX "idx_vc_verification" ON "verifiable_credentials" (verification_request_id);
CREATE INDEX "idx_verification_requests_document_id" ON "verification_requests" (document_id);
CREATE INDEX "idx_verification_requests_enhanced_verification" ON "verification_requests" (enhanced_verification_completed);
CREATE INDEX "idx_verification_requests_live_capture" ON "verification_requests" (live_capture_completed);
CREATE INDEX "idx_verification_requests_liveness" ON "verification_requests" (liveness_score);
CREATE INDEX "idx_verification_requests_parent" ON "verification_requests" (parent_verification_id) WHERE parent_verification_id IS NOT NULL;
CREATE INDEX "idx_verification_requests_photo_consistency" ON "verification_requests" (photo_consistency_score);
CREATE INDEX "idx_verification_requests_selfie_id" ON "verification_requests" (selfie_id);
CREATE INDEX "idx_verification_requests_source_created" ON "verification_requests" (source, created_at) WHERE source = 'demo';
CREATE INDEX "idx_verification_requests_status" ON "verification_requests" (status);
CREATE INDEX "idx_verification_requests_user_id" ON "verification_requests" (user_id);
CREATE INDEX "idx_verification_reviewers_developer" ON "verification_reviewers" (developer_id);
CREATE INDEX "idx_verification_reviewers_email" ON "verification_reviewers" (email);
CREATE INDEX "idx_verification_reviewers_role" ON "verification_reviewers" (role);
CREATE INDEX "idx_vr_client_ip" ON "verification_requests" (client_ip) WHERE client_ip IS NOT NULL;
CREATE INDEX "idx_vr_developer_created" ON "verification_requests" (developer_id, created_at DESC);
CREATE INDEX "idx_vr_session_token" ON "verification_requests" (session_token_hash) WHERE session_token_hash IS NOT NULL;
CREATE INDEX "idx_vr_user_created" ON "verification_requests" (user_id, created_at DESC);
CREATE INDEX "idx_webhook_deliveries_status" ON "webhook_deliveries" (status);
CREATE INDEX "idx_webhook_deliveries_verification" ON "webhook_deliveries" (verification_request_id);
CREATE INDEX "idx_webhook_deliveries_webhook" ON "webhook_deliveries" (webhook_id);
CREATE INDEX "idx_webhooks_api_key_id" ON "webhooks" (api_key_id);
CREATE INDEX "mobile_handoff_sessions_api_key_id_idx" ON "mobile_handoff_sessions" (api_key_id);
CREATE INDEX "mobile_handoff_sessions_expires_at_idx" ON "mobile_handoff_sessions" (expires_at);
CREATE INDEX "mobile_handoff_sessions_token_idx" ON "mobile_handoff_sessions" (token);
CREATE INDEX "verification_contexts_updated_at_idx" ON "verification_contexts" (updated_at);
CREATE INDEX "verification_requests_api_key_id_idx" ON "verification_requests" (api_key_id) WHERE api_key_id IS NOT NULL;
