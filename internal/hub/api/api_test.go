package api

import "testing"

func TestValidNodeAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		valid   bool
	}{
		{name: "ipv4", address: "192.0.2.10", valid: true},
		{name: "ipv6", address: "2001:db8::10", valid: true},
		{name: "hostname", address: "agent.example.com", valid: true},
		{name: "single label hostname", address: "node-01", valid: true},
		{name: "trailing dot hostname", address: "agent.example.com.", valid: true},
		{name: "empty", address: "", valid: false},
		{name: "scheme", address: "https://agent.example.com", valid: false},
		{name: "port", address: "agent.example.com:8443", valid: false},
		{name: "invalid label", address: "-agent.example.com", valid: false},
		{name: "underscore", address: "agent_node.example.com", valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validNodeAddress(test.address); got != test.valid {
				t.Fatalf("validNodeAddress(%q) = %v, want %v", test.address, got, test.valid)
			}
		})
	}
}
