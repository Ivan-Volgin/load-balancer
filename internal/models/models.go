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
	ClientID      string // уникальный ключ (IP или API-ключ)
	Capacity      int64  // ёмкость bucket
	RatePerSecond int64  // скорость пополнения токенов
	Tokens        int64  // текущее количество токенов
	LastRefillAt  int64  // время последнего пополнения (Unix timestamp)
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
