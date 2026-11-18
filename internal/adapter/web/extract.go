package web

import (
	"io"
	"net/http"
	"strings"

	"facial-liveness-verification-go/internal/document/fields"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/ocr"
)

const engineUploadLimit = 10 << 20

// OCRService 是 OCR 引擎的最小接口，便于测试替换。
type OCRService interface {
	RecognizeBytes(data []byte, opts ocr.Options) (ocr.Result, error)
}

type boxJSON struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type lineJSON struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Box        boxJSON `json:"box"`
}

// handleExtractOCR 实现 POST /extract/ocr：返回原始文本、行与置信度。
func (s *Server) handleExtractOCR(w http.ResponseWriter, r *http.Request) {
	if s.ocr == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OCR 未启用"))
		return
	}
	data, form, err := readEngineUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	res, err := s.ocr.RecognizeBytes(data, ocr.Options{Orientation: form["orientation"]})
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "OCR 识别失败"))
		return
	}
	lines := make([]lineJSON, 0)
	for _, line := range res.Lines {
		for _, item := range line {
			lines = append(lines, lineJSON{
				Text: item.Text, Confidence: item.Confidence,
				Box: boxJSON{X: item.Box.X, Y: item.Box.Y, W: item.Box.W, H: item.Box.H},
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"result": map[string]any{
			"raw_text":          res.Text,
			"confidence_scores": map[string]float64{"overall": res.Confidence},
			"confidence":        res.Confidence,
			"orientation":       res.Orientation,
			"lines":             lines,
		},
	})
}

// handleExtractFront 实现 POST /extract/front：OCR 字段 + 人脸特征。
func (s *Server) handleExtractFront(w http.ResponseWriter, r *http.Request) {
	if s.ocr == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OCR 未启用"))
		return
	}
	data, form, err := readEngineUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	res, err := s.ocr.RecognizeBytes(data, ocr.Options{})
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "OCR 识别失败"))
		return
	}
	fields := extractFrontFields(res.Text, form["issuing_country"])
	if missing := fields.missing(); len(missing) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"success":        false,
			"error":          "OCR_FIELDS_MISSING",
			"missing_fields": missing,
		})
		return
	}

	embedding := []float32(nil)
	faceConfidence := 0.0
	if s.face != nil {
		if boxes, err := s.face.DetectAll(data); err == nil {
			for _, box := range boxes {
				if box.Score > faceConfidence {
					faceConfidence = box.Score
				}
			}
		}
		if vec, err := s.face.ExtractLargest([][]byte{data}); err == nil {
			embedding = vec
		} else {
			faceConfidence = 0
		}
	}

	ocrPayload := map[string]any{
		"full_name":     fields.FullName,
		"date_of_birth": fields.DateOfBirth,
		"id_number":     fields.IDNumber,
		"expiry_date":   fields.ExpiryDate,
	}
	if fields.IssuingCountry != "" {
		ocrPayload["issuing_country"] = fields.IssuingCountry
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"result": map[string]any{
			"ocr":             ocrPayload,
			"face_embedding":  embedding,
			"face_confidence": faceConfidence,
			"ocr_confidence":  res.Confidence,
			"mrz_from_front":  nil,
		},
	})
}

// readEngineUpload 解析 multipart 文件与表单，限制 10MB 并校验图片类型。
func readEngineUpload(w http.ResponseWriter, r *http.Request) ([]byte, map[string]string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, engineUploadLimit)
	if err := r.ParseMultipartForm(engineUploadLimit); err != nil {
		return nil, nil, domain.New(domain.CodeDocumentInvalid, "请求体必须为 multipart/form-data 且不超过 10MB")
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, nil, domain.New(domain.CodeDocumentInvalid, "缺少 file 文件字段")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, engineUploadLimit))
	if err != nil {
		return nil, nil, domain.New(domain.CodeDocumentInvalid, "读取上传文件失败")
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, nil, domain.New(domain.CodeDocumentUnsupported, "仅支持 JPEG/PNG 图片")
	}
	form := map[string]string{
		"document_type":   r.FormValue("document_type"),
		"issuing_country": strings.ToUpper(strings.TrimSpace(r.FormValue("issuing_country"))),
		"orientation":     r.FormValue("orientation"),
	}
	return data, form, nil
}

type frontFields struct {
	FullName       string
	DateOfBirth    string
	IDNumber       string
	ExpiryDate     string
	IssuingCountry string
}

// missing 返回缺失的必填字段名。
func (f frontFields) missing() []string {
	out := make([]string, 0, 4)
	if f.FullName == "" {
		out = append(out, "full_name")
	}
	if f.DateOfBirth == "" {
		out = append(out, "date_of_birth")
	}
	if f.IDNumber == "" {
		out = append(out, "id_number")
	}
	if f.ExpiryDate == "" {
		out = append(out, "expiry_date")
	}
	return out
}

// extractFrontFields 调用证件字段规则 v1（MRZ + 国家规则 + 通用兜底）。
func extractFrontFields(text, issuingCountry string) frontFields {
	r := fields.Extract(text, issuingCountry)
	return frontFields{
		FullName:       r.FullName,
		DateOfBirth:    r.DateOfBirth,
		IDNumber:       r.IDNumber,
		ExpiryDate:     r.ExpiryDate,
		IssuingCountry: r.IssuingCountry,
	}
}
