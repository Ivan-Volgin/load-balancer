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

func NewBalancerFactory(logger *zap.SugaredLogger) BalancerFactory {
	return &balancerFactory{logger: logger}
}

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
