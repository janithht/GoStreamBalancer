package config

import "sync"

type Iterator interface {
	Add(item any)
	Next() (item any)
}

type LeastConnectionsIterator struct {
	mu    sync.Mutex
	items []*UpstreamServer
}

func NewLeastConnectionsIterator() *LeastConnectionsIterator {
	return &LeastConnectionsIterator{}
}

func (l *LeastConnectionsIterator) Add(server *UpstreamServer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, server)
}

func (l *LeastConnectionsIterator) Next() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.items) == 0 {
		return nil
	}

	server := l.items[0]
	l.items = append(l.items[1:], server)
	return server
}

func (l *LeastConnectionsIterator) NextHealthy() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.items) == 0 {
		return nil
	}

	var leastConnServer *UpstreamServer
	minConnections := int(^uint(0) >> 1)

	for _, server := range l.items {
		if server.GetStatus() && server.ActiveConnections < minConnections {
			minConnections = server.ActiveConnections
			leastConnServer = server
		}
	}

	if leastConnServer != nil {
		leastConnServer.IncrementConnections()
	}
	return leastConnServer
}
