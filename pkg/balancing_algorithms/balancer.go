package balancing_algorithms

import (
	"go.uber.org/zap"
	m "load-balancer/internal/models"
	"net/http"
	"time"
)

type Balancer interface {
	Next() *m.Backend
}

func StartHealthCheck(backends []*m.Backend, interval time.Duration, logger *zap.SugaredLogger) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for _, backend := range backends {
					ok := isBackendAlive(backend.URL.String())
					backend.Mu.Lock()
					prev := backend.Available
					backend.Available = ok
					backend.Mu.Unlock()

					if prev != ok {
						logger.Infow("Backend status changed",
							"url", backend.URL.String(),
							"available", backend.Available,
						)
					}
				}
			}
		}
	}()
}

func isBackendAlive(url string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	return err == nil && resp.StatusCode == http.StatusOK
}
