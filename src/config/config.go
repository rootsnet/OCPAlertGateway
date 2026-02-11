package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the full application configuration loaded from YAML.
type Config struct {
	// Server configuration section
	Server struct {
		// ListenAddr defines the HTTP listen address (e.g. ":8080")
		ListenAddr string `yaml:"listen_addr"`
		// WebhookPath defines the HTTP path for Alertmanager webhook (e.g. "/webhook")
		WebhookPath string `yaml:"webhook_path"`
	} `yaml:"server"`

	// SynologyChat configuration section
	SynologyChat struct {
		// Enabled determines whether Synology Chat integration is active
		Enabled bool `yaml:"enabled"`
		// WebhookURL is the full Incoming Webhook URL provided by Synology Chat
		WebhookURL string `yaml:"webhook_url"`
		// InsecureSkipVerify controls TLS certificate verification (for internal/self-signed environments)
		InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
	} `yaml:"synology_chat"`

	// Debug enables verbose logging including request/response dumps
	Debug bool `yaml:"debug"`
}

// Load reads the YAML configuration file from the given path,
// unmarshals it into Config struct, and applies default values.
func Load(path string) (*Config, error) {
	// Read configuration file
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal YAML into Config struct
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	// Apply default values if not specified

	// Default HTTP listen address
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}

	// Default webhook path
	if cfg.Server.WebhookPath == "" {
		cfg.Server.WebhookPath = "/webhook"
	}

	return &cfg, nil
}
