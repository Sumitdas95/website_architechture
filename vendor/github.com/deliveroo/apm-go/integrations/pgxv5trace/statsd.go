package pgxv5trace

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonboulle/clockwork"

	apmgo "github.com/deliveroo/apm-go"
)

const (
	tick   = 10 * time.Second
	prefix = "pgxpool."
)

var clock = clockwork.NewRealClock()

// Connect returns a pool that is watched.
// Metrics are transmitted to StatsD every 10 seconds.
func Connect(ctx context.Context, dbName string, cfg *pgxpool.Config, apm apmgo.Service) (*pgxpool.Pool, error) {
	watch(dbName, cfg, apm)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	go monitor(ctx, dbName, pool, apm)
	return pool, err
}

// watch injects functions to transmit metrics about pool usage.
func watch(dbName string, cfg *pgxpool.Config, apm apmgo.Service) {
	incr := func(name string) {
		apm.StatsD().Incr(prefix+dbName+"."+name, 1)
	}
	cfg.AfterConnect = func(context.Context, *pgx.Conn) error {
		incr("pool.added")
		return nil
	}
	cfg.BeforeAcquire = func(context.Context, *pgx.Conn) bool {
		incr("pool.acquired")
		return true
	}
	cfg.AfterRelease = func(*pgx.Conn) bool {
		incr("pool.released")
		return true
	}
}

func monitor(ctx context.Context, dbName string, p *pgxpool.Pool, apm apmgo.Service) {
	gauge := func(name string, value int64) {
		apm.StatsD().Gauge(prefix+dbName+"."+name, float64(value), 1)
	}
	transmit := func() {
		s := p.Stat()
		gauge("connections.active", int64(s.AcquiredConns()))
		gauge("connections.idle", int64(s.IdleConns()))
		gauge("connections.max", int64(s.MaxConns()))
		gauge("connections.total", int64(s.TotalConns()))

		gauge("acquire.blocked", s.EmptyAcquireCount())
		gauge("acquire.canceled", s.CanceledAcquireCount())
		gauge("acquire.count", s.AcquireCount())
		gauge("acquire.duration", int64(s.AcquireDuration()))
	}
	ticker := clock.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.Chan():
			transmit()
		case <-ctx.Done():
			transmit()
			return
		}
	}
}
