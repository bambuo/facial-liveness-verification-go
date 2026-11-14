package domain

import "fmt"

// Code 是对外错误码。
type Code string

const (
	CodeAuthInvalid           Code = "AUTH_INVALID"
	CodeTenantDisabled        Code = "TENANT_DISABLED"
	CodeQuotaExceeded         Code = "QUOTA_EXCEEDED"
	CodeChallengeExpired      Code = "CHALLENGE_EXPIRED"
	CodeChallengeReplay       Code = "CHALLENGE_REPLAY"
	CodeNoFace                Code = "NO_FACE"
	CodeMultiFace             Code = "MULTI_FACE"
	CodeLowQuality            Code = "LOW_QUALITY"
	CodeLivenessFailed        Code = "LIVENESS_FAILED"
	CodeActionMismatch        Code = "ACTION_MISMATCH"
	CodeFaceMismatch          Code = "FACE_MISMATCH"
	CodeFaceNotEnrolled       Code = "FACE_NOT_ENROLLED"
	CodeGalleryNotFound       Code = "GALLERY_NOT_FOUND"
	CodeIdentifyReview        Code = "IDENTIFY_REVIEW"
	CodeDocumentUnsupported   Code = "DOCUMENT_UNSUPPORTED"
	CodeDocumentLowConfidence Code = "DOCUMENT_LOW_CONFIDENCE"
	CodeDocumentInvalid       Code = "DOCUMENT_INVALID"
	CodeServiceUnavailable    Code = "SERVICE_UNAVAILABLE"
	CodeDevDisabled           Code = "DEV_DISABLED"
	CodeOperationNotFound     Code = "OPERATION_NOT_FOUND"
)

// Error 是业务拒绝，由 HTTP 层统一转换为响应信封。
type Error struct {
	Code    Code
	Message string
	Status  int
}

// New 构造业务错误，按错误码映射 HTTP 状态。
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message, Status: statusOf(code)}
}

// Of 构造只带错误码与默认文案的业务错误。
func Of(code Code) *Error { return New(code, string(code)) }

// Error 实现 error 接口。
func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// statusOf 返回错误码对应的 HTTP 状态。
func statusOf(code Code) int {
	switch code {
	case CodeAuthInvalid:
		return 401
	case CodeTenantDisabled:
		return 403
	case CodeQuotaExceeded:
		return 429
	case CodeChallengeExpired, CodeChallengeReplay:
		return 409
	case CodeNoFace, CodeMultiFace, CodeLowQuality, CodeLivenessFailed,
		CodeActionMismatch, CodeFaceMismatch, CodeDocumentUnsupported,
		CodeDocumentLowConfidence, CodeDocumentInvalid:
		return 422
	case CodeFaceNotEnrolled, CodeGalleryNotFound, CodeDevDisabled, CodeOperationNotFound:
		return 404
	case CodeIdentifyReview:
		return 200
	case CodeServiceUnavailable:
		return 503
	default:
		return 500
	}
}
