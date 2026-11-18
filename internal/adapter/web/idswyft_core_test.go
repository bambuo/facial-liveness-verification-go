package web

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/ocr"
	"facial-liveness-verification-go/internal/port"
	"facial-liveness-verification-go/internal/store"
)

// stubIdswyftStore 是核心接口测试的内存仓储。
type stubIdswyftStore struct {
	developerID  uuid.UUID
	verification *gormmodel.VerificationRequest
	documents    map[uuid.UUID]*gormmodel.Document
	selfies      map[uuid.UUID]*gormmodel.Selfie
	handoff      *port.HandoffAuth
	session      *port.SessionAuth
}

func newStubIdswyftStore() *stubIdswyftStore {
	return &stubIdswyftStore{developerID: uuid.New(), documents: map[uuid.UUID]*gormmodel.Document{}, selfies: map[uuid.UUID]*gormmodel.Selfie{}}
}

func (s *stubIdswyftStore) FindAPIKeyByHash(context.Context, string) (*port.APIKeyAuth, error) {
	return &port.APIKeyAuth{ID: uuid.New(), DeveloperID: s.developerID, DeveloperStatus: "active"}, nil
}
func (s *stubIdswyftStore) FindHandoffByToken(context.Context, string) (*port.HandoffAuth, error) {
	if s.handoff == nil {
		return nil, context.Canceled
	}
	return s.handoff, nil
}
func (s *stubIdswyftStore) FindSessionByToken(context.Context, string) (*port.SessionAuth, error) {
	if s.session == nil {
		return nil, context.Canceled
	}
	return s.session, nil
}
func (s *stubIdswyftStore) TouchAPIKey(context.Context, uuid.UUID) error { return nil }
func (s *stubIdswyftStore) CreateVerification(_ context.Context, v *gormmodel.VerificationRequest) error {
	s.verification = v
	return nil
}
func (s *stubIdswyftStore) GetVerification(_ context.Context, id uuid.UUID) (*gormmodel.VerificationRequest, error) {
	if s.verification == nil || s.verification.ID != id {
		return nil, context.Canceled
	}
	return s.verification, nil
}
func (s *stubIdswyftStore) UpdateVerification(_ context.Context, id uuid.UUID, updates map[string]any) error {
	if s.verification != nil && s.verification.ID == id {
		for key, value := range updates {
			switch key {
			case "status":
				s.verification.Status = value.(string)
			case "document_id":
				if v, ok := value.(uuid.UUID); ok {
					s.verification.DocumentID = &v
				}
			}
		}
	}
	return nil
}
func (s *stubIdswyftStore) CreateDocument(_ context.Context, d *gormmodel.Document) error {
	s.documents[d.ID] = d
	return nil
}
func (s *stubIdswyftStore) GetDocument(_ context.Context, id uuid.UUID) (*gormmodel.Document, error) {
	return s.documents[id], nil
}
func (s *stubIdswyftStore) GetBackDocument(context.Context, uuid.UUID) (*gormmodel.Document, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) CreateSelfie(_ context.Context, selfie *gormmodel.Selfie) error {
	s.selfies[selfie.ID] = selfie
	return nil
}
func (s *stubIdswyftStore) GetSelfie(_ context.Context, id uuid.UUID) (*gormmodel.Selfie, error) {
	return s.selfies[id], nil
}
func (s *stubIdswyftStore) UpsertContext(context.Context, uuid.UUID, []byte) error { return nil }
func (s *stubIdswyftStore) GetContext(context.Context, uuid.UUID) ([]byte, error) {
	return []byte("{}"), nil
}
func (s *stubIdswyftStore) CheckPhoneOTPSend(context.Context, uuid.UUID, int, int) (bool, error) {
	return true, nil
}
func (s *stubIdswyftStore) CreatePhoneOTP(context.Context, uuid.UUID, string, string, time.Duration) error {
	return nil
}
func (s *stubIdswyftStore) VerifyPhoneOTP(context.Context, uuid.UUID, string, int) (store.OTPResult, error) {
	return store.OTPResult{Status: "valid"}, nil
}
func (s *stubIdswyftStore) FindDeveloperByEmail(context.Context, string) (*gormmodel.Developer, error) {
	return &gormmodel.Developer{}, nil
}
func (s *stubIdswyftStore) CheckDeveloperOTPSend(context.Context, string, int, int) (bool, error) {
	return true, nil
}
func (s *stubIdswyftStore) CreateDeveloperOTP(context.Context, string, string, time.Duration) error {
	return nil
}
func (s *stubIdswyftStore) VerifyDeveloperOTP(context.Context, string, string, int) (store.OTPResult, error) {
	return store.OTPResult{Status: "valid"}, nil
}
func (s *stubIdswyftStore) WebhookForDeveloper(context.Context, uuid.UUID) (*port.WebhookTarget, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) CreateWebhookDelivery(context.Context, *gormmodel.WebhookDelivery) error {
	return nil
}
func (s *stubIdswyftStore) UpdateWebhookDelivery(context.Context, uuid.UUID, map[string]any) error {
	return nil
}

