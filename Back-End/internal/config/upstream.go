package config

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	ActiveConnections int32
	LastCheck         time.Time
	LastSuccess       bool
	mu                sync.RWMutex
}

type PriorityServer struct {
	server *UpstreamServer
	index  int // index in the heap
}

type ServerHeap []*PriorityServer

func (server *UpstreamServer) SetStatus(status bool) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.Status = status
}

func (server *UpstreamServer) GetStatus() bool {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return server.Status
}

func (server *UpstreamServer) IncrementConnections() {
	atomic.AddInt32(&server.ActiveConnections, 1)
}

func (server *UpstreamServer) DecrementConnections() {
	atomic.AddInt32(&server.ActiveConnections, -1)
}

func (h ServerHeap) Len() int { return len(h) }
func (h ServerHeap) Less(i, j int) bool {
	return h[i].server.ActiveConnections < h[j].server.ActiveConnections
}
func (h ServerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *ServerHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*PriorityServer)
	item.index = n
	*h = append(*h, item)
}

func (h *ServerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
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
		upstreamMap[strings.ToLower(upstream.Name)] = iterator
		upstreamConfigMap[strings.ToLower(upstream.Name)] = upstream
		if upstream.RateLimit.Enabled {
			upstream.Limiter = ratelimits.NewRateLimiter(upstream.RateLimit.Limit, upstream.RateLimit.Interval)
		}
	}
	return upstreamMap, upstreamConfigMap
}

func CollectHealthData(upstreamConfigMap map[string]*Upstream) []UpstreamHealth {
	var upstreamsHealth []UpstreamHealth
	for _, upstream := range upstreamConfigMap {
		var serversHealth []ServerHealth
		for _, server := range upstream.Servers {
			server.mu.Lock()
			serversHealth = append(serversHealth, ServerHealth{
				URL:         server.Url,
				Status:      server.Status,
				LastCheck:   server.LastCheck,
				LastSuccess: server.LastSuccess,
			})
			server.mu.Unlock()
		}
		upstreamsHealth = append(upstreamsHealth, UpstreamHealth{
			Name:    upstream.Name,
			Servers: serversHealth,
		})
	}
	return upstreamsHealth
}

type UpstreamHealth struct {
	Name    string         `json:"name"`
	Servers []ServerHealth `json:"servers"`
}

type ServerHealth struct {
	URL         string    `json:"url"`
	Status      bool      `json:"status"`
	LastCheck   time.Time `json:"lastCheck"`
	LastSuccess bool      `json:"lastSuccess"`
}
