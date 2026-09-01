package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestArtifactCacheSHAFailure 校验失败不得安装
func TestArtifactCacheSHAFailure(t *testing.T) {
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st := newTestAgentStore(t)
	cache := NewArtifactCache(filepath.Join(dir, "artifacts"), st, logger, "")

	// 服务端返回错误内容（SHA 不匹配）
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("corrupted-content"))
	}))
	defer srv.Close()

	wantSHA := strings.Repeat("a", 64)
	err := cache.Prefetch(context.Background(), srv.URL, wantSHA)
	if err == nil {
		t.Fatalf("expected SHA mismatch error")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("expected mismatch error, got: %v", err)
	}
	// 缓存文件不应存在
	if cache.Has(wantSHA) {
		t.Fatalf("corrupted artifact must not be cached")
	}
	// .tmp 应被清理
	matches, _ := filepath.Glob(filepath.Join(dir, "artifacts", "*.tmp"))
	if len(matches) > 0 {
		t.Fatalf("tmp files not cleaned: %v", matches)
	}
}

// TestArtifactCacheSuccess 成功下载并原子缓存
func TestArtifactCacheSuccess(t *testing.T) {
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st := newTestAgentStore(t)
	cache := NewArtifactCache(filepath.Join(dir, "artifacts"), st, logger, "")

	content := []byte("hello-artifact")
	h := sha256.Sum256(content)
	sha := hex.EncodeToString(h[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	if err := cache.Prefetch(context.Background(), srv.URL, sha); err != nil {
		t.Fatalf("prefetch: %v", err)
	}
	if !cache.Has(sha) {
		t.Fatalf("artifact should be cached")
	}
	f, err := cache.Open(sha)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	if string(data) != string(content) {
		t.Fatalf("content mismatch")
	}
	// 再次 Prefetch 应命中缓存
	if err := cache.Prefetch(context.Background(), srv.URL, sha); err != nil {
		t.Fatalf("second prefetch: %v", err)
	}
}

func TestArtifactCacheRejectsHTTPError(t *testing.T) {
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st := newTestAgentStore(t)
	cache := NewArtifactCache(filepath.Join(dir, "artifacts"), st, logger, "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()
	if err := cache.Prefetch(context.Background(), srv.URL, strings.Repeat("b", 64)); err == nil {
		t.Fatal("expected non-200 download to fail")
	}
}

var _ = os.Getenv
