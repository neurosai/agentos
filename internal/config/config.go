package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the agentosd runtime configuration.
type Config struct {
	Listen   string         `yaml:"listen"`
	Database DatabaseConfig `yaml:"database"`
	OPA      OPAConfig      `yaml:"opa"`
	Tools    ToolsConfig    `yaml:"tools"`
	Dev      DevConfig      `yaml:"dev"`
}

// DatabaseConfig holds PostgreSQL settings.
type DatabaseConfig struct {
	URL         string `yaml:"url"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

// OPAConfig holds Open Policy Agent settings.
type OPAConfig struct {
	URL string `yaml:"url"`
}

// ToolsConfig holds tool registry settings.
type ToolsConfig struct {
	Registry string `yaml:"registry"`
}

// DevConfig holds development auth defaults.
type DevConfig struct {
	TenantID     string   `yaml:"tenant_id"`
	SubjectID    string   `yaml:"subject_id"`
	SubjectRoles []string `yaml:"subject_roles"`
	BearerToken  string   `yaml:"bearer_token"`
}

// Load reads configuration from a YAML file.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.Dev.TenantID == "" {
		cfg.Dev.TenantID = "tenant:dev"
	}
	if cfg.Dev.SubjectID == "" {
		cfg.Dev.SubjectID = "user:dev"
	}
	if len(cfg.Dev.SubjectRoles) == 0 {
		cfg.Dev.SubjectRoles = []string{"engineer"}
	}
	if cfg.Dev.BearerToken == "" {
		cfg.Dev.BearerToken = "dev-token"
	}
	return cfg, nil
}
