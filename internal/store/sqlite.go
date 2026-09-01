package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore SQLite 实现
type SQLiteStore struct {
	db *sql.DB
}

type txContextKey struct{}

type txExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type txQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (s *SQLiteStore) txFrom(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sql.Tx)
	return tx
}

func (s *SQLiteStore) execer(ctx context.Context) txExecer {
	if tx := s.txFrom(ctx); tx != nil {
		return tx
	}
	return s.db
}

func (s *SQLiteStore) queryer(ctx context.Context) txQueryer {
	if tx := s.txFrom(ctx); tx != nil {
		return tx
	}
	return s.db
}

// WithTx 在一个 SQLite 事务中执行 Desired State 变更及其 Revision/Audit。
func (s *SQLiteStore) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// OpenSQLite 打开 SQLite 数据库并应用迁移
func OpenSQLite(path string) (*SQLiteStore, error) {
	dsn := path
	if path != ":memory:" {
		dsn = path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

// OpenInMemory 测试用内存库
func OpenInMemory() (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

// DB 暴露底层连接（供测试与扩展使用）
func (s *SQLiteStore) DB() *sql.DB { return s.db }

func (s *SQLiteStore) Close() error { return s.db.Close() }

var _ Store = (*SQLiteStore)(nil)

// now 统一时间格式
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func nullStr(v string) any {
	if v == "" {
		return nil
	}
	return v
}
