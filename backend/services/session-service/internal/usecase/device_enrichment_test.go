package usecase

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestDeviceEnricherBuildsPrivacySafeContext(t *testing.T) {
	enricher, err := NewDeviceEnricher(
		fakeUserAgentParser{},
		fakeGeoResolver{},
		fakePrivacyHasher{},
		DeviceEnricherConfig{
			Enabled:            true,
			StoreUserAgent:     true,
			UserAgentMaxLength: 1024,
			GeoLookupTimeout:   time.Millisecond,
			DefaultChannel:     domain.ChannelUnknown,
			DefaultDevice:      domain.DeviceTypeUnknown,
			DefaultIPVersion:   domain.IPVersionUnknown,
		},
		discardLogger(),
	)
	if err != nil {
		t.Fatalf("new device enricher: %v", err)
	}

	got := enricher.Enrich(context.Background(), DeviceEnrichmentInput{
		SessionID:         "sess_123",
		AnonymousID:       "anon_123",
		UserAgent:         " Mozilla/5.0 Chrome/125.0 ",
		ClientIP:          "203.0.113.10",
		DeviceFingerprint: "raw-fingerprint",
		Channel:           domain.ChannelUserAppWeb,
		Locale:            "en-US,en;q=0.8",
		Timezone:          "Asia/Kolkata",
		ViewportWidth:     390,
		ViewportHeight:    844,
	})

	if got.Device.Type != domain.DeviceTypeMobile || got.Device.Browser == nil || *got.Device.Browser != "Chrome" {
		t.Fatalf("unexpected parsed device: %+v", got.Device)
	}
	if got.IPHash == "" || strings.Contains(got.IPHash, "203.0.113.10") {
		t.Fatalf("expected privacy-safe IP hash, got %q", got.IPHash)
	}
	if got.DeviceFingerprintHash == nil || strings.Contains(*got.DeviceFingerprintHash, "raw-fingerprint") {
		t.Fatalf("expected hashed fingerprint, got %#v", got.DeviceFingerprintHash)
	}
	if got.Client.Locale == nil || *got.Client.Locale != "en-US" {
		t.Fatalf("expected first locale token, got %+v", got.Client)
	}
	if got.Geo.Country == nil || *got.Geo.Country != "IN" || got.Geo.Source != domain.GeoSourceGeoIP {
		t.Fatalf("expected GeoIP result, got %+v", got.Geo)
	}
}

func TestDeviceEnricherClientHintsOverrideParserGaps(t *testing.T) {
	enricher, err := NewDeviceEnricher(
		unknownUserAgentParser{},
		fakeGeoResolver{},
		fakePrivacyHasher{},
		DeviceEnricherConfig{Enabled: true, StoreUserAgent: true},
		discardLogger(),
	)
	if err != nil {
		t.Fatalf("new device enricher: %v", err)
	}

	got := enricher.Enrich(context.Background(), DeviceEnrichmentInput{
		UserAgent: "unknown",
		ClientHints: ClientHints{
			UA:       `"Chromium";v="125", "Google Chrome";v="125"`,
			Platform: `"Android"`,
			Mobile:   "?1",
			Model:    `"Pixel 7"`,
		},
	})

	if got.Device.Type != domain.DeviceTypeMobile {
		t.Fatalf("device type = %s, want mobile", got.Device.Type)
	}
	if got.Device.Browser == nil || *got.Device.Browser != "Google Chrome" {
		t.Fatalf("browser = %#v, want Google Chrome", got.Device.Browser)
	}
	if got.Device.OS == nil || *got.Device.OS != "Android" {
		t.Fatalf("os = %#v, want Android", got.Device.OS)
	}
	if got.Device.Model == nil || *got.Device.Model != "Pixel 7" {
		t.Fatalf("model = %#v, want Pixel 7", got.Device.Model)
	}
}

type fakeUserAgentParser struct{}

func (fakeUserAgentParser) Parse(ctx context.Context, userAgent string) ParsedUserAgent {
	return ParsedUserAgent{
		Browser:        "Chrome",
		BrowserVersion: "125.0",
		OS:             "Android",
		OSVersion:      "14",
		DeviceType:     domain.DeviceTypeMobile,
		Model:          "Pixel 7",
	}
}

type fakeGeoResolver struct{}

func (fakeGeoResolver) Resolve(ctx context.Context, clientIP string) domain.Geo {
	country := "IN"
	city := "Delhi"
	return domain.Geo{
		Country: &country,
		City:    &city,
		Source:  domain.GeoSourceGeoIP,
	}
}

type fakePrivacyHasher struct{}

func (fakePrivacyHasher) Hash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("hashed-%x", len(value))
}
