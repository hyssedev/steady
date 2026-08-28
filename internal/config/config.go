package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/hyssedev/steady/internal/monitor"
	"go.yaml.in/yaml/v4"
)

type rawConfig struct {
	Listen   string `yaml:"listen"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`

	Monitors []monitor.Monitor `yaml:"monitors"`
}

type Config struct {
	Listen   string
	Interval time.Duration
	Timeout  time.Duration

	Monitors []monitor.Monitor
}

func ReadConfig(path string) (Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	return parse(bytes)
}

func parse(data []byte) (Config, error) {
	var raw rawConfig

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if len(raw.Listen) == 0 {
		return Config{}, fmt.Errorf("empty listen")
	}

	interval, err := time.ParseDuration(raw.Interval)
	if err != nil {
		return Config{}, fmt.Errorf("invalid interval %q: %w", raw.Interval, err)
	}

	if interval <= 0 {
		return Config{}, fmt.Errorf("interval must be greater than zero")
	}

	timeout, err := time.ParseDuration(raw.Timeout)
	if err != nil {
		return Config{}, fmt.Errorf("invalid timeout %q: %w", raw.Timeout, err)
	}

	if timeout <= 0 {
		return Config{}, fmt.Errorf("timeout must be greater than zero")
	}

	if len(raw.Monitors) == 0 {
		return Config{}, fmt.Errorf("no monitors present")
	}

	for _, m := range raw.Monitors {
		if _, err := url.ParseRequestURI(m.URL); err != nil {
			return Config{}, fmt.Errorf("monitor %s has invalid URL: %w", m.Name, err)
		}
	}

	return Config{
		Listen:   raw.Listen,
		Interval: interval,
		Timeout:  timeout,
		Monitors: raw.Monitors,
	}, nil
}
