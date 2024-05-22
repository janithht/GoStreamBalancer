package config

import (
	"container/heap"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

type Iterator interface {
	Add(server *UpstreamServer)
	Next() *UpstreamServer
}

type IteratorImpl struct {
	mu           sync.RWMutex
	servers      ServerHeap
	currentIndex int
}

func NewIterator() *IteratorImpl {
	return &IteratorImpl{}
}

func (l *IteratorImpl) Add(server *UpstreamServer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	heap.Push(&l.servers, &PriorityServer{server: server})
}

func (l *IteratorImpl) Next() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.servers) == 0 {
		return nil
	}
	server := l.servers[l.currentIndex].server
	l.currentIndex = (l.currentIndex + 1) % len(l.servers)
	return server
}

func (l *IteratorImpl) NextRR() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.servers) == 0 {
		return nil
	}

	startIndex := l.currentIndex
	numServers := len(l.servers)

	for i := 0; i < numServers; i++ {
		index := (startIndex + i) % numServers
		server := l.servers[index].server
		if server.GetStatus() {
			l.currentIndex = (index + 1) % numServers
			return server
		}
	}
	return nil
}

func (l *IteratorImpl) NextLeastConServer() *UpstreamServer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.servers) == 0 {
		return nil
	}

	// Pop the server with the least connections
	leastConnServer := heap.Pop(&l.servers).(*PriorityServer)
	defer heap.Push(&l.servers, leastConnServer) // Push it back after incrementing connections

	atomic.AddInt32(&leastConnServer.server.ActiveConnections, 1)
	return leastConnServer.server
}

func (iterator *IteratorImpl) MatchServer(clientIP string) *UpstreamServer {
	iterator.mu.Lock()
	defer iterator.mu.Unlock()

	if len(iterator.servers) == 0 {
		return nil
	}

	hasher := fnv.New32()
	hasher.Write([]byte(clientIP))
	index := int(hasher.Sum32()) % len(iterator.servers)

	// Try to find a healthy server starting from the hashed index
	for offset := 0; offset < len(iterator.servers); offset++ {
		currentIndex := (index + offset) % len(iterator.servers)
		server := iterator.servers[currentIndex].server
		if server.GetStatus() {
			return server
		}
	}
	return nil
}
