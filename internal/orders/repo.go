package orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Order struct {
	ID     int
	Status string
}

type Repository interface {
	GetOrder(ctx context.Context, id int) (*Order, error)
}

func NewRepository(readDB, writeDB *pgxpool.Pool) Repository {
	return postgresBackedRepo{writeDB, readDB}
}

type postgresBackedRepo struct {
	writeDB *pgxpool.Pool
	readDB  *pgxpool.Pool
}

func (r postgresBackedRepo) GetOrder(ctx context.Context, id int) (*Order, error) {
	var order Order

	rows, err := r.readDB.Query(ctx, `SELECT id, status FROM orders WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	err = rows.Scan(&order.ID, &order.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve order from database: %w", err)
	}

	return &order, nil
}
