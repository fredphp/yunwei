package ha

import (
        "context"
        "encoding/json"
        "errors"
        "fmt"
        "sync"
        "time"

        "yunwei/global"
        "yunwei/model/ha"
)

// SessionManager 会话管理器
type SessionManager struct {
        mu         sync.RWMutex
        sessions   map[string]*ha.HASession
        nodeID     string
        ttl        time.Duration
        backend    SessionBackend
        cleanupCh  chan struct{}
}

// SessionBackend 会话后端接口
type SessionBackend interface {
        Get(ctx context.Context, sessionID string) (*ha.HASession, error)
        Set(ctx context.Context, session *ha.HASession) error
        Delete(ctx context.Context, sessionID string) error
        Exists(ctx context.Context, sessionID string) (bool, error)
}

// NewSessionManager 创建会话管理器
func NewSessionManager(nodeID string, ttl time.Duration) *SessionManager {
        return &SessionManager{
                sessions:  make(map[string]*ha.HASession),
                nodeID:    nodeID,
                ttl:       ttl,
                cleanupCh: make(chan struct{}),
        }
}

// SetBackend 设置会话后端
func (m *SessionManager) SetBackend(backend SessionBackend) {
        m.backend = backend
}

// ==================== 会话操作 ====================

// CreateSession 创建会话
func (m *SessionManager) CreateSession(ctx context.Context, userID uint, username string, data map[string]interface{}) (*ha.HASession, error) {
        m.mu.Lock()
        defer m.mu.Unlock()

        now := time.Now()
        expiresAt := now.Add(m.ttl)

        dataJSON, _ := json.Marshal(data)

        session := &ha.HASession{
                SessionID:        generateSessionID(),
                UserID:           userID,
                Username:         username,
                Data:             string(dataJSON),
                CreatedNodeID:    m.nodeID,
                LastAccessNodeID: m.nodeID,
                LastAccessAt:     &now,
                ExpiresAt:        &expiresAt,
                IsActive:         true,
        }

        // 保存到后端
        if m.backend != nil {
                if err := m.backend.Set(ctx, session); err != nil {
                        return nil, err
                }
        }

        // 保存到本地缓存
        m.sessions[session.SessionID] = session

        // 保存到数据库
        global.DB.Create(session)

        return session, nil
}

// GetSession 获取会话
func (m *SessionManager) GetSession(ctx context.Context, sessionID string) (*ha.HASession, error) {
        m.mu.RLock()
        session, exists := m.sessions[sessionID]
        m.mu.RUnlock()

        if exists {
                // 检查是否过期
                if session.ExpiresAt != nil && session.ExpiresAt.After(time.Now()) {
                        // 更新访问时间
                        m.touchSession(session)
                        return session, nil
                }
                // 过期，删除
                m.DeleteSession(ctx, sessionID)
                return nil, ErrSessionExpired
        }

        // 从后端获取
        if m.backend != nil {
                session, err := m.backend.Get(ctx, sessionID)
                if err != nil {
                        return nil, err
                }
                if session == nil {
                        return nil, ErrSessionNotFound
                }

                // 检查是否过期
                if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
                        m.DeleteSession(ctx, sessionID)
                        return nil, ErrSessionExpired
                }

                // 缓存到本地
                m.mu.Lock()
                m.sessions[sessionID] = session
                m.mu.Unlock()

                m.touchSession(session)
                return session, nil
        }

        // 从数据库获取
        var dbSession ha.HASession
        err := global.DB.Where("session_id = ? AND is_active = ?", sessionID, true).First(&dbSession).Error
        if err != nil {
                return nil, ErrSessionNotFound
        }

        // 检查是否过期
        if dbSession.ExpiresAt != nil && dbSession.ExpiresAt.Before(time.Now()) {
                m.DeleteSession(ctx, sessionID)
                return nil, ErrSessionExpired
        }

        // 缓存到本地
        m.mu.Lock()
        m.sessions[sessionID] = &dbSession
        m.mu.Unlock()

        m.touchSession(&dbSession)
        return &dbSession, nil
}

// UpdateSession 更新会话
func (m *SessionManager) UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error {
        m.mu.Lock()
        defer m.mu.Unlock()

        session, exists := m.sessions[sessionID]
        if !exists {
                return ErrSessionNotFound
        }

        // 检查是否过期
        if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
                return ErrSessionExpired
        }

        dataJSON, _ := json.Marshal(data)
        session.Data = string(dataJSON)
        now := time.Now()
        session.LastAccessAt = &now
        session.LastAccessNodeID = m.nodeID

        // 更新后端
        if m.backend != nil {
                if err := m.backend.Set(ctx, session); err != nil {
                        return err
                }
        }

        // 更新数据库
        global.DB.Model(&ha.HASession{}).Where("session_id = ?", sessionID).
                Updates(map[string]interface{}{
                        "data":              session.Data,
                        "last_access_at":    now,
                        "last_access_node_id": m.nodeID,
                })

        return nil
}

// DeleteSession 删除会话
func (m *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
        m.mu.Lock()
        defer m.mu.Unlock()

        // 删除本地缓存
        delete(m.sessions, sessionID)

        // 删除后端
        if m.backend != nil {
                m.backend.Delete(ctx, sessionID)
        }

        // 更新数据库
        global.DB.Model(&ha.HASession{}).Where("session_id = ?", sessionID).
                Update("is_active", false)

        return nil
}

