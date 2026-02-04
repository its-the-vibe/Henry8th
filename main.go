package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure
type Config struct {
	Redis        RedisConfig  `yaml:"redis"`
	Lists        []ListConfig `yaml:"lists"`
	PollInterval string       `yaml:"poll_interval"`
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// ListConfig represents a single list to trim
type ListConfig struct {
	Name    string `yaml:"name"`
	MaxSize int64  `yaml:"max_size"`
}

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override Redis password with environment variable if present
	if envPassword := os.Getenv("REDIS_PASSWORD"); envPassword != "" {
		config.Redis.Password = envPassword
	}

	// Parse poll interval
	pollInterval, err := time.ParseDuration(config.PollInterval)
	if err != nil {
		log.Fatalf("Failed to parse poll_interval: %v", err)
	}

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	defer rdb.Close()

	ctx := context.Background()

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis")

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Printf("Starting list trimming service (poll interval: %s)", pollInterval)

	// Initial trim on startup
	trimLists(ctx, rdb, config.Lists)

	// Main loop
	for {
		select {
		case <-ticker.C:
			trimLists(ctx, rdb, config.Lists)
		case <-shutdown:
			log.Println("Shutting down gracefully...")
			return
		}
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if config.Redis.Host == "" {
		return nil, fmt.Errorf("redis host is required")
	}
	if config.Redis.Port == 0 {
		return nil, fmt.Errorf("redis port is required")
	}
	if config.PollInterval == "" {
		return nil, fmt.Errorf("poll_interval is required")
	}
	if len(config.Lists) == 0 {
		return nil, fmt.Errorf("at least one list configuration is required")
	}

	return &config, nil
}

func trimLists(ctx context.Context, rdb *redis.Client, lists []ListConfig) {
	for _, listConfig := range lists {
		if err := trimList(ctx, rdb, listConfig); err != nil {
			log.Printf("Error trimming list %s: %v", listConfig.Name, err)
		}
	}
}

func trimList(ctx context.Context, rdb *redis.Client, listConfig ListConfig) error {
	// LTRIM to keep only the most recent items (tail of the list)
	// Since RPUSH adds to the tail, we want to keep [-(maxSize), -1]
	// This removes old items from the head
	// LTRIM is a no-op if the list is smaller than max_size
	start := -listConfig.MaxSize
	end := int64(-1)

	if err := rdb.LTrim(ctx, listConfig.Name, start, end).Err(); err != nil {
		return fmt.Errorf("failed to trim list: %w", err)
	}

	log.Printf("Applied LTRIM to list %s (max size %d)", listConfig.Name, listConfig.MaxSize)

	return nil
}
