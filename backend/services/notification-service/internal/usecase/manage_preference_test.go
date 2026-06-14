package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type preferenceStore struct {
	preference domain.Preference
	findErr    error
	saved      domain.Preference
}

func (s *preferenceStore) FindPreferenceByUserID(context.Context, string) (domain.Preference, error) {
	return s.preference, s.findErr
}

func (s *preferenceStore) UpsertPreference(_ context.Context, preference domain.Preference) (domain.Preference, error) {
	s.saved = preference
	return preference, nil
}

func TestManagePreferenceReturnsSafeDefaultsWhenNoRecordExists(t *testing.T) {
	t.Parallel()

	service := newPreferenceService(t, &preferenceStore{findErr: domain.ErrPreferenceNotFound})
	preference, err := service.Get(context.Background(), "user_1")
	if err != nil || !preference.EmailEnabled || preference.MarketingEnabled {
		t.Fatalf("Get() = %+v, %v", preference, err)
	}
}

func TestManagePreferencePatchPreservesExplicitFalseAndOmittedFields(t *testing.T) {
	t.Parallel()

	store := &preferenceStore{preference: domain.DefaultPreference("user_1", time.Now())}
	store.preference.MarketingEnabled = true
	service := newPreferenceService(t, store)
	disabled := false
	preference, err := service.Update(context.Background(), "user_1", usecase.PreferencePatch{
		MarketingEnabled: &disabled,
	})
	if err != nil || preference.MarketingEnabled || !preference.EmailEnabled || store.saved.MarketingEnabled {
		t.Fatalf("Update() = %+v, %v; saved = %+v", preference, err, store.saved)
	}
}

func TestManagePreferenceRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	service := newPreferenceService(t, &preferenceStore{})
	_, err := service.Update(context.Background(), "user_1", usecase.PreferencePatch{})
	if !errors.Is(err, domain.ErrInvalidPreference) {
		t.Fatalf("Update() error = %v", err)
	}
}

func newPreferenceService(t *testing.T, store usecase.PreferenceStore) *usecase.ManagePreferenceService {
	t.Helper()
	service, err := usecase.NewManagePreferenceService(store)
	if err != nil {
		t.Fatalf("NewManagePreferenceService() error = %v", err)
	}
	service.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 15, 0, 0, 0, time.UTC)
	})
	return service
}
