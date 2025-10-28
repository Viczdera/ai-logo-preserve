package db

//same package as crud code
//main_test used to set up connection with query object

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCreateJob(t *testing.T) {

	db := testQueries
	job, err := db.CreateJob(context.Background(), CreateJobParams{
		ID:        int64(uuid.New().ID()),
		Status:    "pending",
		S3Key:     int64(uuid.New().ID()),
		UploadUrl: uuid.New().String()[:10],
	})
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}
	t.Logf("Job created: %v", job)

}
