package sessionlink

import (
	"strings"
	"unicode"
)

const (
	maxAnonymousIDLength = 128
	maxMetadataLength    = 128
	maxUserAgentLength   = 1024
)

func NormalizeAnonymousID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxAnonymousIDLength {
		return ""
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		switch r {
		case '_', '-', '.', ':':
			continue
		default:
			return ""
		}
	}
	return value
}

func NormalizeMetadataValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxMetadataLength {
		return value
	}
	return value[:maxMetadataLength]
}

func NormalizeUserAgent(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxUserAgentLength {
		return value
	}
	return value[:maxUserAgentLength]
}

func normalizeLoginEvent(event LoginSucceededEvent) LoginSucceededEvent {
	event.EventID = strings.TrimSpace(event.EventID)
	event.TraceID = NormalizeMetadataValue(event.TraceID)
	event.AccountID = strings.TrimSpace(event.AccountID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.SessionID = strings.TrimSpace(event.SessionID)
	event.AnonymousID = NormalizeAnonymousID(event.AnonymousID)
	event.Roles = normalizeRoles(event.Roles)
	event.SellerID = strings.TrimSpace(event.SellerID)
	event.TenantID = strings.TrimSpace(event.TenantID)
	event.Device = normalizeDeviceInfo(event.Device)
	event.Network.IPHash = strings.TrimSpace(event.Network.IPHash)
	event.Auth.Method = NormalizeMetadataValue(event.Auth.Method)
	if event.Auth.Method == "" {
		event.Auth.Method = AuthMethodPassword
	}
	event.OccurredAt = event.OccurredAt.UTC()
	return event
}

func normalizeSignupEvent(event SignupSucceededEvent) SignupSucceededEvent {
	event.EventID = strings.TrimSpace(event.EventID)
	event.TraceID = NormalizeMetadataValue(event.TraceID)
	event.AccountID = strings.TrimSpace(event.AccountID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.SessionID = strings.TrimSpace(event.SessionID)
	event.AnonymousID = NormalizeAnonymousID(event.AnonymousID)
	event.Roles = normalizeRoles(event.Roles)
	event.SellerID = strings.TrimSpace(event.SellerID)
	event.TenantID = strings.TrimSpace(event.TenantID)
	event.Device = normalizeDeviceInfo(event.Device)
	event.Network.IPHash = strings.TrimSpace(event.Network.IPHash)
	event.Auth.Method = NormalizeMetadataValue(event.Auth.Method)
	if event.Auth.Method == "" {
		event.Auth.Method = AuthMethodSignup
	}
	event.OccurredAt = event.OccurredAt.UTC()
	return event
}

func normalizeLogoutEvent(event LogoutSucceededEvent) LogoutSucceededEvent {
	event.EventID = strings.TrimSpace(event.EventID)
	event.TraceID = NormalizeMetadataValue(event.TraceID)
	event.AccountID = strings.TrimSpace(event.AccountID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.SessionID = strings.TrimSpace(event.SessionID)
	event.Reason = NormalizeMetadataValue(event.Reason)
	if event.Reason == "" {
		event.Reason = LogoutReasonUserRequested
	}
	event.OccurredAt = event.OccurredAt.UTC()
	return event
}

func normalizeRefreshReuseEvent(event RefreshReuseDetectedEvent) RefreshReuseDetectedEvent {
	event.EventID = strings.TrimSpace(event.EventID)
	event.TraceID = NormalizeMetadataValue(event.TraceID)
	event.AccountID = strings.TrimSpace(event.AccountID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.SessionID = strings.TrimSpace(event.SessionID)
	event.TokenID = strings.TrimSpace(event.TokenID)
	event.IPHash = strings.TrimSpace(event.IPHash)
	event.DeviceFingerprintHash = strings.TrimSpace(event.DeviceFingerprintHash)
	event.ActionTaken = NormalizeMetadataValue(event.ActionTaken)
	if event.ActionTaken == "" {
		event.ActionTaken = ActionRevokedSessionFamily
	}
	event.OccurredAt = event.OccurredAt.UTC()
	return event
}

func normalizeDeviceInfo(device DeviceInfo) DeviceInfo {
	return DeviceInfo{
		DeviceFingerprintHash: strings.TrimSpace(device.DeviceFingerprintHash),
		UserAgent:             NormalizeUserAgent(device.UserAgent),
		Channel:               NormalizeMetadataValue(device.Channel),
		Locale:                NormalizeMetadataValue(device.Locale),
	}
}

func normalizeRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}
