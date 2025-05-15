package models

import (
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL       *url.URL
	Available bool
	Mu        sync.Mutex
}

type RateLimitClient struct {
	ClientID      string `json:"client_id"`
	Capacity      int64  `json:"capacity"`
	RatePerSecond int64  `json:"rate_per_second"`
	Tokens        int64  `json:"tokens"`
	LastRefillAt  int64  `json:"last_refill_at"`
}

type RateLimitState struct {
	ClientID      string
	Capacity      int64
	RatePerSecond int64
	Tokens        int64
	LastRefillAt  int64
	Dirty         bool
	LastSeen      time.Time
}
