package ha

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/model/ha"

	"gorm.io/gorm"
)

// DistributedLockService 分布式锁服务
type DistributedLockService struct {
	mu          sync.RWMutex
	localLocks  map[string]*localLock
	nodeID      string
	backend     LockBackend
}

// LockBackend 锁后端接口
type LockBackend interface {
	Acquire(key, value string, ttl time.Duration) (bool, error)
	Release(key, value string) (bool, error)
	Renew(key, value string, ttl time.Duration) (bool, error)
	IsHeld(key string) (bool, string, error)
}

// localLock 本地锁
type localLock struct {
	value      string
	holderNode string
	acquiredAt time.Time
	expiresAt  time.Time
	renewCount int
}

// NewDistributedLockService 创建分布式锁服务
func NewDistributedLockService(nodeID string, backend LockBackend) *DistributedLockService {
	return &DistributedLockService{
		localLocks: make(map[string]*localLock),
		nodeID:     nodeID,
		backend:    backend,
	}
}

// ==================== 锁操作 ====================

// Acquire 获取锁
func (s *DistributedLockService) Acquire(ctx context.Context, key string, ttl time.Duration, opts ...LockOption) (*LockResult, error) {
	options := &lockOptions{
		waitTimeout: 0,
		retryInterval: 100 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 生成锁值
	value := s.generateLockValue()

	// 如果有等待超时，则重试
	if options.waitTimeout > 0 {
		return s.acquireWithWait(ctx, key, value, ttl, options)
	}

	// 直接获取
	acquired, err := s.acquireLock(key, value, ttl)
	if err != nil {
		return nil, err
	}

	return &LockResult{
		Acquired:   acquired,
		LockKey:    key,
		LockValue:  value,
		HolderNode: s.nodeID,
	}, nil
}

// acquireWithWait 带等待的获取锁
func (s *DistributedLockService) acquireWithWait(ctx context.Context, key, value string, ttl time.Duration, opts *lockOptions) (*LockResult, error) {
	timeout := time.NewTimer(opts.waitTimeout)
	defer timeout.Stop()

	ticker := time.NewTicker(opts.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return &LockResult{Acquired: false, LockKey: key}, nil
		case <-ticker.C:
			acquired, err := s.acquireLock(key, value, ttl)
			if err != nil {
				return nil, err
			}
			if acquired {
				return &LockResult{
					Acquired:   true,
					LockKey:    key,
					LockValue:  value,
					HolderNode: s.nodeID,
				}, nil
			}
		}
	}
}

// acquireLock 获取锁内部实现
func (s *DistributedLockService) acquireLock(key, value string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// 检查本地锁
	if localLock, exists := s.localLocks[key]; exists {
		if localLock.expiresAt.After(now) {
			return false, nil
		}
		// 过期，删除
		delete(s.localLocks, key)
	}

	// 使用后端获取锁
	if s.backend != nil {
		acquired, err := s.backend.Acquire(key, value, ttl)
		if err != nil || !acquired {
			return false, err
		}
	}

	// 本地记录
	s.localLocks[key] = &localLock{
		value:      value,
		holderNode: s.nodeID,
		acquiredAt: now,
		expiresAt:  now.Add(ttl),
	}

	// 持久化到数据库
	lockRecord := &ha.DistributedLock{
		LockKey:      key,
		LockValue:    value,
		HolderNodeID: s.nodeID,
		AcquiredAt:   &now,
		ExpiresAt:    ptrTime(now.Add(ttl)),
		Status:       "acquired",
		TTLSeconds:   int(ttl.Seconds()),
	}
	global.DB.Create(lockRecord)

	return true, nil
}

// Release 释放锁
func (s *DistributedLockService) Release(ctx context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查本地锁
	localLock, exists := s.localLocks[key]
	if !exists {
		return ErrLockNotFound
	}

	// 验证锁值
	if localLock.value != value {
		return ErrLockNotOwner
	}

	// 使用后端释放锁
	if s.backend != nil {
		_, err := s.backend.Release(key, value)
		if err != nil {
			return err
		}
	}

	// 删除本地锁
	delete(s.localLocks, key)

	// 更新数据库记录
	now := time.Now()
	global.DB.Model(&ha.DistributedLock{}).
		Where("lock_key = ? AND lock_value = ?", key, value).
		Updates(map[string]interface{}{
			"status":      "released",
			"released_at": now,
		})

	return nil
}

// Renew 续期锁
func (s *DistributedLockService) Renew(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查本地锁
	localLock, exists := s.localLocks[key]
	if !exists {
		return false, ErrLockNotFound
	}

	// 验证锁值
	if localLock.value != value {
		return false, ErrLockNotOwner
	}

	now := time.Now()

	// 使用后端续期
	if s.backend != nil {
		renewed, err := s.backend.Renew(key, value, ttl)
		if err != nil || !renewed {
			return false, err
		}
	}

	// 更新本地锁
	localLock.expiresAt = now.Add(ttl)
	localLock.renewCount++

	// 更新数据库
	global.DB.Model(&ha.DistributedLock{}).
		Where("lock_key = ? AND lock_value = ?", key, value).
		Updates(map[string]interface{}{
			"expires_at":  now.Add(ttl),
			"renew_count": localLock.renewCount,
		})

	return true, nil
}

