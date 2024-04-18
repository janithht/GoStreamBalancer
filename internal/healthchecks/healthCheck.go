package healthchecks

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

type HealthCheckTask struct {
	Server            *config.UpstreamServer
	HealthCheckConfig config.HealthCheck
}

var (
	mutex     = &sync.Mutex{}
	taskQueue []HealthCheckTask
	taskChan  chan HealthCheckTask
)

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
	taskQueue = make([]HealthCheckTask, 0)
	taskChan = make(chan HealthCheckTask, 100)

	for i := 0; i < numWorkers; i++ {
		go worker(ctx, i, taskChan)
	}

	go manageQueue()

	lastScheduled := make(map[string]time.Time)
	for {
		for _, upstream := range cfg.Upstreams {
			now := time.Now()
			if lastTime, ok := lastScheduled[upstream.Name]; !ok || now.Sub(lastTime) >= upstream.HealthCheck.Interval {
				for _, server := range upstream.Servers {
					task := HealthCheckTask{
						Server:            &server,
						HealthCheckConfig: upstream.HealthCheck,
					}
					mutex.Lock()
					taskQueue = append(taskQueue, task)
					mutex.Unlock()
				}
				lastScheduled[upstream.Name] = now
			}
		}
		time.Sleep(time.Second) // Check every second if it's time to schedule next checks
	}
}

func manageQueue() {
	for {
		mutex.Lock()
		if len(taskQueue) > 0 {
			task := taskQueue[0]
			taskQueue = taskQueue[1:]
			mutex.Unlock()
			taskChan <- task
		} else {
			mutex.Unlock()
			time.Sleep(1 * time.Second)
		}
	}
}
