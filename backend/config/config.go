// backend/config/config.go
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
	SESConfigs     []SESConfig
	TrackingDomain string // New field for the tracking domain
}

type AWSCreds struct {
	AccessKey string
	SecretKey string
	Region    string
}

type SESConfig struct {
	Subdomain string
	IP        string
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

			SESConfigs:     generateSESConfigs(),
			TrackingDomain: getEnv("TRACKING_DOMAIN", "track.mail3.example.com"), // Load from env or use default
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

func generateSESConfigs() []SESConfig {
	subdomains := []string{
		"mail1.domainname.com", "mail2.domainname.com", "mail3.domainname.com",
		"mail4.domainname.com", "mail5.domainname.com",
	}
	ips := []string{
		"192.0.2.1", "192.0.2.2", "192.0.2.3", "192.0.2.4", "192.0.2.5",
		"192.0.2.6", "192.0.2.7", "192.0.2.8", "192.0.2.9", "192.0.2.10",
	}

	var configs []SESConfig
	for _, subdomain := range subdomains {
		for _, ip := range ips {
			configs = append(configs, SESConfig{Subdomain: subdomain, IP: ip})
		}
	}
	return configs
}