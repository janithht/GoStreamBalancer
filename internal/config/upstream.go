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

type Iterator interface {
	Add(item any)
	Next() (item any)
}

func NewUpstream(servers []*UpstreamServer, iterator Iterator) *Upstream {
	for _, svr := range servers {
		iterator.Add(svr)
	}
	return &Upstream{
		Servers: servers,
	}
}

type RoundRobinIterator struct {
	mu    sync.Mutex
	items []*UpstreamServer
}

func NewRoundRobinIterator() *RoundRobinIterator {
	return &RoundRobinIterator{}
}

func (r *RoundRobinIterator) Add(item any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if server, ok := item.(*UpstreamServer); ok {
		r.items = append(r.items, server)
	}
}

func (r *RoundRobinIterator) Next() (item any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) == 0 {
		return nil
	}
	item = r.items[0]
	r.items = append(r.items[1:], item.(*UpstreamServer))
	return item
}
