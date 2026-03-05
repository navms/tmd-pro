package main

import (
	"context"
	"embed"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"go.uber.org/zap"

	appPkg "tmd-pro/internal/app"
	AppConfig "tmd-pro/internal/config"
	"tmd-pro/internal/executor"
	"tmd-pro/internal/scheduler"
	"tmd-pro/internal/storage"
	pkgConfig "tmd-pro/pkg/config"
	"tmd-pro/pkg/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	err = pkgConfig.Init("app",
		pkgConfig.WithConfigType("yaml"),
		pkgConfig.WithPaths(filepath.Join(home, ".tmd-pro", "conf")),
		pkgConfig.WithEnvPrefix("APP"),
	)
	if err != nil {
		panic("初始化配置失败: " + err.Error())
	}
	defer pkgConfig.G().Close()

	var g AppConfig.GlobalConfig
	if err := pkgConfig.Unmarshal(&g); err != nil {
		panic("解析配置失败: " + err.Error())
	}
	AppConfig.Init(&g)

	var logConf logger.Config
	if err := pkgConfig.UnmarshalKey("logger", &logConf); err != nil {
		panic("解析日志配置失败: " + err.Error())
	}
	if err := logger.Init(logConf); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("配置信息", zap.Any("config", g))

	rootCmd := &cobra.Command{
		Use:   "tmd-pro",
		Short: "Tmd Pro - Screen Name 管理和定时执行工具",
		Long: `Tmd Pro 是一个用于管理 screen_name 并定时执行 tmd 命令的工具。
功能包括：
- 轮换扫描：每隔固定时间处理下一批次
- 管理 screen_name：添加、列出、删除`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGui()
		},
	}

	rootCmd.AddCommand(
		newRotateCmd(),
		newAddScreenNameCmd(),
		newRemoveScreenNameCmd(),
		newGuiCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		logger.Panic("命令执行失败", zap.Error(err))
		os.Exit(1)
	}
}

func newRotateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate",
		Short: "启动轮转扫描",
		Long:  "启动轮转扫描，根据配置的 scan_interval_minutes 间歇地扫描所有 screen_name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext()
			defer cancel()

			store, err := storage.NewStorage()
			if err != nil {
				return err
			}
			defer store.Close()

			exec := executor.NewExecutor(Tmd)
			sched := scheduler.NewScheduler(store, exec)
			return sched.StartWithRotation(ctx)
		},
	}
}

func newAddScreenNameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <screen_name>...",
		Short: "添加 screen_name",
		Long:  "添加 screen_name，存在则不添加",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext()
			defer cancel()

			store, err := storage.NewStorage()
			if err != nil {
				return err
			}
			defer store.Close()

			for _, arg := range args {
				if err := store.AddScreenName(ctx, strings.TrimSpace(arg)); err != nil {
					logger.Error(err.Error())
				}
			}
			return nil
		},
	}
}

func newRemoveScreenNameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <screen_name>...",
		Short: "删除 screen_name",
		Long:  "删除 screen_name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext()
			defer cancel()

			store, err := storage.NewStorage()
			if err != nil {
				return err
			}
			defer store.Close()

			for _, arg := range args {
				if err := store.RemoveScreenName(ctx, strings.TrimSpace(arg)); err != nil {
					logger.Error(err.Error())
				}
			}
			return nil
		},
	}
}

func newGuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gui",
		Short: "启动图形界面",
		Long:  "启动 Wails 桌面图形界面，可视化管理 screen_name 和扫描任务",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGui()
		},
	}
}

func runGui() error {
	a := appPkg.NewApp(Tmd)
	return wails.Run(&options.App{
		Title:            "TMD Pro",
		Width:            1100,
		Height:           720,
		MinWidth:         900,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 245, G: 245, B: 247, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  a.OnStartup,
		OnShutdown: a.OnShutdown,
		Bind:       []interface{}{a},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				HideTitle: true,
			},
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
}

// signalContext returns a context that is canceled on SIGINT/SIGTERM
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
		signal.Stop(sigChan)
	}()
	return ctx, cancel
}
