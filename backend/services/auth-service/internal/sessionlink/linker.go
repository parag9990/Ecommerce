package sessionlink

import (
	"context"
	"errors"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type Linker interface {
	RecordLoginSucceeded(ctx context.Context, event LoginSucceededEvent) error
	RecordSignupSucceeded(ctx context.Context, event SignupSucceededEvent) error
	RecordLogoutSucceeded(ctx context.Context, event LogoutSucceededEvent) error
	RecordRefreshReuseDetected(ctx context.Context, event RefreshReuseDetectedEvent) error
}

type OutboxRepository interface {
	InsertOutboxEvent(ctx context.Context, event domain.OutboxEvent) error
}

type OutboxLinker struct {
	repo   OutboxRepository
	logger *slog.Logger
}

func NewOutboxLinker(repo OutboxRepository, logger *slog.Logger) (*OutboxLinker, error) {
	if repo == nil {
		return nil, errors.New("outbox repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &OutboxLinker{repo: repo, logger: logger}, nil
}

func (l *OutboxLinker) RecordLoginSucceeded(ctx context.Context, event LoginSucceededEvent) error {
	outboxEvent, err := BuildLoginSucceededOutboxEvent(event)
	if err != nil {
		return err
	}
	if err := l.repo.InsertOutboxEvent(ctx, outboxEvent); err != nil {
		return err
	}
	l.logger.InfoContext(ctx, "auth.session_link.enqueued",
		slog.String("event_type", outboxEvent.EventType),
		slog.String("event_id", outboxEvent.EventID),
		slog.String("account_id", event.AccountID),
		slog.String("session_id", event.SessionID),
	)
	return nil
}

func (l *OutboxLinker) RecordSignupSucceeded(ctx context.Context, event SignupSucceededEvent) error {
	outboxEvent, err := BuildSignupSucceededOutboxEvent(event)
	if err != nil {
		return err
	}
	if err := l.repo.InsertOutboxEvent(ctx, outboxEvent); err != nil {
		return err
	}
	l.logger.InfoContext(ctx, "auth.session_link.enqueued",
		slog.String("event_type", outboxEvent.EventType),
		slog.String("event_id", outboxEvent.EventID),
		slog.String("account_id", event.AccountID),
		slog.String("session_id", event.SessionID),
	)
	return nil
}

func (l *OutboxLinker) RecordLogoutSucceeded(ctx context.Context, event LogoutSucceededEvent) error {
	outboxEvent, err := BuildLogoutSucceededOutboxEvent(event)
	if err != nil {
		return err
	}
	if err := l.repo.InsertOutboxEvent(ctx, outboxEvent); err != nil {
		return err
	}
	l.logger.InfoContext(ctx, "auth.session_link.enqueued",
		slog.String("event_type", outboxEvent.EventType),
		slog.String("event_id", outboxEvent.EventID),
		slog.String("account_id", event.AccountID),
		slog.String("session_id", event.SessionID),
		slog.Bool("all_devices", event.AllDevices),
	)
	return nil
}

func (l *OutboxLinker) RecordRefreshReuseDetected(ctx context.Context, event RefreshReuseDetectedEvent) error {
	outboxEvent, err := BuildRefreshReuseDetectedOutboxEvent(event)
	if err != nil {
		return err
	}
	if err := l.repo.InsertOutboxEvent(ctx, outboxEvent); err != nil {
		return err
	}
	l.logger.InfoContext(ctx, "auth.session_link.enqueued",
		slog.String("event_type", outboxEvent.EventType),
		slog.String("event_id", outboxEvent.EventID),
		slog.String("account_id", event.AccountID),
		slog.String("session_id", event.SessionID),
		slog.String("refresh_token_id", event.TokenID),
	)
	return nil
}

type DisabledLinker struct{}

func NewDisabledLinker() DisabledLinker {
	return DisabledLinker{}
}

func (DisabledLinker) RecordLoginSucceeded(context.Context, LoginSucceededEvent) error {
	return nil
}

func (DisabledLinker) RecordSignupSucceeded(context.Context, SignupSucceededEvent) error {
	return nil
}

func (DisabledLinker) RecordLogoutSucceeded(context.Context, LogoutSucceededEvent) error {
	return nil
}

func (DisabledLinker) RecordRefreshReuseDetected(context.Context, RefreshReuseDetectedEvent) error {
	return nil
}
