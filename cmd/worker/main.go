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
	cfg, err := config.Load()
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

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	cleanup := app.NewCleanupService(repo.Sessions(), repo.Documents(), repo.Profiles(), clock.System{}, cfg.Retention.ProfileDays)
	expiryTicker := time.NewTicker(time.Minute)
	retentionTicker := time.NewTicker(5 * time.Minute)
	defer expiryTicker.Stop()
	defer retentionTicker.Stop()

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

	runExpiry()
	runRetention()
	slog.Info("worker 已启动")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-expiryTicker.C:
			runExpiry()
		case <-retentionTicker.C:
			runRetention()
		}
	}
}
