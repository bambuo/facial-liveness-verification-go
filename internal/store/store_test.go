package store

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const testSchema = `
CREATE TABLE IF NOT EXISTS developer_otp_rate_limits (
  email TEXT PRIMARY KEY, send_count INT NOT NULL, window_start TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL);
CREATE TABLE IF NOT EXISTS developer_otp_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), email TEXT NOT NULL, code_hash TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0, used_at TIMESTAMPTZ, expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE IF NOT EXISTS phone_otp_rate_limits (
  verification_request_id UUID PRIMARY KEY, send_count INT NOT NULL, window_start TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL);
CREATE TABLE IF NOT EXISTS phone_otp_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), verification_request_id UUID NOT NULL, phone_number TEXT NOT NULL DEFAULT '',
  code_hash TEXT NOT NULL, attempts INT NOT NULL DEFAULT 0, used_at TIMESTAMPTZ, expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE IF NOT EXISTS rate_limits (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), identifier TEXT NOT NULL, identifier_type TEXT NOT NULL,
  request_count INT NOT NULL, window_start TIMESTAMPTZ, blocked_until TIMESTAMPTZ, created_at TIMESTAMPTZ);
CREATE TABLE IF NOT EXISTS verification_requests (
  id UUID PRIMARY KEY, status VARCHAR(20) NOT NULL, back_of_id_uploaded BOOLEAN,
  face_match_score NUMERIC(3,2), liveness_score NUMERIC(3,2), cross_validation_score NUMERIC(3,2),
  updated_at TIMESTAMPTZ);
CREATE TABLE IF NOT EXISTS documents (
  id UUID PRIMARY KEY, verification_request_id UUID, is_back_of_id BOOLEAN,
  document_type VARCHAR(50) NOT NULL, back_of_id_document_id UUID, created_at TIMESTAMPTZ);
CREATE TABLE IF NOT EXISTS verification_contexts (
  verification_id UUID, context JSONB NOT NULL, updated_at TIMESTAMPTZ NOT NULL);
CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY, updated_at TIMESTAMPTZ);
CREATE TABLE IF NOT EXISTS developers (id UUID PRIMARY KEY, updated_at TIMESTAMPTZ);
`

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL 未设置")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, stmt := range strings.Split(testSchema, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db.WithContext(context.Background())
}

func reset(t *testing.T, db *gorm.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if err := db.Exec("TRUNCATE " + table).Error; err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}

func TestCheckOTPRateLimit(t *testing.T) {
	db := testDB(t)
	reset(t, db, "developer_otp_rate_limits")
	ctx := context.Background()
	for i, want := range []bool{true, true, false} {
		got, err := CheckOTPRateLimit(ctx, db, "a@b.com", 2, 60)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("send %d: got %v want %v", i+1, got, want)
		}
	}
	if err := db.Table("developer_otp_rate_limits").
		Where("email = ?", "a@b.com").
		Update("window_start", time.Now().Add(-2*time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	got, err := CheckOTPRateLimit(ctx, db, "a@b.com", 2, 60)
	if err != nil || !got {
		t.Fatalf("window reset: got %v err %v", got, err)
	}
}

func TestCheckPhoneOTPRateLimit(t *testing.T) {
	db := testDB(t)
	reset(t, db, "phone_otp_rate_limits")
	ctx := context.Background()
	id := uuid.New()
	for i, want := range []bool{true, false} {
		got, err := CheckPhoneOTPRateLimit(ctx, db, id, 1, 60)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("send %d: got %v want %v", i+1, got, want)
		}
	}
}

