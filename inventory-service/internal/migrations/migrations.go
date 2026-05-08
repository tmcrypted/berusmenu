package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Apply(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		up := upSQL(string(content))
		if strings.TrimSpace(up) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, up); err != nil {
			return fmt.Errorf("apply migration %s: %w", file, err)
		}
	}

	return nil
}

func upSQL(content string) string {
	const down = "-- +goose Down"
	idx := strings.Index(content, down)
	if idx >= 0 {
		content = content[:idx]
	}
	content = strings.Replace(content, "-- +goose Up", "", 1)
	return content
}
