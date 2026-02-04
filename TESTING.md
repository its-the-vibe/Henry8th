# Testing Guide

## Unit Tests

Run the unit tests with:

```bash
go test -v
```

The unit tests validate:
- Configuration file loading
- Configuration validation (required fields)
- Error handling for missing files

## Manual Integration Testing

To manually test the service with a real Redis instance:

### 1. Start a Redis server

```bash
docker run -d --name redis-test -p 6379:6379 redis:latest
```

### 2. Populate test data

Connect to Redis and add test data:

```bash
redis-cli

# Add 150 items to a test list
RPUSH mylist1 "item1"
RPUSH mylist1 "item2"
# ... continue to item150

# Check list length
LLEN mylist1
# Should return 150
```

Or use a script to populate:

```bash
for i in {1..150}; do
  redis-cli RPUSH mylist1 "item$i"
done

redis-cli LLEN mylist1
# Should return 150
```

### 3. Configure the service

Edit `config.yaml`:

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

poll_interval: "10s"

lists:
  - name: "mylist1"
    max_size: 100
```

### 4. Run the service

```bash
./henry8th
```

Expected output:
```
2026/02/04 20:30:00 Successfully connected to Redis
2026/02/04 20:30:00 Starting list trimming service (poll interval: 10s)
2026/02/04 20:30:00 Trimmed list mylist1: removed 50 old items (was 150, now 100)
```

### 5. Verify the results

Check the list length:

```bash
redis-cli LLEN mylist1
# Should return 100
```

Verify the correct items were kept (newest 100):

```bash
redis-cli LRANGE mylist1 0 0
# Should return "item51" (oldest remaining item)

redis-cli LRANGE mylist1 -1 -1
# Should return "item150" (newest item)
```

### 6. Test continuous monitoring

Add more items:

```bash
for i in {151..200}; do
  redis-cli RPUSH mylist1 "item$i"
done

redis-cli LLEN mylist1
# Should return 150 (100 + 50 new items)
```

Wait for the next poll interval (10 seconds), then check:

```bash
redis-cli LLEN mylist1
# Should return 100 again
```

### 7. Clean up

```bash
docker stop redis-test
docker rm redis-test
```

## Docker Testing

### Build and test with Docker

```bash
# Build the image
docker build -t henry8th:latest .

# Start Redis
docker run -d --name redis-test -p 6379:6379 redis:latest

# Update config.yaml to use host.docker.internal (or your Redis host)
# Then run the service
docker run --rm \
  -v $(pwd)/config.yaml:/config.yaml:ro \
  --add-host=host.docker.internal:host-gateway \
  henry8th:latest
```

### Test with docker-compose

```bash
# Start Redis separately
docker run -d --name redis-test -p 6379:6379 redis:latest

# Update config.yaml to point to Redis
# Then start the service
docker-compose up
```

## Expected Behavior

1. **On startup**: Service connects to Redis and immediately trims all configured lists
2. **Periodic polling**: At each poll interval, the service checks and trims lists
3. **Logging**: Clear logs show when lists are trimmed and how many items were removed
4. **Graceful shutdown**: Service responds to SIGTERM/SIGINT signals
5. **Error handling**: Errors are logged but don't crash the service

## Performance Notes

- The service is lightweight and has minimal resource requirements
- Trimming operations are fast (O(n) where n is the number of items to remove)
- Poll intervals should be set based on your list growth rate
- For high-volume lists, consider shorter poll intervals
