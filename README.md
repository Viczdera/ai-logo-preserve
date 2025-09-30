# Logo Preservation System

A comprehensive AI-powered system for detecting, extracting, and recomposing logos on apparel using YOLOv8, SAM, and OpenCV.

## Architecture Overview

```
┌─────────────┐      ┌──────────────────────────────────────────┐
│   Next.js   │◄────►│  REST API (FastAPI + Python)             │
│  Dashboard  │      │  - OAuth2 (JWT via Supabase Auth)        │
│  + Upload   │      │  - Rate limiting (500 images/hour/user)  │
└─────────────┘      │  - Input validation (max 10MB, PNG/JPG)  │
                     └──────────────────┬───────────────────────┘
                                        │
                     ┌──────────────────▼───────────────────────┐
                     │  Image Ingestion Service (Go)            │
                     │  - S3 upload with presigned URLs         │
                     │  - Image preprocessing (resize, format)  │
                     │  - Publishes job to RabbitMQ             │
                     └──────────────────┬───────────────────────┘
                                        │
                     ┌──────────────────▼───────────────────────┐
                     │  Logo Detection Worker (Python + CUDA)   │
                     │  - YOLOv8 (fine-tuned on 10K apparel)    │
                     │  - SAM (Segment Anything) for mask       │
                     │  - Confidence scoring + bounding box     │
                     │  - Batch inference (32 images/GPU)       │
                     └──────────────────┬───────────────────────┘
                                        │
                     ┌──────────────────▼───────────────────────┐
                     │  Logo Extraction Service (Rust)          │
                     │  - OpenCV homography estimation          │
                     │  - Perspective unwarp (transform to rect)│
                     │  - Edge detection + alpha mask creation  │
                     │  - Stores extracted logo in S3           │
                     └──────────────────┬───────────────────────┘
                                        │
       ┌────────────────────────────────┴────────────────────┐
       │                                                      │
┌──────▼─────┐                                        ┌──────▼─────┐
│  Postgres  │                                        │  S3        │
│ (Supabase) │                                        │  Buckets   │
│ - jobs     │                                        │ - original │
│ - logos    │                                        │ - extracted│
│ - brands   │                                        │ - composite│
└────────────┘                                        └────────────┘
       │
┌──────▼───────────────────────────────────────────────────────────┐
│  Logo Re-Composition Worker (Python + OpenCV)                    │
│  - Reads AI-generated target image from S3                       │
│  - Detects garment region (YOLOv8 instance segmentation)         │
│  - Estimates pose/perspective via keypoint detection             │
│  - Applies homography to warp extracted logo                     │
│  - Color correction (histogram matching)                         │
│  - Alpha blending with feathered edges                           │
│  - Optional: Stable Diffusion inpainting for seamless blend      │
└──────────────────────────────────────────────────────────────────┘
```

## Services

- **Frontend**: Next.js dashboard with upload functionality
- **API**: FastAPI backend with authentication and rate limiting
- **Ingestion**: Go service for image processing and job queuing
- **Detection**: Python worker with YOLOv8 and SAM for logo detection
- **Extraction**: Rust service for logo extraction using OpenCV
- **Composition**: Python worker for logo recomposition
- **Infrastructure**: PostgreSQL, S3, RabbitMQ, Redis, Prometheus/Grafana

## Quick Start

1. Clone the repository
2. Set up environment variables (see `.env.example`)
3. Run `docker-compose up` for local development
4. Access the dashboard at `http://localhost:3000`

## Development

Each service has its own directory with specific setup instructions:

- `frontend/` - Next.js application
- `api/` - FastAPI backend
- `ingestion/` - Go image ingestion service
- `detection/` - Python logo detection worker
- `extraction/` - Rust logo extraction service
- `composition/` - Python logo composition worker
- `infrastructure/` - Database schemas and configurations

## License

MIT
# ai-logo-presevere
