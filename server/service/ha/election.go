package ha

import (
        "context"
        "errors"
        "fmt"
        "sync"
        "time"

        "yunwei/global"
        "yunwei/model/ha"

        "gorm.io/gorm"
)

// LeaderElectionService Leader选举服务
type LeaderElectionService struct {
        mu             sync.RWMutex
        nodeID         string
        electionKey    string
        leaseDuration  time.Duration
        renewInterval  time.Duration
        backend        ElectionBackend
        leaderChangeCh chan LeaderChangeEvent
        
        // 当前状态
        isLeader     bool
        leaderNodeID string
        term         int64
        stopCh       chan struct{}
}

// ElectionBackend 选举后端接口
type ElectionBackend interface {
        Campaign(ctx context.Context, key, nodeID string, lease time.Duration) (bool, int64, error)
        Resign(ctx context.Context, key, nodeID string) error
        Proclaim(ctx context.Context, key, nodeID string, lease time.Duration) error
        GetLeader(ctx context.Context, key string) (string, int64, error)
}

// LeaderChangeEvent Leader变更事件
type LeaderChangeEvent struct {
        OldLeader string
        NewLeader string
        Term      int64
        Timestamp time.Time
}

// NewLeaderElectionService 创建Leader选举服务
func NewLeaderElectionService(nodeID, electionKey string, leaseDuration time.Duration) *LeaderElectionService {
        return &LeaderElectionService{
                nodeID:         nodeID,
                electionKey:    electionKey,
                leaseDuration:  leaseDuration,
                renewInterval:  leaseDuration / 2,
                leaderChangeCh: make(chan LeaderChangeEvent, 100),
                stopCh:         make(chan struct{}),
        }
}

// SetBackend 设置选举后端
func (s *LeaderElectionService) SetBackend(backend ElectionBackend) {
        s.backend = backend
}

// Start 启动选举服务
func (s *LeaderElectionService) Start(ctx context.Context) error {
        // 开始竞选
        go s.campaignLoop(ctx)
        
        // 开始监控
        go s.monitorLoop(ctx)
        
        return nil
}

// Stop 停止选举服务
func (s *LeaderElectionService) Stop() {
        close(s.stopCh)
}

// ==================== 竞选循环 ====================

// campaignLoop 竞选循环
func (s *LeaderElectionService) campaignLoop(ctx context.Context) {
        for {
                select {
                case <-s.stopCh:
                        return
                case <-ctx.Done():
                        return
                default:
                        // 尝试竞选
                        s.campaign(ctx)
                        
                        // 如果是 Leader，保持租约
                        if s.IsLeader() {
                                s.maintainLeadership(ctx)
                        } else {
                                // 等待一段时间后重试
                                time.Sleep(s.leaseDuration)
                        }
                }
        }
}

// campaign 竞选
func (s *LeaderElectionService) campaign(ctx context.Context) {
        s.mu.Lock()
        defer s.mu.Unlock()

        // 使用后端竞选
        var won bool
        var term int64
        var err error

        if s.backend != nil {
                won, term, err = s.backend.Campaign(ctx, s.electionKey, s.nodeID, s.leaseDuration)
                if err != nil {
                        // 记录错误
                        s.recordElectionEvent("campaign_failed", err.Error())
                        return
                }
        } else {
                // 使用数据库后端
                won, term, err = s.campaignWithDatabase(ctx)
                if err != nil {
                        return
                }
        }

        if won {
                wasLeader := s.isLeader
                s.isLeader = true
                s.leaderNodeID = s.nodeID
                s.term = term

                // 记录事件
                s.recordElectionEvent("elected", fmt.Sprintf("Node %s became leader, term %d", s.nodeID, term))
                
                // 更新数据库
                s.updateLeaderInDatabase(term)

                // 发送变更通知
                if !wasLeader {
                        s.leaderChangeCh <- LeaderChangeEvent{
                                NewLeader: s.nodeID,
                                Term:      term,
                                Timestamp: time.Now(),
                        }
                }
        }
}

