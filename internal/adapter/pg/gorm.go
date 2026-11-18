package pg

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// migrationLockKey 是启动迁移使用的 PostgreSQL advisory lock 键。
const migrationLockKey int64 = 0x1d5af7

// OpenGORM 打开 GORM 连接池。
func OpenGORM(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Warn),
		PrepareStmt: true,
		NowFunc:     func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}
	return db, nil
}

// AutoMigrate 在 advisory lock 内对全部模型执行增量迁移。
func AutoMigrate(ctx context.Context, pool *gorm.DB, lockTimeout, statementTimeout time.Duration) error {
	sqlDB, err := pool.DB()
	if err != nil {
		return fmt.Errorf("gorm sql db: %w", err)
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migration conn: %w", err)
	}
	defer conn.Close()

	if lockTimeout > 0 {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("SET lock_timeout = '%dms'", lockTimeout.Milliseconds())); err != nil {
			return fmt.Errorf("set lock_timeout: %w", err)
		}
	}
	if statementTimeout > 0 {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("SET statement_timeout = '%dms'", statementTimeout.Milliseconds())); err != nil {
			return fmt.Errorf("set statement_timeout: %w", err)
		}
	}

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.ExecContext(unlockCtx, `SELECT pg_advisory_unlock($1)`, migrationLockKey)
	}()

	migrator, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		NowFunc:                                  func() time.Time { return time.Now().UTC() },
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("migration gorm: %w", err)
	}
	if err := migrator.AutoMigrate(gormmodel.All()...); err != nil {
		return fmt.Errorf("auto migrate tables: %w", err)
	}

	// 第二阶段：表已存在，再创建关联外键，避免环依赖与建表顺序问题。
	constraints, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return fmt.Errorf("constraint gorm: %w", err)
	}
	if err := constraints.AutoMigrate(gormmodel.All()...); err != nil {
		return fmt.Errorf("auto migrate constraints: %w", err)
	}
	return nil
}

// CloseGORM 关闭底层连接池。
func CloseGORM(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
