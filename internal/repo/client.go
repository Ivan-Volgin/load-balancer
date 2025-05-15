package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	m "load-balancer/internal/models"
)

const (
	createClientQuery  = `INSERT INTO clients (client_id, capacity, rate_per_second, tokens, last_refill_at) VALUES ($1, $2, $3, $4, $5)`
	getClientQuery     = `SELECT capacity, rate_per_second, tokens, last_refill_at FROM clients WHERE client_id = $1`
	updateClientQuery  = `UPDATE clients SET capacity = $2, rate_per_second = $3, tokens = $4, last_refill_at = $5 WHERE client_id = $1`
	deleteClientQuery  = `DELETE FROM clients WHERE client_id = $1`
	getAllClientsQuery = `SELECT client_id, capacity, rate_per_second, tokens, last_refill_at FROM clients`
)

// CreateClient — добавляет нового клиента в БД с заданными лимитами и состоянием токенов.
func (r *repository) CreateClient(ctx context.Context, client m.RateLimitClient) error {
	_, err := r.pool.Exec(ctx, createClientQuery, client.ClientID, client.Capacity, client.RatePerSecond, client.Tokens, client.LastRefillAt)
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	return nil
}

// GetClientByID — получает данные клиента по ID из БД, возвращает ошибку, если клиент не найден.
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

// UpdateClient — обновляет данные клиента в БД, проверяет, была ли затронута хотя бы одна строка.
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

// DeleteClient — удаляет клиента из БД, возвращает ошибку, если запись не найдена.
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

// GetAllClients — возвращает список всех клиентов из БД, используется для фонового пополнения токенов.
func (r *repository) GetAllClients(ctx context.Context) ([]*m.RateLimitClient, error) {
	rows, err := r.pool.Query(ctx, getAllClientsQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query all clients")
	}
	defer rows.Close()

	var clients []*m.RateLimitClient
	for rows.Next() {
		var client m.RateLimitClient
		if err := rows.Scan(
			&client.ClientID,
			&client.Capacity,
			&client.RatePerSecond,
			&client.Tokens,
			&client.LastRefillAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		clients = append(clients, &client)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows iteration error")
	}

	return clients, nil
}
