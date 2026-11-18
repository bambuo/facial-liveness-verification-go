package web

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/audio"
	"facial-liveness-verification-go/internal/document/fields"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/ocr"
)

// handleFrontDocument 实现 POST /api/v2/verify/:id/front-document。
func (s *Server) handleFrontDocument(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.DocumentID != nil {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "verification_id": verification.ID.String(), "status": "processing", "message": "front document already uploaded"})
		return
	}
	if s.ocr == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OCR 未启用"))
		return
	}
	data, filename, mime, form, err := readMultipartImage(w, r, "document")
	if err != nil {
		writeError(w, r, err)
		return
	}
	res, err := s.ocr.RecognizeBytes(data, ocr.Options{})
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "OCR 识别失败"))
		return
	}
	country := strings.ToUpper(strings.TrimSpace(form["issuing_country"]))
	extracted := fields.Extract(res.Text, country)
	var frontEmbedding []float32
	faceConfidence := 0.0
	if s.face != nil {
		if boxes, detectErr := s.face.DetectAll(data); detectErr == nil {
			for _, box := range boxes {
				if box.Score > faceConfidence {
					faceConfidence = box.Score
				}
			}
		}
		if embedding, embedErr := s.face.ExtractLargest([][]byte{data}); embedErr == nil {
			frontEmbedding = embedding
		}
	}
	docID := uuid.New()
	ocrPayload := map[string]any{
		"full_name": extracted.FullName, "date_of_birth": extracted.DateOfBirth,
		"id_number": extracted.IDNumber, "expiry_date": extracted.ExpiryDate,
		"issuing_country": extracted.IssuingCountry, "raw_text": res.Text,
		"confidence": res.Confidence,
	}
	raw, _ := json.Marshal(ocrPayload)
	filePath, err := s.storeObject(r.Context(), fmt.Sprintf("verification/%s/front-%s", verification.ID, docID), data, mime)
	if err != nil {
		writeError(w, r, err)
		return
	}
	document := &gormmodel.Document{
		ID: docID, VerificationRequestID: &verification.ID,
		DocumentType:   firstNonEmpty(form["document_type"], "auto"),
		IssuingCountry: strPtr(country), FileName: filename,
		FilePath: filePath,
		FileSize: int32Ptr(len(data)), MimeType: &mime, OCRData: raw,
		OCRExtracted: boolPtr(true), IsBackOfID: boolPtr(false),
	}
	if err := s.idswyft.CreateDocument(r.Context(), document); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{
		"document_id": docID, "status": "processing", "updated_at": time.Now().UTC(),
	}
	if country != "" {
		updates["issuing_country"] = country
	}
	if err := s.idswyft.UpdateVerification(r.Context(), verification.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	requiresBack := !strings.Contains(strings.ToLower(document.DocumentType), "passport")
	_ = s.idswyft.UpsertContext(r.Context(), verification.ID, mustJSON(map[string]any{
		"current_step": stepAfterFront(requiresBack), "total_steps": 5, "document_type": document.DocumentType,
		"front_embedding": frontEmbedding, "face_confidence": faceConfidence,
	}))
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(),
		"verification_mode": stringValue(verification.VerificationMode), "status": "processing",
		"current_step": stepAfterFront(requiresBack), "total_steps": 5,
		"document_id": docID.String(), "ocr_data": ocrPayload,
		"detected_document_type": document.DocumentType, "classification_confidence": res.Confidence,
		"requires_back": requiresBack, "message": "Front document accepted",
	})
}

