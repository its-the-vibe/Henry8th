# Build stage
FROM golang:1.26.1-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o henry8th .

# Runtime stage using scratch
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/henry8th /henry8th

# Run the service
ENTRYPOINT ["/henry8th"]