// TestVerifyInitializeAndStatus 覆盖 API Key 鉴权、初始化与状态查询。
func TestVerifyInitializeAndStatus(t *testing.T) {
	store := newStubIdswyftStore()
	cfg := testConfig()
	cfg.Security.APIKeySecret = "test-secret"
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: store})
	userID := uuid.New()
	body := []byte(`{"user_id":"` + userID.String() + `","document_type":"passport","issuing_country":"US"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/initialize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("initialize status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["verification_id"] == nil || out["session_token"] == nil {
		t.Fatalf("initialize response=%v", out)
	}
	verificationID := out["verification_id"].(string)

	req = httptest.NewRequest(http.MethodGet, "/api/v2/verify/"+verificationID+"/status", nil)
	req.Header.Set("X-API-Key", "test-key")
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// TestFrontDocumentShape 覆盖证件上传与 OCR 字段响应。
func TestFrontDocumentShape(t *testing.T) {
	store := newStubIdswyftStore()
	cfg := testConfig()
	cfg.Security.APIKeySecret = "test-secret"
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: store, OCR: stubOCR{res: ocr.Result{
		Text: "NAME: JOHN DOE\nDOB: 1990-01-01\nID NO: AB1234567\nEXP: 2030-01-01\n", Confidence: 0.95,
	}}})
	verificationID := uuid.New()
	store.verification = &gormmodel.VerificationRequest{ID: verificationID, DeveloperID: &store.developerID, Status: "pending"}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("document", "front.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(testPNG(t))
	_ = writer.WriteField("document_type", "passport")
	_ = writer.WriteField("issuing_country", "us")
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/"+verificationID.String()+"/front-document", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	out := decodeObject(t, rec)
	assertShape(t, out, map[string]string{
		"success": "bool", "verification_id": "string", "status": "string",
		"document_id": "string", "ocr_data": "object", "requires_back": "bool",
	}, "$")
}

// TestVerifyRequiresCredential 验证核心接口缺少凭证时返回 401。
func TestVerifyRequiresCredential(t *testing.T) {
	cfg := testConfig()
	cfg.Security.APIKeySecret = "test-secret"
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: newStubIdswyftStore()})
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/initialize", bytes.NewReader([]byte(`{"user_id":"`+uuid.New().String()+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// TestInternalSessionServiceToken 验证 service token 保护内部 session 端点。
func TestInternalSessionServiceToken(t *testing.T) {
	store := newStubIdswyftStore()
	cfg := testConfig()
	cfg.Security.ServiceToken = "service-secret"
	verificationID := uuid.New()
	store.verification = &gormmodel.VerificationRequest{ID: verificationID, DeveloperID: &store.developerID, Status: "pending"}
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: store})
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/"+verificationID.String()+"/internal/session", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status=%d", rec.Code)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v2/verify/"+verificationID.String()+"/internal/session", nil)
	req.Header.Set("X-Service-Token", "service-secret")
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid token status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// TestHandoffAndSessionAuth 验证 handoff token 与 session token 鉴权路径。
func TestHandoffAndSessionAuth(t *testing.T) {
	store := newStubIdswyftStore()
	cfg := testConfig()
	cfg.Security.APIKeySecret = "test-secret"
	apiKey := port.APIKeyAuth{ID: uuid.New(), DeveloperID: store.developerID, DeveloperStatus: "active"}
	store.handoff = &port.HandoffAuth{APIKey: apiKey, ExpiresAt: time.Now().Add(time.Hour), Status: "pending"}
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: store})
	body := []byte(`{"user_id":"` + uuid.New().String() + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/initialize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Handoff-Token", strings.Repeat("a", 64))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("handoff status=%d body=%s", rec.Code, rec.Body.String())
	}

	verificationID := uuid.New()
	store.verification = &gormmodel.VerificationRequest{ID: verificationID, DeveloperID: &store.developerID, Status: "pending"}
	store.session = &port.SessionAuth{VerificationID: verificationID, DeveloperID: store.developerID, ExpiresAt: ptrTime(time.Now().Add(time.Hour))}
	req = httptest.NewRequest(http.MethodGet, "/api/v2/verify/"+verificationID.String()+"/status", nil)
	req.Header.Set("X-Session-Token", strings.Repeat("b", 64))
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
