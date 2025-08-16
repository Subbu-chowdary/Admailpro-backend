package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI       string
	RedisAddr      string
	JWTSecret      string
	AWSCredentials AWSCreds
}

type AWSCreds struct {
	AccessKey string
	SecretKey string
	Region    string
}

var (
	config *Config
	once   sync.Once
)

// GetConfig returns the singleton config
func GetConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}

		config = &Config{
			MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017"),
			RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
			JWTSecret: getEnv("JWT_SECRET", "super-secret-jwt-key"),

			AWSCredentials: AWSCreds{
				AccessKey: getEnv("AWS_ACCESS_KEY", ""),
				SecretKey: getEnv("AWS_SECRET_KEY", ""),
				Region:    getEnv("AWS_REGION", "us-east-1"),
			},
		}
	})

	return config
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