// handleBackDocument 实现 POST /api/v2/verify/:id/back-document。
func (s *Server) handleBackDocument(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.DocumentID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "verification_id": verification.ID.String(), "message": "front document required"})
		return
	}
	if verification.BackOfIDUploaded != nil && *verification.BackOfIDUploaded {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "verification_id": verification.ID.String(), "message": "back document already uploaded"})
		return
	}
	if s.ocr == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OCR 未启用"))
		return
	}
	data, filename, mime, form, err := readMultipartImage(w, r, "document")
	if err != nil {
		writeError(w, r, err)
		return
	}
	res, err := s.ocr.RecognizeBytes(data, ocr.Options{})
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "OCR 识别失败"))
		return
	}
	mrz, hasMRZ := fields.ExtractMRZ(res.Text)
	docID := uuid.New()
	extracted := fields.Extract(res.Text, strings.ToUpper(strings.TrimSpace(form["issuing_country"])))
	ocrPayload := map[string]any{
		"raw_text": res.Text, "confidence": res.Confidence,
		"full_name": extracted.FullName, "date_of_birth": extracted.DateOfBirth,
		"id_number": extracted.IDNumber, "expiry_date": extracted.ExpiryDate,
	}
	if hasMRZ {
		ocrPayload["mrz"] = map[string]any{"raw_lines": mrz.RawLines, "checksums_valid": mrz.ChecksumsValid}
	}
	raw, _ := json.Marshal(ocrPayload)
	filePath, err := s.storeObject(r.Context(), fmt.Sprintf("verification/%s/back-%s", verification.ID, docID), data, mime)
	if err != nil {
		writeError(w, r, err)
		return
	}
	document := &gormmodel.Document{
		ID: docID, VerificationRequestID: &verification.ID,
		DocumentType: firstNonEmpty(form["document_type"], "auto"), FileName: filename,
		FilePath: filePath,
		FileSize: int32Ptr(len(data)), MimeType: &mime, OCRData: raw,
		OCRExtracted: boolPtr(true), IsBackOfID: boolPtr(true),
	}
	if err := s.idswyft.CreateDocument(r.Context(), document); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{
		"back_of_id_uploaded": true, "status": "processing", "updated_at": time.Now().UTC(),
	}
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, updates)
	_ = s.idswyft.UpsertContext(r.Context(), verification.ID, mustJSON(map[string]any{
		"current_step": "cross_validation", "total_steps": 5, "back_document_id": docID.String(),
	}))
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(),
		"verification_mode": stringValue(verification.VerificationMode), "status": "processing",
		"current_step": "cross_validation", "total_steps": 5, "document_id": docID.String(),
		"barcode_data": nil, "barcode_extraction_failed": true,
		"documents_match": hasMRZ, "message": "Back document accepted",
	})
}

// handleCrossValidation 实现 POST /api/v2/verify/:id/cross-validation。
func (s *Server) handleCrossValidation(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.DocumentID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "verification_id": verification.ID.String(), "message": "front document required"})
		return
	}
	front, err := s.idswyft.GetDocument(r.Context(), *verification.DocumentID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var frontData map[string]any
	_ = json.Unmarshal(front.OCRData, &frontData)
	back, backErr := s.idswyft.GetBackDocument(r.Context(), verification.ID)
	if backErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "verification_id": verification.ID.String(), "message": "back document required"})
		return
	}
	var backData map[string]any
	_ = json.Unmarshal(back.OCRData, &backData)
	score := 0.0
	checks := []struct {
		key    string
		weight float64
	}{
		{"id_number", 0.4}, {"date_of_birth", 0.25}, {"expiry_date", 0.2}, {"full_name", 0.15},
	}
	for _, check := range checks {
		if fieldString(frontData, check.key) != "" && strings.EqualFold(fieldString(frontData, check.key), fieldString(backData, check.key)) {
			score += check.weight
		}
	}
	score = math.Round(score*100) / 100
	match := score >= 0.7
	if err := s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"cross_validation_score": score, "status": "processing", "updated_at": time.Now().UTC(),
	}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(), "status": "processing",
		"current_step": "live_capture", "documents_match": match,
		"cross_validation_results": map[string]any{"score": score},
		"message":                  "Cross validation completed",
	})
}

// handleLiveCapture 实现 POST /api/v2/verify/:id/live-capture。
func (s *Server) handleLiveCapture(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.DocumentID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "verification_id": verification.ID.String(), "message": "front document required"})
		return
	}
	if s.face == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "人脸引擎未启用"))
		return
	}
	data, filename, _, _, err := readMultipartImage(w, r, "selfie")
	if err != nil {
		writeError(w, r, err)
		return
	}
	faceConfidence := 0.0
	if boxes, detectErr := s.face.DetectAll(data); detectErr == nil {
		for _, box := range boxes {
			if box.Score > faceConfidence {
				faceConfidence = box.Score
			}
		}
	}
	embedding, embedErr := s.face.ExtractLargest([][]byte{data})
	faceMatchScore := 0.0
	if embedErr == nil {
		if rawContext, ctxErr := s.idswyft.GetContext(r.Context(), verification.ID); ctxErr == nil {
			var contextData map[string]any
			_ = json.Unmarshal(rawContext, &contextData)
			if values, ok := contextData["front_embedding"].([]any); ok && len(values) > 0 {
				frontEmbedding := make([]float32, len(values))
				for i, value := range values {
					if number, ok := value.(float64); ok {
						frontEmbedding[i] = float32(number)
					}
				}
				if score, compareErr := domain.Compare(frontEmbedding, embedding); compareErr == nil {
					faceMatchScore = score
				}
			}
		}
	}
	livenessScore := 0.0
	livenessPassed := false
	if vote, livenessErr := s.face.CheckLiveness([][]byte{data}, s.cfg.Face.LivenessThreshold, s.cfg.Face.LiveRatioThreshold); livenessErr == nil {
		livenessScore = vote.MinScore
		livenessPassed = vote.Status == domain.LivenessReal
	}
	selfieID := uuid.New()
	selfiePath, err := s.storeObject(r.Context(), fmt.Sprintf("verification/%s/selfie-%s", verification.ID, selfieID), data, http.DetectContentType(data))
	if err != nil {
		writeError(w, r, err)
		return
	}
	selfie := &gormmodel.Selfie{
		ID: selfieID, VerificationRequestID: &verification.ID, FileName: filename,
		FilePath: selfiePath,
		FileSize: int32Ptr(len(data)), FaceDetected: boolPtr(embedErr == nil),
		LivenessScore: &livenessScore,
	}
	if err := s.idswyft.CreateSelfie(r.Context(), selfie); err != nil {
		writeError(w, r, err)
		return
	}
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"selfie_id": selfieID, "liveness_score": livenessScore, "live_capture_completed": true,
		"face_match_score": faceMatchScore, "status": "processing", "updated_at": time.Now().UTC(),
	})
	_ = embedding
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(), "status": "processing",
		"current_step": "finalize", "total_steps": 5, "selfie_id": selfieID.String(),
		"face_match_results": map[string]any{"confidence": faceConfidence},
		"liveness_results":   map[string]any{"liveness_passed": livenessPassed, "liveness_score": livenessScore, "liveness_mode": "passive"},
		"final_result":       nil, "message": "Live capture accepted",
	})
}

