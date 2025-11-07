package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
	Services ServicesConfig
}

type ServerConfig struct {
	Port         string
	Env          string
	ReadTimeout  int
	WriteTimeout int
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
	Prefix   string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type ServicesConfig struct {
	UserService  string
	RoomService  string
	GiftService  string
	AdminService string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Env:          getEnv("SERVER_ENV", "development"),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 60),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 60),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "3306"),
			User:         getEnv("DB_USER", "live_user"),
			Password:     getEnv("DB_PASSWORD", "live_pass123"),
			Database:     getEnv("DB_NAME", "live_platform"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "live_platform_exchange"),
			Prefix:   getEnv("RABBITMQ_QUEUE_PREFIX", "live_platform"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key"),
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		Services: ServicesConfig{
			UserService:  getEnv("USER_SERVICE_ADDR", "localhost:50051"),
			RoomService:  getEnv("ROOM_SERVICE_ADDR", "localhost:50052"),
			GiftService:  getEnv("GIFT_SERVICE_ADDR", "localhost:50053"),
			AdminService: getEnv("ADMIN_SERVICE_ADDR", "localhost:50054"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
