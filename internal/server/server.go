package server

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	logger *zap.SugaredLogger
	port   *int
}

func NewServer(logger *zap.SugaredLogger, port *int) *Server {
	return &Server{
		logger: logger,
		port:   port,
	}
}

func (s *Server) Start() error {
	s.logger.Infow("Starting server", "port", s.port)
	addr := fmt.Sprintf(":%d", *s.port)
	return http.ListenAndServe(addr, nil)
}
