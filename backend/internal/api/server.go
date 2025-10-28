package api

import (
	"net/http"
	"time"

	db "github.com/Viczdera/ai-logo-preserve/backend/internal/db/sqlc"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/queue"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/storage"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/redis/go-redis/v9"
)

var (
	ALLOWED_ORIGINS = []string{"http://localhost:3000"}
)

type Server struct {
	config        utils.Config
	storageClient storage.Client
	store         db.Querier
	redisClient   *redis.Client
	queueClient   queue.Client
	router        *gin.Engine
}

func NewServer(cfg utils.Config, storageClient storage.Client, store db.Querier, redisClient *redis.Client, queueClient queue.Client) *Server {
	server := &Server{
		config:        cfg,
		storageClient: storageClient,
		store:         store,
		redisClient:   redisClient,
		queueClient:   queueClient,
	}

	server.setupRouter()
	return server
}

func (s *Server) setupRouter() {
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiting
	// s.router.Use(ratelimit.RateLimiter(ratelimit.NewRedisRateLimiter(
	// 	s.redisClient,
	// 	ratelimit.RateLimiterConfig{
	// 		Rate:   time.Duration(s.config.Server.RateLimit.RequestsPerHour) * time.Hour,
	// 		Burst:  s.config.Server.RateLimit.Burst,
	// 		Prefix: "logo-preserve:",
	// 	},
	// )))

	router.GET("/health", s.healthCheck)

	api := router.Group("/api/v1")
	{
		api.POST("/upload", s.UploadImage)
		api.POST("/upload/presigned-url", s.GetPresignedUrl)
		// api.GET("/jobs/:id", s.getJobStatus)
		// api.GET("/jobs/:id/result", s.getJobResult)
	}

	s.router = router
}

func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Server.Port)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "logo-preserve-backend",
	})
}

// func (s *Server) storeJobStatusInRedis(job *db.Job) error {
// 	ctx := context.Background()
// 	key := fmt.Sprintf("job:%s", job.ID.String())

// 	// Store job as JSON in Redis
// 	return s.redisClient.HSet(ctx, key, map[string]interface{}{
// 		"id":         job.ID.String(),
// 		"status":     job.Status,
// 		"s3_key":     job.S3Key,
// 		"upload_url": job.UploadUrl.String,
// 		"created_at": job.CreatedAt.Time.Format(time.RFC3339),
// 		"updated_at": job.UpdatedAt.Time.Format(time.RFC3339),
// 	}).Err()
// }

func successResponse(data interface{}, message string) gin.H {
	return gin.H{"success": true, "message": message, "data": data}
}

func errResponse(err error, message string) gin.H {
	return gin.H{"success": false, message: message, "error": err.Error()}
}

func (s *Server) isValidImageType(contentType string) bool {
	for _, allowedType := range s.config.Server.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
