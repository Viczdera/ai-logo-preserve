package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	db "github.com/Viczdera/ai-logo-preserve/backend/internal/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	AllowedTypes = []string{"image/jpeg", "image/png"}
)

func (s *Server) UploadImage(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No image file provided"})
		return
	}
	defer file.Close()

	// Validations
	if !slices.Contains(AllowedTypes, header.Header.Get("Content-Type")) {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid file type. Only JPEG and PNG are allowed"})
		return
	}

	if header.Size > s.config.Server.MaxFileSize {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false, "error": fmt.Sprintf("File too large. Maximum size is %d bytes", s.config.Server.MaxFileSize),
		})
		return
	}

	jobID := uuid.New()
	s3Key := fmt.Sprintf("original/%s/%s", jobID.String(), header.Filename)

	// Upload file to S3
	_, err = s.storageClient.UploadFile(context.Background(), s3Key, file, header.Size)
	if err != nil {
		logrus.WithError(err).Error("Failed to upload file to S3")
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to upload file"})
		return
	}

	// Get presigned URL for accessing the uploaded file
	uploadURL, err := s.storageClient.GetPresignedGetURL(context.Background(), s3Key, 24*time.Hour)
	if err != nil {
		logrus.WithError(err).Error("Failed to get presigned URL")
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to generate file URL"})
		return
	}

	// Create job record in database
	_, err = s.store.CreateJob(context.Background(), db.CreateJobParams{
		ID:        int64(jobID.ID()),
		Status:    "pending",
		S3Key:     s3Key,
		UploadUrl: uploadURL,
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to create job in database")
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create job"})
		return
	}

	// // Store job status in Redis for fast lookups
	// err = s.storeJobStatusInRedis(&job)
	// if err != nil {
	// 	logrus.WithError(err).Error("Failed to store job status in Redis")
	// 	// Don't fail the request, just log the error
	// }

	// // Publish job to queue for processing
	// jobModel := &models.Job{
	// 	ID:        job.ID.String(),
	// 	Status:    job.Status,
	// 	S3Key:     job.S3Key,
	// 	UploadURL: uploadURL,
	// 	CreatedAt: job.CreatedAt.Time,
	// 	UpdatedAt: job.UpdatedAt.Time,
	// }

	// err = s.queueClient.PublishJob(jobModel)
	// if err != nil {
	// 	logrus.WithError(err).Error("Failed to publish job to queue")
	// 	// Update job status to failed
	// 	s.queries.UpdateJobError(context.Background(), db.UpdateJobErrorParams{
	// 		ID:           jobID,
	// 		Status:       "failed",
	// 		ErrorMessage: db.NewNullString("Failed to queue job for processing"),
	// 	})
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to queue job"})
	// 	return
	//}

	ctx.JSON(http.StatusAccepted, gin.H{
		"success":    true,
		"job_id":     jobID.String(),
		"status":     "pending",
		"message":    "Image uploaded successfully. Processing started.",
		"upload_url": uploadURL,
	})
}
