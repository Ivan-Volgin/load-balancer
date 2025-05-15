package service

import (
	"go.uber.org/zap"
	"load-balancer/pkg/balancing_algorithms"
	"net/http"
	"net/http/httputil"
)

type ProxyService struct {
	balancer balancing_algorithms.Balancer
	logger   *zap.SugaredLogger
}

// NewProxyService — создаёт сервис прокси с указанием балансировщика и логгера.
func NewProxyService(balancer balancing_algorithms.Balancer, logger *zap.SugaredLogger) *ProxyService {
	return &ProxyService{
		balancer: balancer,
		logger:   logger,
	}
}

// ProxyHandler — основной обработчик запросов: выбирает бэкенд, проксирует запрос, помечает недоступные бэкенды.
func (ps *ProxyService) ProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := ps.balancer.Next()
		if backend == nil {
			ps.logger.Errorw("There is no available service")
			http.Error(w, "All backends are unavailable", http.StatusServiceUnavailable)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, err error) {
			ps.logger.Errorw("Error while request redirection",
				"service", backend,
				"error", err.Error())

			backend.Mu.Lock()
			backend.Available = false
			backend.Mu.Unlock()
			ps.logger.Infow("Backend status switched to unavailable",
				"service", backend.URL.String())

			http.Error(rw, err.Error(), http.StatusBadGateway)
		}
		proxy.ServeHTTP(w, r)
	}
}
