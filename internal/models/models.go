package models

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL       *url.URL
	Available bool
	Mu        sync.Mutex
}
