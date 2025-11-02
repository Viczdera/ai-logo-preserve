# Logo Detection Worker

Python worker service that consumes logo detection jobs from RabbitMQ, performs YOLOv8 detection on images, and publishes results back to RabbitMQ.

## Features

- Consumes jobs from RabbitMQ `detection-queue`
- Downloads images from Cloudflare R2 (S3-compatible)
- Runs YOLOv8 logo detection with configurable confidence threshold
- Extracts detected logo regions
- Uploads extracted logos to R2
- Publishes results to RabbitMQ `results-queue`
- Retry logic with exponential backoff
- Comprehensive error handling

## Configuration

The worker reads configuration from environment variables. See the "Local Development Setup" section below for all available options and setup instructions.

Required variables:
- `RABBITMQ_URL`: RabbitMQ connection string
- `AWS_ACCESS_KEY_ID`: Cloudflare R2 access key
- `AWS_SECRET_ACCESS_KEY`: Cloudflare R2 secret key
- `AWS_S3_BUCKET`: R2 bucket name
- `CLOUDFARE_ACCOUNT_ID`: Cloudflare account ID (for R2 endpoint)

## Local Development Setup

### Prerequisites

- Python 3.8 or higher
- pip (Python package manager)
- RabbitMQ (can use Docker or install locally)
- Cloudflare R2 account with API credentials

### Step-by-Step Setup

#### 1. Install Python Dependencies

Create a virtual environment (recommended):
```bash
cd detection
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install --upgrade pip
pip install -r requirements.txt
```

#### 2. Set Up RabbitMQ

**Option A: Using Docker Compose (Recommended)**

From the project root directory:
```bash
docker-compose up -d rabbitmq
```

This will start RabbitMQ on:
- Port `5672` (AMQP)
- Port `15672` (Management UI) - http://localhost:15672

**Option B: Install RabbitMQ Locally**

- **macOS**: `brew install rabbitmq`
- **Linux**: `sudo apt-get install rabbitmq-server` (Ubuntu/Debian)
- **Windows**: Download from https://www.rabbitmq.com/download.html

Then start RabbitMQ:
```bash
rabbitmq-server
```

#### 3. Configure Environment Variables

Create a `.env` file in the `detection` directory:

```bash
cd detection
touch .env
```

Add the following variables to `.env`:

```bash
# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=logo-preserve-exchange
RABBITMQ_QUEUE=detection-queue
RABBITMQ_RESULTS_QUEUE=results-queue
RABBITMQ_RESULTS_ROUTING_KEY=job.result

# Cloudflare R2 Configuration (Required)
AWS_ACCESS_KEY_ID=your_r2_access_key
AWS_SECRET_ACCESS_KEY=your_r2_secret_key
AWS_S3_BUCKET=your-bucket-name
CLOUDFARE_ACCOUNT_ID=your_cloudflare_account_id
AWS_REGION=auto

# Detection Configuration (Optional)
CONFIDENCE_THRESHOLD=0.85
YOLO_MODEL_SIZE=n  # Options: n (nano), s (small), m (medium), l (large), x (xlarge)

# Worker Configuration (Optional)
MAX_RETRIES=3
RETRY_BACKOFF_BASE=1.0
```

**Note**: The default RabbitMQ credentials are `guest:guest` for local development. If using docker-compose, check your `env.example` file in the project root for the actual credentials.

#### 4. Load Environment Variables

**Option A: Use python-dotenv (Recommended)**

Install it first:
```bash
pip install python-dotenv
```

Then create a simple script or modify `worker.py` to load `.env`:
```python
from dotenv import load_dotenv
load_dotenv()
```

**Option B: Export Variables Manually**

```bash
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export AWS_ACCESS_KEY_ID="your_key"
export AWS_SECRET_ACCESS_KEY="your_secret"
export AWS_S3_BUCKET="your_bucket"
export CLOUDFARE_ACCOUNT_ID="your_account_id"
```

#### 5. Run the Worker

Make sure your virtual environment is activated and environment variables are set:

```bash
python worker.py
```

You should see output like:
```
2024-01-01 00:00:00 [INFO] Connecting to RabbitMQ at amqp://guest:guest@localhost:5672/
2024-01-01 00:00:00 [INFO] Successfully connected to RabbitMQ
2024-01-01 00:00:00 [INFO] Starting to consume messages from detection queue...
2024-01-01 00:00:00 [INFO] Worker is ready. Waiting for messages...
```

### Troubleshooting

**Connection Errors:**
- Ensure RabbitMQ is running: `docker ps` or check RabbitMQ status
- Verify RabbitMQ URL matches your setup
- Check firewall settings if using non-localhost connection

**Import Errors:**
- Ensure virtual environment is activated
- Reinstall dependencies: `pip install -r requirements.txt`

**YOLO Model Download:**
- On first run, the YOLOv8 model will be downloaded automatically
- This may take a few minutes depending on your internet connection
- Models are cached in `~/.ultralytics/` directory

**Permission Errors:**
- Ensure `/tmp/logo_detection` directory is writable
- Or modify `worker.py` to use a different temp directory

### Testing the Setup

You can verify RabbitMQ is working by:
1. Opening RabbitMQ Management UI: http://localhost:15672 (guest/guest)
2. Checking queues are created when the worker starts
3. Sending a test message to the `detection-queue`

## Docker

Build image:
```bash
docker build -t logo-detection-worker .
```

Run container:
```bash
docker run --env-file .env logo-detection-worker
```

## Message Format

### Input (from detection-queue)
```json
{
  "id": "job-uuid",
  "status": "pending",
  "s3_key": "original/job-uuid/filename.jpg",
  "upload_url": "https://...",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Output (to results-queue)
```json
{
  "job_id": "job-uuid",
  "status": "completed",
  "logos_found": [
    {
      "id": "logo-uuid",
      "job_id": "job-uuid",
      "bounding_box": {"x": 100, "y": 200, "width": 150, "height": 150},
      "confidence": 0.92,
      "logo_type": "unknown",
      "s3_key": "extracted/job-uuid/logo_0.png"
    }
  ],
  "result_url": "",
  "processed_at": "2024-01-01T00:05:00Z"
}
```

## Model

The worker uses YOLOv8n (nano) by default for fast inference. The model is automatically downloaded on first run from Ultralytics.

For MVP, all detected objects with confidence > 0.85 are treated as potential logos.

