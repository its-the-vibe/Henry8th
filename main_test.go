package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	configContent := `redis:
  host: testhost
  port: 6379
  password: testpass
  db: 1

poll_interval: "1m"

lists:
  - name: "test1"
    max_size: 100
  - name: "test2"
    max_size: 200
`

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading configuration
	config, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify Redis config
	if config.Redis.Host != "testhost" {
		t.Errorf("Expected host 'testhost', got '%s'", config.Redis.Host)
	}
	if config.Redis.Port != 6379 {
		t.Errorf("Expected port 6379, got %d", config.Redis.Port)
	}
	if config.Redis.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", config.Redis.Password)
	}
	if config.Redis.DB != 1 {
		t.Errorf("Expected DB 1, got %d", config.Redis.DB)
	}

	// Verify poll interval
	if config.PollInterval != "1m" {
		t.Errorf("Expected poll_interval '1m', got '%s'", config.PollInterval)
	}

	// Verify lists
	if len(config.Lists) != 2 {
		t.Fatalf("Expected 2 lists, got %d", len(config.Lists))
	}

	if config.Lists[0].Name != "test1" {
		t.Errorf("Expected first list name 'test1', got '%s'", config.Lists[0].Name)
	}
	if config.Lists[0].MaxSize != 100 {
		t.Errorf("Expected first list max_size 100, got %d", config.Lists[0].MaxSize)
	}

	if config.Lists[1].Name != "test2" {
		t.Errorf("Expected second list name 'test2', got '%s'", config.Lists[1].Name)
	}
	if config.Lists[1].MaxSize != 200 {
		t.Errorf("Expected second list max_size 200, got %d", config.Lists[1].MaxSize)
	}
}

func TestLoadConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing redis host",
			config: `redis:
  port: 6379
poll_interval: "1m"
lists:
  - name: "test"
    max_size: 100
`,
			expectError: true,
			errorMsg:    "redis host is required",
		},
		{
			name: "missing redis port",
			config: `redis:
  host: localhost
poll_interval: "1m"
lists:
  - name: "test"
    max_size: 100
`,
			expectError: true,
			errorMsg:    "redis port is required",
		},
		{
			name: "missing poll_interval",
			config: `redis:
  host: localhost
  port: 6379
lists:
  - name: "test"
    max_size: 100
`,
			expectError: true,
			errorMsg:    "poll_interval is required",
		},
		{
			name: "missing lists",
			config: `redis:
  host: localhost
  port: 6379
poll_interval: "1m"
lists: []
`,
			expectError: true,
			errorMsg:    "at least one list configuration is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "config-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.config)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			_, err = loadConfig(tmpfile.Name())
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.errorMsg)
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := loadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent config file, got nil")
	}
}