// RenewSession 续期会话
func (m *SessionManager) RenewSession(ctx context.Context, sessionID string) error {
        m.mu.Lock()
        defer m.mu.Unlock()

        session, exists := m.sessions[sessionID]
        if !exists {
                return ErrSessionNotFound
        }

        now := time.Now()
        expiresAt := now.Add(m.ttl)
        session.LastAccessAt = &now
        session.LastAccessNodeID = m.nodeID
        session.ExpiresAt = &expiresAt

        // 更新后端
        if m.backend != nil {
                if err := m.backend.Set(ctx, session); err != nil {
                        return err
                }
        }

        // 更新数据库
        global.DB.Model(&ha.HASession{}).Where("session_id = ?", sessionID).
                Updates(map[string]interface{}{
                        "last_access_at":      now,
                        "last_access_node_id": m.nodeID,
                        "expires_at":          expiresAt,
                })

        return nil
}

// touchSession 更新访问时间
func (m *SessionManager) touchSession(session *ha.HASession) {
        now := time.Now()
        session.LastAccessAt = &now
        session.LastAccessNodeID = m.nodeID

        // 异步更新数据库
        go func() {
                global.DB.Model(&ha.HASession{}).Where("session_id = ?", session.SessionID).
                        Updates(map[string]interface{}{
                                "last_access_at":      now,
                                "last_access_node_id": m.nodeID,
                        })
        }()
}

// ==================== 会话查询 ====================

// ListSessions 列出会话
func (m *SessionManager) ListSessions(filter *SessionFilter) ([]ha.HASession, int64, error) {
        query := global.DB.Model(&ha.HASession{}).Where("is_active = ?", true)

        if filter != nil {
                if filter.UserID > 0 {
                        query = query.Where("user_id = ?", filter.UserID)
                }
                if filter.Username != "" {
                        query = query.Where("username LIKE ?", "%"+filter.Username+"%")
                }
        }

        var total int64
        query.Count(&total)

        var sessions []ha.HASession
        err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&sessions).Error
        return sessions, total, err
}

// SessionFilter 会话过滤器
type SessionFilter struct {
        UserID   uint   `json:"userId"`
        Username string `json:"username"`
        Limit    int    `json:"limit"`
        Offset   int    `json:"offset"`
}

// ==================== 清理过期会话 ====================

// StartCleanup 启动清理
func (m *SessionManager) StartCleanup(ctx context.Context) {
        go func() {
                ticker := time.NewTicker(5 * time.Minute)
                defer ticker.Stop()

                for {
                        select {
                        case <-ctx.Done():
                                return
                        case <-m.cleanupCh:
                                return
                        case <-ticker.C:
                                m.cleanupExpiredSessions()
                        }
                }
        }()
}

// StopCleanup 停止清理
func (m *SessionManager) StopCleanup() {
        close(m.cleanupCh)
}

// cleanupExpiredSessions 清理过期会话
func (m *SessionManager) cleanupExpiredSessions() int {
        m.mu.Lock()
        defer m.mu.Unlock()

        now := time.Now()
        count := 0

        for sessionID, session := range m.sessions {
                if session.ExpiresAt != nil && session.ExpiresAt.Before(now) {
                        delete(m.sessions, sessionID)
                        count++
                }
        }

        // 更新数据库
        global.DB.Model(&ha.HASession{}).
                Where("is_active = ? AND expires_at < ?", true, now).
                Update("is_active", false)

        return count
}

// ==================== 统计 ====================

// GetSessionStats 获取会话统计
func (m *SessionManager) GetSessionStats() *SessionStats {
        m.mu.RLock()
        defer m.mu.RUnlock()

        stats := &SessionStats{}

        for _, session := range m.sessions {
                if session.IsActive {
                        if session.ExpiresAt != nil && session.ExpiresAt.After(time.Now()) {
                                stats.ActiveSessions++
                        } else {
                                stats.ExpiredSessions++
                        }
                }
        }

        global.DB.Model(&ha.HASession{}).Where("is_active = ?", true).Count(&stats.TotalSessions)

        return stats
}

// SessionStats 会话统计
type SessionStats struct {
        TotalSessions    int64 `json:"totalSessions"`
        ActiveSessions   int   `json:"activeSessions"`
        ExpiredSessions  int   `json:"expiredSessions"`
}

// ==================== 错误定义 ====================

var (
        ErrSessionNotFound = errors.New("session not found")
        ErrSessionExpired  = errors.New("session expired")
)

// ==================== 工具函数 ====================

// generateSessionID 生成会话ID
func generateSessionID() string {
        return fmt.Sprintf("sess_%d_%s", time.Now().UnixNano(), randomString(16))
}

// randomString 生成随机字符串
func randomString(n int) string {
        const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        b := make([]byte, n)
        for i := range b {
                b[i] = letters[time.Now().Nanosecond()%len(letters)]
        }
        return string(b)
}

// ==================== Redis 会话后端 ====================

// RedisSessionBackend Redis 会话后端
type RedisSessionBackend struct {
        client interface{}
        prefix string
}

// NewRedisSessionBackend 创建 Redis 会话后端
func NewRedisSessionBackend(client interface{}, prefix string) *RedisSessionBackend {
        return &RedisSessionBackend{
                client: client,
                prefix: prefix,
        }
}

// Get 获取会话
func (b *RedisSessionBackend) Get(ctx context.Context, sessionID string) (*ha.HASession, error) {
        // TODO: 从 Redis 获取
        return nil, nil
}

// Set 设置会话
func (b *RedisSessionBackend) Set(ctx context.Context, session *ha.HASession) error {
        // TODO: 保存到 Redis
        return nil
}

// Delete 删除会话
func (b *RedisSessionBackend) Delete(ctx context.Context, sessionID string) error {
        // TODO: 从 Redis 删除
        return nil
}

// Exists 检查会话是否存在
func (b *RedisSessionBackend) Exists(ctx context.Context, sessionID string) (bool, error) {
        // TODO: 检查 Redis
        return false, nil
}
