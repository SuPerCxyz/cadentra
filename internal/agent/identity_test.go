package agent

import (
	"log/slog"
	"testing"
)

func TestHelloUsesConfiguredNodeIdentity(t *testing.T) {
	a, err := New(Config{
		HubURL:            "ws://127.0.0.1:8443",
		RegistrationToken: "test-token",
		NodeName:          "edge-node-01",
		NodeIP:            "203.0.113.10",
		DataDir:           t.TempDir(),
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	hello := a.helloFn()
	if hello.Hostname != "edge-node-01" || hello.IP != "203.0.113.10" {
		t.Fatalf("unexpected configured identity: hostname=%q ip=%q", hello.Hostname, hello.IP)
	}
}
