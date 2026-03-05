package app

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"tmd-pro/internal/config"
	"tmd-pro/internal/executor"
	"tmd-pro/internal/scheduler"
	"tmd-pro/internal/storage"
	"tmd-pro/pkg/logger"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// ScreenNameItem is a DTO for screen names
type ScreenNameItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ConfigData represents the editable configuration
type ConfigData struct {
	// 扫描配置
	ScanInterval int    `json:"scanInterval"`
	DataDir      string `json:"dataDir"`
	// 代理配置
	HttpProxy  string `json:"httpProxy"`
	HttpsProxy string `json:"httpsProxy"`
	NoProxy    string `json:"noProxy"`
	// 数据库配置
	DbHost     string `json:"dbHost"`
	DbPort     int    `json:"dbPort"`
	DbUsername string `json:"dbUsername"`
	DbPassword string `json:"dbPassword"`
	DbDatabase string `json:"dbDatabase"`
	DbCharset  string `json:"dbCharset"`
}

// App is the Wails application backend
type App struct {
	ctx     context.Context
	store   *storage.Storage
	exec    *executor.Executor
	sched   *scheduler.Scheduler
	tmdBin  embed.FS
	initErr string

	scanMu     sync.Mutex
	isScanning bool
	scanCtx    context.Context
	scanCancel context.CancelFunc
}

// NewApp creates a new App instance
func NewApp(tmdBin embed.FS) *App {
	return &App{tmdBin: tmdBin}
}

// OnStartup is called when the Wails app starts
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	var err error
	a.store, err = storage.NewStorage()
	if err != nil {
		a.initErr = fmt.Sprintf("数据库连接失败: %v", err)
		logger.Errorf("初始化存储失败: %v", err)
		return
	}

	a.exec = executor.NewExecutor(a.tmdBin)
	a.sched = scheduler.NewScheduler(a.store, a.exec)

	logger.AddWriter(&wailsLogWriter{app: a})
}

// OnShutdown is called when the Wails app shuts down
func (a *App) OnShutdown(ctx context.Context) {
	a.StopScan()
	if a.store != nil {
		a.store.Close()
	}
}

// GetInitError returns any startup initialization error
func (a *App) GetInitError() string {
	return a.initErr
}

// --- Screen Name API ---

