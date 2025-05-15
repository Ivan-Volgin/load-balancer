package balancing_algorithms

import (
	"go.uber.org/zap"
	m "load-balancer/internal/models"
	"sync"
)

type RoundRobinBalancer struct {
	backends []*m.Backend
	mu       sync.Mutex
	curIndex int
	log      *zap.SugaredLogger
}

// NewRoundRobinBalancer — создаёт балансировщик с алгоритмом Round Robin и начальным индексом.
func NewRoundRobinBalancer(backends []*m.Backend, logger *zap.SugaredLogger) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: backends,
		curIndex: 0,
		log:      logger,
	}
}

// Next — выбирает следующий доступный бэкенд по кругу, пропуская недоступные, и обновляет текущий индекс.
func (b *RoundRobinBalancer) Next() *m.Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := 0; i < len(b.backends); i++ {
		idx := (b.curIndex + i) % len(b.backends)
		be := b.backends[idx]

		if be.Available {
			b.log.Infow("Backend is chosen", "url", be.URL.String())
			b.curIndex = (idx + 1) % len(b.backends)
			return be
		}
	}

	b.log.Errorw("There are no available backends")
	return nil
}
