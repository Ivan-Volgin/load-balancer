package main

import (
	"flag"
	"net/url"
	"sync"
	"time"

	"load-balancer/internal/config"
	"load-balancer/internal/models"
	"load-balancer/internal/server"
	"load-balancer/pkg/balancing_algorithms"
	"load-balancer/pkg/logger"
)

var configPath = flag.String("config", "config.yaml", "config file path")

func main() {
	flag.Parse()

	logger, err := logger.NewLogger("info")
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalw("failed to load config file", "error", err)
	}

	var backends []*models.Backend
	for _, urlStr := range cfg.Backends {
		url, err := url.Parse(urlStr)
		if err != nil {
			logger.Fatalw("failed to parse url", "url", urlStr, "error", err)
		}
		backends = append(backends, &models.Backend{
			URL:       url,
			Available: true,
			Mu:        sync.Mutex{},
		})
	}

	balancing_algorithms.StartHealthCheck(backends, time.Second*5, logger)

	srv := server.NewServer(
		backends,
		logger,
		cfg.Port,
	)

	logger.Infow("Starting server", "port", cfg.Port)
	if err := srv.Start(); err != nil {
		logger.Fatalw("failed to start server", "error", err)
	}
}
