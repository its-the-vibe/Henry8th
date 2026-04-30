# Henry8th

<img src="https://github.com/its-the-vibe/Henry8th/actions/workflows/ci.yaml/badge.svg">

A simple service in Go that will trim the head of Redis Lists to maintain a configured maximum size.

## Overview

Henry8th is a lightweight Go service that automatically manages Redis list sizes by periodically trimming them to configured maximum lengths. Lists are populated using `RPUSH` (adding to the tail), and this service removes the oldest items from the head to maintain the specified size limits.

## Features

- 🚀 Built with Go 1.24
- 📦 Minimal Docker image using `scratch` base
- ⚙️ YAML-based configuration
- 🔄 Periodic polling and trimming
- 🛡️ Graceful shutdown handling
- 📊 Detailed logging

## Configuration

Create a `config.yaml` file with the following structure:

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

# Poll interval for checking and trimming lists
poll_interval: "30s"

# Lists to trim - each list will be trimmed to its max_size
# Lists are populated with RPUSH, so trimming removes from the head (oldest items)
lists:
  - name: "mylist1"
    max_size: 1000
  - name: "mylist2"
    max_size: 500
  - name: "events"
    max_size: 10000
```

### Configuration Options

- **redis**: Redis connection settings
  - `host`: Redis server hostname
  - `port`: Redis server port
  - `password`: Redis password (leave empty if none)
  - `db`: Redis database number (default: 0)
- **poll_interval**: How often to check and trim lists (e.g., "30s", "1m", "5m")
- **lists**: Array of lists to manage
  - `name`: Redis list key name
  - `max_size`: Maximum number of items to keep in the list

## Running with Docker

### Build the Docker image

```bash
docker build -t henry8th:latest .
```

### Run with Docker Compose

1. Edit `config.yaml` to point to your external Redis instance
2. Run the service:

```bash
docker-compose up -d
```

### Run directly with Docker

```bash
docker run -v $(pwd)/config.yaml:/config.yaml:ro henry8th:latest
```

## Makefile Targets

The project includes a `Makefile` to standardize common development tasks:

| Target | Description |
|--------|-------------|
| `make build` | Compile the Go project |
| `make test` | Run unit tests with coverage |
| `make lint` | Run `golangci-lint` |
| `make ci` | Run build, test, and lint (used by CI) |

## Running Locally

### Prerequisites

- Go 1.24 or later
- Access to a Redis server

### Build

```bash
go build -o henry8th .
```

### Run

```bash
./henry8th
```

By default, the service looks for `config.yaml` in the current directory. You can specify a different path using the `CONFIG_PATH` environment variable:

```bash
CONFIG_PATH=/path/to/config.yaml ./henry8th
```

## How It Works

1. **Startup**: The service reads the configuration file and connects to Redis
2. **Initial Trim**: Immediately trims all configured lists on startup
3. **Periodic Polling**: At each poll interval, the service:
   - Checks the length of each configured list
   - If a list exceeds its `max_size`, it uses `LTRIM` to keep only the most recent items
   - Removes old items from the head of the list
4. **Graceful Shutdown**: Responds to SIGTERM and SIGINT signals

### Trimming Logic

Since lists are populated with `RPUSH` (which adds items to the tail), the oldest items are at the head (index 0). The service uses Redis `LTRIM` with negative indices to keep only the most recent items:

```
LTRIM key -max_size -1
```

This keeps the last `max_size` items and removes everything before them.

## Example

Given a list populated like this:

```bash
RPUSH mylist "item1"
RPUSH mylist "item2"
RPUSH mylist "item3"
# ... continues to item100
```

With `max_size: 10`, the service will keep only items 91-100 (the most recent), removing items 1-90 (the oldest).

## Environment Variables

- `CONFIG_PATH`: Path to configuration file (default: `config.yaml`)
- `REDIS_PASSWORD`: Redis password (overrides the password in config file if set)

## Logging

The service provides detailed logging:
- Connection status
- Trim operations (number of items removed)
- Lists that don't need trimming
- Errors and warnings

## License

MIT