// handleVoiceChallenge 实现 POST /api/v2/verify/:id/voice-challenge。
func (s *Server) handleVoiceChallenge(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	digits, err := randomDigits(6)
	if err != nil {
		writeError(w, r, err)
		return
	}
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"voice_challenge": digits, "voice_challenge_created_at": time.Now().UTC(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(), "challenge_digits": digits,
		"expires_in_seconds": int(defaultPhoneOTPTTL.Seconds()), "message": "Voice challenge generated",
	})
}

// handleVoiceCapture 实现 POST /api/v2/verify/:id/voice-capture。
func (s *Server) handleVoiceCapture(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if s.voice == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "voice engine disabled", "retryable": true, "current_step": "voice_capture"})
		return
	}
	if verification.VoiceChallenge == nil || verification.VoiceChallengeCreatedAt == nil || !verification.VoiceChallengeCreatedAt.Add(120*time.Second).After(time.Now()) {
		writeJSON(w, http.StatusGone, map[string]any{"success": false, "error": "voice challenge expired"})
		return
	}
	samples, err := readVoiceUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	embedding, err := s.voice.SpeakerEmbedding(samples)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "声纹提取失败"))
		return
	}
	transcription, err := s.voice.Transcribe(samples)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "语音转写失败"))
		return
	}
	spoken := audio.DigitsOnly(transcription)
	score := 0.0
	if spoken == *verification.VoiceChallenge {
		score = 1
	}
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"voice_match_score": score, "status": "processing", "updated_at": time.Now().UTC(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(), "status": "processing",
		"current_step": "finalize", "total_steps": 5,
		"voice_match_results": map[string]any{"score": score, "embedding_dimension": len(embedding), "transcription": transcription},
		"final_result":        finalForVoice(score), "message": "Voice capture processed",
	})
}

// handleFinalize 实现 POST /api/v2/verify/:id/finalize。
func (s *Server) handleFinalize(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.LiveCaptureCompleted == nil || !*verification.LiveCaptureCompleted {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "verification_id": verification.ID.String(), "status": verification.Status, "final_result": nil, "message": "live capture required"})
		return
	}
	final := "manual_review"
	status := "manual_review"
	if verification.DocumentID != nil && verification.LiveCaptureCompleted != nil && *verification.LiveCaptureCompleted &&
		verification.LivenessScore != nil && *verification.LivenessScore >= s.cfg.Face.LivenessThreshold &&
		verification.FaceMatchScore != nil && *verification.FaceMatchScore >= s.cfg.Face.MatchThreshold {
		final, status = "verified", "verified"
	}
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"status": status, "processing_completed_at": time.Now().UTC(), "updated_at": time.Now().UTC(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(), "status": status,
		"current_step": "complete", "final_result": final, "message": "Verification finalized",
	})
	s.dispatchVerificationWebhook(verification, status, final)
}

// handleRestart 实现 POST /api/v2/verify/:id/restart。
func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.RetryCount >= 3 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "retry limit reached", "retry_count": verification.RetryCount})
		return
	}
	retry := verification.RetryCount + 1
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"status": "pending", "retry_count": retry, "live_capture_completed": false,
		"document_id": nil, "selfie_id": nil, "liveness_score": nil, "updated_at": time.Now().UTC(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(), "retry_count": retry,
		"message": "Verification restarted",
	})
}

