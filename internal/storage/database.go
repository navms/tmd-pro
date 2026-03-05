package storage

import (
	"context"
	"errors"
	"fmt"
	"tmd-pro/internal/config"
	"tmd-pro/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// Database MySQL 数据库管理
type Database struct {
	db *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase() (*Database, error) {
	charset := config.G.StorageConf.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.G.StorageConf.Username, config.G.StorageConf.Password,
		config.G.StorageConf.Host, config.G.StorageConf.Port, config.G.StorageConf.Database, charset)

	env := config.G.AppConf.Env
	cfg := gorm.Config{}
	if env == "dev" {
		cfg.Logger = glogger.Default.LogMode(glogger.Info)
	}
	db, err := gorm.Open(mysql.Open(dsn), &cfg)
	if err != nil {
		return nil, err
	}
	logger.Info("MySQL 数据库连接成功")

	err = db.AutoMigrate(&ScreenName{}, &ScanStats{})
	if err != nil {
		return nil, err
	}
	logger.Info("MySQL 迁移表结构成功")

	return &Database{db: db}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	db, err := d.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// AddScreenName 添加 screen_name
func (d *Database) AddScreenName(ctx context.Context, name string) error {
	sn := &ScreenName{Name: name}
	result := d.db.WithContext(ctx).FirstOrCreate(sn, ScreenName{Name: name})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// RemoveScreenName 删除 screen_name
func (d *Database) RemoveScreenName(ctx context.Context, name string) error {
	_, err := gorm.G[ScreenName](d.db).Where("name = ?", name).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetScreenNameByLastMaxId 根据lastMaxId和limit进行查询
func (d *Database) GetScreenNameByLastMaxId(ctx context.Context, lastMaxId int64, limit int) ([]ScreenName, error) {
	var screenNames []ScreenName
	err := gorm.G[ScreenName](d.db).Where("id > ?", lastMaxId).Order("id asc").Limit(limit).Scan(ctx, &screenNames)
	if err != nil {
		return nil, err
	}
	return screenNames, nil
}

// GetLastScanId 获取上一次扫描的最后ID
func (d *Database) GetLastScanId(ctx context.Context) (int64, error) {
	scanStats, err := gorm.G[ScanStats](d.db).First(ctx)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	return scanStats.Value, nil
}

// UpdateLastScanId 更新上一次扫描的最后ID
func (d *Database) UpdateLastScanId(ctx context.Context, value int64) error {
	_, err := gorm.G[ScanStats](d.db).Where("id = ?", 1).Update(ctx, "value", value)
	if err != nil {
		return err
	}
	return nil
}

// GetMaxScreenNameId 获取screen_name的maxId
func (d *Database) GetMaxScreenNameId(ctx context.Context) (int64, error) {
	var maxId int64
	err := gorm.G[ScreenName](d.db).Select("max(id)").Scan(ctx, &maxId)
	if err != nil {
		return -1, err
	}
	return maxId, nil
}

// GetAllScreenNames 获取所有 screen_name
func (d *Database) GetAllScreenNames(ctx context.Context) ([]ScreenName, error) {
	var screenNames []ScreenName
	err := gorm.G[ScreenName](d.db).Order("id desc").Scan(ctx, &screenNames)
	if err != nil {
		return nil, err
	}
	return screenNames, nil
}
