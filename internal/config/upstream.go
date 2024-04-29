package config

import (
	"strings"
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

func BuildUpstreamMap(upstreams []Upstream) map[string]*RoundRobinIterator {
	upstreamMap := make(map[string]*RoundRobinIterator)
	for i := range upstreams {
		upstream := &upstreams[i]
		iterator := NewRoundRobinIterator()
		for _, server := range upstream.Servers {
			iterator.Add(server) // Add all servers initially
		}
		upstreamMap[strings.ToLower(upstream.Name)] = iterator
	}
	return upstreamMap
}
