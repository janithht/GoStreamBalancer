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
