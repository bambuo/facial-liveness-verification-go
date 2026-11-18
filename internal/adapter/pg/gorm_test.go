package pg

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestAutoMigrateConcurrent 验证两个副本同时启动时只有一个执行迁移、另一个等待成功。
func TestAutoMigrateConcurrent(t *testing.T) {
	dsn := os.Getenv("TEST_GORM_CONCURRENT_URL")
	if dsn == "" {
		t.Skip("TEST_GORM_CONCURRENT_URL 未设置")
	}
	ctx := context.Background()
	db1, err := OpenGORM(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CloseGORM(db1) }()
	db2, err := OpenGORM(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CloseGORM(db2) }()

	start := make(chan struct{})
	errs := make(chan error, 2)
	go func() {
		<-start
		errs <- AutoMigrate(ctx, db1, 30*time.Second, 120*time.Second)
	}()
	go func() {
		<-start
		errs <- AutoMigrate(ctx, db2, 30*time.Second, 120*time.Second)
	}()
	close(start)
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent migrate: %v", err)
		}
	}

	sqlDB, err := db1.DB()
	if err != nil {
		t.Fatal(err)
	}
	var tables, fks int
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`).Scan(&fks); err != nil {
		t.Fatal(err)
	}
	if tables != 31 || fks != 43 {
		t.Fatalf("tables=%d fks=%d, want 31/43", tables, fks)
	}
}

// TestAutoMigrateSmoke 在空库上执行两次 AutoMigrate 并核对 31 表/329 列/43 外键。
func TestAutoMigrateSmoke(t *testing.T) {
	dsn := os.Getenv("TEST_GORM_URL")
	if dsn == "" {
		t.Skip("TEST_GORM_URL 未设置")
	}
	ctx := context.Background()
	db, err := OpenGORM(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = CloseGORM(db) })

	for i := 0; i < 2; i++ {
		if err := AutoMigrate(ctx, db, 30*time.Second, 120*time.Second); err != nil {
			t.Fatalf("migrate #%d: %v", i+1, err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	var tables, columns, fks, checks int
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema='public'`).Scan(&columns); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`).Scan(&fks); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM pg_constraint WHERE contype='c' AND connamespace='public'::regnamespace`).Scan(&checks); err != nil {
		t.Fatal(err)
	}
	if tables != 31 || columns != 329 {
		t.Fatalf("tables=%d columns=%d, want 31/329", tables, columns)
	}
	if fks != 43 {
		t.Fatalf("fks=%d, want 43", fks)
	}
	if checks != 9 {
		t.Fatalf("checks=%d, want 9", checks)
	}
}
