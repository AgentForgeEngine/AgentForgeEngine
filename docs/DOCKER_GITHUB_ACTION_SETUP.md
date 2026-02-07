# Docker & GitHub Action Setup

This guide covers containerizing AgentForgeEngine and setting up CI/CD pipelines using Docker and GitHub Actions.

## Table of Contents

- [Docker Setup](#docker-setup)
  - [Dockerfile](#dockerfile)
  - [Docker Compose](#docker-compose)
  - [Building the Image](#building-the-image)
  - [Running the Container](#running-the-container)
- [GitHub Actions CI/CD](#github-actions-cicd)
  - [Build Pipeline](#build-pipeline)
  - [Release Pipeline](#release-pipeline)
  - [Security Scanning](#security-scanning)
- [Deployment Strategies](#deployment-strategies)
  - [Single Node Deployment](#single-node-deployment)
  - [Multi-Stage Deployment](#multi-stage-deployment)
- [Configuration Management](#configuration-management)
  - [Environment Variables](#environment-variables)
  - [Config Maps and Secrets](#config-maps-and-secrets)
- [Monitoring and Logging](#monitoring-and-logging)
  - [Health Checks](#health-checks)
  - [Log Aggregation](#log-aggregation)
  - [Metrics Collection](#metrics-collection)

## Docker Setup

### Dockerfile

Create a multi-stage Dockerfile for efficient builds:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o afe ./cmd/agentforge

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1001 -S afe && \
    adduser -u 1001 -S afe -G afe

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/afe .

# Copy configuration template
COPY configs/agentforge.yaml.template ./configs/

# Create necessary directories
RUN mkdir -p /app/logs /app/cache /app/accounts && \
    chown -R afe:afe /app

# Switch to non-root user
USER afe

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./afe status || exit 1

# Default command
CMD ["./afe", "start", "--config", "./configs/agentforge.yaml"]
```

### Docker Compose

Create a `docker-compose.yml` for local development:

```yaml
version: '3.8'

services:
  afe:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: afe
    ports:
      - "8080:8080"
    environment:
      - AFE_LOG_LEVEL=info
      - AFE_HOST=0.0.0.0
      - AFE_PORT=8080
    volumes:
      - afe_data:/app/data
      - afe_logs:/app/logs
      - afe_cache:/app/cache
      - ./configs:/app/configs:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./afe", "status"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Optional: Add a reverse proxy
  nginx:
    image: nginx:alpine
    container_name: afe-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - afe
    restart: unless-stopped

volumes:
  afe_data:
    driver: local
  afe_logs:
    driver: local
  afe_cache:
    driver: local
```

### Building the Image

Build the Docker image:

```bash
# Build for current platform
docker build -t agentforge-engine:latest .

# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 -t agentforge-engine:latest --push .
```

### Running the Container

Run the container with various options:

```bash
# Basic run
docker run -d \
  --name afe \
  -p 8080:8080 \
  -v afe_data:/app/data \
  agentforge-engine:latest

# Run with environment variables
docker run -d \
  --name afe \
  -p 8080:8080 \
  -e AFE_LOG_LEVEL=debug \
  -e AFE_HOST=0.0.0.0 \
  -v $(pwd)/configs:/app/configs:ro \
  agentforge-engine:latest

# Run with Docker Compose
docker-compose up -d

# Check container logs
docker logs -f afe

# Execute commands in container
docker exec -it afe ./afe status
docker exec -it afe ./afe build all
```

## GitHub Actions CI/CD

### Build Pipeline

Create `.github/workflows/build.yml`:

```yaml
name: Build and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.24'
  REGISTRY: ghcr.io
  IMAGE_NAME: agentforge-engine

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
          ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella

    - name: Run integration tests
      run: ./scripts/test_agents.sh integration

  build:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push'
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ github.repository }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
```

### Release Pipeline

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

env:
  GO_VERSION: '1.24'
  REGISTRY: ghcr.io
  IMAGE_NAME: agentforge-engine

jobs:
  create-release:
    name: Create Release
    runs-on: ubuntu-latest
    outputs:
      upload_url: ${{ steps.create_release.outputs.upload_url }}
    steps:
    - name: Create Release
      id: create_release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref }}
        release_name: Release ${{ github.ref }}
        draft: false
        prerelease: false

  build-and-upload:
    name: Build and Upload
    runs-on: ubuntu-latest
    needs: create-release
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Build binary
      env:
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
      run: |
        binary_name="afe-${{ matrix.goos }}-${{ matrix.goarch }}"
        if [ "${{ matrix.goos }}" = "windows" ]; then
          binary_name="${binary_name}.exe"
        fi
        CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} \
          go build -ldflags="-s -w" -o ${binary_name} ./cmd/agentforge

    - name: Create archive
      run: |
        binary_name="afe-${{ matrix.goos }}-${{ matrix.goarch }}"
        if [ "${{ matrix.goos }}" = "windows" ]; then
          binary_name="${binary_name}.exe"
        fi
        tar -czf "${binary_name}.tar.gz" "${binary_name}"
        shasum -a 256 "${binary_name}.tar.gz" > "${binary_name}.tar.gz.sha256"

    - name: Upload Release Asset
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ needs.create-release.outputs.upload_url }}
        asset_path: ./afe-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz
        asset_name: afe-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz
        asset_content_type: application/gzip

    - name: Upload Checksum
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ needs.create-release.outputs.upload_url }}
        asset_path: ./afe-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz.sha256
        asset_name: afe-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz.sha256
        asset_content_type: text/plain

  docker-release:
    name: Docker Release
    runs-on: ubuntu-latest
    needs: create-release
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ github.repository }}
        tags: |
          type=ref,event=tag
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
```

### Security Scanning

Create `.github/workflows/security.yml`:

```yaml
name: Security

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Monday at 2 AM

jobs:
  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-no-fail -fmt sarif -out results.sarif ./...'

    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: results.sarif

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results to GitHub Security tab
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'
```

## Deployment Strategies

### Single Node Deployment

Simple single-node deployment using Docker Compose:

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  afe:
    image: agentforge-engine:latest
    container_name: afe-prod
    ports:
      - "8080:8080"
    environment:
      - AFE_LOG_LEVEL=info
      - AFE_HOST=0.0.0.0
      - AFE_PORT=8080
    volumes:
      - afe_prod_data:/app/data
      - afe_prod_logs:/app/logs
      - afe_prod_cache:/app/cache
      - ./production-configs:/app/configs:ro
    restart: always
    healthcheck:
      test: ["CMD", "./afe", "status"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  afe_prod_data:
    driver: local
  afe_prod_logs:
    driver: local
  afe_prod_cache:
    driver: local
```

### Multi-Stage Deployment

Deployment with build, staging, and production environments:

```yaml
# docker-compose.staging.yml
version: '3.8'

services:
  afe:
    image: agentforge-engine:${STAGE_VERSION}
    container_name: afe-staging
    ports:
      - "8081:8080"
    environment:
      - AFE_LOG_LEVEL=debug
      - AFE_HOST=0.0.0.0
      - AFE_PORT=8080
      - AFE_ENV=staging
    volumes:
      - afe_staging_data:/app/data
      - afe_staging_logs:/app/logs
      - afe_staging_cache:/app/cache
      - ./staging-configs:/app/configs:ro
    restart: unless-stopped

volumes:
  afe_staging_data:
  afe_staging_logs:
  afe_staging_cache:
```

## Configuration Management

### Environment Variables

Key environment variables for Docker deployments:

```bash
# Basic configuration
AFE_LOG_LEVEL=info
AFE_HOST=0.0.0.0
AFE_PORT=8080
AFE_ENV=production

# Security
AFE_TLS_ENABLED=true
AFE_TLS_CERT_PATH=/app/configs/tls.crt
AFE_TLS_KEY_PATH=/app/configs/tls.key

# Performance
AFE_MAX_WORKERS=10
AFE_CACHE_SIZE=100
AFE_TIMEOUT=30s

# Monitoring
AFE_METRICS_ENABLED=true
AFE_METRICS_PORT=9090
```

### Config Maps and Secrets

For Kubernetes deployments:

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: afe-config
data:
  agentforge.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
    models:
      - name: "qwen3"
        type: "http"
        endpoint: "http://qwen3:8000"
    agents:
      local:
        - name: "ls"
          path: "./agents/ls"
          config:
            timeout: 30s

---
apiVersion: v1
kind: Secret
metadata:
  name: afe-secrets
type: Opaque
data:
  api-key: <base64-encoded-api-key>
  db-password: <base64-encoded-password>
```

## Monitoring and Logging

### Health Checks

Implement comprehensive health checks:

```yaml
# In Dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./afe status --verbose || exit 1

# In docker-compose.yml
healthcheck:
  test: ["CMD", "./afe", "status"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### Log Aggregation

Configure structured logging for container environments:

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  afe:
    image: agentforge-engine:latest
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service,environment"

  # Optional: Logstash for log aggregation
  logstash:
    image: docker.elastic.co/logstash/logstash:8.5.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf:ro
    ports:
      - "5044:5044"
```

### Metrics Collection

Add metrics collection using Prometheus:

```yaml
# monitoring.yml
version: '3.8'

services:
  afe:
    image: agentforge-engine:latest
    environment:
      - AFE_METRICS_ENABLED=true
      - AFE_METRICS_PORT=9090
    ports:
      - "9090:9090"  # Metrics endpoint

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'agentforge'
    static_configs:
      - targets: ['afe:9090']
    metrics_path: /metrics
    scrape_interval: 5s

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx:9113']
```

This setup provides a complete containerization and CI/CD solution for AgentForgeEngine without Redis dependencies, using the correct container name `afe` and focusing on practical deployment scenarios.