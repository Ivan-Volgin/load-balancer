package repo

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"load-balancer/internal/config"
	m "load-balancer/internal/models"
)

type repository struct {
	pool *pgxpool.Pool
}

type Repository interface {
	CreateClient(ctx context.Context, client m.RateLimitClient) error
	GetClientByID(ctx context.Context, id string) (*m.RateLimitClient, error)
	UpdateClient(ctx context.Context, client m.RateLimitClient) error
	DeleteClient(ctx context.Context, id string) error
	GetAllClients(ctx context.Context) ([]*m.RateLimitClient, error)
	Close()
}

// NewRepository — создаёт и настраивает пул подключений к PostgreSQL, используя переданные параметры из конфига.
func NewRepository(ctx context.Context, config config.PostgreSQL) (Repository, error) {
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s
       pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.SSLMode,
		config.PoolMaxConns,
		config.PoolMaxConnLifetime.String(),
		config.PoolMaxConnIdleTime.String(),
	)

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	return &repository{pool}, nil
}

// Close закрывает соединение с базой данных. Функция была созданна для graceful shutdown.
func (r *repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
