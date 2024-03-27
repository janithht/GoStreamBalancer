package main

import (
	"fmt"
	"log"
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
}
