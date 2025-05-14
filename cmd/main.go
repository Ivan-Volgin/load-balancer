package main

import (
	"context"
	"flag"
	"load-balancer/internal/rate_limit"
	"load-balancer/internal/repo"
	"load-balancer/internal/server/middleware"
	"load-balancer/internal/service"
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
	logger.Infow("config loaded", "config", cfg)

	dbRepo, err := repo.NewRepository(context.Background(), cfg.PostgreSQL)
	if err != nil {
		logger.Fatalw("failed to connect to database", "error", err)
	}
	logger.Infow("connected to database")

	var backends []*models.Backend
	for _, urlStr := range cfg.Backends {
		url, err := url.Parse(urlStr)
		if err != nil {
			logger.Fatalw("failed to parse url", "url", urlStr, "error", err)
			continue
		}
		backends = append(backends, &models.Backend{
			URL:       url,
			Available: true,
			Mu:        sync.Mutex{},
		})
	}

	balancerFactory := balancing_algorithms.NewBalancerFactory(logger)
	balancer := balancerFactory.Create(backends, cfg.BalanceStrategy)
	balancing_algorithms.StartHealthCheck(backends, time.Second*5, logger)

	proxyService := service.NewProxyService(balancer, logger)

	clientService := service.NewClientService(dbRepo)

	tokenBucket := rate_limit.NewTokenBucket(dbRepo, logger)
	go tokenBucket.StartBackgroundSync(context.Background(), dbRepo, time.Second*7)
	go tokenBucket.StartBackgroundSync(context.Background(), dbRepo, time.Minute)
	go tokenBucket.ReplenishAll(context.Background(), time.Second*5)

	rateLimiter := middleware.NewRateLimitMiddleware(tokenBucket, dbRepo)

	server.RegisterRoutes(proxyService, clientService, rateLimiter)

	srv := server.NewServer(
		logger,
		cfg.Port,
	)

	logger.Infow("Starting server", "port", cfg.Port)
	if err := srv.Start(); err != nil {
		logger.Fatalw("failed to start server", "error", err)
	}
}
