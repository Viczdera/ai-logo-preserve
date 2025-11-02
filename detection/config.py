"""Configuration management for the detection worker."""
import os
from typing import Optional


class Config:
    """Worker configuration loaded from environment variables."""

    def __init__(self):
        # RabbitMQ Configuration
        self.rabbitmq_url: str = os.getenv(
            "RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"
        )
        self.rabbitmq_exchange: str = os.getenv(
            "RABBITMQ_EXCHANGE", "logo-preserve-exchange"
        )
        self.detection_queue: str = os.getenv(
            "RABBITMQ_QUEUE", "detection-queue"
        )
        self.results_queue: str = os.getenv(
            "RABBITMQ_RESULTS_QUEUE", "results-queue"
        )
        self.results_routing_key: str = os.getenv(
            "RABBITMQ_RESULTS_ROUTING_KEY", "job.result"
        )

        # S3/Cloudflare R2 Configuration
        self.aws_access_key_id: str = os.getenv("BUCKET_ACCESS_KEY_ID", "")
        self.aws_secret_access_key: str = os.getenv("BUCKET_SECRET_KEY", "")
        self.aws_s3_bucket: str = os.getenv("BUCKET_NAME", "")
        self.cloudflare_account_id: str = os.getenv("CLOUDFARE_ACCOUNT_ID", "")
        self.aws_region: str = os.getenv("AWS_REGION", "auto")
        self.aws_endpoint_url: Optional[str] = None
        if self.cloudflare_account_id:
            self.aws_endpoint_url = f"https://{self.cloudflare_account_id}.r2.cloudflarestorage.com"

        # Redis Configuration (optional)
        self.redis_url: Optional[str] = os.getenv("REDIS_URL")

        # Detection Configuration
        self.confidence_threshold: float = float(
            os.getenv("CONFIDENCE_THRESHOLD", "0.85")
        )
        self.model_size: str = os.getenv("YOLO_MODEL_SIZE", "n")  # nano

        # Worker Configuration
        self.max_retries: int = int(os.getenv("MAX_RETRIES", "3"))
        self.retry_backoff_base: float = float(os.getenv("RETRY_BACKOFF_BASE", "1.0"))

    def validate(self) -> list[str]:
        """Validate configuration and return list of missing required fields."""
        errors = []
        
        if not self.rabbitmq_url:
            errors.append("RABBITMQ_URL is required")
        
        if not self.aws_access_key_id:
            errors.append("BUCKET_ACCESS_KEY_ID is required")
        
        if not self.aws_secret_access_key:
            errors.append("BUCKET_SECRET_KEY is required")
        
        if not self.aws_s3_bucket:
            errors.append("BUCKET_NAME is required")
        
        if not self.cloudflare_account_id:
            errors.append("CLOUDFARE_ACCOUNT_ID is required")
        
        return errors


def load_config() -> Config:
    """Load and validate configuration."""
    config = Config()
    errors = config.validate()
    
    if errors:
        raise ValueError(f"Configuration errors: {', '.join(errors)}")
    
    return config