func TestVerifyOTPAtomicEmail(t *testing.T) {
	db := testDB(t)
	reset(t, db, "developer_otp_codes")
	ctx := context.Background()
	email := "otp@example.com"
	if err := db.Table("developer_otp_codes").Create(map[string]any{
		"email": email, "code_hash": "good", "attempts": 0,
		"expires_at": time.Now().Add(5 * time.Minute),
	}).Error; err != nil {
		t.Fatal(err)
	}
	verify := func(code string) OTPResult {
		t.Helper()
		var res OTPResult
		err := db.Transaction(func(tx *gorm.DB) error {
			var err error
			res, err = VerifyOTPAtomic(ctx, tx, email, code, 2)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	if res := verify("bad"); res.Status != "invalid" || res.AttemptsLeft != 1 {
		t.Fatalf("wrong #1: %+v", res)
	}
	if res := verify("bad"); res.Status != "invalid" || res.AttemptsLeft != 0 {
		t.Fatalf("wrong #2: %+v", res)
	}
	var used sql.NullTime
	if err := db.Table("developer_otp_codes").Select("used_at").Where("email = ?", email).Scan(&used).Error; err != nil {
		t.Fatal(err)
	}
	if !used.Valid {
		t.Fatal("email otp should be marked used after exhausting attempts")
	}
	if res := verify("good"); res.Status != "not_found" {
		t.Fatalf("after used: %+v", res)
	}

	// 并发验证：同一验证码只允许一个成功。
	reset(t, db, "developer_otp_codes")
	if err := db.Table("developer_otp_codes").Create(map[string]any{
		"email": email, "code_hash": "good", "attempts": 0,
		"expires_at": time.Now().Add(5 * time.Minute),
	}).Error; err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan OTPResult, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var res OTPResult
			err := db.Transaction(func(tx *gorm.DB) error {
				var err error
				res, err = VerifyOTPAtomic(ctx, tx, email, "good", 3)
				return err
			})
			if err != nil {
				res = OTPResult{Status: "error"}
			}
			results <- res
		}()
	}
	wg.Wait()
	close(results)
	valid, notFound := 0, 0
	for res := range results {
		switch res.Status {
		case "valid":
			valid++
		case "not_found":
			notFound++
		}
	}
	if valid != 1 || notFound != 1 {
		t.Fatalf("concurrent valid=%d not_found=%d", valid, notFound)
	}
}

func TestVerifyPhoneOTPAtomic(t *testing.T) {
	db := testDB(t)
	reset(t, db, "phone_otp_codes")
	ctx := context.Background()
	vr := uuid.New()
	if err := db.Table("phone_otp_codes").Create(map[string]any{
		"verification_request_id": vr, "code_hash": "good", "attempts": 3,
		"expires_at": time.Now().Add(5 * time.Minute),
	}).Error; err != nil {
		t.Fatal(err)
	}
	verify := func(code string) OTPResult {
		t.Helper()
		var res OTPResult
		err := db.Transaction(func(tx *gorm.DB) error {
			var err error
			res, err = VerifyPhoneOTPAtomic(ctx, tx, vr, code, 3)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	if res := verify("good"); res.Status != "exhausted" {
		t.Fatalf("expected exhausted, got %+v", res)
	}
	var used sql.NullTime
	if err := db.Table("phone_otp_codes").Select("used_at").Where("verification_request_id = ?", vr).Scan(&used).Error; err != nil {
		t.Fatal(err)
	}
	if used.Valid {
		t.Fatal("phone otp exhausted must not mark used")
	}
	if err := db.Table("phone_otp_codes").Where("verification_request_id = ?", vr).Update("attempts", 2).Error; err != nil {
		t.Fatal(err)
	}
	if res := verify("bad"); res.Status != "invalid" || res.AttemptsLeft != 0 {
		t.Fatalf("last wrong attempt: %+v", res)
	}
	used = sql.NullTime{}
	if err := db.Table("phone_otp_codes").Select("used_at").Where("verification_request_id = ?", vr).Scan(&used).Error; err != nil {
		t.Fatal(err)
	}
	if used.Valid {
		t.Fatal("phone otp wrong attempt must not mark used")
	}
}

func TestCleanupOldRateLimits(t *testing.T) {
	db := testDB(t)
	reset(t, db, "rate_limits")
	ctx := context.Background()
	old := time.Now().Add(-49 * time.Hour)
	blockedUntil := time.Now().Add(time.Hour)
	rows := []map[string]any{
		{"identifier": "old", "identifier_type": "ip", "request_count": 1, "window_start": old},
		{"identifier": "blocked", "identifier_type": "ip", "request_count": 1, "window_start": old, "blocked_until": blockedUntil},
		{"identifier": "new", "identifier_type": "ip", "request_count": 1, "window_start": time.Now()},
	}
	if err := db.Table("rate_limits").Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	n, err := CleanupOldRateLimits(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("deleted %d, want 1", n)
	}
	var count int64
	if err := db.Table("rate_limits").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("remaining %d, want 2", count)
	}
}

func TestUpdateVerificationWithStateCheck(t *testing.T) {
	db := testDB(t)
	reset(t, db, "verification_requests")
	ctx := context.Background()
	id := uuid.New()
	if err := db.Table("verification_requests").Create(map[string]any{
		"id": id, "status": "pending", "updated_at": time.Now().Add(-time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}
	face := 0.9
	var ok bool
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		ok, err = UpdateVerificationWithStateCheck(ctx, tx, id, []string{"pending"}, "verified", VerificationUpdate{FaceMatchScore: &face})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected transition to succeed")
	}
	var row struct {
		Status string  `gorm:"column:status"`
		Score  float64 `gorm:"column:face_match_score"`
	}
	if err := db.Table("verification_requests").Select("status", "face_match_score").Where("id = ?", id).Take(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Status != "verified" || row.Score != 0.9 {
		t.Fatalf("status=%s score=%v", row.Status, row.Score)
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		var err error
		ok, err = UpdateVerificationWithStateCheck(ctx, tx, id, []string{"pending"}, "failed", VerificationUpdate{})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected transition to be rejected")
	}
}

func TestTouchUpdatedAtAndContext(t *testing.T) {
	db := testDB(t)
	reset(t, db, "users", "verification_contexts")
	ctx := context.Background()
	id := uuid.New()
	if err := db.Table("users").Create(map[string]any{
		"id": id, "updated_at": time.Now().Add(-time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := TouchUpdatedAt(ctx, db, "users", id); err != nil {
		t.Fatal(err)
	}
	var updatedAt time.Time
	if err := db.Table("users").Select("updated_at").Where("id = ?", id).Scan(&updatedAt).Error; err != nil {
		t.Fatal(err)
	}
	if !updatedAt.After(time.Now().Add(-5 * time.Second)) {
		t.Fatal("users.updated_at not touched")
	}
	if err := TouchUpdatedAt(ctx, db, "documents; DROP TABLE users", id); err == nil {
		t.Fatal("unsafe table name must be rejected")
	}

	vID := uuid.New()
	if err := db.Table("verification_contexts").Create(map[string]any{
		"verification_id": vID, "context": `{}`, "updated_at": time.Now().Add(-time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := TouchVerificationContext(ctx, db, vID); err != nil {
		t.Fatal(err)
	}
	if err := db.Table("verification_contexts").Select("updated_at").Where("verification_id = ?", vID).Scan(&updatedAt).Error; err != nil {
		t.Fatal(err)
	}
	if !updatedAt.After(time.Now().Add(-5 * time.Second)) {
		t.Fatal("verification_contexts.updated_at not touched")
	}
}

func TestAfterDocumentInsert(t *testing.T) {
	db := testDB(t)
	reset(t, db, "verification_requests", "documents")
	ctx := context.Background()
	vr := uuid.New()
	front := uuid.New()
	back := uuid.New()
	if err := db.Table("verification_requests").Create(map[string]any{
		"id": vr, "status": "pending", "updated_at": time.Now(),
	}).Error; err != nil {
		t.Fatal(err)
	}
	docs := []map[string]any{
		{"id": front, "verification_request_id": vr, "is_back_of_id": false, "document_type": "passport", "created_at": time.Now().Add(-time.Minute)},
		{"id": back, "verification_request_id": vr, "is_back_of_id": true, "document_type": "passport", "created_at": time.Now()},
	}
	if err := db.Table("documents").Create(&docs).Error; err != nil {
		t.Fatal(err)
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		return AfterDocumentInsert(ctx, tx, DocumentInput{ID: back, VerificationRequestID: vr, DocumentType: "passport", IsBack: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	var uploaded bool
	if err := db.Table("verification_requests").Select("back_of_id_uploaded").Where("id = ?", vr).Scan(&uploaded).Error; err != nil {
		t.Fatal(err)
	}
	if !uploaded {
		t.Fatal("back_of_id_uploaded not set")
	}
	var backLink, frontLink string
	if err := db.Table("documents").Select("back_of_id_document_id").Where("id = ?", back).Scan(&backLink).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("documents").Select("back_of_id_document_id").Where("id = ?", front).Scan(&frontLink).Error; err != nil {
		t.Fatal(err)
	}
	if backLink != front.String() || frontLink != back.String() {
		t.Fatalf("links not set: back=%v front=%v", backLink, frontLink)
	}
}
