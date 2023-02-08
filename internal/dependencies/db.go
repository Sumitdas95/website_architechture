package dependencies

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"

	"github.com/deliveroo/apm-go"
	"github.com/deliveroo/apm-go/integrations/pgxv5trace"
)

// InitDatabase initializes a Postgres database connection.
func InitDatabase(url string, apm apm.Service) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	// Add github.com/google/uuid type support
	pgxConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pgxConnPool, err := pgxv5trace.Connect(context.Background(), pgxConfig.ConnConfig.Database, pgxConfig, apm)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return pgxConnPool, nil
}

// CloseDatabaseConnection cleans up the connection to the db
func CloseDatabaseConnection(pgxConnPool *pgxpool.Pool) {
	pgxConnPool.Close()
}
