package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/document/fields"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/ocr"
)

type batchItemInput struct {
	UserID           string `json:"user_id"`
	DocumentType     string `json:"document_type"`
	FrontDocumentURL string `json:"front_document_url"`
	BackDocumentURL  string `json:"back_document_url"`
}

// handleBatchUpload 实现 POST /api/v2/batch/upload。
func (s *Server) handleBatchUpload(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		Items []batchItemInput `json:"items"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if len(req.Items) == 0 || len(req.Items) > 1000 {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "items 必须为 1-1000 条"))
		return
	}
	now := time.Now().UTC()
	job := &gormmodel.BatchJob{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, Status: "pending",
		TotalItems: int32(len(req.Items)), CreatedAt: &now,
	}
	items := make([]*gormmodel.BatchItem, 0, len(req.Items))
	for _, item := range req.Items {
		userID, err := uuid.Parse(item.UserID)
		if err != nil {
			writeError(w, r, domain.New(domain.CodeDocumentInvalid, "items.user_id 必须是 UUID"))
			return
		}
		if item.DocumentType != "" && item.DocumentType != "passport" && item.DocumentType != "drivers_license" && item.DocumentType != "national_id" {
			writeError(w, r, domain.New(domain.CodeDocumentInvalid, "document_type 无效"))
			return
		}
		for _, target := range []string{item.FrontDocumentURL, item.BackDocumentURL} {
			if target != "" && !safeHTTPURL(target) {
				writeError(w, r, domain.New(domain.CodeDocumentInvalid, "证件 URL 无效或不允许"))
				return
			}
		}
		raw, _ := json.Marshal(item)
		items = append(items, &gormmodel.BatchItem{
			ID: uuid.New(), UserID: &userID, InputData: raw, Status: "pending", CreatedAt: &now,
		})
	}
	if err := s.idswyft.CreateBatchJob(r.Context(), job, items); err != nil {
		writeError(w, r, err)
		return
	}
	go s.processBatch(job.ID, principal)
	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true, "batch_id": job.ID.String(), "status": "pending",
		"total_items": job.TotalItems, "message": "Batch created and processing started",
	})
}

// processBatch 顺序处理批量条目；失败条目不会中断其他条目。
func (s *Server) processBatch(jobID uuid.UUID, principal *idswyftPrincipal) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	job, err := s.idswyft.GetBatchJob(ctx, principal.DeveloperID, jobID)
	if err != nil {
		return
	}
	items, err := s.idswyft.ListBatchItems(ctx, jobID)
	if err != nil {
		return
	}
	var processed, succeeded, failed int32
	for _, item := range items {
		var input batchItemInput
		_ = json.Unmarshal(item.InputData, &input)
		verificationID, processErr := s.processBatchItem(ctx, input, principal)
		status, itemErr := "completed", ""
		if processErr != nil {
			status, itemErr = "failed", processErr.Error()
		}
		updates := map[string]any{"status": status}
		if processErr != nil {
			updates["error"] = itemErr
			failed++
		} else {
			succeeded++
		}
		if verificationID != nil {
			updates["verification_id"] = *verificationID
		}
		_ = s.idswyft.UpdateBatchItem(ctx, item.ID, updates)
		processed++
	}
	completedAt := time.Now().UTC()
	_ = s.idswyft.UpdateBatchJob(ctx, job.ID, map[string]any{
		"status": "completed", "processed_items": processed, "succeeded_items": succeeded,
		"failed_items": failed, "completed_at": completedAt,
	})
	_ = job
}

func (s *Server) processBatchItem(ctx context.Context, input batchItemInput, principal *idswyftPrincipal) (*uuid.UUID, error) {
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}
	sandbox := principal.Sandbox
	verification := &gormmodel.VerificationRequest{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, APIKeyID: &principal.APIKeyID,
		UserID: &userID, Status: "manual_review", Source: "api", IsSandbox: &sandbox,
		VerificationMode: strPtr("full"), VerificationType: strPtr("document"),
	}
	if err := s.idswyft.CreateVerification(ctx, verification); err != nil {
		return nil, err
	}
	if input.FrontDocumentURL != "" {
		data, err := s.downloadBatchFile(ctx, input.FrontDocumentURL)
		if err != nil {
			return &verification.ID, err
		}
		if s.ocr != nil {
			res, ocrErr := s.ocr.RecognizeBytes(data, ocr.Options{})
			if ocrErr == nil {
				extracted := fields.Extract(res.Text, "")
				raw, _ := json.Marshal(map[string]any{"full_name": extracted.FullName, "date_of_birth": extracted.DateOfBirth, "id_number": extracted.IDNumber, "expiry_date": extracted.ExpiryDate, "raw_text": res.Text})
				document := &gormmodel.Document{
					ID: uuid.New(), VerificationRequestID: &verification.ID,
					DocumentType: firstNonEmpty(input.DocumentType, "drivers_license"),
					FileName:     "front_document", FilePath: fmt.Sprintf("batch/%s/front-%s", verification.ID, uuid.New()),
					FileSize: int32Ptr(len(data)), OCRData: raw, OCRExtracted: boolPtr(true), IsBackOfID: boolPtr(false),
				}
				_ = s.idswyft.CreateDocument(ctx, document)
				_ = s.idswyft.UpdateVerification(ctx, verification.ID, map[string]any{"document_id": document.ID})
			}
		}
	}
	return &verification.ID, nil
}

func allowedIP(ip net.IP) bool {
	return !(ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsLinkLocalMulticast())
}

func (s *Server) downloadBatchFile(ctx context.Context, target string) ([]byte, error) {
	if !safeHTTPURL(target) {
		return nil, domain.New(domain.CodeDocumentInvalid, "文件 URL 不允许")
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("下载文件失败: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 10<<20 {
		return nil, fmt.Errorf("文件超过 10MB")
	}
	return data, nil
}

func safeHTTPURL(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return false
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return allowedIP(ip)
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if !allowedIP(ip) {
			return false
		}
	}
	return true
}

// handleBatchStatus 实现 GET /api/v2/batch/:id/status。
func (s *Server) handleBatchStatus(w http.ResponseWriter, r *http.Request) {
	job, ok := s.batchJob(w, r)
	if !ok {
		return
	}
	progress := int32(0)
	if job.TotalItems > 0 {
		progress = job.ProcessedItems * 100 / job.TotalItems
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"batch_id": job.ID.String(), "status": job.Status, "total_items": job.TotalItems,
		"processed_items": job.ProcessedItems, "succeeded_items": job.SucceededItems,
		"failed_items": job.FailedItems, "progress_percentage": progress,
		"created_at": job.CreatedAt, "completed_at": job.CompletedAt,
	})
}

// handleBatchResults 实现 GET /api/v2/batch/:id/results。
func (s *Server) handleBatchResults(w http.ResponseWriter, r *http.Request) {
	job, ok := s.batchJob(w, r)
	if !ok {
		return
	}
	items, err := s.idswyft.ListBatchItems(r.Context(), job.ID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	results := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry := map[string]any{"item_id": item.ID.String(), "status": item.Status}
		if item.VerificationID != nil {
			entry["verification_id"] = item.VerificationID.String()
		}
		if item.Error != nil {
			entry["error"] = *item.Error
		}
		results = append(results, entry)
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

// handleBatchCancel 实现 POST /api/v2/batch/:id/cancel。
func (s *Server) handleBatchCancel(w http.ResponseWriter, r *http.Request) {
	job, ok := s.batchJob(w, r)
	if !ok {
		return
	}
	if job.Status == "completed" || job.Status == "cancelled" || job.Status == "canceled" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Cannot cancel batch", "message": "Batch is already " + job.Status})
		return
	}
	if err := s.idswyft.UpdateBatchJob(r.Context(), job.ID, map[string]any{"status": "cancelled"}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Batch cancellation initiated"})
}

// handleBatchList 实现 GET /api/v2/batch。
func (s *Server) handleBatchList(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	jobs, total, err := s.idswyft.ListBatchJobs(r.Context(), principal.DeveloperID, page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"jobs":       jobs,
		"pagination": map[string]any{"page": page, "limit": limit, "total": total, "pages": (total + int64(limit) - 1) / int64(limit)},
	})
}

func (s *Server) batchJob(w http.ResponseWriter, r *http.Request) (*gormmodel.BatchJob, bool) {
	principal := principalFrom(r)
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return nil, false
	}
	job, err := s.idswyft.GetBatchJob(r.Context(), principal.DeveloperID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Batch job not found"})
		return nil, false
	}
	return job, true
}
