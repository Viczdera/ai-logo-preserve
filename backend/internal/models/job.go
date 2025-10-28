package models

import "time"

type Job struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"` // pending, processing, completed, failed
	S3Key       string     `json:"s3_key"`
	UploadURL   string     `json:"upload_url"`
	ResultURL   string     `json:"result_url,omitempty"`
	LogosFound  int        `json:"logos_found,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type LogoDetection struct {
	ID          string  `json:"id"`
	JobID       string  `json:"job_id"`
	BoundingBox BBox    `json:"bounding_box"`
	Confidence  float64 `json:"confidence"`
	LogoType    string  `json:"logo_type"`
	S3Key       string  `json:"s3_key"`
}

type BBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ProcessingResult struct {
	JobID       string          `json:"job_id"`
	Status      string          `json:"status"`
	LogosFound  []LogoDetection `json:"logos_found"`
	ResultURL   string          `json:"result_url"`
	ProcessedAt time.Time       `json:"processed_at"`
}
