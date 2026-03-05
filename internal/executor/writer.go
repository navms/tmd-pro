package executor

import (
	"tmd-pro/pkg/logger"

	"go.uber.org/zap"
)

// StdLoggerWriter 适配 io.Writer 接口，将标准输出/标准错误内容写入日志
type StdLoggerWriter struct{}

// Write 实现 io.Writer 接口，将字节数据转为字符串写入Info日志
func (w *StdLoggerWriter) Write(p []byte) (n int, err error) {
	content := string(p)
	if content != "" && content != "\n" {
		logger.Info("tmd: ", zap.String("output", content))
	}
	return len(p), nil
}
