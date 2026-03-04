package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"

	"github.com/pathplanner-lab/backend/internal/config"
	"github.com/pathplanner-lab/backend/internal/repository"
	"github.com/pathplanner-lab/backend/internal/worker"
)

func main() {
	cfg := config.Load()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("connect to db", zap.Error(err))
	}
	defer pool.Close()

	sqsClient, err := buildSQSClient(ctx, cfg)
	if err != nil {
		logger.Fatal("build sqs client", zap.Error(err))
	}

	source := worker.NewSQSSource(sqsClient, cfg.SQSQueueURL)
	runRepo := repository.NewRunRepository(pool)
	expRepo := repository.NewExperimentRepository(pool)
	mapRepo := repository.NewMapRepository(pool)

	w := worker.New(source, runRepo, expRepo, mapRepo, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("received shutdown signal")
		cancel()
	}()

	w.Run(ctx)
	logger.Info("worker exited")
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
