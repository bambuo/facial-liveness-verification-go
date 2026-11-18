package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"facial-liveness-verification-go/internal/ocr"
)

// TestExtractOCRRealEngine 用真实模型验证 /extract/ocr；未设置 OCR_TEST_MODEL_DIR 时跳过。
func TestExtractOCRRealEngine(t *testing.T) {
	dir := os.Getenv("OCR_TEST_MODEL_DIR")
	if dir == "" {
		t.Skip("OCR_TEST_MODEL_DIR 未设置")
	}
	engine, err := ocr.New(ocr.Config{
		DetModel:    filepath.Join(dir, "det.onnx"),
		RecModel:    filepath.Join(dir, "rec.onnx"),
		DictPath:    filepath.Join(dir, "dict.txt"),
		WidthBucket: 32,
	})
	if err != nil {
		t.Fatalf("load ocr: %v", err)
	}
	defer engine.Close()
	image, err := os.ReadFile(filepath.Join(dir, "demo.jpg"))
	if err != nil {
		t.Fatalf("read demo: %v", err)
	}

	srv := NewServer(Deps{Cfg: testConfig(), OCR: engine})
	body, contentType := multipartBody(t, image, nil)
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
	text, _ := result["raw_text"].(string)
	if len(text) < 50 || !strings.Contains(text, "44,900") {
		t.Fatalf("unexpected OCR text: %q", text)
	}
}
