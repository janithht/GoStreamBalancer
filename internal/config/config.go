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
	Enabled  bool          `yaml:"enabled"`
}

type RateLimit struct {
	Limit    int           `yaml:"limit"`
	Interval time.Duration `yaml:"interval"`
}
