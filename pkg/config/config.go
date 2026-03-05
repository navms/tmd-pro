package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"tmd-pro/pkg/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 配置管理器
type Config struct {
	v          *viper.Viper
	name       string        // 配置文件名
	configType string        // 配置文件类型
	paths      []string      // 配置文件搜索路径
	prefix     string        // 环境变量前缀
	watcher    chan struct{} // 配置热更新通知通道
	callbacks  []func()      // 配置变化回调函数
	mu         sync.RWMutex  // 回调函数锁
}

// Option 配置选项函数
type Option func(*Config)

// 全局配置实例
var (
	globalConfig *Config
	once         sync.Once
)

// New 创建配置实例
func New(name string, opts ...Option) (*Config, error) {
	c := &Config{
		v:          viper.New(),
		name:       name,
		configType: "yaml",
		paths: []string{
			"./",
			"./config/",
			"./configs/",
		},
		watcher:   make(chan struct{}, 1),
		callbacks: make([]func(), 0),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(c)
	}

	// 初始化 Viper
	c.initViper()

	// 读取配置文件
	if err := c.v.ReadInConfig(); err != nil {
		var fileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &fileNotFoundError) {
			return nil, fmt.Errorf("read config file failed: %w", err)
		}
	}

	return c, nil
}

// MustNew 创建配置实例，失败则 panic
func MustNew(name string, opts ...Option) *Config {
	c, err := New(name, opts...)
	if err != nil {
		panic(fmt.Sprintf("create config failed: %v", err))
	}
	return c
}

// Init 初始化全局配置实例（单例模式）
func Init(name string, opts ...Option) error {
	var err error
	once.Do(func() {
		globalConfig, err = New(name, opts...)
	})
	return err
}

// G 返回全局配置实例
func G() *Config {
	return globalConfig
}

// initViper 初始化 Viper 核心配置
func (c *Config) initViper() {
	c.v.SetConfigName(c.name)
	c.v.SetConfigType(c.configType)

	for _, path := range c.paths {
		resolvedPath := os.ExpandEnv(path)
		c.v.AddConfigPath(resolvedPath)
	}

	if c.prefix != "" {
		c.v.SetEnvPrefix(c.prefix)
		c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		c.v.AutomaticEnv()
	}

	c.v.AllowEmptyEnv(false)
}

// ======================== 配置选项 ========================

// WithConfigType 设置配置文件类型
func WithConfigType(typ string) Option {
	return func(c *Config) {
		c.configType = typ
	}
}

// WithPaths 设置配置文件搜索路径
func WithPaths(paths ...string) Option {
	return func(c *Config) {
		c.paths = paths
	}
}

// WithAddPaths 添加配置文件搜索路径
func WithAddPaths(paths ...string) Option {
	return func(c *Config) {
		c.paths = append(c.paths, paths...)
	}
}

// WithEnvPrefix 设置环境变量前缀
func WithEnvPrefix(prefix string) Option {
	return func(c *Config) {
		c.prefix = prefix
	}
}

// WithDefaults 批量设置默认值
func WithDefaults(defaults map[string]any) Option {
	return func(c *Config) {
		for key, value := range defaults {
			c.v.SetDefault(key, value)
		}
	}
}

// ======================== 读取方法 ========================

// Get 获取配置值
func (c *Config) Get(key string) any {
	return c.v.Get(key)
}

// GetString 获取字符串配置
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt 获取整型配置
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetInt64 获取 int64 配置
func (c *Config) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

// GetFloat64 获取浮点型配置
func (c *Config) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

// GetBool 获取布尔型配置
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetDuration 获取时间间隔配置
func (c *Config) GetDuration(key string) time.Duration {
	return c.v.GetDuration(key)
}

// GetStringSlice 获取字符串切片配置
func (c *Config) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// GetIntSlice 获取整型切片配置
func (c *Config) GetIntSlice(key string) []int {
	return c.v.GetIntSlice(key)
}

// GetStringMap 获取 map[string]any 配置
func (c *Config) GetStringMap(key string) map[string]any {
	return c.v.GetStringMap(key)
}

// GetStringMapString 获取 map[string]string 配置
func (c *Config) GetStringMapString(key string) map[string]string {
	return c.v.GetStringMapString(key)
}

// IsSet 检查配置键是否存在
func (c *Config) IsSet(key string) bool {
	return c.v.IsSet(key)
}

// AllSettings 获取所有配置
func (c *Config) AllSettings() map[string]any {
	return c.v.AllSettings()
}

// AllKeys 获取所有配置键
func (c *Config) AllKeys() []string {
	return c.v.AllKeys()
}

// ======================== 解析方法 ========================

// Unmarshal 将所有配置解析到结构体
func (c *Config) Unmarshal(obj any) error {
	return c.v.Unmarshal(obj)
}

// UnmarshalKey 将指定 key 的配置解析到结构体
func (c *Config) UnmarshalKey(key string, obj any) error {
	return c.v.UnmarshalKey(key, obj)
}

// ======================== 热更新 ========================

// Watch 监听配置文件变化
// 返回通知通道，配置变化时会发送信号
func (c *Config) Watch() <-chan struct{} {
	c.v.WatchConfig()
	c.v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("[Config] file changed: ", zap.String("path", e.Name))

		// 执行回调函数
		c.mu.RLock()
		callbacks := make([]func(), len(c.callbacks))
		copy(callbacks, c.callbacks)
		c.mu.RUnlock()

		for _, cb := range callbacks {
			cb()
		}

		// 非阻塞发送通知
		select {
		case c.watcher <- struct{}{}:
		default:
		}
	})
	return c.watcher
}

// OnChange 注册配置变化回调函数
func (c *Config) OnChange(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callbacks = append(c.callbacks, callback)
}

// ======================== 设置方法 ========================

// Set 手动设置配置值
func (c *Config) Set(key string, value any) {
	c.v.Set(key, value)
}

// SetDefault 设置默认值
func (c *Config) SetDefault(key string, value any) {
	c.v.SetDefault(key, value)
}

// ======================== 辅助方法 ========================

// ConfigFileUsed 获取当前使用的配置文件路径
func (c *Config) ConfigFileUsed() string {
	return c.v.ConfigFileUsed()
}

// Viper 返回底层 viper 实例（用于高级操作）
func (c *Config) Viper() *viper.Viper {
	return c.v
}

// Close 关闭配置监听
func (c *Config) Close() {
	close(c.watcher)
}

// ======================== 全局便捷方法 ========================

// Get 全局获取配置
func Get(key string) any {
	return globalConfig.Get(key)
}

// GetString 全局获取字符串配置
func GetString(key string) string {
	return globalConfig.GetString(key)
}

// GetInt 全局获取整型配置
func GetInt(key string) int {
	return globalConfig.GetInt(key)
}

// GetBool 全局获取布尔型配置
func GetBool(key string) bool {
	return globalConfig.GetBool(key)
}

// GetDuration 全局获取时间间隔配置
func GetDuration(key string) time.Duration {
	return globalConfig.GetDuration(key)
}

// IsSet 全局检查配置是否存在
func IsSet(key string) bool {
	return globalConfig.IsSet(key)
}

// Unmarshal 全局解析配置到结构体
func Unmarshal(obj any) error {
	return globalConfig.Unmarshal(obj)
}

// UnmarshalKey 全局解析指定 key 到结构体
func UnmarshalKey(key string, obj any) error {
	return globalConfig.UnmarshalKey(key, obj)
}
