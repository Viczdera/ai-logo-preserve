"""S3/R2 client wrapper for downloading and uploading files."""
import boto3
from botocore.config import Config
from botocore.exceptions import ClientError
import logging
from typing import Optional
import os

logger = logging.getLogger(__name__)


class S3Client:
    """S3 client for Cloudflare R2 (S3-compatible) operations."""

    def __init__(
        self,
        access_key_id: str,
        secret_access_key: str,
        bucket_name: str,
        endpoint_url: Optional[str] = None,
        region: str = "auto",
    ):
        """Initialize S3 client.
        
        Args:
            access_key_id: AWS access key ID
            secret_access_key: AWS secret access key
            bucket_name: S3 bucket name
            endpoint_url: Custom endpoint URL (for Cloudflare R2)
            region: AWS region (default: "auto" for R2)
        """
        self.bucket_name = bucket_name
        
        # Configure boto3 for R2
        boto_config = Config(
            region_name=region,
            signature_version="s3v4",
        )
        
        self.s3_client = boto3.client(
            "s3",
            endpoint_url=endpoint_url,
            aws_access_key_id=access_key_id,
            aws_secret_access_key=secret_access_key,
            config=boto_config,
        )
        
        logger.info(f"S3 client initialized for bucket: {bucket_name}")

    def download_file(self, s3_key: str, local_path: str) -> None:
        """Download file from S3 to local path.
        
        Args:
            s3_key: S3 object key
            local_path: Local file path to save to
            
        Raises:
            ClientError: If download fails
        """
        try:
            logger.info(f"Downloading {s3_key} to {local_path}")
            self.s3_client.download_file(self.bucket_name, s3_key, local_path)
            logger.info(f"Successfully downloaded {s3_key}")
        except ClientError as e:
            logger.error(f"Failed to download {s3_key}: {e}")
            raise

    def upload_file(self, local_path: str, s3_key: str, content_type: Optional[str] = None) -> str:
        """Upload local file to S3.
        
        Args:
            local_path: Local file path to upload
            s3_key: S3 object key
            content_type: Optional content type (default: inferred from extension)
            
        Returns:
            S3 key of uploaded file
            
        Raises:
            ClientError: If upload fails
        """
        try:
            extra_args = {}
            if content_type:
                extra_args["ContentType"] = content_type
            elif local_path.endswith(".png"):
                extra_args["ContentType"] = "image/png"
            elif local_path.endswith(".jpg") or local_path.endswith(".jpeg"):
                extra_args["ContentType"] = "image/jpeg"
            
            logger.info(f"Uploading {local_path} to {s3_key}")
            self.s3_client.upload_file(
                local_path,
                self.bucket_name,
                s3_key,
                ExtraArgs=extra_args,
            )
            logger.info(f"Successfully uploaded {s3_key}")
            return s3_key
        except ClientError as e:
            logger.error(f"Failed to upload {local_path} to {s3_key}: {e}")
            raise

    def get_presigned_url(self, s3_key: str, expiration: int = 3600) -> str:
        """Generate presigned URL for S3 object.
        
        Args:
            s3_key: S3 object key
            expiration: URL expiration time in seconds (default: 1 hour)
            
        Returns:
            Presigned URL string
            
        Raises:
            ClientError: If URL generation fails
        """
        try:
            url = self.s3_client.generate_presigned_url(
                "get_object",
                Params={"Bucket": self.bucket_name, "Key": s3_key},
                ExpiresIn=expiration,
            )
            return url
        except ClientError as e:
            logger.error(f"Failed to generate presigned URL for {s3_key}: {e}")
            raise


