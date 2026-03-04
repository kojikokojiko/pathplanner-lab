package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL         string
	JWTSecret           string
	SQSQueueURL         string
	SQSEndpoint         string
	AWSRegion           string
	Port                string
	AnthropicAPIKey     string
	LogLevel            string
	WorkerConcurrency   int
	CognitoUserPoolID   string
	CognitoClientID     string
	CognitoRegion       string
}

func Load() *Config {
	concurrency := 4
	if v := os.Getenv("WORKER_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			concurrency = n
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	cognitoRegion := os.Getenv("COGNITO_REGION")
	if cognitoRegion == "" {
		cognitoRegion = region
	}
	return &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		SQSQueueURL:       os.Getenv("SQS_QUEUE_URL"),
		SQSEndpoint:       os.Getenv("SQS_ENDPOINT"),
		AWSRegion:         region,
		Port:              port,
		AnthropicAPIKey:   os.Getenv("ANTHROPIC_API_KEY"),
		LogLevel:          logLevel,
		WorkerConcurrency: concurrency,
		CognitoUserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		CognitoClientID:   os.Getenv("COGNITO_CLIENT_ID"),
		CognitoRegion:     cognitoRegion,
	}
}
