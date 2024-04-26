package config

import "sync"

type Iterator interface {
	Add(item any)
	Next() (item any)
}

type RoundRobinIterator struct {
	mu    sync.Mutex
	items []*UpstreamServer
}

func NewRoundRobinIterator() *RoundRobinIterator {
	return &RoundRobinIterator{}
}

func (r *RoundRobinIterator) Add(server *UpstreamServer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, server)
}

func (r *RoundRobinIterator) Next() *UpstreamServer {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.items) == 0 {
		return nil
	}

	server := r.items[0]
	r.items = append(r.items[1:], server)
	return server
}

func (r *RoundRobinIterator) NextHealthy() *UpstreamServer {
	r.mu.Lock()
	defer r.mu.Unlock()

	originalCount := len(r.items)
	for count := 0; count < originalCount; count++ {
		server := r.items[0]
		r.items = append(r.items[1:], server)

		if server.GetStatus() {
			return server
		}
	}
	return nil
}
