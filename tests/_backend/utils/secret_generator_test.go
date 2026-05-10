package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRandomSecret(t *testing.T) {
	secret, err := GenerateRandomSecret(50)
	if err != nil {
		t.Fatalf("GenerateRandomSecret returned error: %v", err)
	}
	if len(secret) != 50 {
		t.Fatalf("secret length = %d, want 50", len(secret))
	}
	if strings.ContainsAny(secret, "+/=") {
		t.Fatalf("secret should be raw URL-safe base64, got %q", secret)
	}
}

func TestEnsureJWTSecretKeepsExistingSecret(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	existing := "JWT_SECRET=already-configured-secret-value-1234567890\nOTHER=value\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureJWTSecret(); err != nil {
		t.Fatalf("EnsureJWTSecret returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != existing {
		t.Fatalf("existing secret should remain unchanged, got:\n%s", string(content))
	}
}

func TestEnsureJWTSecretCreatesSecret(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	if err := EnsureJWTSecret(); err != nil {
		t.Fatalf("EnsureJWTSecret returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, "JWT_SECRET=") {
		t.Fatalf("expected generated JWT_SECRET, got:\n%s", text)
	}
	if len(strings.TrimSpace(os.Getenv("JWT_SECRET"))) != 50 {
		t.Fatalf("expected JWT_SECRET env var length 50, got %d", len(strings.TrimSpace(os.Getenv("JWT_SECRET"))))
	}
}
