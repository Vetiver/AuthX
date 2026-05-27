package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ProfilerAddr string
	HTTPAddr     string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	KafkaBrokers            string
	KafkaAuthEventsTopic    string
	KafkaAuthEventsDLQTopic string
	KafkaAuditConsumerGroup string
	KafkaClientID           string
	KafkaRequiredAcks       string
}

func NewConfig() *Config {
	return &Config{
		ProfilerAddr: fromEnv("PROFILER_ADDR", "0.0.0.0:6060"),
		HTTPAddr:     fromEnv("HTTP_ADDR", "0.0.0.0:8080"),

		DatabaseURL: fromEnv("DATABASE_URL", "postgres://authx:authx_password@localhost:5432/authx?sslmode=disable"),

		RedisAddr:     fromEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: fromEnv("REDIS_PASSWORD", ""),
		RedisDB:       fromEnvAsInt("REDIS_DB", 0),

		JWTSecret:     fromEnv("JWT_SECRET", "local-dev-secret"),
		JWTAccessTTL:  fromEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: fromEnvAsDuration("JWT_REFRESH_TTL", 168*time.Hour),

		KafkaBrokers:            fromEnv("KAFKA_BROKERS", "localhost:29092"),
		KafkaAuthEventsTopic:    fromEnv("KAFKA_AUTH_EVENTS_TOPIC", "authx.auth.events"),
		KafkaAuthEventsDLQTopic: fromEnv("KAFKA_AUTH_EVENTS_DLQ_TOPIC", "authx.auth.events.dlq"),
		KafkaAuditConsumerGroup: fromEnv("KAFKA_AUDIT_CONSUMER_GROUP", "authx-audit-consumer"),
		KafkaClientID:           fromEnv("KAFKA_CLIENT_ID", "authx"),
		KafkaRequiredAcks:       fromEnv("KAFKA_REQUIRED_ACKS", "all"),
	}
}

func fromEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func fromEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func fromEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}
