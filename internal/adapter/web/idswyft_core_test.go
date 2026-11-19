package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	developerID   uuid.UUID
	verification  *gormmodel.VerificationRequest
	ensuredUser   uuid.UUID
	ensureUserErr error
	documents     map[uuid.UUID]*gormmodel.Document
	selfies       map[uuid.UUID]*gormmodel.Selfie
	handoff       *port.HandoffAuth
	session       *port.SessionAuth
	handoffs      map[string]*gormmodel.MobileHandoffSession
}

func newStubIdswyftStore() *stubIdswyftStore {
	return &stubIdswyftStore{developerID: uuid.New(), documents: map[uuid.UUID]*gormmodel.Document{}, selfies: map[uuid.UUID]*gormmodel.Selfie{}, handoffs: map[string]*gormmodel.MobileHandoffSession{}}
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
func (s *stubIdswyftStore) EnsureUser(_ context.Context, id uuid.UUID) error {
	s.ensuredUser = id
	return s.ensureUserErr
}
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
func (s *stubIdswyftStore) GetDeveloperByID(context.Context, uuid.UUID) (*gormmodel.Developer, error) {
	return &gormmodel.Developer{}, nil
}
func (s *stubIdswyftStore) GetDeveloperByAPIKey(context.Context, uuid.UUID) (*gormmodel.Developer, error) {
	return &gormmodel.Developer{}, nil
}
func (s *stubIdswyftStore) CreateBatchJob(context.Context, *gormmodel.BatchJob, []*gormmodel.BatchItem) error {
	return nil
}
func (s *stubIdswyftStore) GetBatchJob(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.BatchJob, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) UpdateBatchJob(context.Context, uuid.UUID, map[string]any) error {
	return nil
}
func (s *stubIdswyftStore) ListBatchJobs(context.Context, uuid.UUID, int, int) ([]gormmodel.BatchJob, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) ListBatchItems(context.Context, uuid.UUID) ([]gormmodel.BatchItem, error) {
	return nil, nil
}
func (s *stubIdswyftStore) UpdateBatchItem(context.Context, uuid.UUID, map[string]any) error {
	return nil
}
func (s *stubIdswyftStore) CreateHandoff(_ context.Context, session *gormmodel.MobileHandoffSession) error {
	s.handoffs[session.Token] = session
	return nil
}
func (s *stubIdswyftStore) GetHandoffByToken(_ context.Context, hash string) (*gormmodel.MobileHandoffSession, error) {
	session := s.handoffs[hash]
	if session == nil {
		return nil, context.Canceled
	}
	return session, nil
}
func (s *stubIdswyftStore) UpdateHandoff(_ context.Context, hash string, updates map[string]any) (int64, error) {
	if session := s.handoffs[hash]; session != nil {
		if status, ok := updates["status"].(string); ok {
			session.Status = status
		}
	}
	return 1, nil
}
func (s *stubIdswyftStore) UpdatePendingHandoff(_ context.Context, hash string, updates map[string]any) (int64, error) {
	session := s.handoffs[hash]
	if session == nil || session.Status != "pending" {
		return 0, nil
	}
	if value, ok := updates["verification_id"].(string); ok {
		session.VerificationID = &value
	}
	if value, ok := updates["status"].(string); ok {
		session.Status = value
	}
	if value, ok := updates["result"].(json.RawMessage); ok {
		session.Result = value
	}
	return 1, nil
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
	if store.ensuredUser != userID {
		t.Fatalf("initialize did not ensure user %s; got %s", userID, store.ensuredUser)
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

// TestVerifyInitializeStopsWhenUserEnsureFails 不创建缺少父用户行的验证请求。
func TestVerifyInitializeStopsWhenUserEnsureFails(t *testing.T) {
	store := newStubIdswyftStore()
	store.ensureUserErr = errors.New("ensure user failed")
	cfg := testConfig()
	cfg.Security.APIKeySecret = "test-secret"
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: store})
	body := []byte(`{"user_id":"` + uuid.New().String() + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/initialize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if store.verification != nil {
		t.Fatal("verification was created after ensuring user failed")
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

// TestVerificationStatusIncludesOCRData 覆盖正面证件上传后 status 返回 ocr_data，
// 防止演示页轮询卡在「正在提取证件文字」。
func TestVerificationStatusIncludesOCRData(t *testing.T) {
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
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/verify/"+verificationID.String()+"/front-document", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("front upload status=%d body=%s", rec.Code, rec.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/v2/verify/"+verificationID.String()+"/status", nil)
	statusReq.Header.Set("X-API-Key", "test-key")
	statusRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", statusRec.Code, statusRec.Body.String())
	}
	out := decodeObject(t, statusRec)
	assertShape(t, out, map[string]string{
		"front_document_uploaded": "bool", "back_document_uploaded": "bool",
		"live_capture_uploaded": "bool", "ocr_data": "object",
	}, "$")
	ocrData, ok := out["ocr_data"].(map[string]any)
	if !ok || len(ocrData) == 0 {
		t.Fatalf("ocr_data 应为非空对象，得到 %v", out["ocr_data"])
	}
	if text, _ := ocrData["raw_text"].(string); !strings.Contains(text, "JOHN DOE") {
		t.Fatalf("ocr_data.raw_text 应包含识别文本，得到 %q", text)
	}
	if uploaded, _ := out["front_document_uploaded"].(bool); !uploaded {
		t.Fatal("front_document_uploaded 应为 true")
	}
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

// TestHandoffFlow 覆盖 handoff 创建、读取、link、complete 与 status。
func TestHandoffFlow(t *testing.T) {
	store := newStubIdswyftStore()
	cfg := testConfig()
	cfg.Security.APIKeySecret = "test-secret"
	srv := NewServer(Deps{Cfg: cfg, IdswyftStore: store})
	body := []byte(`{"api_key":"ik_test","user_id":"user-1","source":"api"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/verify/handoff/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	token, _ := created["token"].(string)
	if token == "" {
		t.Fatalf("missing token: %v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/verify/handoff/"+token+"/session", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", rec.Code, rec.Body.String())
	}

	verificationID := uuid.New()
	linkBody := []byte(`{"verification_id":"` + verificationID.String() + `"}`)
	req = httptest.NewRequest(http.MethodPatch, "/api/verify/handoff/"+token+"/link", bytes.NewReader(linkBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("link status=%d body=%s", rec.Code, rec.Body.String())
	}

	completeBody := []byte(`{"status":"completed","result":{"final_result":"verified"}}`)
	req = httptest.NewRequest(http.MethodPatch, "/api/verify/handoff/"+token+"/complete", bytes.NewReader(completeBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/verify/handoff/"+token+"/status", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "completed") {
		t.Fatalf("status status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func (s *stubIdswyftStore) GetVaultByVerification(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.IdentityVault, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) CreateVaultEntry(context.Context, *gormmodel.IdentityVault) error {
	return nil
}
func (s *stubIdswyftStore) GetVaultByToken(context.Context, uuid.UUID, string) (*gormmodel.IdentityVault, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) ListVault(context.Context, uuid.UUID, int, int) ([]gormmodel.IdentityVault, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) DeleteVault(context.Context, uuid.UUID, string) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) CreateShareLink(context.Context, *gormmodel.VaultShareLink) error {
	return nil
}
func (s *stubIdswyftStore) GetShareLink(context.Context, string) (*gormmodel.VaultShareLink, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) IncrementShareAccess(context.Context, uuid.UUID) (int64, error) {
	return 1, nil
}
func (s *stubIdswyftStore) GetVaultByID(context.Context, uuid.UUID) (*gormmodel.IdentityVault, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) TouchVaultAccess(context.Context, uuid.UUID) error { return nil }
func (s *stubIdswyftStore) GetActiveCredential(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.VerifiableCredential, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) CreateCredential(context.Context, *gormmodel.VerifiableCredential) error {
	return nil
}
func (s *stubIdswyftStore) GetCredentialByJTI(context.Context, string) (*gormmodel.VerifiableCredential, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) RevokeCredential(context.Context, uuid.UUID, string, string) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) WebhooksByDeveloper(context.Context, uuid.UUID) ([]gormmodel.Webhook, error) {
	return nil, nil
}
func (s *stubIdswyftStore) WebhookByID(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.Webhook, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) WebhookByURL(context.Context, uuid.UUID, string, bool) (*gormmodel.Webhook, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) CreateWebhook(context.Context, *gormmodel.Webhook) error { return nil }
func (s *stubIdswyftStore) UpdateWebhook(context.Context, uuid.UUID, uuid.UUID, map[string]any) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) DeleteWebhook(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) WebhookDeliveries(context.Context, uuid.UUID, int, int) ([]gormmodel.WebhookDelivery, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) CreateSchedule(context.Context, *gormmodel.ReverificationSchedule) error {
	return nil
}
func (s *stubIdswyftStore) ListSchedules(context.Context, uuid.UUID, string, int, int) ([]gormmodel.ReverificationSchedule, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) GetSchedule(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.ReverificationSchedule, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) CancelSchedule(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) ExpiringDocuments(context.Context, uuid.UUID, int, int, int) ([]gormmodel.ExpiryAlert, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) CreateVAASUser(context.Context, *gormmodel.User) error { return nil }
func (s *stubIdswyftStore) ListDevelopers(context.Context, string, string, int, int) ([]gormmodel.Developer, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) SetDeveloperStatus(context.Context, uuid.UUID, string) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) CountAPIKeys(context.Context, uuid.UUID) (int64, error) { return 0, nil }
func (s *stubIdswyftStore) CountVerifications(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func (s *stubIdswyftStore) DeveloperAPIKeys(context.Context, uuid.UUID) ([]gormmodel.APIKey, error) {
	return nil, nil
}
func (s *stubIdswyftStore) CreateDeveloperAPIKey(context.Context, *gormmodel.APIKey) error {
	return nil
}
func (s *stubIdswyftStore) RevokeDeveloperAPIKey(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) GetDeveloperAPIKey(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.APIKey, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) UpdateDeveloper(context.Context, uuid.UUID, map[string]any) error {
	return nil
}
func (s *stubIdswyftStore) DeleteDeveloper(context.Context, uuid.UUID) (int64, error) { return 0, nil }

func (s *stubIdswyftStore) CreateRuleset(context.Context, *gormmodel.ComplianceRuleset) error {
	return nil
}
func (s *stubIdswyftStore) ListRulesets(context.Context, uuid.UUID) ([]gormmodel.ComplianceRuleset, error) {
	return nil, nil
}
func (s *stubIdswyftStore) GetRuleset(context.Context, uuid.UUID, uuid.UUID) (*gormmodel.ComplianceRuleset, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) UpdateRuleset(context.Context, uuid.UUID, uuid.UUID, map[string]any) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) DeleteRuleset(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) RulesByRuleset(context.Context, uuid.UUID) ([]gormmodel.ComplianceRule, error) {
	return nil, nil
}
func (s *stubIdswyftStore) CreateRule(context.Context, *gormmodel.ComplianceRule) error { return nil }
func (s *stubIdswyftStore) GetRule(context.Context, uuid.UUID) (*gormmodel.ComplianceRule, error) {
	return nil, context.Canceled
}
func (s *stubIdswyftStore) UpdateRule(context.Context, uuid.UUID, map[string]any) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) DeleteRule(context.Context, uuid.UUID) (int64, error) { return 0, nil }

func (s *stubIdswyftStore) CreateDeveloper(context.Context, *gormmodel.Developer) error { return nil }
func (s *stubIdswyftStore) CreateReviewer(context.Context, *gormmodel.VerificationReviewer) error {
	return nil
}
func (s *stubIdswyftStore) Reviewers(context.Context, uuid.UUID) ([]gormmodel.VerificationReviewer, error) {
	return nil, nil
}
func (s *stubIdswyftStore) DeleteReviewer(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *stubIdswyftStore) AdminVerifications(context.Context, string, *uuid.UUID, int, int) ([]gormmodel.VerificationRequest, int64, error) {
	return nil, 0, nil
}
func (s *stubIdswyftStore) AdminStatusCounts(context.Context, time.Time) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (s *stubIdswyftStore) AdminDeveloperCount(context.Context) (int64, error) { return 0, nil }
func (s *stubIdswyftStore) DuplicateVerifications(context.Context, *gormmodel.VerificationRequest) ([]gormmodel.VerificationRequest, error) {
	return nil, nil
}
func (s *stubIdswyftStore) ReviewVerification(context.Context, uuid.UUID, string, string, string) error {
	return nil
}
func (s *stubIdswyftStore) ProviderMetricSummary(context.Context, time.Time) ([]map[string]any, error) {
	return nil, nil
}
func (s *stubIdswyftStore) AuditLogs(context.Context, time.Time, time.Time, *uuid.UUID, *int32) ([]gormmodel.APIActivityLog, error) {
	return nil, nil
}
func (s *stubIdswyftStore) DeleteUserData(context.Context, uuid.UUID) (int64, error) { return 0, nil }
func (s *stubIdswyftStore) VerificationStatusCounts(context.Context, uuid.UUID, time.Time) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (s *stubIdswyftStore) RecentActivity(context.Context, uuid.UUID, int) ([]gormmodel.APIActivityLog, error) {
	return nil, nil
}
func (s *stubIdswyftStore) VerificationByID(_ context.Context, id uuid.UUID) (*gormmodel.VerificationRequest, error) {
	return s.GetVerification(context.Background(), id)
}
