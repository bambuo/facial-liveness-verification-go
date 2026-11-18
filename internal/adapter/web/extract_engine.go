package web

import (
	"io"
	"net/http"

	"facial-liveness-verification-go/internal/audio"
	"facial-liveness-verification-go/internal/document/fields"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/ocr"
)

// handleExtractBack 实现 POST /extract/back：OCR 后解析 MRZ；无条码能力时返回空条码字段。
func (s *Server) handleExtractBack(w http.ResponseWriter, r *http.Request) {
	if s.ocr == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OCR 未启用"))
		return
	}
	data, _, err := readEngineUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	res, err := s.ocr.RecognizeBytes(data, ocr.Options{})
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "OCR 识别失败"))
		return
	}
	result := map[string]any{
		"qr_payload":       nil,
		"mrz_result":       nil,
		"barcode_format":   nil,
		"raw_barcode_data": nil,
	}
	if mrz, ok := fields.ExtractMRZ(res.Text); ok {
		fieldsMap := map[string]string{
			"full_name":       mrz.Fields.FullName,
			"document_number": mrz.Fields.IDNumber,
			"date_of_birth":   mrz.Fields.DateOfBirth,
			"expiry_date":     mrz.Fields.ExpiryDate,
		}
		if mrz.Fields.IssuingCountry != "" {
			fieldsMap["issuing_country"] = mrz.Fields.IssuingCountry
		}
		result["qr_payload"] = map[string]any{
			"full_name":     mrz.Fields.FullName,
			"date_of_birth": mrz.Fields.DateOfBirth,
			"id_number":     mrz.Fields.IDNumber,
			"expiry_date":   mrz.Fields.ExpiryDate,
		}
		result["mrz_result"] = map[string]any{
			"raw_lines":       mrz.RawLines,
			"fields":          fieldsMap,
			"checksums_valid": mrz.ChecksumsValid,
		}
		result["barcode_format"] = "MRZ_TD3"
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "result": result})
}

// handleExtractLive 实现 POST /extract/live：人脸检测、特征与被动活体。
func (s *Server) handleExtractLive(w http.ResponseWriter, r *http.Request) {
	if s.face == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "人脸引擎未启用"))
		return
	}
	data, _, err := readEngineUpload(w, r)
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
	var embedding []float32
	if vec, embedErr := s.face.ExtractLargest([][]byte{data}); embedErr == nil {
		embedding = vec
	} else {
		faceConfidence = 0
	}
	livenessScore := 0.0
	livenessPassed := false
	if vote, livenessErr := s.face.CheckLiveness([][]byte{data}, s.cfg.Face.LivenessThreshold, s.cfg.Face.LiveRatioThreshold); livenessErr == nil {
		livenessScore = vote.MinScore
		livenessPassed = vote.Status == domain.LivenessReal
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"result": map[string]any{
			"face_embedding":  embedding,
			"face_confidence": faceConfidence,
			"liveness_passed": livenessPassed,
			"liveness_score":  livenessScore,
		},
	})
}

// handleVoiceEnroll 实现 POST /extract/voice-enroll：返回 CAM++ 声纹向量。
func (s *Server) handleVoiceEnroll(w http.ResponseWriter, r *http.Request) {
	if s.voice == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "语音引擎未启用"))
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
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"result": map[string]any{
			"speaker_embedding":   embedding,
			"embedding_dimension": len(embedding),
		},
	})
}

// handleVoiceVerify 实现 POST /extract/voice-verify：声纹向量 + 数字转写。
func (s *Server) handleVoiceVerify(w http.ResponseWriter, r *http.Request) {
	if s.voice == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "语音引擎未启用"))
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
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"result": map[string]any{
			"speaker_embedding":   embedding,
			"embedding_dimension": len(embedding),
			"transcription":       transcription,
		},
	})
}

// readVoiceUpload 读取 multipart 音频并转成 16kHz 单声道样本；当前接受 WAV。
func readVoiceUpload(w http.ResponseWriter, r *http.Request) ([]float32, error) {
	r.Body = http.MaxBytesReader(w, r.Body, engineUploadLimit)
	if err := r.ParseMultipartForm(engineUploadLimit); err != nil {
		return nil, domain.New(domain.CodeDocumentInvalid, "请求体必须为 multipart/form-data 且不超过 10MB")
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, domain.New(domain.CodeDocumentInvalid, "缺少 file 文件字段")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, engineUploadLimit))
	if err != nil {
		return nil, domain.New(domain.CodeDocumentInvalid, "读取上传音频失败")
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, domain.New(domain.CodeDocumentUnsupported, "仅支持 WAV 音频")
	}
	samples, err := audio.DecodeUploadAudio(data)
	if err != nil {
		return nil, domain.New(domain.CodeDocumentInvalid, "音频解码失败")
	}
	return samples, nil
}
