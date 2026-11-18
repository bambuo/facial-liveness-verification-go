package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/ocr"
)

const testMRZText = "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\n" +
	"L898902C36UTO7408122F1204159ZE184226B<<<<<10\n"

// stubFace 返回固定人脸与活体结果。
type stubFace struct{}

func (stubFace) ModelVersion() string { return "stub" }
func (stubFace) DetectAll([]byte) ([]domain.FaceBox, error) {
	return []domain.FaceBox{{Score: 0.97}}, nil
}
func (stubFace) ExtractLargest([][]byte) ([]float32, error) { return []float32{0.1, 0.2}, nil }
func (stubFace) CheckLiveness([][]byte, float64, float64) (domain.LivenessVote, error) {
	return domain.LivenessVote{Status: domain.LivenessReal, MinScore: 0.9}, nil
}
func (stubFace) Analyze([][]byte) ([]domain.FrameMetric, error) { return nil, nil }
func (stubFace) Compare(_, _ []float32) (float64, error)        { return 1, nil }

// TestExtractBackContractShape 校验 /extract/back 的 MRZ 响应形状。
func TestExtractBackContractShape(t *testing.T) {
	srv := NewServer(Deps{
		Cfg: testConfig(),
		OCR: stubOCR{res: ocr.Result{Text: testMRZText, Confidence: 0.99}},
	})
	body, contentType := multipartBody(t, testPNG(t), nil)
	req := httptest.NewRequest(http.MethodPost, "/extract/back", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	out := decodeObject(t, rec)
	result := out["result"].(map[string]any)
	assertShape(t, result, map[string]string{
		"qr_payload":       "object",
		"mrz_result":       "object",
		"barcode_format":   "string",
		"raw_barcode_data": "null",
	}, "$.result")
	mrz := result["mrz_result"].(map[string]any)
	assertShape(t, mrz, map[string]string{
		"raw_lines":       "array",
		"fields":          "object",
		"checksums_valid": "bool",
	}, "$.result.mrz_result")
	if mrz["checksums_valid"] != true {
		t.Fatalf("expected valid MRZ checksums: %v", mrz)
	}
	if result["barcode_format"] != "MRZ_TD3" {
		t.Fatalf("barcode_format=%v", result["barcode_format"])
	}
}

// TestExtractLiveContractShape 校验 /extract/live 的人脸与活体响应形状。
func TestExtractLiveContractShape(t *testing.T) {
	srv := NewServer(Deps{Cfg: testConfig(), Face: stubFace{}})
	body, contentType := multipartBody(t, testPNG(t), nil)
	req := httptest.NewRequest(http.MethodPost, "/extract/live", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	out := decodeObject(t, rec)
	result := out["result"].(map[string]any)
	assertShape(t, result, map[string]string{
		"face_embedding":  "array",
		"face_confidence": "number",
		"liveness_passed": "bool",
		"liveness_score":  "number",
	}, "$.result")
	if result["liveness_passed"] != true {
		t.Fatalf("expected liveness passed: %v", result)
	}
}

// TestVoiceEndpointsDisabled 校验未启用语音引擎时返回 503。
func TestVoiceEndpointsDisabled(t *testing.T) {
	srv := NewServer(Deps{Cfg: testConfig()})
	for _, path := range []string{"/extract/voice-enroll", "/extract/voice-verify"} {
		body, contentType := multipartBody(t, testPNG(t), nil)
		req := httptest.NewRequest(http.MethodPost, path, body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}
