package config

import (
	"sync"
)

type Upstream struct {
	Name        string            `yaml:"name"`
	Servers     []*UpstreamServer `yaml:"servers"`
	HealthCheck HealthCheck       `yaml:"healthCheck"`
	RateLimit   RateLimit         `yaml:"rateLimit"`
}

type UpstreamServer struct {
	Url    string `yaml:"url"`
	Status bool   `yaml:"status"`
	mu     sync.Mutex
}

func (server *UpstreamServer) SetStatus(status bool) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.Status = status
}

func (server *UpstreamServer) GetStatus() bool {
	server.mu.Lock()
	defer server.mu.Unlock()
	return server.Status
}
