package executor

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"tmd-pro/internal/config"

	"tmd-pro/pkg/logger"
)

// ErrCanceled 表示操作被取消
var ErrCanceled = errors.New("操作被取消")

// Executor tmd 命令执行器
type Executor struct {
	tmd     embed.FS
	tmdPath string
}

// NewExecutor 创建执行器
func NewExecutor(tmd embed.FS) *Executor {
	return &Executor{tmd: tmd,
		tmdPath: filepath.Join(config.G.AppConf.DataDir, config.G.AppConf.TmdBinaryName)}
}

// EnsureTmdBinary 确保 tmd 二进制文件存在
func (e *Executor) EnsureTmdBinary() error {
	// 检查是否已存在
	if _, err := os.Stat(e.tmdPath); err == nil {
		return nil
	}

	// 从嵌入的文件系统读取
	data, err := e.tmd.ReadFile("lib/tmd")
	if err != nil {
		return fmt.Errorf("读取嵌入的 tmd 二进制失败: %w", err)
	}

	// 写入到目标路径
	if err := os.WriteFile(e.tmdPath, data, 0755); err != nil {
		return fmt.Errorf("写入 tmd 二进制失败: %w", err)
	}

	logger.Infof("已释放 tmd 二进制到: %s", e.tmdPath)
	return nil
}

// Execute 执行 tmd 命令
func (e *Executor) Execute(ctx context.Context, screenName string) error {
	stdLogger := &StdLoggerWriter{}
	multiStdout := io.MultiWriter(os.Stdout, stdLogger)
	return e.ExecuteWithOutput(ctx, screenName, multiStdout, multiStdout)
}

// ExecuteWithOutput 执行 tmd 命令，指定输出
func (e *Executor) ExecuteWithOutput(ctx context.Context, screenName string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, e.tmdPath, "--user", screenName)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// 设置环境变量
	cmd.Env = e.buildEnv()

	logger.Infof("执行命令: %s --user %s", e.tmdPath, screenName)

	if err := cmd.Run(); err != nil {
		// 检查是否是因为 context 被取消
		if ctx.Err() != nil {
			return fmt.Errorf("%w: %v", ErrCanceled, ctx.Err())
		}
		return fmt.Errorf("执行 tmd 失败: %w", err)
	}
	return nil
}

// buildEnv 构建环境变量
func (e *Executor) buildEnv() []string {
	// 继承当前进程的所有环境变量
	env := os.Environ()

	// 如果配置了代理，则添加代理环境变量
	proxyConf := config.G.AppConf.Proxy
	if proxyConf == nil {
		return env
	}

	if proxyConf.HttpProxy != "" {
		env = append(env, "HTTP_PROXY="+proxyConf.HttpProxy)
		env = append(env, "http_proxy="+proxyConf.HttpProxy)
		logger.Infof("设置 HTTP_PROXY: %s", proxyConf.HttpProxy)
	}
	if proxyConf.HttpsProxy != "" {
		env = append(env, "HTTPS_PROXY="+proxyConf.HttpsProxy)
		env = append(env, "https_proxy="+proxyConf.HttpsProxy)
		logger.Infof("设置 HTTPS_PROXY: %s", proxyConf.HttpsProxy)
	}
	if proxyConf.NoProxy != "" {
		env = append(env, "NO_PROXY="+proxyConf.NoProxy)
		env = append(env, "no_proxy="+proxyConf.NoProxy)
	}

	return env
}
