package hub

import (
	"context"
	"errors"
	"time"

	"github.com/cadentra/cadentra/internal/models"
	"github.com/cadentra/cadentra/internal/store"
	"github.com/google/uuid"
)

// ErrNodeNotFound 节点不存在
var ErrNodeNotFound = errors.New("node not found")

// NodeManager 节点管理
type NodeManager struct {
	store     store.Store
	revisions *RevisionManager
	syncMgr   *SyncManager
}

// SetSyncManager 注入同步管理器，供节点目标变化触发全局收敛。
func (nm *NodeManager) SetSyncManager(sm *SyncManager) { nm.syncMgr = sm }

// NewNodeManager 创建节点管理器
func NewNodeManager(st store.Store, rm *RevisionManager) *NodeManager {
	return &NodeManager{store: st, revisions: rm}
}

// RegisterAgent 注册新 Agent（返回 node/agent id 与凭证）
func (nm *NodeManager) RegisterAgent(ctx context.Context, hostname, ip, os, arch, agentVersion, mode string, hostInt bool, caps map[string]bool) (*models.Node, string, error) {
	nodeID := uuid.NewString()
	agentID := "agent-" + uuid.NewString()
	cred := "cred-" + uuid.NewString()
	n := &models.Node{
		ID:              nodeID,
		AgentID:         agentID,
		Hostname:        hostname,
		IP:              ip,
		OS:              os,
		Arch:            arch,
		AgentVersion:    agentVersion,
		DeploymentMode:  mode,
		HostIntegration: hostInt,
		Status:          models.NodeStatusOnline,
		Labels:          map[string]string{},
		Capabilities:    caps,
		LastSeen:        time.Now(),
		FirstSeen:       time.Now(),
		SyncStatus:      "outdated",
	}
	if err := nm.store.UpsertNode(ctx, n); err != nil {
		return nil, "", err
	}
	nm.store.SetSetting(ctx, "cred:"+nodeID, cred)
	return n, cred, nil
}

// AuthenticateAgent 校验 Agent 凭证
func (nm *NodeManager) AuthenticateAgent(ctx context.Context, nodeID, cred string) bool {
	stored, err := nm.store.GetSetting(ctx, "cred:"+nodeID)
	if err != nil {
		return false
	}
	return stored == cred
}

// RegisterOrUpdate 注册或更新节点（幂等）
func (nm *NodeManager) RegisterOrUpdate(ctx context.Context, agentID, hostname, ip, os, arch, agentVersion, mode string, hostInt bool, caps map[string]bool) (*models.Node, string, bool, error) {
	existing, err := nm.store.GetNodeByAgentID(ctx, agentID)
	if err == nil {
		existing.Hostname = hostname
		existing.IP = ip
		existing.OS = os
		existing.Arch = arch
		existing.AgentVersion = agentVersion
		existing.Capabilities = caps
		existing.DeploymentMode = mode
		existing.HostIntegration = hostInt
		existing.LastSeen = time.Now()
		existing.Status = models.NodeStatusOnline
		if err := nm.store.UpsertNode(ctx, existing); err != nil {
			return nil, "", false, err
		}
		cred, _ := nm.store.GetSetting(ctx, "cred:"+existing.ID)
		return existing, cred, false, nil
	}
	n, cred, err := nm.RegisterAgent(ctx, hostname, ip, os, arch, agentVersion, mode, hostInt, caps)
	if err != nil {
		return nil, "", false, err
	}
	return n, cred, true, nil
}

// GetNode 获取节点
func (nm *NodeManager) GetNode(ctx context.Context, id string) (*models.Node, error) {
	return nm.store.GetNode(ctx, id)
}

// ListNodes 节点列表
func (nm *NodeManager) ListNodes(ctx context.Context) ([]*models.Node, error) {
	return nm.store.ListNodes(ctx)
}

// SetNodeStatus 设置节点状态
func (nm *NodeManager) SetNodeStatus(ctx context.Context, id, status string) error {
	return nm.store.UpdateNodeStatus(ctx, id, status)
}

// UpdateHeartbeat 心跳更新
func (nm *NodeManager) UpdateHeartbeat(ctx context.Context, id string, lastSeen time.Time) error {
	return nm.store.UpdateNodeHeartbeat(ctx, id, lastSeen)
}

// UpdateSyncState 同步状态更新
func (nm *NodeManager) UpdateSyncState(ctx context.Context, id string, rev int64, syncStatus string) error {
	return nm.store.UpdateNodeSyncState(ctx, id, rev, syncStatus)
}