// campaignWithDatabase 使用数据库竞选
func (s *LeaderElectionService) campaignWithDatabase(ctx context.Context) (bool, int64, error) {
        now := time.Now()
        expiresAt := now.Add(s.leaseDuration)

        // 检查当前 Leader
        var election ha.LeaderElection
        err := global.DB.Where("election_key = ?", s.electionKey).First(&election).Error

        if err != nil {
                // 没有记录，创建新的
                election = ha.LeaderElection{
                        ElectionKey:  s.electionKey,
                        LeaderNodeID: s.nodeID,
                        Term:         1,
                        AcquiredAt:   &now,
                        ExpiresAt:    &expiresAt,
                        Status:       "active",
                }
                if err := global.DB.Create(&election).Error; err != nil {
                        return false, 0, err
                }
                return true, 1, nil
        }

        // 检查是否过期或自己是 Leader
        if election.ExpiresAt == nil || election.ExpiresAt.Before(now) || election.LeaderNodeID == s.nodeID {
                // 可以竞选
                term := election.Term
                if election.LeaderNodeID != s.nodeID {
                        term++
                }

                result := global.DB.Model(&ha.LeaderElection{}).
                        Where("election_key = ? AND (expires_at IS NULL OR expires_at < ? OR leader_node_id = ?)",
                                s.electionKey, now, s.nodeID).
                        Updates(map[string]interface{}{
                                "leader_node_id": s.nodeID,
                                "term":           term,
                                "acquired_at":    now,
                                "expires_at":     expiresAt,
                                "status":         "active",
                        })

                if result.RowsAffected > 0 {
                        return true, term, nil
                }
        }

        // 获取当前 Leader 信息
        s.mu.Lock()
        s.leaderNodeID = election.LeaderNodeID
        s.term = election.Term
        s.isLeader = false
        s.mu.Unlock()

        return false, election.Term, nil
}

// maintainLeadership 保持 Leader 地位
func (s *LeaderElectionService) maintainLeadership(ctx context.Context) {
        ticker := time.NewTicker(s.renewInterval)
        defer ticker.Stop()

        for {
                select {
                case <-s.stopCh:
                        s.resign()
                        return
                case <-ctx.Done():
                        s.resign()
                        return
                case <-ticker.C:
                        if !s.renew(ctx) {
                                // 续期失败，失去 Leader
                                s.mu.Lock()
                                wasLeader := s.isLeader
                                s.isLeader = false
                                s.mu.Unlock()

                                if wasLeader {
                                        s.recordElectionEvent("lost", "Failed to renew leadership")
                                        s.leaderChangeCh <- LeaderChangeEvent{
                                                OldLeader: s.nodeID,
                                                Term:      s.term,
                                                Timestamp: time.Now(),
                                        }
                                }
                                return
                        }
                }
        }
}

// renew 续期
func (s *LeaderElectionService) renew(ctx context.Context) bool {
        s.mu.Lock()
        defer s.mu.Unlock()

        if !s.isLeader {
                return false
        }

        now := time.Now()
        expiresAt := now.Add(s.leaseDuration)

        // 使用后端续期
        if s.backend != nil {
                if err := s.backend.Proclaim(ctx, s.electionKey, s.nodeID, s.leaseDuration); err != nil {
                        return false
                }
        }

        // 更新数据库
        result := global.DB.Model(&ha.LeaderElection{}).
                Where("election_key = ? AND leader_node_id = ?", s.electionKey, s.nodeID).
                Updates(map[string]interface{}{
                        "expires_at":  expiresAt,
                        "renew_count": gorm.Expr("renew_count + 1"),
                })

        return result.RowsAffected > 0
}

// resign 辞职
func (s *LeaderElectionService) resign() {
        s.mu.Lock()
        defer s.mu.Unlock()

        if !s.isLeader {
                return
        }

        // 使用后端辞职
        if s.backend != nil {
                s.backend.Resign(context.Background(), s.electionKey, s.nodeID)
        }

        // 更新数据库
        now := time.Now()
        global.DB.Model(&ha.LeaderElection{}).
                Where("election_key = ? AND leader_node_id = ?", s.electionKey, s.nodeID).
                Updates(map[string]interface{}{
                        "status":      "resigned",
                        "expires_at":  now,
                })

        s.isLeader = false
        s.recordElectionEvent("resigned", fmt.Sprintf("Node %s resigned", s.nodeID))
}

// ==================== 监控循环 ====================

// monitorLoop 监控循环
func (s *LeaderElectionService) monitorLoop(ctx context.Context) {
        ticker := time.NewTicker(s.leaseDuration / 2)
        defer ticker.Stop()

        for {
                select {
                case <-s.stopCh:
                        return
                case <-ctx.Done():
                        return
                case <-ticker.C:
                        s.checkLeader(ctx)
                }
        }
}

