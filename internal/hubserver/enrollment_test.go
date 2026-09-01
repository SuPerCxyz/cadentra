package hubserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cadentra/cadentra/internal/models"
)

func TestNodeEnrollmentMetadata(t *testing.T) {
	dir := t.TempDir()
	h, err := New(Config{
		RegistrationToken: "registration-test",
		DataDir:           dir,
		ArtifactDir:       dir + "/artifacts",
		BaseURL:           "http://hub.example:8080",
		GatewayAddr:       ":8443",
		HeartbeatTimeout:  30 * time.Second,
		AdminUsername:     "admin",
		AdminPassword:     "admin123",
		SessionTTL:        time.Hour,
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	if err := h.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h.ServeMux())
	defer ts.Close()
	login, err := http.Post(ts.URL+"/api/login", "application/json", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer login.Body.Close()
	var session struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(login.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	params := url.Values{
		"node_name":   {"edge-node-01"},
		"node_ip":     {"203.0.113.10"},
		"hub_address": {"http://public.example:8080"},
	}
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/nodes/enrollment?"+params.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("enrollment status %d: %s", resp.StatusCode, body)
	}
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if result["gateway_url"] != "ws://public.example:8443" ||
		!strings.Contains(result["native"], "node_name: \"edge-node-01\"") ||
		!strings.Contains(result["native"], "node_ip: \"203.0.113.10\"") ||
		!strings.Contains(result["native"], "registration_token: \"registration-test\"") ||
		!strings.Contains(result["docker_run"], "CADENTRA_REGISTRATION_TOKEN='registration-test'") ||
		!strings.Contains(result["docker_run"], "CADENTRA_NODE_NAME='edge-node-01'") ||
		!strings.Contains(result["docker_run"], "CADENTRA_NODE_IP='203.0.113.10'") ||
		!strings.Contains(result["docker_run"], "CADENTRA_DEPLOYMENT_MODE=docker") ||
		!strings.Contains(result["docker_compose"], "CADENTRA_NODE_NAME: \"edge-node-01\"") ||
		!strings.Contains(result["docker_compose"], "CADENTRA_NODE_IP: \"203.0.113.10\"") ||
		!strings.Contains(result["docker_compose"], "CADENTRA_DEPLOYMENT_MODE: docker") {
		t.Fatalf("unexpected enrollment metadata: %+v", result)
	}
}

func TestNodeEnrollmentRejectsInvalidIdentity(t *testing.T) {
	dir := t.TempDir()
	h, err := New(Config{
		RegistrationToken: "registration-test",
		DataDir:           dir,
		ArtifactDir:       dir + "/artifacts",
		BaseURL:           "http://hub.example:8080",
		GatewayAddr:       ":8443",
		HeartbeatTimeout:  30 * time.Second,
		AdminUsername:     "admin",
		AdminPassword:     "admin123",
		SessionTTL:        time.Hour,
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	if err := h.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h.ServeMux())
	defer ts.Close()
	login, err := http.Post(ts.URL+"/api/login", "application/json", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		Token string `json:"token"`
	}
	json.NewDecoder(login.Body).Decode(&session)
	login.Body.Close()
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/nodes/enrollment?node_name=bad%20name&node_ip=not-an-ip", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", resp.StatusCode)
	}
}

func TestFileTransferAPI(t *testing.T) {
	dir := t.TempDir()
	h, err := New(Config{
		RegistrationToken: "registration-test",
		DataDir:           dir,
		ArtifactDir:       dir + "/artifacts",
		BaseURL:           "http://hub.example:8080",
		GatewayAddr:       ":8443",
		HeartbeatTimeout:  30 * time.Second,
		AdminUsername:     "admin",
		AdminPassword:     "admin123",
		SessionTTL:        time.Hour,
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	ctx := context.Background()
	if err := h.EnsureDefaultAdmin(ctx); err != nil {
		t.Fatal(err)
	}
	source, _, _, err := h.Nodes().RegisterOrUpdate(ctx, "source-agent", "source", "10.0.0.1", "linux", "amd64", "1", models.DeploymentModeNative, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	target, _, _, err := h.Nodes().RegisterOrUpdate(ctx, "target-agent", "target", "10.0.0.2", "linux", "amd64", "1", models.DeploymentModeNative, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h.ServeMux())
	defer ts.Close()
	login, err := http.Post(ts.URL+"/api/login", "application/json", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		Token string `json:"token"`
	}
	json.NewDecoder(login.Body).Decode(&session)
	login.Body.Close()
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/transfers", strings.NewReader(`{"source_node_id":"`+source.ID+`","source_path":"/source","targets":[{"node_id":"`+target.ID+`","destination_path":"/target"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var transfer models.FileTransfer
	json.NewDecoder(resp.Body).Decode(&transfer)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || transfer.ID == "" || transfer.Status != models.FileTransferPending {
		t.Fatalf("create transfer failed: status=%d transfer=%+v", resp.StatusCode, transfer)
	}
	unauth, err := http.Get(ts.URL + "/api/transfers")
	if err != nil {
		t.Fatal(err)
	}
	unauth.Body.Close()
	if unauth.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized transfer list, got %d", unauth.StatusCode)
	}
}
