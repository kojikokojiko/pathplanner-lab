package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"

	"github.com/pathplanner-lab/backend/internal/config"
	"github.com/pathplanner-lab/backend/internal/handler"
	"github.com/pathplanner-lab/backend/internal/llm"
	"github.com/pathplanner-lab/backend/internal/middleware"
	"github.com/pathplanner-lab/backend/internal/repository"
	"github.com/pathplanner-lab/backend/internal/service"
	workerSvc "github.com/pathplanner-lab/backend/internal/worker"
)

func main() {
	cfg := config.Load()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DB
	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("connect to db", zap.Error(err))
	}
	defer pool.Close()

	// SQS
	var sqsPub service.SQSPublisher
	if cfg.SQSQueueURL != "" {
		sqsClient, err := buildSQSClient(ctx, cfg)
		if err != nil {
			logger.Warn("SQS not available", zap.Error(err))
		} else {
			sqsPub = workerSvc.NewSQSPublisher(sqsClient, cfg.SQSQueueURL)
		}
	}

	// Repos
	userRepo := repository.NewUserRepository(pool)
	mapRepo := repository.NewMapRepository(pool)
	expRepo := repository.NewExperimentRepository(pool)
	runRepo := repository.NewRunRepository(pool)
	llmRepo := repository.NewLLMReportRepository(pool)
	idempRepo := repository.NewIdempotencyRepository(pool)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	mapSvc := service.NewMapService(mapRepo)
	expSvc := service.NewExperimentService(expRepo, runRepo, mapRepo, sqsPub)
	runSvc := service.NewRunService(runRepo, expRepo)
	compareSvc := service.NewCompareService(runRepo)

	var llmSvc *service.LLMService
	if cfg.AnthropicAPIKey != "" {
		llmClient := llm.NewClient(cfg.AnthropicAPIKey)
		llmSvc = service.NewLLMService(llmClient, runRepo, expRepo, mapRepo, llmRepo)
	}

	// Auth0 JWKS (optional — local dev uses HMAC JWT when not configured)
	var auth0KF *middleware.Auth0KeyFunc
	var upsertUser middleware.UpsertUserFunc
	if cfg.Auth0Domain != "" {
		auth0KF = middleware.NewAuth0KeyFunc(cfg.Auth0Domain, cfg.Auth0Audience)
		upsertUser = func(ctx context.Context, id, email string) error {
			return userRepo.Upsert(ctx, id, email)
		}
	}

	deps := handler.Deps{
		AuthSvc:      authSvc,
		MapSvc:       mapSvc,
		ExpSvc:       expSvc,
		RunSvc:       runSvc,
		CompareSvc:   compareSvc,
		LLMSvc:       llmSvc,
		IdempRepo:    idempRepo,
		JWTSecret:    cfg.JWTSecret,
		Auth0KeyFunc: auth0KF,
		UpsertUser:   upsertUser,
	}

	router := handler.NewRouter(deps)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("starting API server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}
}

func buildSQSClient(ctx context.Context, cfg *config.Config) (*sqs.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.AWSRegion),
	}
	if cfg.SQSEndpoint != "" {
		opts = append(opts,
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		)
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	var sqsOpts []func(*sqs.Options)
	if cfg.SQSEndpoint != "" {
		sqsOpts = append(sqsOpts, func(o *sqs.Options) {
			o.BaseEndpoint = &cfg.SQSEndpoint
		})
	}
	return sqs.NewFromConfig(awsCfg, sqsOpts...), nil
}
