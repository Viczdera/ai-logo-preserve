package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Viczdera/ai-logo-preserve/backend/internal/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client interface {
	UploadFile(ctx context.Context, key string, file io.Reader, size int64) (*s3.PutObjectOutput, error)
	GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	GetPresignedGetURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	DeleteFile(ctx context.Context, key string) error
}

type S3Client struct {
	client     *s3.Client
	bucketName string
}

func NewS3Client(envVar utils.CloudflareConfig) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(envVar.BucketAccessKeyID, envVar.BucketSecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", envVar.CloudflareAccountID))
	})

	return &S3Client{
		client:     client,
		bucketName: envVar.BucketName,
	}, nil
}

func (c *S3Client) UploadFile(ctx context.Context, key string, file io.Reader, size int64) (*s3.PutObjectOutput, error) {
	output, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return output, nil

}

func (c *S3Client) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)
	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL for PutObject: %w", err)
	}

	fmt.Println(presignResult.URL)
	return presignResult.URL, nil
}

func (c *S3Client) GetPresignedGetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL for GetObject: %w", err)
	}

	return presignResult.URL, nil
}

func (c *S3Client) DeleteFile(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
