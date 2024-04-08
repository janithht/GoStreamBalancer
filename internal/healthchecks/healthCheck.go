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
func performHealthCheckForUpstream(upstream config.Upstream) {
	for {
		fmt.Println()
		for i := len(upstream.Servers) - 1; i >= 0; i-- {
			server := upstream.Servers[i]
			if !checkServerHealth(server, upstream.HealthCheck) {
				log.Printf("Server %s failed health check, removing from pool\n", server)
				// Properly synchronize server removal here.
			}
		}
		time.Sleep(upstream.HealthCheck.Interval)
	}
}

func PerformHealthChecks(config *config.Config) {
	for _, upstream := range config.Upstreams {
		go performHealthCheckForUpstream(upstream)
	}

	select {}
}
