<!-- 077c14af-bf4b-4700-9f4b-1814ad781141 6d4ccfbf-c3db-4945-abcc-4d7a0fec3830 -->
# Logo Preserve MVP - Portfolio Project

## Architecture Summary

- **Go Backend**: REST API + image ingestion service (already scaffolded)
- **Python Worker**: YOLOv8 logo detection + bounding-box extraction (single container, sequential processing)
- **Message Queue**: RabbitMQ for async job processing
- **Storage**: Cloudflare R2 (S3-compatible, already configured)
- **State Management**: PostgreSQL (persistent jobs/logos) + Redis (real-time status/cache)
- **Frontend**: Next.js with drag-and-drop upload UI

## Implementation Plan

### 1. Backend Infrastructure Setup

**Complete Go Backend API** (`backend/internal/api/`)

- Fix Redis client initialization in `server.go` (currently commented out)
- Uncomment and fix RabbitMQ queue client initialization in `main.go`
- Update `upload.go` to publish jobs to RabbitMQ after S3 upload
- Fix import path in `queue/rabbitmq.go` (references wrong module path)
- Add missing config loading for RabbitMQ and Redis in `utils/config.go`

**Database Layer** (`backend/internal/db/`)

- Create PostgreSQL schema: `jobs`, `logos`, `detections` tables
- Implement repository pattern with `pgx` driver
- Add database migrations (use `golang-migrate` or embed SQL)
- Store job creation, status updates, and detection results in Postgres
- Keep Redis for fast status lookups (cache-aside pattern)

**API Endpoints** (already scaffolded, need completion)

- `POST /api/v1/upload` - Upload image → S3 → publish to RabbitMQ ✓ (mostly done)
- `GET /api/v1/jobs/:id` - Get job status from Redis/PostgreSQL
- `GET /api/v1/jobs/:id/result` - Get detection results + presigned URLs for extracted logos
- `POST /api/v1/upload/presigned-url` - Generate presigned URL for client-side upload ✓ (done)

### 2. Python CV Worker

**Setup** (`workers/detection/`)

- Create Dockerfile with Python 3.11 + PyTorch + YOLOv8 (ultralytics)
- Use pre-trained YOLOv8 model (no fine-tuning for MVP)
- Install dependencies: `ultralytics`, `opencv-python`, `pika` (RabbitMQ), `boto3` (S3)

**Worker Logic** (`worker.py`)

```python
# Consume jobs from RabbitMQ "detection" queue
# 1. Download image from S3 (original/{job_id}/{filename})
# 2. Run YOLOv8 detection (detect logos/objects)
# 3. Extract bounding boxes (simple crop, no perspective correction)
# 4. Upload extracted logos to S3 (extracted/{job_id}/logo_{idx}.png)
# 5. Publish results back to RabbitMQ "results" queue
# 6. Update job status in Redis + PostgreSQL
```

**Detection Logic**

- Use YOLOv8n (nano) for speed - fine-tune later
- Filter detections by confidence threshold (>0.85)
- For MVP: Treat any detected object as potential logo (relax class restrictions)
- Save bounding box coordinates + confidence scores

### 3. Message Queue Integration

**RabbitMQ Setup** (`docker-compose.yml`)

- Add RabbitMQ service with management UI (port 15672)
- Create exchanges: `logo-preserve-exchange` (direct)
- Create queues: `detection-queue`, `results-queue`
- Bind routing keys: `job.detect`, `job.result`

**Go Publisher** (already in `queue/rabbitmq.go`, needs fixes)

- Fix import paths
- Publish job to `detection-queue` after S3 upload
- Include job metadata: job_id, s3_key, filename, upload_time

**Python Consumer**

- Connect to RabbitMQ, consume from `detection-queue`
- Process job, publish results to `results-queue`
- Implement retry logic (3 attempts, exponential backoff)

**Go Results Consumer** (new: `backend/internal/workers/results_consumer.go`)

- Consume from `results-queue`
- Update job status in PostgreSQL + Redis
- Store detection metadata (bounding boxes, confidence, S3 keys)

### 4. Frontend Upload Interface

**Upload Page** (`frontend/app/page.tsx`)

- Drag-and-drop file upload component (use `react-dropzone` or native)
- Image preview before upload
- Upload progress indicator
- Display job ID and link to results page

**Results Page** (`frontend/app/results/[jobId]/page.tsx`)

- Poll job status endpoint every 2s (use `useEffect` + `setInterval`)
- Show processing stages: "Uploaded" → "Processing" → "Completed"
- Display detection results:
  - Original image with bounding boxes overlaid
  - Extracted logo thumbnails
  - Confidence scores
  - Download links for extracted logos
