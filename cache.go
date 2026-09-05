package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/pressly/goose/v3"

	_ "github.com/malikbenkirane/atty/migrations"
)

type repo interface {
	Log(title string) error
	MostRecent(limit int) ([]titleAccess, error)
	Close() error
	Path() string
}

type db struct {
	*sql.DB
	path string
}

//go:embed migrations
var migrations embed.FS

func initCache(ctx context.Context) (repo, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("os: user cache dir: %w", err)
	}
	cacheDir = filepath.Join(cacheDir, "atty")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("os: mkdir all: %q: %w", cacheDir, err)
	}
	cacheFile := filepath.Join(cacheDir, "cache.db")
	conn, err := sql.Open("sqlite", cacheFile+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("sql: open: %w", err)
	}
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("fs: sub: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, conn, fsys)
	if err != nil {
		return nil, fmt.Errorf("goose: new provider: %w", err)
	}
	_, err = provider.Up(ctx)
	return &db{
		DB:   conn,
		path: cacheFile,
	}, nil
}

func (db *db) Log(title string) error {
	unixTimestamp := time.Now().Unix()
	db.Exec(`
		INSERT INTO access_history (title, accessed_at) VALUES (?, ?)
	`, title, unixTimestamp)
	return nil
}

type titleAccess struct {
	title      string
	accessedAt time.Time
}

func (db *db) MostRecent(limit int) ([]titleAccess, error) {
	rows, err := db.Query(`
		SELECT title, accessed_at FROM access_history
		ORDER BY accessed_at DESC  
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("sql: query: %w", err)
	}
	accesses := make([]titleAccess, 0, limit)
	for rows.Next() {
		var ta titleAccess
		if err := rows.Scan(&ta.title, &ta.accessedAt); err != nil {
			return nil, fmt.Errorf("sql: scan row: %w", err)
		}
		accesses = append(accesses, ta)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sql: next row: %w", err)
	}
	return accesses, nil
}

func (db *db) Close() error {
	return db.DB.Close()
}

func (db *db) Path() string {
	return db.path
}
