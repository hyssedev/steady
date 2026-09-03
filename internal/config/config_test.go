package config

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hyssedev/steady/internal/monitor"
)

const validConfig = `
listen: ":8080"
interval: 1m
timeout: 5s
monitors:
  - name: Main
    url: https://example.com
`

func TestParse(t *testing.T) {
	want := Config{
		Listen:   ":8080",
		Interval: time.Minute,
		Timeout:  5 * time.Second,
		Monitors: []monitor.Monitor{{Name: "Main", URL: &url.URL{Scheme: "https", Host: "example.com"}}},
	}
	tests := []struct {
		name    string
		data    string
		want    Config
		wantErr bool
	}{
		{
			name: "valid",
			data: validConfig,
			want: want,
		},
		{
			name:    "malformed yaml",
			data:    "listen: [",
			wantErr: true,
		},
		{
			name:    "empty listen",
			data:    strings.Replace(validConfig, "listen: \":8080\"", "listen: \"\"", 1),
			wantErr: true,
		},
		{
			name:    "invalid interval",
			data:    strings.Replace(validConfig, "interval: 1m", "interval: never", 1),
			wantErr: true,
		},
		{
			name:    "zero interval",
			data:    strings.Replace(validConfig, "interval: 1m", "interval: 0s", 1),
			wantErr: true,
		},
		{
			name:    "invalid timeout",
			data:    strings.Replace(validConfig, "timeout: 5s", "timeout: never", 1),
			wantErr: true,
		},
		{
			name:    "zero timeout",
			data:    strings.Replace(validConfig, "timeout: 5s", "timeout: 0s", 1),
			wantErr: true,
		},
		{
			name:    "no monitors",
			data:    strings.Replace(validConfig, "monitors:\n  - name: Main\n    url: https://example.com", "monitors: []", 1),
			wantErr: true,
		},
		{
			name:    "invalid monitor URL",
			data:    strings.Replace(validConfig, "https://example.com", "://invalid", 1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestReadConfig(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "valid.yaml")
	invalidPath := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(validPath, []byte(validConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(invalidPath, []byte("listen: ["), 0o600); err != nil {
		t.Fatal(err)
	}

	want := Config{
		Listen:   ":8080",
		Interval: time.Minute,
		Timeout:  5 * time.Second,
		Monitors: []monitor.Monitor{{Name: "Main", URL: &url.URL{Scheme: "https", Host: "example.com"}}},
	}
	tests := []struct {
		name    string
		path    string
		want    Config
		wantErr bool
	}{
		{
			name: "valid",
			path: validPath,
			want: want,
		},
		{
			name:    "missing file",
			path:    filepath.Join(dir, "missing.yaml"),
			wantErr: true,
		},
		{
			name:    "invalid config",
			path:    invalidPath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadConfig(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
