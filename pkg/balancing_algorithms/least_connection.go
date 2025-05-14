package balancing_algorithms

import (
	"go.uber.org/zap"
	"load-balancer/internal/models"
	"sync"
)

type LeastConnectionsBalancer struct {
	backends []*leastConnsBackend
	mu       sync.Mutex
	log      *zap.SugaredLogger
}

type leastConnsBackend struct {
	*models.Backend
	activeConnections int64
}

func NewLeastConnectionsBalancer(backends []*models.Backend, logger *zap.SugaredLogger) *LeastConnectionsBalancer {
	wrapped := make([]*leastConnsBackend, len(backends))
	for i, b := range backends {
		wrapped[i] = &leastConnsBackend{
			Backend:           b,
			activeConnections: 0,
		}
	}
	return &LeastConnectionsBalancer{
		backends: wrapped,
		log:      logger,
	}
}

func (b *LeastConnectionsBalancer) Next() *models.Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

	var selected *leastConnsBackend
	for _, backend := range b.backends {
		if !backend.Available {
			continue
		}
		if selected == nil || backend.activeConnections < selected.activeConnections {
			selected = backend
		}
	}

	if selected == nil {
		b.log.Errorw("There are no available backends")
		return nil
	}

	selected.activeConnections++

	b.log.Infow("Backend is chosen", "url", selected.URL.String())
	return selected.Backend
}
