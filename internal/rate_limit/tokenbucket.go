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

// NewTokenBucket — создаёт новый Token Bucket с подключением к репозиторию и логгером, инициализирует внутреннее хранилище клиентов.
func NewTokenBucket(repo repo.Repository, logger *zap.SugaredLogger) *TokenBucket {
	return &TokenBucket{
		repo:    repo,
		logger:  logger,
		clients: make(map[string]*models.RateLimitState),
	}
}

// Allow — проверяет, есть ли у клиента токены для выполнения запроса, обновляет состояние bucket'а, загружает данные
// из БД при необходимости.
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

// StartBackgroundSync — запускает фоновую горутину, которая периодически сохраняет изменения по клиентам в БД.
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

// syncToDB — записывает в БД только изменённые (грязные) клиенты, помечая их как синхронизированные.
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

// StartInactiveCleaner — запускает фоновый процесс очистки неактивных клиентов из памяти через заданный интервал.
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

// cleanupInactiveClients — удаляет из памяти клиентов, которые давно не делали запросов.
func (tb *TokenBucket) cleanupInactiveClients(timeout time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for id, state := range tb.clients {
		if time.Since(state.LastSeen) > timeout {
			delete(tb.clients, id)
		}
	}
}

// ReplenishAll — фоновая задача, которая регулярно пополняет токены всех клиентов напрямую из БД.
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

// replenishTick — выполняет одно пополнение токенов всех клиентов в БД, исходя из времени последнего пополнения.
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
