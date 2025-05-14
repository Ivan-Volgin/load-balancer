package middleware

import (
	"fmt"
	"load-balancer/internal/rate_limit"
	"load-balancer/internal/repo"
	"load-balancer/internal/service"
	"net/http"
)

type RateLimitMiddleware struct {
	tokenBucket *rate_limit.TokenBucket
	repo        repo.Repository
}

func NewRateLimitMiddleware(tb *rate_limit.TokenBucket, repo repo.Repository) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		tokenBucket: tb,
		repo:        repo,
	}
}

func (middleware *RateLimitMiddleware) Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := r.Header.Get("X-API-KEY")
		if clientID == "" {
			service.WriteJSONError(w, http.StatusUnauthorized, "missing X-API-KEY header")
			return
		}

		allowed, err := middleware.tokenBucket.Allow(ctx, clientID)
		if err != nil {
			service.WriteJSONError(w, http.StatusInternalServerError, fmt.Sprintf("rate limit error: %v", err))
			return
		}

		if !allowed {
			service.WriteJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	}
}