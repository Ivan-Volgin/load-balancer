package rate_limit

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"load-balancer/internal/models"
	"load-balancer/internal/repo"
)

type TokenBucket struct {
	repo    repo.Repository
	mu      sync.Mutex
	clients map[string]*models.RateLimitState
	logger  *zap.SugaredLogger
}

func NewTokenBucket(repo repo.Repository, logger *zap.SugaredLogger) *TokenBucket {
	return &TokenBucket{
		repo:    repo,
		logger:  logger,
		clients: make(map[string]*models.RateLimitState),
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, clientID string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now().Unix()
	cl, ok := tb.clients[clientID]

	if !ok {
		dbClient, err := tb.repo.GetClientByID(ctx, clientID)
		if err != nil {
			return false, errors.Wrap(err, "error getting client")
		}
		cl = &models.RateLimitState{
			ClientID:      dbClient.ClientID,
			Capacity:      dbClient.Capacity,
			RatePerSecond: dbClient.RatePerSecond,
			Tokens:        dbClient.Tokens,
			LastRefillAt:  dbClient.LastRefillAt,
			Dirty:         false,
		}
		tb.clients[clientID] = cl
	}

	interval := now - cl.LastRefillAt
	newTokens := interval * cl.RatePerSecond
	cl.Tokens = min(cl.Tokens+newTokens, cl.Capacity)
	cl.LastRefillAt = now
	cl.LastSeen = time.Now()

	if cl.Tokens < 1 {
		return false, nil
	}

	cl.Tokens--
	cl.Dirty = true

	return true, nil
}

func (tb *TokenBucket) StartBackgroundSync(ctx context.Context, repo repo.Repository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			tb.syncToDB(repo)
		}
	}
}

func (tb *TokenBucket) syncToDB(repo repo.Repository) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	var toUpdate []*models.RateLimitState
	for _, state := range tb.clients {
		if state.Dirty {
			toUpdate = append(toUpdate, state)
			state.Dirty = false
		}
	}

	for _, state := range toUpdate {
		err := repo.UpdateClient(context.Background(), models.RateLimitClient{
			ClientID:      state.ClientID,
			Capacity:      state.Capacity,
			RatePerSecond: state.RatePerSecond,
			Tokens:        state.Tokens,
			LastRefillAt:  state.LastRefillAt,
		})
		if err != nil {
			state.Dirty = true
		}
	}
}

func (tb *TokenBucket) StartInactiveCleaner(ctx context.Context, inactiveTimeout time.Duration, tickerInterval time.Duration) {
	ticker := time.NewTicker(tickerInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			tb.cleanupInactiveClients(inactiveTimeout)
		}
	}
}

func (tb *TokenBucket) cleanupInactiveClients(timeout time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for id, state := range tb.clients {
		if time.Since(state.LastSeen) > timeout {
			delete(tb.clients, id)
		}
	}
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