// SetLabels 设置标签
func (nm *NodeManager) SetLabels(ctx context.Context, id string, labels map[string]string) error {
	var rev int64
	if err := runMutationTx(ctx, nm.store, func(txctx context.Context) error {
		if err := nm.store.SetNodeLabels(txctx, id, labels); err != nil {
			return err
		}
		var err error
		rev, err = nm.recordTargetChange(txctx, models.ObjectNode, id, "labels")
		return err
	}); err != nil {
		return err
	}
	_ = rev
	if nm.syncMgr != nil {
		nm.syncMgr.NotifyAll(ctx)
	}
	return nil
}

// CreateGroup 创建节点组并推进 Desired State Revision。
func (nm *NodeManager) CreateGroup(ctx context.Context, g *models.Group) error {
	var rev int64
	if err := runMutationTx(ctx, nm.store, func(txctx context.Context) error {
		if err := nm.store.CreateGroup(txctx, g); err != nil {
			return err
		}
		var err error
		rev, err = nm.recordTargetChange(txctx, models.ObjectGroup, g.ID, "create")
		return err
	}); err != nil {
		return err
	}
	_ = rev
	if nm.syncMgr != nil {
		nm.syncMgr.NotifyAll(ctx)
	}
	return nil
}

// UpdateGroup 更新节点组并推进 Desired State Revision。
func (nm *NodeManager) UpdateGroup(ctx context.Context, g *models.Group) error {
	var rev int64
	if err := runMutationTx(ctx, nm.store, func(txctx context.Context) error {
		if err := nm.store.UpdateGroup(txctx, g); err != nil {
			return err
		}
		var err error
		rev, err = nm.recordTargetChange(txctx, models.ObjectGroup, g.ID, "update")
		return err
	}); err != nil {
		return err
	}
	_ = rev
	if nm.syncMgr != nil {
		nm.syncMgr.NotifyAll(ctx)
	}
	return nil
}

// DeleteGroup 删除节点组并推进 Desired State Revision。
func (nm *NodeManager) DeleteGroup(ctx context.Context, id string) error {
	var rev int64
	if err := runMutationTx(ctx, nm.store, func(txctx context.Context) error {
		if err := nm.store.DeleteGroup(txctx, id); err != nil {
			return err
		}
		var err error
		rev, err = nm.recordTargetChange(txctx, models.ObjectGroup, id, "delete")
		return err
	}); err != nil {
		return err
	}
	_ = rev
	if nm.syncMgr != nil {
		nm.syncMgr.NotifyAll(ctx)
	}
	return nil
}

func (nm *NodeManager) recordTargetChange(ctx context.Context, objectType, objectID, operation string) (int64, error) {
	if nm.syncMgr == nil {
		return 0, nil
	}
	rev, err := nm.revisions.Next(ctx)
	if err != nil {
		return 0, err
	}
	if err := nm.revisions.RecordChange(ctx, rev, objectType, objectID, 0, operation); err != nil {
		return 0, err
	}
	if err := recordMutationAudit(ctx, nm.store, objectType, objectID, operation); err != nil {
		return 0, err
	}
	return rev, nil
}

// RevokeCredential 撤销节点当前 Agent Credential。
func (nm *NodeManager) RevokeCredential(ctx context.Context, nodeID string) error {
	return nm.store.DeleteSetting(ctx, "cred:"+nodeID)
}

// ResolveTarget 解析任务目标为节点 ID 集合
func (nm *NodeManager) ResolveTarget(ctx context.Context, tgt models.Target) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	add := func(ids []string) {
		for _, id := range ids {
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	switch tgt.Type {
	case "node":
		add(tgt.NodeIDs)
	case "group":
		for _, gid := range tgt.GroupIDs {
			g, err := nm.store.GetGroup(ctx, gid)
			if err != nil {
				return nil, err
			}
			if g.Type == "label" {
				// 标签分组为动态成员，按 label 解析命中节点
				nodes, err := nm.store.ListNodes(ctx)
				if err != nil {
					return nil, err
				}
				for _, n := range nodes {
					if v, ok := n.Labels[g.LabelKey]; ok && (g.LabelValue == "" || v == g.LabelValue) {
						add([]string{n.ID})
					}
				}
			} else {
				members, err := nm.store.GroupMemberIDs(ctx, gid)
				if err != nil {
					return nil, err
				}
				add(members)
			}
		}
	case "label":
		nodes, err := nm.store.ListNodes(ctx)
		if err != nil {
			return nil, err
		}
		for _, n := range nodes {
			if v, ok := n.Labels[tgt.LabelKey]; ok && (tgt.LabelValue == "" || v == tgt.LabelValue) {
				add([]string{n.ID})
			}
		}
	default:
		add(tgt.NodeIDs)
	}
	return out, nil
}
