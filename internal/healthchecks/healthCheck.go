package healthchecks

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

type HealthCheckTask struct {
	Server            string
	HealthCheckConfig config.HealthCheck
}

func checkServerHealth(server string, healthCheckConfig config.HealthCheck) bool {
	client := http.Client{
		Timeout: healthCheckConfig.Timeout,
	}
	res, err := client.Get(server + healthCheckConfig.Url)
	return err == nil && res.StatusCode == 200
}

func worker(id int, tasks <-chan HealthCheckTask) {
	for task := range tasks {
		if !checkServerHealth(task.Server, task.HealthCheckConfig) {
			log.Printf("[Worker %d] Server %s failed health check, removing from pool\n", id, task.Server)
		} else {
			log.Printf("[Worker %d] Server %s passed health check\n", id, task.Server)
		}
	}
}

func PerformHealthChecks(cfg *config.Config) {
	const numWorkers = 10
	tasks := make(chan HealthCheckTask, 100)

	for i := 0; i < numWorkers; i++ {
		go worker(i, tasks)
	}

	for {
		fmt.Println()
		for _, upstream := range cfg.Upstreams {
			for _, server := range upstream.Servers {
				task := HealthCheckTask{
					Server:            server,
					HealthCheckConfig: upstream.HealthCheck,
				}
				tasks <- task
			}
			time.Sleep(upstream.HealthCheck.Interval)
		}
	}
}
