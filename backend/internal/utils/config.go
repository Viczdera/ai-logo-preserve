package utils

import (
	"log"

	"github.com/spf13/viper"
)

var DBDriver = "postgres"

type Config struct {
	Server     ServerConfig
	Cloudflare CloudflareConfig
	RabbitMQ   RabbitMQConfig
	Redis      RedisConfig
	Database   DatabaseConfig
}

type ServerConfig struct {
	Port         string   `mapstructure:"PORT"`
	MaxFileSize  int64    `mapstructure:"MAX_FILE_SIZE"`
	AllowedTypes []string `mapstructure:"ALLOWED_TYPES"`
	RateLimit    RateLimitConfig
}

type RateLimitConfig struct {
	RequestsPerHour int
	Burst           int
}

type CloudflareConfig struct {
	BucketName          string `mapstructure:"BUCKET_NAME"`
	BucketAccessKeyID   string `mapstructure:"BUCKET_ACCESS_KEY_ID"`
	BucketSecretKey     string `mapstructure:"BUCKET_SECRET_KEY"`
	BucketTokenValue    string `mapstructure:"BUCKET_TOKEN_VALUE"`
	CloudflareAccountID string `mapstructure:"CLOUDFARE_ACCOUNT_ID"`
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
	Queue    string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type DatabaseConfig struct {
	DBDriver string `mapstructure:"DB_DRIVER"`
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     string `mapstructure:"POSTGRES_PORT"`
	User     string `mapstructure:"POSTGRES_USER"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	DBName   string `mapstructure:"POSTGRES_DB"`
	SSLMode  string `mapstructure:"POSTGRES_SSLMODE"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	// Cloudflare configuration
	config.Cloudflare.BucketName = viper.GetString("BUCKET_NAME")
	config.Cloudflare.BucketAccessKeyID = viper.GetString("BUCKET_ACCESS_KEY_ID")
	config.Cloudflare.BucketSecretKey = viper.GetString("BUCKET_SECRET_KEY")
	config.Cloudflare.BucketTokenValue = viper.GetString("BUCKET_TOKEN_VALUE")
	config.Cloudflare.CloudflareAccountID = viper.GetString("CLOUDFARE_ACCOUNT_ID")

	// Server configuration
	config.Server.Port = viper.GetString("PORT")
	config.Server.MaxFileSize = viper.GetInt64("MAX_FILE_SIZE")
	config.Server.AllowedTypes = []string{"image/jpeg", "image/png"}
	config.Server.RateLimit.RequestsPerHour = viper.GetInt("RATE_LIMIT_REQUESTS_PER_HOUR")
	config.Server.RateLimit.Burst = viper.GetInt("RATE_LIMIT_BURST")

	// RabbitMQ configuration
	config.RabbitMQ.URL = viper.GetString("RABBITMQ_URL")
	config.RabbitMQ.Exchange = viper.GetString("RABBITMQ_EXCHANGE")
	config.RabbitMQ.Queue = viper.GetString("RABBITMQ_QUEUE")

	// Redis configuration
	config.Redis.Addr = viper.GetString("REDIS_ADDR")
	config.Redis.Password = viper.GetString("REDIS_PASSWORD")
	config.Redis.DB = viper.GetInt("REDIS_DB")

	// PostgreSQL configuration
	config.Database.DBDriver = "postgres"
	config.Database.Host = viper.GetString("POSTGRES_HOST")
	config.Database.Port = viper.GetString("POSTGRES_PORT")
	config.Database.User = viper.GetString("POSTGRES_USER")
	config.Database.Password = viper.GetString("POSTGRES_PASSWORD")
	config.Database.DBName = viper.GetString("POSTGRES_DB")
	config.Database.SSLMode = viper.GetString("POSTGRES_SSLMODE")

	return
}