// IsHeld 检查锁是否被持有
func (s *DistributedLockService) IsHeld(ctx context.Context, key string) (bool, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	// 检查本地锁
	if localLock, exists := s.localLocks[key]; exists {
		if localLock.expiresAt.After(now) {
			return true, localLock.holderNode, nil
		}
	}

	// 使用后端检查
	if s.backend != nil {
		return s.backend.IsHeld(key)
	}

	return false, "", nil
}

// ForceRelease 强制释放锁（仅管理员）
func (s *DistributedLockService) ForceRelease(ctx context.Context, key string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 使用后端释放
	if s.backend != nil {
		// 对于 Redis 后端，直接删除 key
		if redisBackend, ok := s.backend.(*RedisLockBackend); ok {
			redisBackend.Delete(key)
		}
	}

	// 删除本地锁
	delete(s.localLocks, key)

	// 更新数据库
	now := time.Now()
	global.DB.Model(&ha.DistributedLock{}).
		Where("lock_key = ?", key).
		Updates(map[string]interface{}{
			"status":      "force_released",
			"released_at": now,
		})

	return nil
}

// ==================== 锁监控 ====================

// GetLockInfo 获取锁信息
func (s *DistributedLockService) GetLockInfo(key string) (*LockInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	localLock, exists := s.localLocks[key]
	if !exists {
		return nil, ErrLockNotFound
	}

	now := time.Now()
	return &LockInfo{
		LockKey:     key,
		LockValue:   localLock.value,
		HolderNode:  localLock.holderNode,
		AcquiredAt:  localLock.acquiredAt,
		ExpiresAt:   localLock.expiresAt,
		IsExpired:   localLock.expiresAt.Before(now),
		RenewCount:  localLock.renewCount,
		TTLSeconds:  int(localLock.expiresAt.Sub(now).Seconds()),
	}, nil
}

// ListLocks 列出所有锁
func (s *DistributedLockService) ListLocks(filter *LockFilter) ([]LockInfo, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []LockInfo
	now := time.Now()

	for key, lock := range s.localLocks {
		if filter != nil && filter.HolderNode != "" && lock.holderNode != filter.HolderNode {
			continue
		}
		if filter != nil && filter.OnlyActive && lock.expiresAt.Before(now) {
			continue
		}

		result = append(result, LockInfo{
			LockKey:    key,
			LockValue:  lock.value,
			HolderNode: lock.holderNode,
			AcquiredAt: lock.acquiredAt,
			ExpiresAt:  lock.expiresAt,
			IsExpired:  lock.expiresAt.Before(now),
			RenewCount: lock.renewCount,
		})
	}

	return result, int64(len(result)), nil
}

// ==================== 自动续期 ====================

// StartAutoRenew 启动自动续期
func (s *DistributedLockService) StartAutoRenew(ctx context.Context, key, value string, ttl time.Duration) (context.CancelFunc, error) {
	// 检查锁是否存在
	if _, err := s.GetLockInfo(key); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(ttl / 2) // 每半个 TTL 续期一次
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, err := s.Renew(ctx, key, value, ttl)
				if err != nil {
					// 续期失败，停止自动续期
					return
				}
			}
		}
	}()

	return cancel, nil
}

// ==================== 工具方法 ====================

// generateLockValue 生成锁值
func (s *DistributedLockService) generateLockValue() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%s-%s-%d", s.nodeID, hex.EncodeToString(b), time.Now().UnixNano())
}

// CleanupExpiredLocks 清理过期锁
func (s *DistributedLockService) CleanupExpiredLocks() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for key, lock := range s.localLocks {
		if lock.expiresAt.Before(now) {
			delete(s.localLocks, key)
			count++
		}
	}

	// 清理数据库中的过期锁
	global.DB.Model(&ha.DistributedLock{}).
		Where("status = ? AND expires_at < ?", "acquired", now).
		Update("status", "expired")

	return count
}

// ==================== 类型定义 ====================

// LockOption 锁选项
type LockOption func(*lockOptions)

type lockOptions struct {
	waitTimeout    time.Duration
	retryInterval  time.Duration
}

