package web

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"facial-liveness-verification-go/internal/config"
	"facial-liveness-verification-go/internal/ocr"
)

type stubOCR struct {
	res ocr.Result
	err error
}

// RecognizeBytes 返回固定结果，供端点测试使用。
func (s stubOCR) RecognizeBytes(data []byte, opts ocr.Options) (ocr.Result, error) {
	return s.res, s.err
}

// TestExtractFrontFields 覆盖标签优先与正则兜底的字段抽取。
func TestExtractFrontFields(t *testing.T) {
	text := "NAME: JOHN DOE\nDOB: 1990-01-01\nID NO: AB1234567\nEXP: 2030-01-01\n"
	got := extractFrontFields(text, "us")
	if got.FullName != "JOHN DOE" || got.DateOfBirth != "1990-01-01" ||
		got.IDNumber != "AB1234567" || got.ExpiryDate != "2030-01-01" || got.IssuingCountry != "US" {
		t.Fatalf("unexpected fields: %+v", got)
	}
	if missing := got.missing(); len(missing) != 0 {
		t.Fatalf("unexpected missing: %v", missing)
	}

	empty := extractFrontFields("hello world", "")
	if missing := empty.missing(); len(missing) != 4 {
		t.Fatalf("expected 4 missing fields, got %v", missing)
	}
}

// TestExtractOCREndpoint 验证 /extract/ocr 返回原始文本与行。
func TestExtractOCREndpoint(t *testing.T) {
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
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	result, _ := out["result"].(map[string]any)
	if result["raw_text"] != "HELLO" {
		t.Fatalf("unexpected result: %v", out)
	}
}

// TestExtractFrontMissingFields 验证必填字段缺失时返回 422。
func TestExtractFrontMissingFields(t *testing.T) {
	srv := NewServer(Deps{
		Cfg: testConfig(),
		OCR: stubOCR{res: ocr.Result{Text: "HELLO", Confidence: 0.9}},
	})
	body, contentType := multipartBody(t, testPNG(t), nil)
	req := httptest.NewRequest(http.MethodPost, "/extract/front", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	missing, _ := out["missing_fields"].([]any)
	if len(missing) != 4 {
		t.Fatalf("expected 4 missing fields, got %v", out)
	}
}

// TestExtractFrontSuccess 验证字段齐全时返回 FrontExtractionResult 子集。
func TestExtractFrontSuccess(t *testing.T) {
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
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	result, _ := out["result"].(map[string]any)
	ocrPayload, _ := result["ocr"].(map[string]any)
	if ocrPayload["full_name"] != "JOHN DOE" || ocrPayload["issuing_country"] != "US" {
		t.Fatalf("unexpected ocr payload: %v", ocrPayload)
	}
	if result["face_embedding"] != nil || result["face_confidence"].(float64) != 0 {
		t.Fatalf("expected nil embedding, got %v", result)
	}
}

func testConfig() *config.Config {
	return &config.Config{
		Env:       "test",
		Security:  config.Security{MaxUploadBytes: 10 << 20},
		RateLimit: config.RateLimit{PermitPerMinute: 100},
	}
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func multipartBody(t *testing.T, file []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "doc.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(file); err != nil {
		t.Fatal(err)
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}
