"""YOLOv8 detection wrapper for logo detection."""
import logging
from pathlib import Path
from typing import List, Dict, Any
from ultralytics import YOLO
import cv2
import numpy as np

logger = logging.getLogger(__name__)


class LogoDetector:
    """YOLOv8-based logo detector."""

    def __init__(self, model_size: str = "n", confidence_threshold: float = 0.85):
        """Initialize detector with YOLOv8 model.
        
        Args:
            model_size: Model size (n=nano, s=small, m=medium, l=large, x=xlarge)
            confidence_threshold: Minimum confidence for detections
        """
        self.confidence_threshold = confidence_threshold
        model_name = f"yolov8{model_size}.pt"
        
        logger.info(f"Loading YOLOv8 model: {model_name}")
        try:
            self.model = YOLO(model_name)
            logger.info(f"Successfully loaded {model_name}")
        except Exception as e:
            logger.error(f"Failed to load YOLOv8 model: {e}")
            raise

    def detect(self, image_path: str) -> List[Dict[str, Any]]:
        """Detect objects/logos in image.
        
        Args:
            image_path: Path to image file
            
        Returns:
            List of detections with keys: bbox (dict with x, y, width, height),
            confidence (float), class_name (str), class_id (int)
        """
        try:
            logger.info(f"Running detection on {image_path}")
            
            # Run YOLOv8 inference
            results = self.model(image_path, conf=self.confidence_threshold)
            
            detections = []
            
            # Process results (YOLOv8 returns list of Results objects)
            for result in results:
                boxes = result.boxes
                
                # Get image dimensions for bounding box conversion
                img = cv2.imread(image_path)
                if img is None:
                    logger.error(f"Failed to read image: {image_path}")
                    return []
                
                img_height, img_width = img.shape[:2]
                
                for box in boxes:
                    # Extract bounding box coordinates (xyxy format)
                    x1, y1, x2, y2 = box.xyxy[0].cpu().numpy()
                    
                    # Convert to x, y, width, height format
                    x = int(x1)
                    y = int(y1)
                    width = int(x2 - x1)
                    height = int(y2 - y1)
                    
                    # Get confidence and class
                    confidence = float(box.conf[0].cpu().numpy())
                    class_id = int(box.cls[0].cpu().numpy())
                    class_name = self.model.names[class_id] if hasattr(self.model, 'names') else "unknown"
                    
                    # Filter by confidence threshold
                    if confidence >= self.confidence_threshold:
                        detections.append({
                            "bbox": {
                                "x": x,
                                "y": y,
                                "width": width,
                                "height": height,
                            },
                            "confidence": confidence,
                            "class_name": class_name,
                            "class_id": class_id,
                        })
                        logger.debug(
                            f"Detection: {class_name} at ({x}, {y}, {width}, {height}) "
                            f"with confidence {confidence:.2f}"
                        )
            
            logger.info(f"Found {len(detections)} detections in {image_path}")
            return detections
            
        except Exception as e:
            logger.error(f"Error during detection: {e}")
            raise

    def extract_logo(self, image_path: str, bbox: Dict[str, int]) -> np.ndarray:
        """Extract logo region from image using bounding box.
        
        Args:
            image_path: Path to source image
            bbox: Bounding box dict with x, y, width, height
            
        Returns:
            Cropped image as numpy array
        """
        try:
            img = cv2.imread(image_path)
            if img is None:
                raise ValueError(f"Failed to read image: {image_path}")
            
            x = bbox["x"]
            y = bbox["y"]
            width = bbox["width"]
            height = bbox["height"]
            
            # Ensure coordinates are within image bounds
            x = max(0, min(x, img.shape[1]))
            y = max(0, min(y, img.shape[0]))
            width = min(width, img.shape[1] - x)
            height = min(height, img.shape[0] - y)
            
            # Crop the region
            cropped = img[y:y+height, x:x+width]
            return cropped
            
        except Exception as e:
            logger.error(f"Error extracting logo: {e}")
            raise

