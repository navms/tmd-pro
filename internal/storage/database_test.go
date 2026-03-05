package storage

import (
	"context"
	"os"
	"testing"
	AppConfig "tmd-pro/internal/config"
	"tmd-pro/pkg/config"
	"tmd-pro/pkg/logger"

	"github.com/stretchr/testify/assert"
)

var storage Storage

func TestAddScreenName(t *testing.T) {
	err := storage.AddScreenName(context.Background(), "testxxx")
	assert.Nil(t, err)
}

func TestRemoveScreenName(t *testing.T) {
	err := storage.RemoveScreenName(context.Background(), "testxxx")
	assert.Nil(t, err)
}

func TestGetScreenNameByLastMaxId(t *testing.T) {
	screenNames, err := storage.GetScreenNameByLastMaxId(context.Background(), 0, 50)
	assert.Nil(t, err)
	assert.Equal(t, len(screenNames), 50)
}

func TestGetLastScanId(t *testing.T) {
	maxScanId, err := storage.GetLastScanId(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, maxScanId, int64(100))
}

func TestGetMaxScreenNameId(t *testing.T) {
	maxId, err := storage.GetMaxScreenNameId(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, maxId, int64(166))
}

func TestUpdateLastScanId(t *testing.T) {
	err := storage.UpdateLastScanId(context.Background(), 100)
	assert.Nil(t, err)
}

func TestMain(m *testing.M) {
	err := config.Init("app",
		config.WithConfigType("yaml"),
		config.WithPaths("/Users/hejin/workspace/tmd-pro/"),
		config.WithEnvPrefix("APP"),
	)
	if err != nil {
		panic("初始化配置失败: " + err.Error())
	}
	defer config.G().Close()

	var g AppConfig.GlobalConfig
	if err := config.Unmarshal(&g); err != nil {
		panic("解析配置失败: " + err.Error())
	}
	AppConfig.Init(&g)

	var logConf logger.Config
	if err := config.UnmarshalKey("logger", &logConf); err != nil {
		panic("解析日志配置失败: " + err.Error())
	}
	if err := logger.Init(logConf); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Sync()

	store, err := NewStorage()
	if err != nil {
		panic(err)
	}
	storage = *store
	defer store.Close()

	os.Exit(m.Run())
}
