package web

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"facial-liveness-verification-go/internal/domain"
)

func multipartFiles(t *testing.T, contents [][]byte) []*multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for _, content := range contents {
		f, err := w.CreateFormFile("frames", "frame.jpg")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/", &body)
	r.Header.Set("Content-Type", w.FormDataContentType())
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	return r.MultipartForm.File["frames"]
}

func TestReadFramesRejectsOversizedFrame(t *testing.T) {
	files := multipartFiles(t, [][]byte{[]byte("12345")})
	_, err := readFrames(files, 4)
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeLowQuality {
		t.Fatalf("expected LOW_QUALITY, got %v", err)
	}
}

func TestReadFramesAcceptsWithinLimit(t *testing.T) {
	files := multipartFiles(t, [][]byte{[]byte("1234")})
	out, err := readFrames(files, 4)
	if err != nil || len(out) != 1 || string(out[0]) != "1234" {
		t.Fatalf("out=%q err=%v", out, err)
	}
}

func TestReadFramesRejectsTooMany(t *testing.T) {
	items := make([][]byte, 16)
	for i := range items {
		items[i] = []byte{1}
	}
	_, err := readFrames(multipartFiles(t, items), 1024)
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeAuthInvalid {
		t.Fatalf("expected AUTH_INVALID, got %v", err)
	}
}

func TestUploadErrorMapsMaxBytes(t *testing.T) {
	err := uploadError(&http.MaxBytesError{Limit: 1})
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeServiceUnavailable {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %v", err)
	}
}

func TestNormalizeRequestID(t *testing.T) {
	valid := "req_0123456789abcdef0123"
	if got := normalizeRequestID(valid); got != valid {
		t.Fatalf("valid id changed: %s", got)
	}
	for _, in := range []string{"", "3e30e6fb-b465-4ea1-b119-5ec74620ca0f", "req_nothex", "req_0123"} {
		got := normalizeRequestID(in)
		if len(got) != 24 || got[:4] != "req_" {
			t.Fatalf("normalized id invalid for %q: %s", in, got)
		}
	}
}
