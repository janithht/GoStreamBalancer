package healthchecks

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

type HealthCheckTask struct {
	Server            *config.UpstreamServer
	HealthCheckConfig config.HealthCheck
}

func checkServerHealth(server string, healthCheckConfig config.HealthCheck) bool {
	client := http.Client{
		Timeout: healthCheckConfig.Timeout,
	}
	res, err := client.Get(server + healthCheckConfig.Url)
	return err == nil && res.StatusCode == 200
}

func worker(ctx context.Context, id int, tasks <-chan HealthCheckTask) {
	for {
		select {
		case task := <-tasks:
			healthStatus := checkServerHealth(task.Server.Url, task.HealthCheckConfig)
			task.Server.Status = healthStatus

			if !healthStatus {
				log.Printf("[Worker %d] Server %s failed health check, removing from pool\n", id, task.Server.Url)
			} else {
				log.Printf("[Worker %d] Server %s passed health check\n", id, task.Server.Url)
			}
		case <-ctx.Done():
			log.Printf("[Worker %d] Exiting due to context cancellation.\n", id)
			return
		}
	}
}

func PerformHealthChecks(ctx context.Context, cfg *config.Config) {
	const numWorkers = 10
	tasks := make(chan HealthCheckTask, 100)

	for i := 0; i < numWorkers; i++ {
		go worker(ctx, i, tasks)
	}

	for {
		for _, upstream := range cfg.Upstreams {
			for _, server := range upstream.Servers {
				task := HealthCheckTask{
					Server:            &server,
					HealthCheckConfig: upstream.HealthCheck,
				}
				tasks <- task
			}
			time.Sleep(upstream.HealthCheck.Interval)
		}
	}
}