- Error handling (failed jobs, not found)

**API Client** (`frontend/lib/api.ts`)

- Typed fetch wrappers for all endpoints
- Error handling and retry logic
- Types matching Go backend models

### 5. Docker Compose Orchestration

**Services** (`docker-compose.yml`)

```yaml
services:
  backend:        # Go API (port 8083)
  worker:         # Python CV worker
  rabbitmq:       # Message queue (5672, 15672)
  redis:          # Cache (6379)
  postgres:       # Database (5432)
  frontend:       # Next.js (3000)
```

**Networking**

- All services on same Docker network
- Backend connects to: RabbitMQ, Redis, PostgreSQL, S3 (external)
- Worker connects to: RabbitMQ, Redis, PostgreSQL, S3 (external)
- Frontend connects to: Backend only (via localhost:8083)

### 6. Configuration & Environment

**Backend** (`backend/app.env`)

- Add RabbitMQ URL, Redis URL, PostgreSQL connection string
- Keep existing Cloudflare R2 credentials
- Add allowed CORS origins

**Worker** (`workers/detection/.env`)

- RabbitMQ URL, Redis URL, PostgreSQL URL
- Cloudflare R2 credentials (read/write to buckets)
- Model path, confidence threshold

**Frontend** (`frontend/.env.local`)

- `NEXT_PUBLIC_API_URL=http://localhost:8083`

### 7. Testing & Demo Preparation

**Seed Data**

- Prepare 5-10 apparel images with visible logos (Nike, Adidas, etc.)
- Store in `demo/images/` directory

**Integration Test**

- Upload image → verify job created in PostgreSQL
- Check RabbitMQ queue has message
- Verify worker processes job
- Confirm extracted logos in S3
- Validate results returned to frontend

**Demo Script**

1. Start all services: `docker-compose up`
2. Open frontend: `http://localhost:3000`
3. Upload Nike t-shirt image
4. Show real-time status updates
5. Display detection results with bounding boxes
6. Download extracted logo
7. Show RabbitMQ management UI with message flow

## Key Files to Create/Modify

**Backend**

- `backend/internal/db/postgres.go` (new - DB connection)
- `backend/internal/db/repository.go` (new - CRUD operations)
- `backend/internal/db/migrations/001_initial.sql` (new - schema)
- `backend/internal/workers/results_consumer.go` (new - consume results)
- `backend/internal/api/upload.go` (modify - add queue publish)
- `backend/internal/queue/rabbitmq.go` (fix - import paths)
- `backend/internal/utils/config.go` (modify - add DB/Queue config)
- `backend/main.go` (modify - wire up all services)

**Python Worker**

- `workers/detection/worker.py` (new - main worker logic)
- `workers/detection/detector.py` (new - YOLOv8 wrapper)
- `workers/detection/requirements.txt` (new - dependencies)
- `workers/detection/Dockerfile` (new - container image)

**Frontend**

- `frontend/app/page.tsx` (modify - upload UI)
- `frontend/app/results/[jobId]/page.tsx` (new - results page)
- `frontend/lib/api.ts` (new - API client)
- `frontend/components/FileUpload.tsx` (new - upload component)
- `frontend/components/DetectionResults.tsx` (new - results display)

**Infrastructure**

- `docker-compose.yml` (modify - add all services)
- `backend/Dockerfile` (verify/update)
- `.env.example` (update with all required vars)

## Success Criteria

✅ Upload image via frontend

✅ Job persisted in PostgreSQL, status in Redis

✅ RabbitMQ delivers message to Python worker

✅ YOLOv8 detects logos (>0.85 confidence)

✅ Extracted logos saved to S3 `extracted/` bucket

✅ Frontend displays results with bounding boxes

✅ End-to-end processing time <30 seconds

✅ All services run via `docker-compose up`

✅ Demo-ready with 5 sample images

### To-dos

- [ ] Set up backend infrastructure (Redis, RabbitMQ clients, DB layer)
- [ ] Create PostgreSQL schema, migrations, and repository layer
- [ ] Fix and complete RabbitMQ integration in Go backend
- [ ] Complete job status and results API endpoints
- [ ] Build Python CV worker with YOLOv8 detection and RabbitMQ consumer
- [ ] Create Go results consumer to process worker output
- [ ] Build Next.js upload interface with drag-and-drop
- [ ] Build results page with status polling and detection visualization
- [ ] Create docker-compose.yml with all services (backend, worker, RabbitMQ, Redis, PostgreSQL, frontend)
- [ ] Test end-to-end flow and prepare demo