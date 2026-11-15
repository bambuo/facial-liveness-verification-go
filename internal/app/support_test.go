package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
)

type fakeGuard struct {
	locked  bool
	resets  int
	records int
}

func (f *fakeGuard) IsLocked(context.Context, uuid.UUID, string, string) (bool, error) {
	return f.locked, nil
}

func (f *fakeGuard) RecordFailure(context.Context, uuid.UUID, string, string) (bool, error) {
	f.records++
	return f.locked, nil
}

func (f *fakeGuard) Reset(context.Context, uuid.UUID, string, string) error {
	f.resets++
	return nil
}

func guardAuth() *AuthContext {
	return &AuthContext{
		Tenant: &domain.Tenant{ID: uuid.New(), Code: "t"},
		Claims: domain.EndpointClaims{Subject: "subject-1"},
	}
}

func TestRecordLivenessResetsOnReal(t *testing.T) {
	guard := &fakeGuard{}
	support := NewVerificationSupport(nil, nil, nil, nil, nil, nil, guard)
	support.RecordLiveness(context.Background(), guardAuth(), "device", domain.LivenessReal)
	if guard.resets != 1 || guard.records != 0 {
		t.Fatalf("resets=%d records=%d", guard.resets, guard.records)
	}
}

func TestRecordLivenessCountsFailures(t *testing.T) {
	guard := &fakeGuard{}
	support := NewVerificationSupport(nil, nil, nil, nil, nil, nil, guard)
	support.RecordLiveness(context.Background(), guardAuth(), "device", domain.LivenessSpoof)
	support.RecordLiveness(context.Background(), guardAuth(), "device", domain.LivenessFuzzy)
	if guard.records != 2 || guard.resets != 0 {
		t.Fatalf("resets=%d records=%d", guard.resets, guard.records)
	}
}

func TestBeginUploadRejectsLockedGuard(t *testing.T) {
	guard := &fakeGuard{locked: true}
	support := NewVerificationSupport(nil, nil, nil, nil, nil, nil, guard)
	_, _, err := support.BeginUpload(context.Background(), guardAuth(), "ch-1", "device", domain.ModeVerify)
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeQuotaExceeded {
		t.Fatalf("expected QUOTA_EXCEEDED, got %v", err)
	}
}
