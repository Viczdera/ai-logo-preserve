package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Viczdera/ai-logo-preserve/backend/internal/api"
	db "github.com/Viczdera/ai-logo-preserve/backend/internal/db/sqlc"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/queue"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/storage"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/utils"
	_ "github.com/lib/pq"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}
	if config.Server.Port == "" {
		log.Fatal("Port is not set")
	}
	serverSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName, config.Database.SSLMode)

	//initialize storage client
	storageClient, err := storage.NewS3Client(config.Cloudflare)
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}

	//initialize database
	connDB, err := sql.Open(utils.DBDriver, serverSource)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer connDB.Close()

	// Initialize sqlc queries
	queries := db.NewStore(connDB)
	if err != nil {
		log.Fatal("Failed to initialize store queries:", err)
	}
	// Initialize Redis client
	redisClient := queue.NewRedisClient(config.Redis)
	defer redisClient.Close()

	// Initialize message queue (RabbitMQ)
	queueClient, err := queue.NewRabbitMQClient(config.RabbitMQ)
	if err != nil {
		log.Fatal("Failed to initialize message queue:", err)
	}
	defer queueClient.Close()

	// Initialize server with all dependencies
	server := api.NewServer(config, storageClient, queries, redisClient, queueClient)

	log.Printf("Starting server on port %s", config.Server.Port)
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
