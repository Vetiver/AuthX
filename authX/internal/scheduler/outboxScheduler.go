package scheduler

import (
	"context"
	"time"

	"authX/internal/kafka"

	"go.uber.org/zap"
)

const (
	checkInterval = 1 * time.Second
	batchSize     = 100
)

type OutboxRepo interface {
	GetAndPublishPendingEvents(ctx context.Context, limit int, producer *kafka.Producer, logger *zap.Logger) error
}

type OutboxScheduler struct {
	repo     OutboxRepo
	producer *kafka.Producer
	logger   *zap.Logger
}

func NewOutboxScheduler(repo OutboxRepo, producer *kafka.Producer, logger *zap.Logger) *OutboxScheduler {
	return &OutboxScheduler{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (s *OutboxScheduler) ProcessPendingEvents(ctx context.Context) {
	err := s.repo.GetAndPublishPendingEvents(ctx, batchSize, s.producer, s.logger)
	if err != nil {
		s.logger.Error("Failed to get pending events", zap.Error(err))
		return
	}
}

func (s *OutboxScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	s.logger.Info("Outbox scheduler started")

	for {
		select {
		case <-ticker.C:
			s.ProcessPendingEvents(ctx)
		case <-ctx.Done():
			s.logger.Info("Outbox scheduler stopped")
			return
		}
	}
}
