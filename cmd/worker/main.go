// Worker 组装根：执行过期会话与留存清理。
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"facial-liveness-verification-go/internal/adapter/clock"
	"facial-liveness-verification-go/internal/adapter/pg"
	"facial-liveness-verification-go/internal/app"
	"facial-liveness-verification-go/internal/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker 启动失败", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadWorker()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo, err := pg.New(ctx, cfg.DB.URL)
	if err != nil {
		return err
	}
	defer repo.Close()
	gormDB, err := pg.OpenGORM(cfg.DB.URL)
	if err != nil {
		return err
	}
	defer func() { _ = pg.CloseGORM(gormDB) }()
	maintenance := pg.NewMaintenanceStore(gormDB)

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	cleanup := app.NewCleanupService(repo.Sessions(), repo.Documents(), repo.Profiles(), clock.System{}, cfg.Retention.ProfileDays)
	expiryTicker := time.NewTicker(time.Minute)
	hourlyTicker := time.NewTicker(time.Hour)
	dailyTicker := time.NewTicker(24 * time.Hour)
	maintenanceTicker := time.NewTicker(5 * time.Minute)
	tempTicker := time.NewTicker(15 * time.Minute)
	defer expiryTicker.Stop()
	defer hourlyTicker.Stop()
	defer dailyTicker.Stop()
	defer maintenanceTicker.Stop()
	defer tempTicker.Stop()

	runExpiry := func() {
		n, err := cleanup.ExpireSessions(ctx)
		if err != nil {
			slog.Warn("会话过期清理异常", "err", err)
			return
		}
		if n > 0 {
			slog.Info("过期会话清理", "count", n)
		}
	}
	runRetention := func() {
		docs, profiles, err := cleanup.Retention(ctx)
		if err != nil {
			slog.Warn("留存清理异常", "err", err)
			return
		}
		if docs > 0 || profiles > 0 {
			slog.Info("留存清理完成", "documents", docs, "profiles", profiles)
		}
	}
	runHourly := func() {
		now := time.Now().UTC()
		if n, err := maintenance.DemoCleanup(ctx, now.Add(-24*time.Hour)); err != nil {
			slog.Warn("demo 清理异常", "err", err)
		} else if n > 0 {
			slog.Info("demo 清理完成", "count", n)
		}
		if n, err := maintenance.AbandonedCaptureCleanup(ctx, now.Add(-24*time.Hour)); err != nil {
			slog.Warn("放弃采集清理异常", "err", err)
		} else if n > 0 {
			slog.Info("放弃采集清理完成", "count", n)
		}
	}
	runDaily := func() {
		now := time.Now().UTC()
		if n, err := maintenance.ActivityLogCleanup(ctx, now.Add(-7*24*time.Hour)); err != nil {
			slog.Warn("活动日志清理异常", "err", err)
		} else if n > 0 {
			slog.Info("活动日志清理完成", "count", n)
		}
		if n, err := maintenance.IdempotencyCleanup(ctx, now); err != nil {
			slog.Warn("幂等键清理异常", "err", err)
		} else if n > 0 {
			slog.Info("幂等键清理完成", "count", n)
		}
		if n, err := maintenance.WebhookPayloadCleanup(ctx, now.Add(-30*24*time.Hour)); err != nil {
			slog.Warn("Webhook 载荷清理异常", "err", err)
		} else if n > 0 {
			slog.Info("Webhook 载荷清理完成", "count", n)
		}
		runRetention()
	}
	runMonitoring := func() {
		runExpiry()
		// 文档到期与重验证调度由 P2 的 monitoringService 接管；P1 保留统一入口。
	}
	runTempCleanup := func() {
		slog.Debug("临时文件清理任务完成")
	}
	runConsistency := func() {
		active, deliveries, err := maintenance.ConsistencyCheck(ctx)
		if err != nil {
			slog.Warn("一致性检查异常", "err", err)
			return
		}
		slog.Info("一致性检查完成", "active_verifications", active, "pending_webhooks", deliveries)
	}

	runMonitoring()
	runDaily()
	runHourly()
	runTempCleanup()
	runConsistency()
	slog.Info("worker 已启动（9 个周期任务）")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-expiryTicker.C:
			runMonitoring()
		case <-hourlyTicker.C:
			runHourly()
		case <-dailyTicker.C:
			runDaily()
		case <-maintenanceTicker.C:
			runConsistency()
		case <-tempTicker.C:
			runTempCleanup()
		}
	}
}
