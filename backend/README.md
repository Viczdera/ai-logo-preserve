# Go backend service for logo preservation system

## Overview
This Go service provides the unified backend for the logo preservation system, handling:
- REST API endpoints for image upload and job management
- Input validation and rate limiting
- S3 integration for image storage
- RabbitMQ job queuing
- Redis for caching and rate limiting

## Features
- **REST API**: Gin-based HTTP server with CORS support
- **File Upload**: Multipart form handling with size and type validation
- **Rate Limiting**: Redis-based rate limiting (500 requests/hour per user)
- **R2 Storage**: Cloudflare R2 integration for image storage with presigned URLs
- **Job Queuing**: RabbitMQ for asynchronous job processing
- **Health Checks**: Built-in health check endpoint

## API Endpoints

### POST /api/v1/upload
Upload an image for logo detection and extraction.

**Request:**
- Content-Type: multipart/form-data
- Field: `image` (file)
- Max file size: 10MB
- Allowed formats: JPEG, PNG

**Response:**
```json
{
  "job_id": "uuid",
  "status": "pending",
  "upload_url": "https://s3-presigned-url",
  "message": "Image uploaded successfully. Processing started."
}
```

### GET /api/v1/jobs/:id
Get the status of a processing job.

**Response:**
```json
{
  "job_id": "uuid",
  "status": "pending|processing|completed|failed",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /api/v1/jobs/:id/result
Get the result of a completed job.

**Response:**
```json
{
  "job_id": "uuid",
  "status": "completed",
  "result_url": "https://s3-presigned-url",
  "logos_found": 3,
  "created_at": "2024-01-01T00:00:00Z",
  "completed_at": "2024-01-01T00:05:00Z"
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "service": "logo-preserve-backend"
}
```

## Configuration

Copy `env.example` to `.env` and configure:

```bash
# Server
SERVER_PORT=8080
MAX_FILE_SIZE=10485760

# Rate Limiting
RATE_LIMIT_PER_HOUR=500
RATE_LIMIT_BURST=10

# AWS/Cloudflare R2
AWS_REGION=auto
AWS_ACCESS_KEY_ID=your_r2_access_key
AWS_SECRET_ACCESS_KEY=your_r2_secret_key
S3_BUCKET=logo-preserve-images
R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=logo-processing
RABBITMQ_QUEUE=detection-jobs

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Development

### Prerequisites
- Go 1.21+
- Docker and Docker Compose
- Cloudflare R2 bucket
- RabbitMQ server
- Redis server

### Local Development

1. **Start dependencies:**
```bash
docker-compose up -d rabbitmq redis
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Run the service:**
```bash
go run main.go
```

### Docker Development

1. **Build and run:**
```bash
docker build -t logo-preserve-backend .
docker run -p 8080:8080 --env-file .env logo-preserve-backend
```

## Architecture

The service follows a clean architecture pattern:

```
main.go
├── internal/
│   ├── api/          # HTTP handlers and routing
│   ├── config/       # Configuration management
│   ├── models/       # Data models
│   ├── storage/      # S3 storage interface
│   └── queue/        # RabbitMQ and Redis clients
```

## Error Handling

The service includes comprehensive error handling:
- Input validation errors (400)
- File size/type validation (400)
- Rate limiting (429)
- Server errors (500)
- Job not found (404)

## Monitoring

- Health check endpoint for monitoring
- Structured logging with logrus
- Rate limiting metrics via Redis
- Job status tracking via Redis
