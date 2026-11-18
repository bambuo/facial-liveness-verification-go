package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/document/fields"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/ocr"
)

var addressDocumentTypes = map[string]bool{"utility_bill": true, "bank_statement": true, "tax_document": true}

// handleAddressDocument 实现 POST /api/v2/verify/:verification_id/address-document。
func (s *Server) handleAddressDocument(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if s.ocr == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "OCR unavailable"})
		return
	}
	data, filename, mime, form, err := readMultipartImage(w, r, "document")
	if err != nil {
		writeError(w, r, err)
		return
	}
	documentType := form["document_type"]
	if !addressDocumentTypes[documentType] {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "document_type must be one of: utility_bill, bank_statement, tax_document"})
		return
	}
	res, err := s.ocr.RecognizeBytes(data, ocr.Options{})
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "OCR 识别失败"))
		return
	}
	extracted := fields.Extract(res.Text, "")
	identityName := s.identityName(r, verification)
	address := strings.TrimSpace(res.Text)
	nameMatch := 0.0
	if identityName != "" && strings.Contains(strings.ToUpper(address), strings.ToUpper(identityName)) {
		nameMatch = 1
	}
	score := 0.5*nameMatch + 0.5*res.Confidence
	verdict := "REVIEW"
	if nameMatch == 1 && res.Confidence >= 0.5 {
		verdict = "PASS"
	} else if res.Confidence < 0.2 {
		verdict = "FAIL"
	}
	docID := uuid.New()
	payload := map[string]any{
		"document_type": documentType, "document_id": docID.String(),
		"extraction": map[string]any{"name": extracted.FullName, "address": address, "confidence": res.Confidence},
		"validation": map[string]any{"name_match_score": nameMatch, "overall_score": score, "verdict": verdict, "document_fresh": true},
	}
	raw, _ := json.Marshal(payload)
	document := &gormmodel.Document{
		ID: docID, VerificationRequestID: &verification.ID, DocumentType: documentType,
		FileName: filename, FilePath: fmt.Sprintf("address/%s/%s", verification.ID, docID),
		FileSize: int32Ptr(len(data)), MimeType: &mime, OCRData: raw, OCRExtracted: boolPtr(true),
	}
	if err := s.idswyft.CreateDocument(r.Context(), document); err != nil {
		writeError(w, r, err)
		return
	}
	_ = s.idswyft.UpdateVerification(r.Context(), verification.ID, map[string]any{
		"address_verification_status": verdict, "address_data": raw, "address_match_score": score,
		"updated_at": time.Now().UTC(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(),
		"address_verification": map[string]any{
			"status": verdict, "score": score, "name_match_score": nameMatch, "address": address,
			"document_fresh": true, "reasons": []string{},
		},
	})
}

// handleAddressStatus 实现 GET /api/v2/verify/:verification_id/address-status。
func (s *Server) handleAddressStatus(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.AddressVerificationStatus == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"verification_id": verification.ID.String(), "address_verification": nil,
			"message": "No address document has been uploaded for this verification",
		})
		return
	}
	var addressData map[string]any
	_ = json.Unmarshal(verification.AddressData, &addressData)
	validation, _ := addressData["validation"].(map[string]any)
	writeJSON(w, http.StatusOK, map[string]any{
		"verification_id": verification.ID.String(),
		"address_verification": map[string]any{
			"status": *verification.AddressVerificationStatus, "score": verification.AddressMatchScore,
			"name_match_score": validation["name_match_score"], "address": validation["address"],
			"document_type": addressData["document_type"], "document_fresh": validation["document_fresh"],
			"reasons": []string{},
		},
	})
}

func (s *Server) identityName(r *http.Request, verification *gormmodel.VerificationRequest) string {
	if verification.DocumentID == nil {
		return ""
	}
	document, err := s.idswyft.GetDocument(r.Context(), *verification.DocumentID)
	if err != nil {
		return ""
	}
	var data map[string]any
	_ = json.Unmarshal(document.OCRData, &data)
	value, _ := data["full_name"].(string)
	return value
}