// checkLeader 检查 Leader 状态
func (s *LeaderElectionService) checkLeader(ctx context.Context) {
        s.mu.RLock()
        wasLeader := s.isLeader
        s.mu.RUnlock()

        // 从数据库获取 Leader 信息
        var election ha.LeaderElection
        err := global.DB.Where("election_key = ?", s.electionKey).First(&election).Error
        if err != nil {
                return
        }

        s.mu.Lock()
        s.leaderNodeID = election.LeaderNodeID
        s.term = election.Term

        // 检查自己是否还是 Leader
        if wasLeader && election.LeaderNodeID != s.nodeID {
                s.isLeader = false
                s.recordElectionEvent("lost", "Lost leadership to " + election.LeaderNodeID)
                s.leaderChangeCh <- LeaderChangeEvent{
                        OldLeader: s.nodeID,
                        NewLeader: election.LeaderNodeID,
                        Term:      election.Term,
                        Timestamp: time.Now(),
                }
        }
        s.mu.Unlock()
}

// ==================== 状态查询 ====================

// IsLeader 是否是 Leader
func (s *LeaderElectionService) IsLeader() bool {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.isLeader
}

// GetLeader 获取当前 Leader
func (s *LeaderElectionService) GetLeader() string {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.leaderNodeID
}

// GetTerm 获取当前任期
func (s *LeaderElectionService) GetTerm() int64 {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.term
}

// GetLeaderChangeChannel 获取 Leader 变更通道
func (s *LeaderElectionService) GetLeaderChangeChannel() <-chan LeaderChangeEvent {
        return s.leaderChangeCh
}

// ==================== 数据库操作 ====================

// updateLeaderInDatabase 更新数据库中的 Leader 信息
func (s *LeaderElectionService) updateLeaderInDatabase(term int64) {
        now := time.Now()
        expiresAt := now.Add(s.leaseDuration)

        global.DB.Where("election_key = ?", s.electionKey).
                Assign(map[string]interface{}{
                        "leader_node_id": s.nodeID,
                        "term":           term,
                        "acquired_at":    now,
                        "expires_at":     expiresAt,
                        "status":         "active",
                }).
                FirstOrCreate(&ha.LeaderElection{
                        ElectionKey: s.electionKey,
                })

        // 更新节点状态
        global.DB.Model(&ha.ClusterNode{}).
                Where("node_id = ?", s.nodeID).
                Updates(map[string]interface{}{
                        "is_leader": true,
                        "role":      "leader",
                })
}

// recordElectionEvent 记录选举事件
func (s *LeaderElectionService) recordElectionEvent(eventType, detail string) {
        event := &ha.ClusterEvent{
                EventType: "leader_election",
                NodeID:    s.nodeID,
                Title:     eventType,
                Detail:    detail,
                Level:     "info",
                Source:    "leader_election_service",
        }
        global.DB.Create(event)
}

// ==================== 手动操作 ====================

// Resign 辞职（手动）
func (s *LeaderElectionService) Resign() error {
        s.mu.RLock()
        if !s.isLeader {
                s.mu.RUnlock()
                return errors.New("not leader")
        }
        s.mu.RUnlock()

        s.resign()
        return nil
}

// ForceLeader 强制指定 Leader（管理员操作）
func (s *LeaderElectionService) ForceLeader(nodeID string) error {
        s.mu.Lock()
        defer s.mu.Unlock()

        now := time.Now()
        expiresAt := now.Add(s.leaseDuration)

        // 更新数据库
        result := global.DB.Model(&ha.LeaderElection{}).
                Where("election_key = ?", s.electionKey).
                Updates(map[string]interface{}{
                        "leader_node_id": nodeID,
                        "term":           s.term + 1,
                        "acquired_at":    now,
                        "expires_at":     expiresAt,
                        "status":         "active",
                })

        if result.RowsAffected == 0 {
                return errors.New("election not found")
        }

        // 更新状态
        oldLeader := s.leaderNodeID
        s.leaderNodeID = nodeID
        s.isLeader = nodeID == s.nodeID
        s.term++

        // 记录事件
        s.recordElectionEvent("force_leader", fmt.Sprintf("Force set leader to %s", nodeID))

        // 发送变更通知
        s.leaderChangeCh <- LeaderChangeEvent{
                OldLeader: oldLeader,
                NewLeader: nodeID,
                Term:      s.term,
                Timestamp: now,
        }

        return nil
}
