package rate_limit

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"load-balancer/internal/repo"
	"github.com/pkg/errors"
)

type TokenBucket struct {
	repo   repo.Repository
	mu     sync.Mutex
	logger *zap.SugaredLogger
}

func NewTokenBucket(repo repo.Repository, logger zap.SugaredLogger) *TokenBucket {
	return &TokenBucket{
		repo:   repo,
		logger: &logger,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, clientID string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	client, err := tb.repo.GetClientByID(ctx, clientID)
	if err != nil {
		return false, err
	}

	if client == nil {
		return false, errors.New("client not found")
	}

	now := time.Now().Unix()
	interval := now - client.LastRefillAt
	newTokens := interval * client.RatePerSecond

	client.Tokens = min(client.Tokens+newTokens, client.Capacity)
	client.LastRefillAt = now

	if client.Tokens < 1 {
		return false, nil
	}

	client.Tokens--

	err = tb.repo.UpdateClient(ctx, *client)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (tb *TokenBucket) ReplenishAll(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := tb.replenishTick(ctx); err != nil {
				tb.logger.Errorf("Error during token replenishment: %v", err)
			}
		}
	}
}

func (tb *TokenBucket) replenishTick(ctx context.Context) error {
	clients, err := tb.repo.GetAllClients(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get all clients")
	}

	now := time.Now().Unix()

	for _, client := range clients {
		intervalSec := now - client.LastRefillAt
		newTokens := intervalSec * client.RatePerSecond

		client.Tokens = min(client.Tokens+newTokens, client.Capacity)
		client.LastRefillAt = now

		if err := tb.repo.UpdateClient(ctx, *client); err != nil {
			return err
		}
	}

	return nil
}
