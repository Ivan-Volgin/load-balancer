package balancing_algorithms

import (
	"go.uber.org/zap"
	"load-balancer/internal/models"
	"math/rand"
	"sync"
	"time"
)

type RandomBalancer struct {
	backends []*models.Backend
	mu       sync.Mutex
	log      *zap.SugaredLogger
	rand     *rand.Rand
}

// NewRandomBalancer — создаёт новый экземпляр балансировщика, который случайным образом выбирает из доступных бэкендов.
func NewRandomBalancer(backends []*models.Backend, logger *zap.SugaredLogger) *RandomBalancer {
	return &RandomBalancer{
		backends: backends,
		log:      logger,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Next — выбирает случайный доступный бэкенд и логирует выбор.
func (b *RandomBalancer) Next() *models.Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

	available := make([]*models.Backend, 0)
	for _, be := range b.backends {
		if be.Available {
			available = append(available, be)
		}
	}

	if len(available) == 0 {
		b.log.Errorw("There are no available backends")
		return nil
	}

	selected := available[b.rand.Intn(len(available))]
	b.log.Infow("Backend is chosen", "url", selected.URL.String())
	return selected
}
