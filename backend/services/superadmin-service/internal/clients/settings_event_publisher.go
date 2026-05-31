package clients

import (
	"context"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

const DefaultPlatformSettingsEventTopic = "platform.settings.updated"

type LoggingSettingsEventPublisher struct {
	topic  string
	logger logging.Logger
}

func NewLoggingSettingsEventPublisher(topic string, logger logging.Logger) *LoggingSettingsEventPublisher {
	if strings.TrimSpace(topic) == "" {
		topic = DefaultPlatformSettingsEventTopic
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	return &LoggingSettingsEventPublisher{topic: strings.TrimSpace(topic), logger: logger}
}

func (p *LoggingSettingsEventPublisher) PublishPlatformSettingUpdated(ctx context.Context, event domain.PlatformSettingUpdatedEvent) error {
	p.logger.Info(ctx, "platform setting updated event",
		"topic", p.topic,
		"event_id", event.EventID,
		"event_type", event.EventType,
		"setting_key", event.SettingKey,
		"version", event.Version,
		"updated_by_admin_id", event.UpdatedByAdminID,
		"updated_at", event.UpdatedAt,
	)
	return nil
}
