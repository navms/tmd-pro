package storage

import (
	"time"

	"gorm.io/gorm"
)

// ScreenName 用户记录
type ScreenName struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;comment:主键"`
	Name        string         `gorm:"type:varchar(64);not null;uniqueIndex;comment:用户名称"`
	CreatedAt   time.Time      `gorm:"type:datetime;not null;autoCreateTime;comment:创建时间"`
	LastScanned *time.Time     `gorm:"type:datetime;comment:上次扫描时间"`
	DeletedAt   gorm.DeletedAt `gorm:"type:datetime;index;comment:逻辑删除时间"`
}

// ScanStats 扫描状态
type ScanStats struct {
	ID    int64 `gorm:"primaryKey;comment:主键"`
	Value int64 `gorm:"not null;default:0;comment:上一次扫描的最大ID"`
}

// TableName 表名
func (ScreenName) TableName() string {
	return "screen_name"
}

// TableName 表名
func (ScanStats) TableName() string {
	return "scan_stats"
}
