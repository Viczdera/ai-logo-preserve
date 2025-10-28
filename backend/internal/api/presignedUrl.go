package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type getPresignedUrlRequest struct {
	FileName string `json:"file_name"`
}

func (s *Server) GetPresignedUrl(ctx *gin.Context) {
	var req getPresignedUrlRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err, "Invalid request"))
		return
	}

	if req.FileName == "" {
		ctx.JSON(http.StatusBadRequest, errResponse(errors.New("key is required"), "Key is required"))
		return
	}

	presignedUrl, err := s.storageClient.GetPresignedURL(ctx.Request.Context(), req.FileName, 1*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err, "Failed to get presigned URL"))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Presigned URL generated successfully", presignedUrl))
}
