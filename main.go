package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Upstreams []Upstream `yaml:"upstreams"`
}

type Upstream struct {
	Name        string   `yaml:"name"`
	Servers     []string `yaml:"servers"`
	HealthCheck `yaml:"healthCheck"`
	RateLimit   `yaml:"rateLimit"`
}

type HealthCheck struct {
	Url      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

type RateLimit struct {
	Limit    int           `yaml:"limit"`
	Interval time.Duration `yaml:"interval"`
}

func main() {
	var config Config

	data, err := os.ReadFile("config.yaml") // Read the file
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(data, &config) // Unmarshal the data to strcuts defined above
	if err != nil {
		log.Fatalf("error parsing config file: %v", err)
	}

	fmt.Println("Config parsed successfully:", config)

	go performHealthChecks(&config)
	startServer(&config)
}

var currentServerIndex int = 0

func startServer(config *Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstream := config.Upstreams[0]

		target := upstream.Servers[currentServerIndex]
		currentServerIndex = (currentServerIndex + 1) % len(upstream.Servers)

		url, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	})

	fmt.Printf("Load Balancer started on port 3000\n")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func checkServerHealth(server string, healthCheckConfig HealthCheck) bool {
	client := http.Client{
		Timeout: healthCheckConfig.Timeout,
	}
	res, err := client.Get(server + healthCheckConfig.Url)
	return err == nil && res.StatusCode == 200
}

func performHealthChecks(config *Config) {
	for {
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
		time.Sleep(10 * time.Second)
	}
}
