package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromFlags(t *testing.T) {
	// Save original args
	originalArgs := os.Args

	// Test basic flag parsing
	os.Args = []string{"yamcha", "-url", "http://example.com", "-req", "50", "-rate", "15"}

	cfg, err := LoadFromFlags()
	if err != nil {
		t.Fatalf("Failed to load config from flags: %v", err)
	}

	if cfg.Target.URL != "http://example.com" {
		t.Errorf("Expected URL to be http://example.com, got %s", cfg.Target.URL)
	}

	if cfg.Load.Requests != 50 {
		t.Errorf("Expected requests to be 50, got %d", cfg.Load.Requests)
	}

	if cfg.Load.Rate != 15 {
		t.Errorf("Expected rate to be 15, got %d", cfg.Load.Rate)
	}

	// Restore original args
	os.Args = originalArgs
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}
	err := cfg.SetDefaults()
	if err != nil {
		t.Fatalf("Failed to set defaults: %v", err)
	}

	if cfg.Target.Method != "GET" {
		t.Errorf("Expected default method to be GET, got %s", cfg.Target.Method)
	}

	if cfg.Load.AttackType != "steady" {
		t.Errorf("Expected default attack type to be steady, got %s", cfg.Load.AttackType)
	}

	if cfg.Load.Requests != 100 {
		t.Errorf("Expected default requests to be 100, got %d", cfg.Load.Requests)
	}

	if cfg.Load.Rate != 20 {
		t.Errorf("Expected default rate to be 20, got %d", cfg.Load.Rate)
	}

	if time.Duration(cfg.HTTP.Timeout) != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", cfg.HTTP.Timeout)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Target: TargetConfig{URL: "http://example.com"},
				Load:   LoadConfig{Requests: 10, Rate: 5, AttackType: "steady"},
			},
			expectErr: false,
		},
		{
			name: "empty URL",
			config: Config{
				Load: LoadConfig{Requests: 10, Rate: 5, AttackType: "steady"},
			},
			expectErr: true,
		},
		{
			name: "invalid rate",
			config: Config{
				Target: TargetConfig{URL: "http://example.com"},
				Load:   LoadConfig{Requests: 10, Rate: 0, AttackType: "steady"},
			},
			expectErr: true,
		},
		{
			name: "invalid attack type",
			config: Config{
				Target: TargetConfig{URL: "http://example.com"},
				Load:   LoadConfig{Requests: 10, Rate: 5, AttackType: "invalid"},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got %v", err)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	yamlContent := `
target:
  url: "http://test.com"
  method: "POST"
load:
  attack_type: "burst"
  requests: 25
  rate: 10
`

	tmpFile, err := os.CreateTemp("", "yamcha-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if cfg.Target.URL != "http://test.com" {
		t.Errorf("Expected URL to be http://test.com, got %s", cfg.Target.URL)
	}

	if cfg.Target.Method != "POST" {
		t.Errorf("Expected method to be POST, got %s", cfg.Target.Method)
	}

	if cfg.Load.AttackType != "burst" {
		t.Errorf("Expected attack type to be burst, got %s", cfg.Load.AttackType)
	}
}
