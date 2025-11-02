"""Main worker loop for processing logo detection jobs from RabbitMQ."""
import json
import logging
import os
import sys
import time
import uuid
from datetime import datetime
from pathlib import Path
from typing import Dict, Any

import cv2
import pika
from pika.exceptions import AMQPConnectionError
from dotenv import load_dotenv

from config import load_config, Config
from detector import LogoDetector
from s3_client import S3Client

env_path= Path("app.env")
load_dotenv(dotenv_path=env_path)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)


class DetectionWorker:
    """Worker that consumes detection jobs from RabbitMQ and processes them."""

    def __init__(self, config: Config):
        """Initialize worker with configuration."""
        self.config = config
        
        # Initialize S3 client
        self.s3_client = S3Client(
            access_key_id=config.aws_access_key_id,
            secret_access_key=config.aws_secret_access_key,
            bucket_name=config.aws_s3_bucket,
            endpoint_url=config.aws_endpoint_url,
        )
        
        # Initialize detector
        self.detector = LogoDetector(
            model_size=config.model_size,
            confidence_threshold=config.confidence_threshold,
        )
        
        # RabbitMQ connection
        self.connection = None
        self.channel = None
        
        # Temporary directory for processing
        self.temp_dir = Path("/tmp/logo_detection")
        self.temp_dir.mkdir(exist_ok=True)

    def connect_rabbitmq(self) -> None:
        """Connect to RabbitMQ and set up queues."""
        try:
            logger.info(f"Connecting to RabbitMQ at {self.config.rabbitmq_url}")
            self.connection = pika.BlockingConnection(
                pika.URLParameters(self.config.rabbitmq_url)
            )
            self.channel = self.connection.channel()
            
            # Declare exchange
            self.channel.exchange_declare(
                exchange=self.config.rabbitmq_exchange,
                exchange_type="direct",
                durable=True,
            )
            
            # Declare and bind detection queue
            self.channel.queue_declare(
                queue=self.config.detection_queue,
                durable=True,
            )
            self.channel.queue_bind(
                exchange=self.config.rabbitmq_exchange,
                queue=self.config.detection_queue,
                routing_key="detection",
            )
            
            # Declare results queue (if it doesn't exist, the Go consumer will create it)
            self.channel.queue_declare(
                queue=self.config.results_queue,
                durable=True,
            )
            self.channel.queue_bind(
                exchange=self.config.rabbitmq_exchange,
                queue=self.config.results_queue,
                routing_key=self.config.results_routing_key,
            )
            
            # Set QoS to process one message at a time
            self.channel.basic_qos(prefetch_count=1)
            
            logger.info("Successfully connected to RabbitMQ")
            
        except AMQPConnectionError as e:
            logger.error(f"Failed to connect to RabbitMQ: {e}")
            raise

    def process_job(self, job: Dict[str, Any]) -> Dict[str, Any]:
        """Process a single detection job.
        
        Args:
            job: Job dictionary with keys: id, s3_key, status, etc.
            
        Returns:
            ProcessingResult dictionary
        """
        job_id = job.get("id")
        s3_key = job.get("s3_key")
        
        logger.info(f"Processing job {job_id} with S3 key {s3_key}")
        
        temp_image_path = None
        temp_logo_paths = []
        
        try:
            # Download image from S3
            temp_image_path = self.temp_dir / f"{job_id}_{int(time.time())}.jpg"
            print(f"Downloading image from S3: {s3_key} to {temp_image_path}")
            self.s3_client.download_file(s3_key, str(temp_image_path))
            logger.info(f"Downloaded image to {temp_image_path}")
            
            # Run detection
            detections = self.detector.detect(str(temp_image_path))
            logger.info(f"Found {len(detections)} detections")
            
            # Extract and upload logos
            logos_found = []
            for idx, detection in enumerate(detections):
                try:
                    # Extract logo region
                    logo_image = self.detector.extract_logo(
                        str(temp_image_path),
                        detection["bbox"],
                    )
                    
                    # Save to temp file
                    temp_logo_path = self.temp_dir / f"{job_id}_logo_{idx}.png"
                    cv2.imwrite(str(temp_logo_path), logo_image)
                    temp_logo_paths.append(temp_logo_path)
                    
                    # Upload to S3
                    logo_s3_key = f"extracted/{job_id}/logo_{idx}.png"
                    self.s3_client.upload_file(str(temp_logo_path), logo_s3_key)
                    
                    # Build LogoDetection object
                    logo_detection = {
                        "id": str(uuid.uuid4()),
                        "job_id": job_id,
                        "bounding_box": detection["bbox"],
                        "confidence": detection["confidence"],
                        "logo_type": detection["class_name"],
                        "s3_key": logo_s3_key,
                    }
                    logos_found.append(logo_detection)
                    
                except Exception as e:
                    logger.error(f"Error processing detection {idx}: {e}")
                    # Continue with other detections
                    continue
            
            # Build processing result
            result = {
                "job_id": job_id,
                "status": "completed",
                "logos_found": logos_found,
                "result_url": "",  # Optional, can be added later
                "processed_at": datetime.utcnow().isoformat() + "Z",
            }
            
            logger.info(
                f"Job {job_id} completed successfully with {len(logos_found)} logos"
            )
            return result
            
        except Exception as e:
            logger.error(f"Error processing job {job_id}: {e}", exc_info=True)
            
            # Return failed result
            return {
                "job_id": job_id,
                "status": "failed",
                "logos_found": [],
                "result_url": "",
                "processed_at": datetime.utcnow().isoformat() + "Z",
                "error": str(e),
            }
            
        finally:
            # Clean up temp files
            if temp_image_path and temp_image_path.exists():
                temp_image_path.unlink()
            for logo_path in temp_logo_paths:
                if logo_path.exists():
                    logo_path.unlink()

    def publish_result(self, result: Dict[str, Any]) -> None:
        """Publish processing result to results queue.
        
        Args:
            result: ProcessingResult dictionary
        """
        try:
            body = json.dumps(result)
            
            self.channel.basic_publish(
                exchange=self.config.rabbitmq_exchange,
                routing_key=self.config.results_routing_key,
                body=body,
                properties=pika.BasicProperties(
                    delivery_mode=2,  # Make message persistent
                    content_type="application/json",
                ),
            )
            
            logger.info(f"Published result for job {result['job_id']} to results queue")
            
        except Exception as e:
            logger.error(f"Failed to publish result: {e}")
            raise

    def process_message(self, ch, method, properties, body: bytes) -> None:
        """Callback for processing RabbitMQ messages.
        
        Args:
            ch: Channel
            method: Method frame
            properties: Properties
            body: Message body (JSON string)
        """
        job = None
        attempts = 0
        max_attempts = self.config.max_retries
        
        while attempts < max_attempts:
            try:
                # Parse job JSON
                job = json.loads(body.decode("utf-8"))
                logger.info(f"Received job: {job.get('id')}")
                
                # Process job
                result = self.process_job(job)
                
                # Publish result
                self.publish_result(result)
                
                # Acknowledge message
                ch.basic_ack(delivery_tag=method.delivery_tag)
                logger.info(f"Successfully processed job {job.get('id')}")
                return
                
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse job JSON: {e}")
                # Reject and don't requeue malformed messages
                ch.basic_nack(
                    delivery_tag=method.delivery_tag,
                    requeue=False,
                )
                return
                
            except Exception as e:
                attempts += 1
                logger.error(
                    f"Error processing job (attempt {attempts}/{max_attempts}): {e}",
                    exc_info=True,
                )
                
                if attempts < max_attempts:
                    # Exponential backoff
                    backoff_time = self.config.retry_backoff_base * (2 ** (attempts - 1))
                    logger.info(f"Retrying in {backoff_time} seconds...")
                    time.sleep(backoff_time)
                else:
                    # Final attempt failed - publish failed result
                    if job:
                        result = {
                            "job_id": job.get("id", "unknown"),
                            "status": "failed",
                            "logos_found": [],
                            "result_url": "",
                            "processed_at": datetime.utcnow().isoformat() + "Z",
                            "error": f"Processing failed after {max_attempts} attempts: {str(e)}",
                        }
                        try:
                            self.publish_result(result)
                        except Exception as pub_error:
                            logger.error(f"Failed to publish failed result: {pub_error}")
                    
                    # Reject and requeue (or don't requeue based on error type)
                    ch.basic_nack(
                        delivery_tag=method.delivery_tag,
                        requeue=True,  # Requeue for manual inspection
                    )

    def start(self) -> None:
        """Start consuming messages from RabbitMQ."""
        try:
            self.connect_rabbitmq()
            
            logger.info("Starting to consume messages from detection queue...")
            
            self.channel.basic_consume(
                queue=self.config.detection_queue,
                on_message_callback=self.process_message,
            )
            
            logger.info("Worker is ready. Waiting for messages...")
            self.channel.start_consuming()
            
        except KeyboardInterrupt:
            logger.info("Received interrupt signal. Stopping worker...")
            self.stop()
        except Exception as e:
            logger.error(f"Worker error: {e}", exc_info=True)
            raise

    def stop(self) -> None:
        """Stop consuming messages and close connections."""
        if self.channel:
            self.channel.stop_consuming()
        if self.connection and not self.connection.is_closed:
            self.connection.close()
        logger.info("Worker stopped")


def main():
    """Main entry point."""
    try:
        config = load_config()
        worker = DetectionWorker(config)
        worker.start()
    except Exception as e:
        logger.error(f"Failed to start worker: {e}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    main()

