package hub

import (
	"context"
	"log/slog"
	"testing"

	"github.com/cadentra/cadentra/internal/models"
	"github.com/cadentra/cadentra/internal/protocol"
	"github.com/cadentra/cadentra/internal/store"
)

type fakeAgentConn struct {
	nodeID string
}

func (f *fakeAgentConn) NodeID() string { return f.nodeID }
func (f *fakeAgentConn) Send(msg protocol.Envelope) error {
	return nil
}
func (f *fakeAgentConn) Close() error { return nil }

func newTestGateway(t *testing.T) (*Gateway, *NodeManager, *SessionManager) {
	t.Helper()
	st, err := store.OpenInMemory()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	revisions := NewRevisionManager(st)
	nm := NewNodeManager(st, revisions)
	sm := NewSessionManager()
	g := NewGateway(GatewayConfig{}, nm, sm, nil, nil, slog.New(slog.DiscardHandler))
	return g, nm, sm
}

// TestHandleDisconnectMarksOffline 验证 Bug F：Agent 连接断开后
// 节点状态应被标记为 offline（此前仅 Unregister 会话，状态滞留 online）。
func TestHandleDisconnectMarksOffline(t *testing.T) {
	ctx := context.Background()
	g, nm, sm := newTestGateway(t)

	node, _, _, err := nm.RegisterOrUpdate(ctx, "agent-1", "host-a", "10.0.0.1", "linux", "amd64", "1.0", "native", false, nil)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if node.Status != models.NodeStatusOnline {
		t.Fatalf("expect online after register, got %s", node.Status)
	}

	conn := &fakeAgentConn{nodeID: node.ID}
	sm.Register(conn)
	g.handleDisconnect(node.ID, conn)

	got, err := nm.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if got.Status != models.NodeStatusOffline {
		t.Fatalf("expect offline after disconnect, got %s", got.Status)
	}
	if _, ok := sm.Get(node.ID); ok {
		t.Fatalf("session should be unregistered after disconnect")
	}
}

// TestHandleDisconnectSkipsActiveReconnect 验证重连竞态：旧连接断开时
// 若已有新连接替换为当前活跃会话，不得把节点误标为 offline。
func TestHandleDisconnectSkipsActiveReconnect(t *testing.T) {
	ctx := context.Background()
	g, nm, sm := newTestGateway(t)

	node, _, _, err := nm.RegisterOrUpdate(ctx, "agent-2", "host-b", "10.0.0.2", "linux", "amd64", "1.0", "native", false, nil)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	oldConn := &fakeAgentConn{nodeID: node.ID}
	sm.Register(oldConn)
	// 新连接建立（SessionManager 注册时旧连接被替换）
	newConn := &fakeAgentConn{nodeID: node.ID}
	sm.Register(newConn)

	// 旧连接断开，但当前活跃会话已是新连接
	g.handleDisconnect(node.ID, oldConn)

	got, err := nm.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if got.Status != models.NodeStatusOnline {
		t.Fatalf("expect node stays online with active session, got %s", got.Status)
	}
	if _, ok := sm.Get(node.ID); !ok {
		t.Fatalf("new session should remain registered")
	}
}
