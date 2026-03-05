package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Storage 存储管理
type Storage struct {
	db *Database
}

// NewStorage 创建存储管理器
func NewStorage() (*Storage, error) {
	db, err := NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}
	return &Storage{db: db}, nil
}

// Close 关闭存储
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// AddScreenName 添加 screen_name
func (s *Storage) AddScreenName(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("screen_name 不能为空")
	}
	return s.db.AddScreenName(ctx, name)
}

// RemoveScreenName 删除 screen_name
func (s *Storage) RemoveScreenName(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("screen_name 不能为空")
	}
	return s.db.RemoveScreenName(ctx, name)
}

// GetScreenNameByLastMaxId 根据lastMaxId和limit进行查询
func (s *Storage) GetScreenNameByLastMaxId(ctx context.Context, lastMaxId int64, limit int) ([]ScreenName, error) {
	if limit <= 0 {
		return nil, errors.New("limit 必须大于 0")
	}
	if lastMaxId < 0 {
		return nil, errors.New("lastMaxId 必须大于 0")
	}
	return s.db.GetScreenNameByLastMaxId(ctx, lastMaxId, limit)
}

// GetMaxScreenNameId 获取screen_name的maxId
func (s *Storage) GetMaxScreenNameId(ctx context.Context) (int64, error) {
	return s.db.GetMaxScreenNameId(ctx)
}

// GetLastScanId 获取上一次扫描的最后ID
func (s *Storage) GetLastScanId(ctx context.Context) (int64, error) {
	return s.db.GetLastScanId(ctx)
}

// UpdateLastScanId 更新上一次扫描的最后ID
func (s *Storage) UpdateLastScanId(ctx context.Context, value int64) error {
	if value < 0 {
		return errors.New("扫描 ID 不能为负数")
	}
	return s.db.UpdateLastScanId(ctx, value)
}

// GetAllScreenNames 获取所有 screen_name
func (s *Storage) GetAllScreenNames(ctx context.Context) ([]ScreenName, error) {
	return s.db.GetAllScreenNames(ctx)
}
