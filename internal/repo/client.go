package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	m "load-balancer/internal/models"
)

const (
	createClientQuery = `INSERT INTO clients (client_id, capacity, rate_per_second, tokens, last_refill_at) VALUES ($1, $2, $3, $4, $5)`
	getClientQuery    = `SELECT client_id, capacity, rate_per_second, tokens, last_refill_at FROM clients WHERE client_id = $1`
	updateClientQuery = `UPDATE clients SET capacity = $2, rate_per_second = $3, tokens = $4, last_refill_at = $5 WHERE client_id = $1`
	deleteClientQuery = `DELETE FROM clients WHERE client_id = $1`
)

func (r *repository) CreateClient(ctx context.Context, client m.RateLimitClient) error {
	_, err := r.pool.Exec(ctx, createClientQuery, client.ClientID, client.Capacity, client.RatePerSecond, client.Tokens, client.LastRefillAt)
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	return nil
}

func (r *repository) GetClientByID(ctx context.Context, id string) (*m.RateLimitClient, error) {
	client := &m.RateLimitClient{ClientID: id}
	err := r.pool.QueryRow(ctx, getClientQuery, id).Scan(&client.Capacity, &client.RatePerSecond, &client.Tokens, &client.LastRefillAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(err, "failed to query client")
		}
	}
	return client, nil
}

func (r *repository) UpdateClient(ctx context.Context, client m.RateLimitClient) error {
	commandTag, err := r.pool.Exec(ctx, updateClientQuery, client.ClientID, client.Capacity, client.RatePerSecond, client.Tokens, client.LastRefillAt)
	if err != nil {
		return errors.Wrap(err, "failed to update client")
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("No rows were affected, client with given id not found")
	}

	return nil
}

func (r *repository) DeleteClient(ctx context.Context, id string) error {
	commandTag, err := r.pool.Exec(ctx, deleteClientQuery, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete client")
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("No rows were affected, client with given id not found")
	}

	return nil
}
