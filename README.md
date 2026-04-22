# Media Microservice

A lightweight Go microservice that hosts a **MinIO** container as a local S3-compatible object storage service, exposing 2 endpoints for file upload and download. Built on a clean architecture blueprint with structured logging, rate limiting, circuit breaker pattern, and HTTP request handling via `chi` router.

## Architecture

The project follows a layered architecture:

```
cmd/api/main.go          → Entry point, wires dependencies
internal/handler         → HTTP handlers, routing, and middleware
internal/service         → Business logic layer
internal/repo            → Data access layer (S3 operations via MinIO SDK)
internal/logging         → Structured logging setup
internal/models          → Domain models
internal/helper          → Utility functions
internal/config          → YAML configuration loader
```

## Tech Stack

- **Router**: [go-chi/chi/v5](https://github.com/go-chi/chi)
- **Logging**: [go-chi/httplog/v3](https://github.com/go-chi/httplog) + `log/slog`
- **Object Storage**: [minio/minio-go/v7](https://github.com/minio/minio-go) (S3-compatible SDK)
- **Container**: [minio/minio](https://hub.docker.com/r/minio/minio) (local S3 via Docker)
- **Validation**: [go-playground/validator/v10](https://github.com/go-playground/validator)

## Features

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/upload` | Upload a file to the MinIO S3 bucket. Accepts `multipart/form-data` with a `file` field. Returns the object key and bucket name. |
| `GET` | `/download/{bucket}/{key}` | Download a file from the MinIO S3 bucket by bucket and object key. Returns the file stream with appropriate `Content-Type` and `Content-Disposition`. |

### Security
- **Security Headers**: X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, HSTS, CSP
- **Request ID**: Unique request tracking for debugging and logging

### Resilience
- **Rate Limiting**: Token bucket algorithm with configurable requests/second and burst
- **Circuit Breaker**: Automatic failure detection with half-open state for recovery
- **Graceful Shutdown**: 30-second timeout for in-flight requests

### Configuration
- **YAML Config**: All settings loaded from `config.yaml` with environment variable expansion
- **No hardcoded values**: Server port, timeouts, rate limits, MinIO credentials all configurable

## Prerequisites

- Go 1.25.7+
- Docker & Docker Compose

## Getting Started

### 1. Set Environment Variables

```bash
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="minioadmin"
export MINIO_SECRET_KEY="minioadmin"
export MINIO_BUCKET="media"
export MINIO_USE_SSL="false"
```

### 2. Start MinIO with Docker Compose

```bash
docker-compose up -d
```

This spins up a MinIO container with:
- **API port**: `9000` (S3-compatible endpoint)
- **Console port**: `9001` (MinIO Web UI)

### 3. Configure the Service

Edit `config.yaml` to customize:
- Server host/port and timeouts
- Logging level and format
- Rate limiting parameters
- Circuit breaker settings
- Health check paths

### 4. Run the Service

```bash
go run app/cmd/api/main.go
```

The server starts on `localhost:8080` (or configured port).

### 5. Test the Health Endpoint

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status": "ok"}
```

### 6. Upload a File

```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/image.png"
```

Response:
```json
{
  "bucket": "media",
  "key": "image.png"
}
```

### 7. Download a File

```bash
curl -O http://localhost:8080/download/media/image.png
```

## Docker

### Build the Image

```bash
docker build -t media-microservice .
```

### Run with Docker

```bash
docker run -p 8080:8080 \
  -e MINIO_ENDPOINT="minio:9000" \
  -e MINIO_ACCESS_KEY="minioadmin" \
  -e MINIO_SECRET_KEY="minioadmin" \
  -e MINIO_BUCKET="media" \
  -e MINIO_USE_SSL="false" \
  media-microservice
```

### Docker Compose

Use `docker-compose.yml` to spin up MinIO as a local S3 container:

```bash
docker-compose up -d
```

## Project Structure

| Path | Description |
|------|-------------|
| `app/cmd/api/main.go` | Application entry point. Wires together MinIO client, repository, service, and handler layers, then starts the HTTP server. |
| `app/internal/config/` | YAML configuration loader with environment variable expansion. |
| `app/internal/handler/` | HTTP handlers, request routing (`chi`), and middleware (security, rate limiting). |
| `app/internal/service/` | Business logic layer. Defines service interfaces and implements use cases. |
| `app/internal/repo/` | Data access layer. Handles all S3 operations via the MinIO Go SDK. |
| `app/internal/logging/` | Structured JSON logger configuration using `slog` and `httplog`. |
| `app/internal/models/` | Domain models and data structures shared across layers. |
| `app/internal/helper/` | Utility/helper functions including comprehensive error definitions. |
| `config.yaml` | Service configuration file. |
| `Dockerfile` | Multi-stage Docker build with healthcheck. |

## API Endpoints

### Public Routes (No Authentication)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check endpoint. Returns service health status. |
| `GET` | `/ready` | Readiness check. Verifies MinIO connectivity. |

### Media Routes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/upload` | Upload a file to the MinIO S3 bucket. Accepts `multipart/form-data`. |
| `GET` | `/download/{bucket}/{key}` | Download a file from the MinIO S3 bucket. |

## Configuration Reference

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

logging:
  level: "info"
  format: "json"

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200

circuit_breaker:
  enabled: true
  max_failures: 5
  timeout: 30s

health:
  path: "/health"
  ready_path: "/ready"
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MINIO_ENDPOINT` | MinIO S3 endpoint address | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` |
| `MINIO_BUCKET` | Default bucket name for uploads | `media` |
| `MINIO_USE_SSL` | Enable SSL for MinIO connection | `false` |

## Adding New Features

1. **Models**: Define structs in `app/internal/models/models.go`
2. **Repository**: Add S3 operation methods to `app/internal/repo/repo.go`
3. **Service**: Add business logic methods to `app/internal/service/service.go` (update the `Service` interface)
4. **Handler**: Add HTTP handler functions to `app/internal/handler/handlers.go`
5. **Routing**: Register new routes in `app/internal/handler/routing.go`
6. **Configuration**: Add any new config options to `config.yaml` and `app/internal/config/config.go`

## Error Handling

The service uses a comprehensive error system defined in `app/internal/helper/util.go`:

- **General errors**: `ErrInternalServer`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, etc.
- **Service errors**: `ErrServiceUnavailable`, `ErrCreateFailed`, `ErrProcessingFailed`, etc.
- **Validation errors**: `ErrInvalidInput`, `ErrMissingField`, `ErrInvalidFormat`, etc.

Use `helper.MapError()` in the repository layer to convert raw errors to application sentinel errors.