// GetAllScreenNames returns all screen names from storage
func (a *App) GetAllScreenNames() ([]ScreenNameItem, error) {
	if a.store == nil {
		return nil, fmt.Errorf(a.initErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	names, err := a.store.GetAllScreenNames(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]ScreenNameItem, len(names))
	for i, sn := range names {
		items[i] = ScreenNameItem{ID: sn.ID, Name: sn.Name}
	}
	return items, nil
}

// AddScreenNames adds one or more screen names (comma/newline/space separated)
func (a *App) AddScreenNames(input string) (int, error) {
	if a.store == nil {
		return 0, fmt.Errorf(a.initErr)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("请输入用户名")
	}

	input = strings.ReplaceAll(input, "\n", ",")
	input = strings.ReplaceAll(input, " ", ",")
	parts := strings.Split(input, ",")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count := 0
	for _, name := range parts {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if err := a.store.AddScreenName(ctx, name); err != nil {
			logger.Errorf("添加用户失败: %s, %v", name, err)
		} else {
			count++
		}
	}
	return count, nil
}

// DeleteScreenName deletes a screen name by name
func (a *App) DeleteScreenName(name string) error {
	if a.store == nil {
		return fmt.Errorf(a.initErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.store.RemoveScreenName(ctx, name)
}

// --- Scanner API ---

// IsScanning returns whether a scan is currently running
func (a *App) IsScanning() bool {
	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	return a.isScanning
}

// StartScan starts the rotation scan
func (a *App) StartScan() error {
	if a.store == nil {
		return fmt.Errorf(a.initErr)
	}
	a.scanMu.Lock()
	if a.isScanning {
		a.scanMu.Unlock()
		return fmt.Errorf("扫描已在运行中")
	}
	a.isScanning = true
	a.scanCtx, a.scanCancel = context.WithCancel(context.Background())
	a.scanMu.Unlock()

	a.sched.SetOutputWriter(&wailsLogWriter{app: a})

	go func() {
		defer func() {
			a.scanMu.Lock()
			a.isScanning = false
			a.scanMu.Unlock()
			wailsRuntime.EventsEmit(a.ctx, "scan:status", false)
			a.emitLog("扫描已停止")
		}()

		a.emitLog("正在启动轮转扫描...")
		wailsRuntime.EventsEmit(a.ctx, "scan:status", true)

		if err := a.sched.StartWithRotation(a.scanCtx); err != nil && a.scanCtx.Err() == nil {
			a.emitLog(fmt.Sprintf("扫描出错: %v", err))
		}
	}()

	return nil
}

// StopScan stops the current scan
func (a *App) StopScan() {
	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	if a.scanCancel != nil {
		a.emitLog("正在停止扫描...")
		a.scanCancel()
		a.scanCancel = nil
	}
}

// RunOnce executes a single scan batch
func (a *App) RunOnce() error {
	if a.store == nil {
		return fmt.Errorf(a.initErr)
	}
	if a.IsScanning() {
		return fmt.Errorf("扫描正在运行中，请先停止")
	}

	go func() {
		wailsRuntime.EventsEmit(a.ctx, "scan:status", true)
		defer func() {
			wailsRuntime.EventsEmit(a.ctx, "scan:status", false)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		lastMaxId, err := a.store.GetLastScanId(ctx)
		if err != nil {
			a.emitLog(fmt.Sprintf("获取扫描状态失败: %v", err))
			return
		}

		names, err := a.store.GetScreenNameByLastMaxId(ctx, lastMaxId, 50)
		if err != nil {
			a.emitLog(fmt.Sprintf("获取用户列表失败: %v", err))
			return
		}

		if len(names) == 0 {
			a.emitLog("没有待扫描的用户，已重置扫描位置")
			_ = a.store.UpdateLastScanId(ctx, 0)
			return
		}

		a.emitLog(fmt.Sprintf("本次扫描 %d 个用户", len(names)))

		successCount, failCount := 0, 0
		for i, sn := range names {
			select {
			case <-ctx.Done():
				a.emitLog("扫描被中断")
				return
			default:
			}

			a.emitLog(fmt.Sprintf("[%d/%d] 正在处理: %s", i+1, len(names), sn.Name))
			lw := &wailsLogWriter{app: a, prefix: sn.Name}
			if err := a.exec.ExecuteWithOutput(ctx, sn.Name, lw, lw); err != nil {
				a.emitLog(fmt.Sprintf("处理 %s 失败: %v", sn.Name, err))
				failCount++
			} else {
				a.emitLog(fmt.Sprintf("处理 %s 成功", sn.Name))
				successCount++
			}
		}

		if len(names) > 0 {
			_ = a.store.UpdateLastScanId(ctx, names[len(names)-1].ID)
		}
		a.emitLog(fmt.Sprintf("扫描完成，成功: %d，失败: %d", successCount, failCount))
	}()

	return nil
}

// --- Settings API ---

// GetConfig returns the current configuration
func (a *App) GetConfig() ConfigData {
	cfg := config.G
	data := ConfigData{
		ScanInterval: cfg.AppConf.ScanIntervalMinutes,
		DataDir:      cfg.AppConf.DataDir,
		DbHost:       cfg.StorageConf.Host,
		DbPort:       cfg.StorageConf.Port,
		DbUsername:   cfg.StorageConf.Username,
		DbPassword:   cfg.StorageConf.Password,
		DbDatabase:   cfg.StorageConf.Database,
		DbCharset:    cfg.StorageConf.Charset,
	}
	if cfg.AppConf.Proxy != nil {
		data.HttpProxy = cfg.AppConf.Proxy.HttpProxy
		data.HttpsProxy = cfg.AppConf.Proxy.HttpsProxy
		data.NoProxy = cfg.AppConf.Proxy.NoProxy
	}
	return data
}

// SaveConfig saves the configuration to file
func (a *App) SaveConfig(data ConfigData) error {
	if data.ScanInterval <= 0 {
		return fmt.Errorf("扫描间隔必须是大于 0 的整数")
	}
	if data.DbHost == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	if data.DbPort <= 0 || data.DbPort > 65535 {
		return fmt.Errorf("数据库端口无效（1-65535）")
	}
	if data.DbDatabase == "" {
		return fmt.Errorf("数据库名称不能为空")
	}

	// 扫描配置
	config.G.AppConf.ScanIntervalMinutes = data.ScanInterval
	if data.DataDir != "" {
		config.G.AppConf.DataDir = data.DataDir
	}

	// 代理配置
	if data.HttpProxy != "" || data.HttpsProxy != "" || data.NoProxy != "" {
		if config.G.AppConf.Proxy == nil {
			config.G.AppConf.Proxy = &config.ProxyConfig{}
		}
		config.G.AppConf.Proxy.HttpProxy = data.HttpProxy
		config.G.AppConf.Proxy.HttpsProxy = data.HttpsProxy
		config.G.AppConf.Proxy.NoProxy = data.NoProxy
	} else {
		config.G.AppConf.Proxy = nil
	}

	// 数据库配置
	config.G.StorageConf.Host = data.DbHost
	config.G.StorageConf.Port = data.DbPort
	config.G.StorageConf.Username = data.DbUsername
	config.G.StorageConf.Password = data.DbPassword
	config.G.StorageConf.Database = data.DbDatabase
	config.G.StorageConf.Charset = data.DbCharset

	return a.saveConfigToFile()
}

// OpenConfigDir opens the config directory in the system file manager
func (a *App) OpenConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".tmd-pro", "conf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", configDir)
	case "windows":
		cmd = exec.Command("explorer", configDir)
	default:
		cmd = exec.Command("xdg-open", configDir)
	}
	return cmd.Start()
}

func (a *App) saveConfigToFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	configPath := filepath.Join(home, ".tmd-pro", "conf", "app.example.yaml")

	var cfgMap map[string]interface{}
	if raw, err := os.ReadFile(configPath); err == nil {
		_ = yaml.Unmarshal(raw, &cfgMap)
	}
	if cfgMap == nil {
		cfgMap = make(map[string]interface{})
	}

	appMap := map[string]interface{}{
		"name":                  config.G.AppConf.Name,
		"env":                   config.G.AppConf.Env,
		"data_dir":              config.G.AppConf.DataDir,
		"tmd_binary_name":       config.G.AppConf.TmdBinaryName,
		"scan_interval_minutes": config.G.AppConf.ScanIntervalMinutes,
	}
	if config.G.AppConf.Proxy != nil {
		appMap["proxy"] = map[string]interface{}{
			"http_proxy":  config.G.AppConf.Proxy.HttpProxy,
			"https_proxy": config.G.AppConf.Proxy.HttpsProxy,
			"no_proxy":    config.G.AppConf.Proxy.NoProxy,
		}
	}
	cfgMap["app"] = appMap
	cfgMap["storage"] = map[string]interface{}{
		"host":     config.G.StorageConf.Host,
		"port":     config.G.StorageConf.Port,
		"username": config.G.StorageConf.Username,
		"password": config.G.StorageConf.Password,
		"database": config.G.StorageConf.Database,
		"charset":  config.G.StorageConf.Charset,
	}

	raw, err := yaml.Marshal(cfgMap)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(configPath, raw, 0644)
}

func (a *App) emitLog(msg string) {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", msg)
	}
}

// wailsLogWriter forwards log output to the Wails event bus
type wailsLogWriter struct {
	app    *App
	prefix string
}

func (w *wailsLogWriter) Write(p []byte) (n int, err error) {
	content := strings.TrimSpace(string(p))
	if content == "" {
		return len(p), nil
	}
	content = ansiRegex.ReplaceAllString(content, "")
	if w.prefix != "" {
		w.app.emitLog(fmt.Sprintf("[%s] %s", w.prefix, content))
	} else {
		w.app.emitLog(content)
	}
	return len(p), nil
}