// handlePhoneOTPSend 实现 POST /api/v2/verify/:id/phone-otp/send。
func (s *Server) handlePhoneOTPSend(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if !phonePattern.MatchString(req.PhoneNumber) {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "phone_number 格式无效"))
		return
	}
	allowed, err := s.idswyft.CheckPhoneOTPSend(r.Context(), verification.ID, defaultPhoneOTPSends, int(defaultPhoneOTPWindow.Seconds()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !allowed {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"success": false, "message": "too many OTP requests"})
		return
	}
	code, err := randomDigits(6)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := s.idswyft.CreatePhoneOTP(r.Context(), verification.ID, req.PhoneNumber, hashToken(s.apiKeySecret(), code), defaultPhoneOTPTTL); err != nil {
		writeError(w, r, err)
		return
	}
	if s.cfg.SMS.Enabled && s.sms != nil {
		if err := s.sms.Send(r.Context(), req.PhoneNumber, "idswyft verification code: "+code); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "message": "SMS delivery failed"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "OTP sent"})
}

// handlePhoneOTPVerify 实现 POST /api/v2/verify/:id/phone-otp/verify。
func (s *Server) handlePhoneOTPVerify(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if !otpPattern.MatchString(req.Code) {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "code 必须是 6 位数字"))
		return
	}
	result, err := s.idswyft.VerifyPhoneOTP(r.Context(), verification.ID, hashToken(s.apiKeySecret(), req.Code), defaultPhoneOTPAttempts)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if result.Status != "valid" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "invalid code"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "phone verified", "phone_verified": true})
}

func fieldString(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

// storeObject 在启用 S3 时上传对象并返回对象键；未启用时返回本地占位键。
func (s *Server) storeObject(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	if s.storage == nil || !s.cfg.Storage.Enabled {
		return key, nil
	}
	if err := s.storage.Put(ctx, key, data, contentType); err != nil {
		return "", err
	}
	return key, nil
}

func readMultipartImage(w http.ResponseWriter, r *http.Request, field string) ([]byte, string, string, map[string]string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, engineUploadLimit)
	if err := r.ParseMultipartForm(engineUploadLimit); err != nil {
		return nil, "", "", nil, domain.New(domain.CodeDocumentInvalid, "请求体必须为 multipart/form-data 且不超过 10MB")
	}
	file, header, err := r.FormFile(field)
	if err != nil {
		return nil, "", "", nil, domain.New(domain.CodeDocumentInvalid, "缺少 "+field+" 文件字段")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, engineUploadLimit))
	if err != nil {
		return nil, "", "", nil, domain.New(domain.CodeDocumentInvalid, "读取上传文件失败")
	}
	mime := http.DetectContentType(data)
	if mime != "image/jpeg" && mime != "image/png" {
		return nil, "", "", nil, domain.New(domain.CodeDocumentUnsupported, "仅支持 JPEG/PNG 图片")
	}
	form := map[string]string{
		"document_type":   r.FormValue("document_type"),
		"issuing_country": r.FormValue("issuing_country"),
	}
	return data, header.Filename, mime, form, nil
}

func stepAfterFront(requiresBack bool) string {
	if requiresBack {
		return "back_document"
	}
	return "live_capture"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func int32Ptr(value int) *int32 {
	v := int32(value)
	return &v
}

func mustJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}

func randomDigits(n int) (string, error) {
	const digits = "0123456789"
	out := make([]byte, n)
	for i := range out {
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			return "", err
		}
		out[i] = digits[int(b[0])%len(digits)]
	}
	return string(out), nil
}

func finalForVoice(score float64) any {
	if score >= 1 {
		return "verified"
	}
	return nil
}

// dispatchVerificationWebhook 在响应后异步投递 webhook，失败落库但不阻塞主请求。
func (s *Server) dispatchVerificationWebhook(verification *gormmodel.VerificationRequest, status, final string) {
	if s.webhooks == nil || s.idswyft == nil || verification.DeveloperID == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		target, err := s.idswyft.WebhookForDeveloper(ctx, *verification.DeveloperID)
		if err != nil {
			return
		}
		payload := map[string]any{
			"event": "verification.completed", "verification_id": verification.ID.String(),
			"status": status, "final_result": final,
		}
		raw, _ := json.Marshal(payload)
		now := time.Now().UTC()
		delivery := &gormmodel.WebhookDelivery{
			ID: uuid.New(), VerificationRequestID: verification.ID, WebhookID: target.ID,
			Payload: raw, Status: "pending", CreatedAt: &now, Attempts: int32Ptr(1),
		}
		if err := s.idswyft.CreateWebhookDelivery(ctx, delivery); err != nil {
			return
		}
		code, err := s.webhooks.Deliver(ctx, target.URL, target.Secret, payload)
		update := map[string]any{"response_status": code, "updated_at": time.Now().UTC()}
		if err == nil {
			update["status"], update["delivered_at"] = "delivered", time.Now().UTC()
		} else {
			update["status"] = "failed"
		}
		_ = s.idswyft.UpdateWebhookDelivery(ctx, delivery.ID, update)
	}()
}
