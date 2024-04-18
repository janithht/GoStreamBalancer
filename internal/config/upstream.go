package config

import (
	"sync"
)

type Upstream struct {
	Name        string           `yaml:"name"`
	Servers     []UpstreamServer `yaml:"servers"`
	HealthCheck HealthCheck      `yaml:"healthCheck"`
	RateLimit   RateLimit        `yaml:"rateLimit"`
}

type UpstreamServer struct {
	Url    string `yaml:"url"`
	Status bool   `yaml:"status"`
}

type Iterator interface {
	Add(item any)
	Next() (item any)
}

func NewUpstream(servers []UpstreamServer, iterator Iterator) *Upstream {
	for _, svr := range servers {
		iterator.Add(svr)
	}
	return &Upstream{
		Servers: servers,
	}
}

type RoundRobinIterator struct {
	mu    sync.Mutex
	items []any
}

func NewRoundRobinIterator() *RoundRobinIterator {
	return &RoundRobinIterator{}
}

func (r *RoundRobinIterator) Add(item any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, item)
}

func (r *RoundRobinIterator) Next() (item any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) == 0 {
		return nil
	}
	item = r.items[0]
	r.items = append(r.items[1:], item) // Rotate the list
	return item                         // Always return the next item, regardless of status
}
