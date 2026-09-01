package hubserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestAPILifecycle 通过真实 HTTP 测试 API 全流程
func TestAPILifecycle(t *testing.T) {
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

	// Login
	loginResp, err := http.Post(base+"/api/login", "application/json",
		strings.NewReader(`{"username":"admin","password":"admin123"}`))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	json.NewDecoder(loginResp.Body).Decode(&loginBody)
	loginResp.Body.Close()
	if loginBody.Token == "" {
		t.Fatalf("no token returned")
	}

	do := func(method, path, body string, wantStatus int) string {
		var req *http.Request
		var err error
		if body == "" {
			req, err = http.NewRequest(method, base+path, nil)
		} else {
			req, err = http.NewRequest(method, base+path, strings.NewReader(body))
		}
		if err != nil {
			t.Fatalf("req: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+loginBody.Token)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("do %s %s: %v", method, path, err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if resp.StatusCode != wantStatus {
			t.Fatalf("%s %s: expected %d, got %d (body %s)", method, path, wantStatus, resp.StatusCode, string(b))
		}
		return string(b)
	}

	// Script CRUD
	sc := do("POST", "/api/scripts", `{"name":"test-script","interpreter":"shell","content":"echo test","enabled":true}`, 200)
	var script struct {
		ID string `json:"id"`
	}
	json.Unmarshal([]byte(sc), &script)
	if script.ID == "" {
		t.Fatalf("script id empty")
	}
	do("GET", "/api/scripts", "", 200)
	do("PUT", "/api/scripts/"+script.ID, `{"name":"test-script","interpreter":"shell","content":"echo updated","enabled":true}`, 200)
	do("GET", "/api/scripts/"+script.ID+"/revisions/1", "", 200)

	// Task CRUD
	tk := do("POST", "/api/tasks", `{"name":"test-task","type":"script","script_id":"`+script.ID+`","target":{"type":"label"},"enabled":true,"timeout":30}`, 200)
	var task struct {
		ID string `json:"id"`
	}
	json.Unmarshal([]byte(tk), &task)

	// 引用不存在的 script 应失败
	do("POST", "/api/tasks", `{"name":"bad","type":"script","script_id":"nonexistent","enabled":true}`, 400)

	// Schedules
	do("POST", "/api/schedules", `{"task_id":"`+task.ID+`","type":"interval","interval_sec":60,"timezone":"UTC","execution_owner":"agent","enabled":true}`, 200)

	// Settings
	do("GET", "/api/settings", "", 200)
	do("PUT", "/api/settings", `{"max_log_bytes":"1048576"}`, 200)
	app := do("POST", "/api/applications", `{"name":"state-app","version":"1.0","binary_path":"/usr/local/bin/state-app"}`, 200)
	var application struct {
		ID string `json:"id"`
	}
	json.Unmarshal([]byte(app), &application)
	do("GET", "/api/applications/"+application.ID+"/state", "", 200)
	do("GET", "/api/applications/"+application.ID+"/executions", "", 200)

	// Audit
	do("GET", "/api/audit", "", 200)

	// Health
	do("GET", "/healthz", "", 200)
	do("GET", "/readyz", "", 200)

	// 未认证访问应 401
	unauth, err := http.Get(base + "/api/nodes")
	if err != nil {
		t.Fatalf("unauth req: %v", err)
	}
	unauth.Body.Close()
	if unauth.StatusCode != 401 {
		t.Fatalf("expected 401 for unauth, got %d", unauth.StatusCode)
	}

	// 删除
	do("DELETE", "/api/tasks/"+task.ID, "", 200)
	do("DELETE", "/api/scripts/"+script.ID, "", 200)
	do("DELETE", "/api/applications/"+application.ID, "", 200)
}
