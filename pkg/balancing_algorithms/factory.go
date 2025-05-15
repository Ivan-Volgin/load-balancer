package balancing_algorithms

import (
	"go.uber.org/zap"
	"load-balancer/internal/models"
)

type BalancerFactory interface {
	Create(backends []*models.Backend, strategy string) Balancer
}

type balancerFactory struct {
	logger *zap.SugaredLogger
}

// NewBalancerFactory — создаёт фабрику балансировщиков, которая возвращает нужный тип стратегии по названию.
func NewBalancerFactory(logger *zap.SugaredLogger) BalancerFactory {
	return &balancerFactory{logger: logger}
}

// Create — метод фабрики, выбирающий конкретную реализацию балансировщика (Round Robin, Least Connections, Random).
func (f *balancerFactory) Create(backends []*models.Backend, strategy string) Balancer {
	switch strategy {
	case "round_robin":
		return NewRoundRobinBalancer(backends, f.logger)
	case "least_connections":
		return NewLeastConnectionsBalancer(backends, f.logger)
	case "random":
		return NewRandomBalancer(backends, f.logger)
	default:
		return NewRoundRobinBalancer(backends, f.logger)
	}
}
