package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestNewMongoNotificationRepositoryValidatesCollectionMapping(t *testing.T) {
	t.Parallel()

	database := localDatabaseHandle(t)
	tests := []struct {
		name           string
		templates      string
		deliveries     string
		preferences    string
		providerEvents string
	}{
		{name: "missing template collection", deliveries: "deliveries", preferences: "preferences", providerEvents: "events"},
		{name: "reserved collection", templates: "system.templates", deliveries: "deliveries", preferences: "preferences", providerEvents: "events"},
		{name: "same collection", templates: "documents", deliveries: "documents", preferences: "preferences", providerEvents: "events"},
		{name: "duplicate preference collection", templates: "templates", deliveries: "deliveries", preferences: "deliveries", providerEvents: "events"},
		{name: "duplicate provider event collection", templates: "templates", deliveries: "deliveries", preferences: "preferences", providerEvents: "deliveries"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewMongoNotificationRepository(database, test.templates, test.deliveries, test.preferences, test.providerEvents); err == nil {
				t.Fatal("NewMongoNotificationRepository() returned nil error")
			}
		})
	}
}

func TestInsertDeliveryValidatesBeforeDatabaseWrite(t *testing.T) {
	t.Parallel()

	repo := &MongoNotificationRepository{}
	err := repo.InsertDelivery(context.Background(), domain.Delivery{})
	if !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("InsertDelivery() error = %v, want %v", err, domain.ErrInvalidDelivery)
	}
}

func TestEventDeliveryMethodsValidateBeforeDatabaseWrite(t *testing.T) {
	t.Parallel()

	repo := &MongoNotificationRepository{}
	if _, err := repo.ClaimEventDelivery(context.Background(), domain.Delivery{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("ClaimEventDelivery() error = %v, want ErrInvalidDelivery", err)
	}
	if err := repo.RecordEventDeliveryOutcome(context.Background(), "", domain.EventDeliveryOutcome{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("RecordEventDeliveryOutcome() error = %v, want ErrInvalidDelivery", err)
	}
}

func TestRetryDeliveryMethodsValidateBeforeDatabaseWrite(t *testing.T) {
	t.Parallel()

	repo := &MongoNotificationRepository{}
	if _, err := repo.PrepareEventDelivery(context.Background(), domain.Delivery{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("PrepareEventDelivery() error = %v, want ErrInvalidDelivery", err)
	}
	if _, err := repo.BeginAttempt(context.Background(), "", "", 0, time.Time{}, time.Time{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("BeginAttempt() error = %v, want ErrInvalidDelivery", err)
	}
	if err := repo.MarkRetryScheduled(context.Background(), "", 0, time.Time{}, "", time.Time{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("MarkRetryScheduled() error = %v, want ErrInvalidDelivery", err)
	}
	if err := repo.MarkDeadLettered(context.Background(), "", 0, "", time.Time{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("MarkDeadLettered() error = %v, want ErrInvalidDelivery", err)
	}
	if err := repo.MarkSuppressed(context.Background(), "", 0, "", time.Time{}); !errors.Is(err, domain.ErrInvalidDelivery) {
		t.Fatalf("MarkSuppressed() error = %v, want ErrInvalidDelivery", err)
	}
}

func TestPreferenceMethodsValidateBeforeDatabaseWrite(t *testing.T) {
	t.Parallel()

	repo := &MongoNotificationRepository{}
	if _, err := repo.FindPreferenceByUserID(context.Background(), ""); !errors.Is(err, domain.ErrInvalidPreference) {
		t.Fatalf("FindPreferenceByUserID() error = %v, want ErrInvalidPreference", err)
	}
	if _, err := repo.UpsertPreference(context.Background(), domain.Preference{}); !errors.Is(err, domain.ErrInvalidPreference) {
		t.Fatalf("UpsertPreference() error = %v, want ErrInvalidPreference", err)
	}
}

func TestAnalyticsMethodsValidateBeforeDatabaseWrite(t *testing.T) {
	t.Parallel()

	repo := &MongoNotificationRepository{}
	if _, err := repo.FindDeliveryByProviderMessageID(context.Background(), "", "", ""); !errors.Is(err, domain.ErrInvalidProviderEvent) {
		t.Fatalf("FindDeliveryByProviderMessageID() error = %v, want ErrInvalidProviderEvent", err)
	}
	if _, err := repo.RecordEventAndApplyMilestone(context.Background(), domain.ProviderEvent{}); !errors.Is(err, domain.ErrInvalidProviderEvent) {
		t.Fatalf("RecordEventAndApplyMilestone() error = %v, want ErrInvalidProviderEvent", err)
	}
}

func TestRepositoryHonorsCanceledContextBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &MongoNotificationRepository{}
	err := repo.InsertTemplate(ctx, domain.Template{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("InsertTemplate() error = %v, want context canceled", err)
	}
}

func localDatabaseHandle(t *testing.T) *mongo.Database {
	t.Helper()
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		t.Fatalf("mongo.Connect() returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Errorf("mongo Disconnect() returned error: %v", err)
		}
	})
	return client.Database("notification_repository_test")
}
