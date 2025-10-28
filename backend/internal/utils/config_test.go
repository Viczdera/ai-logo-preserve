package utils

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	bucketName          = "test-bucket"
	bucketAccessKeyID   = "test-access-key"
	bucketSecretKey     = "test-secret-key"
	cloudflareAccountID = "test-account-id"
	port                = "8080"
	maxFileSize         = 1048576
)

func TestLoadConfig(t *testing.T) {

	os.Setenv("BUCKET_NAME", bucketName)
	os.Setenv("BUCKET_ACCESS_KEY_ID", bucketAccessKeyID)
	os.Setenv("BUCKET_SECRET_KEY", bucketSecretKey)
	os.Setenv("CLOUDFARE_ACCOUNT_ID", cloudflareAccountID)
	os.Setenv("PORT", port)
	os.Setenv("MAX_FILE_SIZE", strconv.Itoa(maxFileSize))

	defer func() {
		os.Unsetenv("BUCKET_NAME")
		os.Unsetenv("BUCKET_ACCESS_KEY_ID")
		os.Unsetenv("BUCKET_SECRET_KEY")
		os.Unsetenv("CLOUDFARE_ACCOUNT_ID")
		os.Unsetenv("PORT")
		os.Unsetenv("MAX_FILE_SIZE")
	}()

	_, err := LoadConfig("../../")
	require.NoError(t, err)

	assert.Equal(t, bucketName, os.Getenv("BUCKET_NAME"))
	assert.Equal(t, bucketAccessKeyID, os.Getenv("BUCKET_ACCESS_KEY_ID"))
	assert.Equal(t, bucketSecretKey, os.Getenv("BUCKET_SECRET_KEY"))
	assert.Equal(t, cloudflareAccountID, os.Getenv("CLOUDFARE_ACCOUNT_ID"))
	assert.Equal(t, strconv.Itoa(maxFileSize), os.Getenv("MAX_FILE_SIZE"))
}

func TestLoadConfig_WithFile(t *testing.T) {
	config, err := LoadConfig("../../")
	require.NoError(t, err)

	assert.Equal(t, "8083", config.Server.Port)
	assert.Equal(t, int64(1002688), config.Server.MaxFileSize)
}
