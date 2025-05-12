package server

import (
	"fmt"
	"go.uber.org/zap"
	"load-balancer/internal/models"
	"load-balancer/pkg/balancing_algorithms"
	"net/http"
	"net/http/httputil"
)

type Server struct {
	balancer balancing_algorithms.Balancer
	logger   *zap.SugaredLogger
	port     int
}

func NewServer(backends []*models.Backend, logger *zap.SugaredLogger, port int) *Server {
	return &Server{
		balancer: balancing_algorithms.NewRoundRobinBalancer(backends, logger),
		logger:   logger,
		port:     port,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.proxyHandler())
	s.logger.Infow("Starting server", "port", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) proxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := s.balancer.Next()
		if backend == nil {
			s.logger.Errorw("There is no available backend")
			http.Error(w, "All backends are unavailable", http.StatusServiceUnavailable)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		proxy.ServeHTTP(w, r)
	}
}
