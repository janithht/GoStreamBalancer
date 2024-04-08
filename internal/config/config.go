package config

import (
	"time"
)

type Config struct {
	Upstreams []Upstream `yaml:"upstreams"`
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

type Upstream struct {
	Name        string      `yaml:"name"`
	Servers     []string    `yaml:"servers"`
	HealthCheck HealthCheck `yaml:"healthCheck"`
	RateLimit   RateLimit   `yaml:"rateLimit"`
}
