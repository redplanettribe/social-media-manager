package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	DB          DBConfig
	ObjectStore ObjectStoreConfig
	JWT         JWTConfig
	Logger      LoggerConfig
	SSL         SSLConfig
	Scheduler   SchedulerConfig
	Publisher   PublisherConfig
}

type ObjectStoreConfig struct {
	Endpoint         string // including port
	AccessKey        string
	SecretAccessKey  string
	Bucket           string
	Region           string
	UseSSL           bool
	S3ForcePathStyle bool
}
type PublisherConfig struct {
	WorkerNum     int
	RetryNum      int
	PublishBuffer int
	RetryBuffer   int
}

type SchedulerConfig struct {
	Interval      time.Duration
	ChannelBuffer int
}

type AppConfig struct {
	Env    string
	Port   string
	Origen string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	SecretKey string
}

type LoggerConfig struct {
	Level string
}

type SSLConfig struct {
	CertPath string
	KeyPath  string
}

func LoadConfig() (*Config, error) {
	// Load .env file in development
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found")
		}
	}

	appPort := getEnv("APP_PORT", "8080")
	urlOrigen := getEnv("ORIGEN", "http://localhost:"+appPort)

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, err
	}

	config := &Config{
		App: AppConfig{
			Env:    getEnv("APP_ENV", "development"),
			Port:   appPort,
			Origen: urlOrigen,
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("POSTGRES_USER", "myuser"),
			Password: getEnv("POSTGRES_PASSWORD", "mypassword"),
			Name:     getEnv("POSTGRES_DB", "mydatabase"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		ObjectStore: ObjectStoreConfig{
			Endpoint:         getEnv("OBJECT_STORE_ENDPOINT", "http://localhost:9000"),
			AccessKey:        getEnv("OBJECT_STORE_ACCESS_KEY", "minio"),
			SecretAccessKey:  getEnv("OBJECT_STORE_SECRET_ACCESS_KEY", ""),
			Bucket:           getEnv("OBJECT_STORE_BUCKET", "post-media"),
			Region:           getEnv("OBJECT_STORE_REGION", "us-east-1"),
			UseSSL:           getEnv("OBJECT_STORE_USE_SSL", "true") == "true",
			S3ForcePathStyle: getEnv("OBJECT_STORE_S3_FORCE_PATH_STYLE", "true") == "true",
		},

		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "your_secret_key"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
		},
		SSL: SSLConfig{
			CertPath: getEnv("SSL_CERT_PATH", ""),
			KeyPath:  getEnv("SSL_KEY_PATH", ""),
		},
		Scheduler: SchedulerConfig{
			Interval:      10 * time.Second,
			ChannelBuffer: 100,
		},
		Publisher: PublisherConfig{
			WorkerNum:     5,
			RetryNum:      3,
			PublishBuffer: 100,
			RetryBuffer:   100,
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
