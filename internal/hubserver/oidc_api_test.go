package hubserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cadentra/cadentra/internal/hub/auth"
)

// newMockDiscovery 只提供 discovery 文档的 mock IdP（NewOIDC 构造时会请求）
func newMockDiscovery(t *testing.T) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"issuer":                  srv.URL,
				"authorization_endpoint":  srv.URL + "/authorize",
				"token_endpoint":          srv.URL + "/token",
				"jwks_uri":                srv.URL + "/jwks",
				"subject_types_supported": []string{"public"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestOIDCStateAndLocalLogin 未配置 OIDC 时：state=false，本地登录正常
func TestOIDCStateAndLocalLogin(t *testing.T) {
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	h, err := New(Config{
		WebAddr:           "127.0.0.1:0",
		GatewayAddr:       "127.0.0.1:0",
		RegistrationToken: "test-token",
		DataDir:           dir,
		ArtifactDir:       dir + "/artifacts",
		BaseURL:           "http://127.0.0.1:18080",
		HeartbeatTimeout:  30 * time.Second,
		AdminUsername:     "admin",
		AdminPassword:     "admin123",
		SessionTTL:        time.Hour,
		RevisionCheckSec:  5,
		ChangelogWindow:   5000,
	}, logger)
	if err != nil {
		t.Fatalf("new hub: %v", err)
	}
	t.Cleanup(func() { h.Close() })
	if err := h.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}
	ts := httptest.NewServer(h.ServeMux())
	defer ts.Close()
	base := ts.URL

	// OIDC state = false
	resp, err := http.Get(base + "/api/oidc/state")
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		Enabled bool `json:"enabled"`
	}
	json.NewDecoder(resp.Body).Decode(&state)
	resp.Body.Close()
	if state.Enabled {
		t.Fatal("oidc state should be false when disabled")
	}

	// 本地登录仍可用
	loginResp, err := http.Post(base+"/api/login", "application/json",
		strings.NewReader(`{"username":"admin","password":"admin123"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != 200 {
		t.Fatalf("local login expected 200, got %d", loginResp.StatusCode)
	}

	// oidc/login 未启用应 404
	oidcLogin, err := http.Get(base + "/api/oidc/login")
	if err != nil {
		t.Fatal(err)
	}
	oidcLogin.Body.Close()
	if oidcLogin.StatusCode != 404 {
		t.Fatalf("oidc/login expected 404 when disabled, got %d", oidcLogin.StatusCode)
	}
}

// TestOIDCEnabled 配置 OIDC 后：state=true，本地登录被 403 拒绝
func TestOIDCEnabled(t *testing.T) {
	idp := newMockDiscovery(t)
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	h, err := New(Config{
		WebAddr:           "127.0.0.1:0",
		GatewayAddr:       "127.0.0.1:0",
		RegistrationToken: "test-token",
		DataDir:           dir,
		ArtifactDir:       dir + "/artifacts",
		BaseURL:           "http://127.0.0.1:18080",
		HeartbeatTimeout:  30 * time.Second,
		AdminUsername:     "admin",
		AdminPassword:     "admin123",
		SessionTTL:        time.Hour,
		RevisionCheckSec:  5,
		ChangelogWindow:   5000,
		OIDC: auth.OIDCConfig{
			Issuer:   idp.URL,
			ClientID: "test-client",
		},
	}, logger)
	if err != nil {
		t.Fatalf("new hub: %v", err)
	}
	t.Cleanup(func() { h.Close() })
	if err := h.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}
	ts := httptest.NewServer(h.ServeMux())
	defer ts.Close()
	base := ts.URL

	// OIDC state = true
	resp, err := http.Get(base + "/api/oidc/state")
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		Enabled bool `json:"enabled"`
	}
	json.NewDecoder(resp.Body).Decode(&state)
	resp.Body.Close()
	if !state.Enabled {
		t.Fatal("oidc state should be true when configured")
	}

	// 本地登录被 403 拒绝
	loginResp, err := http.Post(base+"/api/login", "application/json",
		strings.NewReader(`{"username":"admin","password":"admin123"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != 403 {
		t.Fatalf("local login expected 403 when oidc enabled, got %d", loginResp.StatusCode)
	}

	// oidc/login 应 302 跳转（无 IdP 授权端点则由 discovery 提供 /authorize，返回 404）
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	oidcLogin, err := client.Get(base + "/api/oidc/login")
	if err != nil {
		t.Fatal(err)
	}
	oidcLogin.Body.Close()
	if oidcLogin.StatusCode != http.StatusFound {
		t.Fatalf("oidc/login expected 302, got %d", oidcLogin.StatusCode)
	}
	if loc := oidcLogin.Header.Get("Location"); !strings.HasPrefix(loc, idp.URL+"/authorize") {
		t.Fatalf("redirect location = %q, want %q prefix", loc, idp.URL+"/authorize")
	}
}
