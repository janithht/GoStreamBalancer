package healthchecks

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

func checkServerHealth(server string, healthCheckConfig config.HealthCheck) bool {
	client := http.Client{
		Timeout: healthCheckConfig.Timeout,
	}
	res, err := client.Get(server + healthCheckConfig.Url)
	return err == nil && res.StatusCode == 200
}

func PerformHealthChecks(config *config.Config) {
	for {
		fmt.Println()
		for _, upstream := range config.Upstreams {
			for i := len(upstream.Servers) - 1; i >= 0; i-- {
				server := upstream.Servers[i]
				if !checkServerHealth(server, upstream.HealthCheck) {
					log.Printf("Server %s failed health check, removing from pool\n", server)
					// Remove server from slice safely.
					upstream.Servers = append(upstream.Servers[:i], upstream.Servers[i+1:]...)
				}
			}
		}
		time.Sleep(config.Upstreams[0].HealthCheck.Interval)
	}
}