// WithWaitTimeout 设置等待超时
func WithWaitTimeout(timeout time.Duration) LockOption {
	return func(o *lockOptions) {
		o.waitTimeout = timeout
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) LockOption {
	return func(o *lockOptions) {
		o.retryInterval = interval
	}
}

// LockResult 锁结果
type LockResult struct {
	Acquired   bool   `json:"acquired"`
	LockKey    string `json:"lockKey"`
	LockValue  string `json:"lockValue"`
	HolderNode string `json:"holderNode"`
}

// LockInfo 锁信息
type LockInfo struct {
	LockKey     string    `json:"lockKey"`
	LockValue   string    `json:"lockValue"`
	HolderNode  string    `json:"holderNode"`
	AcquiredAt  time.Time `json:"acquiredAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	IsExpired   bool      `json:"isExpired"`
	RenewCount  int       `json:"renewCount"`
	TTLSeconds  int       `json:"ttlSeconds"`
}

// LockFilter 锁过滤器
type LockFilter struct {
	HolderNode string `json:"holderNode"`
	OnlyActive bool   `json:"onlyActive"`
}

// 错误定义
var (
	ErrLockNotFound = errors.New("lock not found")
	ErrLockNotOwner = errors.New("not lock owner")
	ErrLockTimeout  = errors.New("lock acquire timeout")
)

// ==================== Redis 锁后端 ====================

// RedisLockBackend Redis 锁后端
type RedisLockBackend struct {
	client interface{} // Redis 客户端
}

// NewRedisLockBackend 创建 Redis 锁后端
func NewRedisLockBackend(client interface{}) *RedisLockBackend {
	return &RedisLockBackend{client: client}
}

// Acquire 获取锁
func (b *RedisLockBackend) Acquire(key, value string, ttl time.Duration) (bool, error) {
	// TODO: 实现 Redis SET NX EX
	// result, err := b.client.SetNX(context.Background(), key, value, ttl).Result()
	// return result, err
	return true, nil
}

// Release 释放锁
func (b *RedisLockBackend) Release(key, value string) (bool, error) {
	// TODO: 使用 Lua 脚本确保只有锁的持有者才能释放
	// script := "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"
	return true, nil
}

// Renew 续期锁
func (b *RedisLockBackend) Renew(key, value string, ttl time.Duration) (bool, error) {
	// TODO: 使用 Lua 脚本续期
	return true, nil
}

// IsHeld 检查锁是否被持有
func (b *RedisLockBackend) IsHeld(key string) (bool, string, error) {
	// TODO: 从 Redis 获取锁信息
	return false, "", nil
}

// Delete 强制删除锁
func (b *RedisLockBackend) Delete(key string) error {
	// TODO: 删除 Redis key
	return nil
}

// ==================== 数据库锁后端 ====================

// DatabaseLockBackend 数据库锁后端
type DatabaseLockBackend struct {
	db *gorm.DB
}

// NewDatabaseLockBackend 创建数据库锁后端
func NewDatabaseLockBackend(db *gorm.DB) *DatabaseLockBackend {
	return &DatabaseLockBackend{db: db}
}

// Acquire 获取锁
func (b *DatabaseLockBackend) Acquire(key, value string, ttl time.Duration) (bool, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	lock := &ha.DistributedLock{
		LockKey:      key,
		LockValue:    value,
		Status:       "acquired",
		AcquiredAt:   &now,
		ExpiresAt:    &expiresAt,
		TTLSeconds:   int(ttl.Seconds()),
	}

	// 尝试创建锁记录
	if err := b.db.Create(lock).Error; err != nil {
		// 检查是否已存在
		var existing ha.DistributedLock
		if err := b.db.Where("lock_key = ?", key).First(&existing).Error; err == nil {
			// 检查是否过期
			if existing.ExpiresAt != nil && existing.ExpiresAt.After(now) {
				return false, nil // 锁被持有
			}
			// 过期，尝试更新
			result := b.db.Model(&ha.DistributedLock{}).
				Where("lock_key = ? AND (expires_at IS NULL OR expires_at < ?)", key, now).
				Updates(map[string]interface{}{
					"lock_value":  value,
					"status":      "acquired",
					"acquired_at": now,
					"expires_at":  expiresAt,
				})
			return result.RowsAffected > 0, result.Error
		}
		return false, err
	}

	return true, nil
}

// Release 释放锁
func (b *DatabaseLockBackend) Release(key, value string) (bool, error) {
	now := time.Now()
	result := b.db.Model(&ha.DistributedLock{}).
		Where("lock_key = ? AND lock_value = ?", key, value).
		Updates(map[string]interface{}{
			"status":      "released",
			"released_at": now,
		})
	return result.RowsAffected > 0, result.Error
}

// Renew 续期锁
func (b *DatabaseLockBackend) Renew(key, value string, ttl time.Duration) (bool, error) {
	expiresAt := time.Now().Add(ttl)
	result := b.db.Model(&ha.DistributedLock{}).
		Where("lock_key = ? AND lock_value = ? AND status = ?", key, value, "acquired").
		Updates(map[string]interface{}{
			"expires_at":  expiresAt,
			"renew_count": gorm.Expr("renew_count + 1"),
		})
	return result.RowsAffected > 0, result.Error
}

// IsHeld 检查锁是否被持有
func (b *DatabaseLockBackend) IsHeld(key string) (bool, string, error) {
	var lock ha.DistributedLock
	err := b.db.Where("lock_key = ? AND status = ?", key, "acquired").First(&lock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", err
	}

	// 检查是否过期
	if lock.ExpiresAt != nil && lock.ExpiresAt.After(time.Now()) {
		return true, lock.HolderNodeID, nil
	}

	return false, "", nil
}

// ==================== 辅助函数 ====================

func ptrTime(t time.Time) *time.Time {
	return &t
}
