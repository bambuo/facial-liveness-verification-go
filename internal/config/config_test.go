package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestLoadRejectsInvalidValues(t *testing.T) {
	base := map[string]string{
		"APP_ENV":                       "development",
		"AUTH_DEV_SIGNING_KEY":          base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		"AUTH_RESULT_TOKEN_SIGNING_KEY": "result-key",
		"DATABASE_URL":                  "host=localhost",
		"REDIS_ADDR":                    "localhost:6379",
	}
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"invalid int", map[string]string{"FACE_LIVENESS_THRESHOLD": "abc"}, "FACE_LIVENESS_THRESHOLD 必须是数字"},
		{"negative retention", map[string]string{"RETENTION_PROFILE_DAYS": "-1"}, "RETENTION_PROFILE_DAYS 不能为负数"},
		{"frame overflow", map[string]string{"FACE_FRAMES_PER_ACTION": "8", "FACE_ACTIONS_COUNT": "3"}, "不能超过 15"},
		{"missing production master key", map[string]string{"APP_ENV": "production"}, "SECURITY_MASTER_KEY"},
		{"missing result key", map[string]string{"AUTH_RESULT_TOKEN_SIGNING_KEY": ""}, "AUTH_RESULT_TOKEN_SIGNING_KEY"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range base {
				t.Setenv(k, v)
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestLoadValidDevelopmentConfig(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("AUTH_DEV_SIGNING_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	t.Setenv("AUTH_RESULT_TOKEN_SIGNING_KEY", "result-key")
	t.Setenv("DATABASE_URL", "host=localhost")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRejectsLabModeInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("LAB_MODE", "true")
	t.Setenv("DATABASE_URL", "host=localhost")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	if _, err := Load(); err == nil {
		t.Fatal("production + LAB_MODE should fail")
	}
}

func TestLoadWorkerNeedsNoAuthKeys(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "host=localhost")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	if _, err := LoadWorker(); err != nil {
		t.Fatal(err)
	}
}

func TestFingerprintChangesWithThresholds(t *testing.T) {
	base := Config{Face: Face{LivenessThreshold: 0.85, ActionsCount: 3}, LabMode: false}
	a := base.Fingerprint()
	b := base.Fingerprint()
	if a != b {
		t.Fatal("fingerprint should be stable")
	}
	base.Face.LivenessThreshold = 0.91
	if a == base.Fingerprint() {
		t.Fatal("fingerprint should change with liveness threshold")
	}
}
