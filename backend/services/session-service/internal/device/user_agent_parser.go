package device

import (
	"context"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
	ua "github.com/mileusna/useragent"
)

type UserAgentParser struct{}

func (UserAgentParser) Parse(ctx context.Context, raw string) usecase.ParsedUserAgent {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return usecase.ParsedUserAgent{DeviceType: domain.DeviceTypeUnknown}
	}
	parsed := ua.Parse(raw)
	return usecase.ParsedUserAgent{
		Browser:        parsed.Name,
		BrowserVersion: parsed.Version,
		OS:             parsed.OS,
		OSVersion:      parsed.OSVersion,
		DeviceType:     normalizeDeviceType(parsed),
		Model:          parsed.Device,
		IsBot:          parsed.Bot,
	}
}

func normalizeDeviceType(parsed ua.UserAgent) domain.DeviceType {
	switch {
	case parsed.Bot:
		return domain.DeviceTypeBot
	case parsed.Tablet:
		return domain.DeviceTypeTablet
	case parsed.Mobile:
		return domain.DeviceTypeMobile
	case parsed.Desktop:
		return domain.DeviceTypeDesktop
	default:
		return domain.DeviceTypeUnknown
	}
}
