package scheduler

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"
	"tmd-pro/internal/config"
	"tmd-pro/pkg/logger"

	"tmd-pro/internal/executor"
	"tmd-pro/internal/storage"
)

// Scheduler 调度器
type Scheduler struct {
	storage      *storage.Storage
	executor     *executor.Executor
	ticker       *time.Ticker
	outputWriter io.Writer
}

// NewScheduler 创建调度器
func NewScheduler(storage *storage.Storage, executor *executor.Executor) *Scheduler {
	return &Scheduler{
		storage:  storage,
		executor: executor,
	}
}

// SetOutputWriter 设置自定义输出 Writer
func (s *Scheduler) SetOutputWriter(w io.Writer) {
	s.outputWriter = w
}

// StartWithRotation 启动轮换扫描模式
func (s *Scheduler) StartWithRotation(ctx context.Context) error {
	logger.Info("启动轮换扫描模式...")

	// 确保 tmd 二进制存在
	if err := s.executor.EnsureTmdBinary(); err != nil {
		logger.Errorf("准备 tmd 二进制失败: %v", err)
		return err
	}

	// 创建定时器
	interval := time.Duration(config.G.AppConf.ScanIntervalMinutes) * time.Minute
	s.ticker = time.NewTicker(interval)
	defer func() {
		if s.ticker != nil {
			s.ticker.Stop()
		}
	}()

	fix := s.fixLastScanId(ctx)
	if !fix {
		return nil
	}

	// 立即执行第一次扫描
	if !s.runRotationTask(ctx) {
		logger.Info("首次扫描未发现待处理数据，关闭定时器")
		s.ticker.Stop()
		return nil
	}

	logger.Infof("轮换扫描已启动，每隔 %v 分钟处理下一批次...", config.G.AppConf.ScanIntervalMinutes)
	logger.Info("按 Ctrl+C 停止服务")

	for {
		select {
		case <-s.ticker.C:
			if !s.runRotationTask(ctx) {
				logger.Info("未发现待处理数据，关闭轮换扫描定时器")
				s.ticker.Stop()
				return nil
			}
		case <-ctx.Done():
			logger.Info("收到停止信号，正在关闭...")
			return nil
		}
	}
}

// scanScreenNames 扫描screen_name，并执行tmd
func (s *Scheduler) scanScreenNames(ctx context.Context, batchNames []storage.ScreenName) error {
	logger.Info("开始扫描...")

	// 确保 tmd 二进制存在
	if err := s.executor.EnsureTmdBinary(); err != nil {
		logger.Errorf("准备 tmd 二进制失败: %v", err)
		return err
	}

	if len(batchNames) == 0 {
		logger.Warn("没有发现 screen_name")
		return nil
	}
	logger.Infof("共读取到 %d 个 screen_name", len(batchNames))

	// 开始执行 tmd
	totalSuccess := 0
	totalFail := 0
	startTime := time.Now()
	for i, sn := range batchNames {
		// 检查 context 是否已取消
		if ctx.Err() != nil {
			logger.Infof("扫描被中断，已完成: %d/%d", i, len(batchNames))
			return ctx.Err()
		}

		name := strings.TrimSpace(sn.Name)
		if name == "" {
			continue
		}

		logger.Infof("[%d/%d] 正在处理: %s", i+1, len(batchNames), name)

		var err error
		if s.outputWriter != nil {
			err = s.executor.ExecuteWithOutput(ctx, name, s.outputWriter, s.outputWriter)
		} else {
			err = s.executor.Execute(ctx, name)
		}
		if err != nil {
			if errors.Is(err, executor.ErrCanceled) || ctx.Err() != nil {
				logger.Infof("扫描被中断，已完成: %d/%d", i, len(batchNames))
				return ctx.Err()
			}
			logger.Errorf("处理 %s 失败: %v", name, err)
			totalFail++
		} else {
			logger.Infof("处理 %s 成功", name)
			totalSuccess++
		}
	}

	duration := time.Since(startTime)
	logger.Infof("扫描完成，耗时: %v，成功: %d，失败: %d",
		duration.Round(time.Second), totalSuccess, totalFail)

	return nil
}

// runRotationTask 执行轮换任务
// 返回值：true-继续后续扫描，false-停止定时器
func (s *Scheduler) runRotationTask(ctx context.Context) bool {
	// 获取上次扫描的最大 ID
	lastMaxId, err := s.storage.GetLastScanId(ctx)
	if err != nil {
		logger.Errorf("获取上次扫描最大ID失败: %v", err)
		return true
	}

	// 获取本次需要扫描的用户
	names, err := s.storage.GetScreenNameByLastMaxId(ctx, lastMaxId, 50)
	if err != nil {
		logger.Errorf("获取本次需要扫描的用户失败: %v", err)
		return true
	}

	if len(names) == 0 {
		logger.Info("本次未查询到待扫描的用户，轮换扫描完成")
		return false
	}

	// 开始扫描
	err = s.scanScreenNames(ctx, names)
	if err != nil {
		logger.Errorf("本次扫描操作失败: %v", err)
		return true
	}

	lastMaxId = names[len(names)-1].ID

	err = s.storage.UpdateLastScanId(ctx, lastMaxId)
	if err != nil {
		logger.Errorf("更新LastMaxId失败: %v", err)
	}

	return true
}

// fixLastScanId 更新lastScanId
func (s *Scheduler) fixLastScanId(ctx context.Context) bool {
	// 获取上次扫描的 maxId
	lastMaxId, err := s.storage.GetLastScanId(ctx)
	if err != nil {
		logger.Errorf("获取LastScanId失败: %v", err)
		return false
	}

	maxId, err := s.storage.GetMaxScreenNameId(ctx)
	if err != nil {
		logger.Errorf("获取maxId失败: %v", err)
		return false
	}

	if lastMaxId == maxId {
		err := s.storage.UpdateLastScanId(ctx, 0)
		if err != nil {
			logger.Errorf("更新LastScanId失败: %v", err)
			return false
		}
	}
	return true
}
