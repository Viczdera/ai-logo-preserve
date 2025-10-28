package storage

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/Viczdera/ai-logo-preserve/backend/internal/utils"
	"github.com/stretchr/testify/require"
)

// var bucketName = "test-bucket"
// var accountId = "test-account-id"
// var accessKeyId = "test-access-key-id"
// var accessKeySecret = "test-access-key-secret"

func newS3Client() (*S3Client, error) {

	config, err := utils.LoadConfig("../../")
	if err != nil {
		return nil, err
	}
	cloudFlareConfig := utils.CloudflareConfig{
		BucketName:          config.Cloudflare.BucketName,
		CloudflareAccountID: config.Cloudflare.CloudflareAccountID,
		BucketAccessKeyID:   config.Cloudflare.BucketAccessKeyID,
		BucketSecretKey:     config.Cloudflare.BucketSecretKey,
	}

	return NewS3Client(cloudFlareConfig)
}

func TestNewS3Client(t *testing.T) {
	storageClient, err := newS3Client()
	require.NoError(t, err)
	require.NotNil(t, storageClient)
}

func TestGetPresignedURL(t *testing.T) {
	storageClient, _ := newS3Client()

	presignedURL, err := storageClient.GetPresignedURL(context.Background(), "test.jpg", 1*time.Hour)
	log.Println(presignedURL)
	require.NoError(t, err)
	require.NotEmpty(t, presignedURL)
}

func TestUploadFile(t *testing.T) {
	storageClient, _ := newS3Client()
	dummyFile := io.Reader(nil)

	presignedURL, err := storageClient.UploadFile(context.Background(), "test.jpg", dummyFile, 256*1024)
	log.Println(presignedURL)
	require.NoError(t, err)
	require.NotEmpty(t, presignedURL)
}
