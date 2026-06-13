package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSessionValidateAcceptsActiveGuestSession(t *testing.T) {
	session := validSession()

	if err := session.Validate(); err != nil {
		t.Fatalf("expected valid session, got %v", err)
	}
}

func TestSessionValidateRejectsRawIPHash(t *testing.T) {
	session := validSession()
	session.IPHash = "203.0.113.10"

	err := session.Validate()
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected invalid session error, got %v", err)
	}
	if !strings.Contains(err.Error(), "ip_hash") {
		t.Fatalf("expected ip_hash violation, got %v", err)
	}
}

func TestSessionValidateRejectsActiveSessionWithEndMetadata(t *testing.T) {
	session := validSession()
	endedAt := session.StartedAt.Add(5 * time.Minute)
	reason := EndReasonLogout
	session.EndedAt = &endedAt
	session.EndReason = &reason

	err := session.Validate()
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected invalid session error, got %v", err)
	}
	if !strings.Contains(err.Error(), "ended_at") || !strings.Contains(err.Error(), "end_reason") {
		t.Fatalf("expected active lifecycle violations, got %v", err)
	}
}

func TestSessionValidateExpiredRequiresInactivityTimeout(t *testing.T) {
	session := validSession()
	endedAt := session.StartedAt.Add(35 * time.Minute)
	reason := EndReasonLogout
	session.Status = SessionStatusExpired
	session.EndedAt = &endedAt
	session.EndReason = &reason
	session.LastSeenAt = endedAt
	session.UpdatedAt = endedAt

	err := session.Validate()
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected invalid session error, got %v", err)
	}
	if !strings.Contains(err.Error(), "end_reason") {
		t.Fatalf("expected end_reason violation, got %v", err)
	}
}

func TestSessionLifecycleHelpersKeepModelValid(t *testing.T) {
	session := validSession()
	userID := "user_123"
	linkAt := session.StartedAt.Add(2 * time.Minute)

	if err := session.LinkUser(userID, linkAt); err != nil {
		t.Fatalf("link user: %v", err)
	}
	if session.UserID == nil || *session.UserID != userID {
		t.Fatalf("expected user id to be linked, got %#v", session.UserID)
	}

	expireAt := session.StartedAt.Add(31 * time.Minute)
	exitPage := "/checkout"
	if err := session.Expire(expireAt, &exitPage); err != nil {
		t.Fatalf("expire session: %v", err)
	}
	if session.Status != SessionStatusExpired || session.EndReason == nil || *session.EndReason != EndReasonInactivityTimeout {
		t.Fatalf("expected expired inactivity session, got status=%s reason=%v", session.Status, session.EndReason)
	}

	session = validSession()
	revokeAt := session.StartedAt.Add(4 * time.Minute)
	if err := session.Revoke(EndReasonSecurityRevoke, revokeAt); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	if session.Status != SessionStatusRevoked || session.RevokedAt == nil {
		t.Fatalf("expected revoked session, got status=%s revoked_at=%v", session.Status, session.RevokedAt)
	}
}

func validSession() Session {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	browser := "Chrome"
	return Session{
		SessionID:     "sess_01HX9ZK6N5M2V4B7S8T9Q0R1P2",
		AnonymousID:   "anon_01HX9ZJY35F7K2M4N6P8R0T1V2",
		SchemaVersion: CurrentSessionSchemaVersion,
		Status:        SessionStatusActive,
		Channel:       ChannelUserAppWeb,
		EntryPage:     "/",
		UserAgent:     "Mozilla/5.0 Chrome/125.0 Safari/537.36",
		Device: Device{
			Type:    DeviceTypeDesktop,
			Browser: &browser,
		},
		IPHash:      "7c9e6679f7425d42a01e0cfbdac7c7b39f4f5f9c31f5b22d9efb9b9f7436e75a",
		IPVersion:   IPVersionIPv4,
		RiskLevel:   RiskLevelLow,
		RiskReasons: []string{},
		StartedAt:   now,
		LastSeenAt:  now.Add(30 * time.Second),
		CreatedAt:   now,
		UpdatedAt:   now.Add(30 * time.Second),
	}
}
