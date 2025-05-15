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

// NewServer — создаёт новый сервер с указанием порта и логгером.
func NewServer(logger *zap.SugaredLogger, port *int) *Server {
	return &Server{
		logger: logger,
		port:   port,
	}
}

// Start — запускает HTTP-сервер на указанном порту и выводит сообщение о старте.
func (s *Server) Start() error {
	s.logger.Infow("Starting server", "port", s.port)
	addr := fmt.Sprintf(":%d", *s.port)
	return http.ListenAndServe(addr, nil)
}
