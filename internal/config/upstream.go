package config

import (
	"strings"
	"sync"

	"github.com/janithht/GoStreamBalancer/internal/ratelimits"
)

type Upstream struct {
	Name        string            `yaml:"name"`
	Servers     []*UpstreamServer `yaml:"servers"`
	HealthCheck HealthCheck       `yaml:"healthCheck"`
	RateLimit   RateLimit         `yaml:"rateLimit"`
	LbType      string            `yaml:"lbType"`
	Limiter     *ratelimits.RateLimiter
}

type UpstreamServer struct {
	Url               string `yaml:"url"`
	Status            bool   `yaml:"status"`
	ActiveConnections int
	mu                sync.Mutex
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

func (server *UpstreamServer) IncrementConnections() {
	server.mu.Lock()
	server.ActiveConnections++
	server.mu.Unlock()
}

func (server *UpstreamServer) DecrementConnections() {
	server.mu.Lock()
	if server.ActiveConnections > 0 {
		server.ActiveConnections--
	}
	server.mu.Unlock()
}

func BuildUpstreamConfigs(upstreams []Upstream) (map[string]*IteratorImpl, map[string]*Upstream) {
	upstreamMap := make(map[string]*IteratorImpl)
	upstreamConfigMap := make(map[string]*Upstream)

	for i := range upstreams {
		upstream := &upstreams[i]
		iterator := NewIterator()
		for _, server := range upstream.Servers {
			iterator.Add(server)
		}
		upstream.Limiter = ratelimits.NewRateLimiter(upstream.RateLimit.Limit, upstream.RateLimit.Interval)
		upstreamMap[strings.ToLower(upstream.Name)] = iterator
		upstreamConfigMap[strings.ToLower(upstream.Name)] = upstream
	}
	return upstreamMap, upstreamConfigMap
}
