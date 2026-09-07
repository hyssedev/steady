package database

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hyssedev/steady/internal/monitor"
)

func BenchmarkSaveCheck(b *testing.B) {
	ctx := context.Background()

	db, err := open(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	if err := createTables(ctx, db); err != nil {
		b.Fatal(err)
	}

	monitors, err := (Database{DB: db}).SyncMonitors(ctx, []monitor.Monitor{
		{
			Name: "Benchmark",
			URL:  mustURL(b, "https://example.com"),
		},
	})
	if err != nil {
		b.Fatal(err)
	}

	database := Database{DB: db}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := database.SaveCheck(
				ctx,
				monitors[0].ID,
				true,
				statusCode(http.StatusOK),
				(12 * time.Millisecond).Milliseconds(),
				nil,
			); err != nil {
				b.Error(err)
			}
		}
	})

	b.StopTimer()

	stats := db.Stats()
	b.Logf(
		"writes/sec=%.0f wait_count=%d wait_duration=%s open=%d in_use=%d idle=%d",
		float64(b.N)/b.Elapsed().Seconds(),
		stats.WaitCount,
		stats.WaitDuration,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
	)
}

func BenchmarkSaveCheckSequential(b *testing.B) {
	ctx := context.Background()

	db, err := open(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	if err := createTables(ctx, db); err != nil {
		b.Fatal(err)
	}

	monitors, err := (Database{DB: db}).SyncMonitors(ctx, []monitor.Monitor{
		{
			Name: "Benchmark",
			URL:  mustURL(b, "https://example.com"),
		},
	})
	if err != nil {
		b.Fatal(err)
	}

	database := Database{DB: db}

	b.ResetTimer()
	for b.Loop() {
		if err := database.SaveCheck(
			ctx,
			monitors[0].ID,
			true,
			statusCode(http.StatusOK),
			(12 * time.Millisecond).Milliseconds(),
			nil,
		); err != nil {
			b.Fatal(err)
		}
	}
}

func mustURL(b *testing.B, rawURL string) *url.URL {
	b.Helper()

	u, err := url.Parse(rawURL)
	if err != nil {
		b.Fatal(err)
	}

	return u
}

func statusCode(code int) *int {
	return &code
}
