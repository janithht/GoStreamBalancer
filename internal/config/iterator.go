package config

import (
	"hash/fnv"
	"sync"
)

type Iterator interface {
	Add(server *UpstreamServer)
	Next() *UpstreamServer
}

type IteratorImpl struct {
	mu      sync.Mutex
	servers []*UpstreamServer
}

func NewIterator() *IteratorImpl {
	return &IteratorImpl{}
}

func (l *IteratorImpl) Add(server *UpstreamServer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.servers = append(l.servers, server)
}

func (l *IteratorImpl) Next() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.servers) == 0 {
		return nil
	}

	server := l.servers[0]
	l.servers = append(l.servers[1:], server)
	return server
}

func (l *IteratorImpl) NextHealthy() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.servers) == 0 {
		return nil
	}

	var leastConnServer *UpstreamServer
	minConnections := int(^uint(0) >> 1)

	for _, server := range l.servers {
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

func (iterator *IteratorImpl) MatchServer(clientIP string) *UpstreamServer {
	iterator.mu.Lock()
	defer iterator.mu.Unlock()

	if len(iterator.servers) == 0 {
		return nil
	}

	// Calculate the hash of the client's IP address
	hasher := fnv.New32()
	hasher.Write([]byte(clientIP))
	index := int(hasher.Sum32()) % len(iterator.servers)

	return iterator.servers[index]
}
