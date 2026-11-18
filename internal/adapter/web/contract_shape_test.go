package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"facial-liveness-verification-go/internal/ocr"
)

// TestExtractOCRContractShape 校验 /extract/ocr 响应形状符合引擎契约。
func TestExtractOCRContractShape(t *testing.T) {
	srv := NewServer(Deps{
		Cfg: testConfig(),
		OCR: stubOCR{res: ocr.Result{Text: "HELLO", Confidence: 0.9}},
	})
	body, contentType := multipartBody(t, testPNG(t), nil)
	req := httptest.NewRequest(http.MethodPost, "/extract/ocr", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	out := decodeObject(t, rec)
	assertShape(t, out, map[string]string{"success": "bool", "result": "object"}, "$")
	result := out["result"].(map[string]any)
	assertShape(t, result, map[string]string{
		"raw_text":          "string",
		"confidence_scores": "object",
		"confidence":        "number",
		"orientation":       "number",
		"lines":             "array",
	}, "$.result")
}

// TestExtractFrontContractShape 校验 /extract/front 成功响应形状。
func TestExtractFrontContractShape(t *testing.T) {
	srv := NewServer(Deps{
		Cfg: testConfig(),
		OCR: stubOCR{res: ocr.Result{
			Text:       "NAME: JOHN DOE\nDOB: 1990-01-01\nID NO: AB1234567\nEXP: 2030-01-01\n",
			Confidence: 0.95,
		}},
	})
	body, contentType := multipartBody(t, testPNG(t), map[string]string{"issuing_country": "us"})
	req := httptest.NewRequest(http.MethodPost, "/extract/front", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	out := decodeObject(t, rec)
	result := out["result"].(map[string]any)
	assertShape(t, result, map[string]string{
		"ocr":             "object",
		"face_embedding":  "null",
		"face_confidence": "number",
		"ocr_confidence":  "number",
		"mrz_from_front":  "null",
		"authenticity":    "object",
	}, "$.result")
	ocrPayload := result["ocr"].(map[string]any)
	assertShape(t, ocrPayload, map[string]string{
		"full_name":     "string",
		"date_of_birth": "string",
		"id_number":     "string",
		"expiry_date":   "string",
	}, "$.result.ocr")
}

func decodeObject(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	return out
}

// assertShape 按期望类型校验对象字段，类型名为 JSON 类型。
func assertShape(t *testing.T, value any, want map[string]string, path string) {
	t.Helper()
	obj, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s 不是 object: %T", path, value)
	}
	for key, typ := range want {
		field, exists := obj[key]
		if !exists {
			t.Fatalf("%s 缺少字段 %q", path, key)
		}
		if !matchesType(field, typ) {
			t.Fatalf("%s.%s 类型 %T 不符合 %s", path, key, field, typ)
		}
	}
}

func matchesType(value any, want string) bool {
	switch want {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}
