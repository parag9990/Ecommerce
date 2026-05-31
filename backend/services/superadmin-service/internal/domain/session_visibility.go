package domain

import "time"

type SessionAnalyticsRoute string

const (
	SessionAnalyticsRouteLive     SessionAnalyticsRoute = "analytics.live"
	SessionAnalyticsRouteSessions SessionAnalyticsRoute = "analytics.sessions"
	SessionAnalyticsRouteJourney  SessionAnalyticsRoute = "analytics.journey"
	SessionAnalyticsRouteFunnels  SessionAnalyticsRoute = "analytics.funnels"
	SessionAnalyticsRouteHeatmaps SessionAnalyticsRoute = "analytics.heatmaps"
)

func (r SessionAnalyticsRoute) Valid() bool {
	switch r {
	case SessionAnalyticsRouteLive,
		SessionAnalyticsRouteSessions,
		SessionAnalyticsRouteJourney,
		SessionAnalyticsRouteFunnels,
		SessionAnalyticsRouteHeatmaps:
		return true
	default:
		return false
	}
}

func KnownSessionAnalyticsRoutes() []SessionAnalyticsRoute {
	return []SessionAnalyticsRoute{
		SessionAnalyticsRouteLive,
		SessionAnalyticsRouteSessions,
		SessionAnalyticsRouteJourney,
		SessionAnalyticsRouteFunnels,
		SessionAnalyticsRouteHeatmaps,
	}
}

type SessionAnalyticsDeviceType string

const (
	SessionAnalyticsDeviceDesktop SessionAnalyticsDeviceType = "desktop"
	SessionAnalyticsDeviceMobile  SessionAnalyticsDeviceType = "mobile"
	SessionAnalyticsDeviceTablet  SessionAnalyticsDeviceType = "tablet"
)

func (d SessionAnalyticsDeviceType) Valid() bool {
	switch d {
	case "", SessionAnalyticsDeviceDesktop, SessionAnalyticsDeviceMobile, SessionAnalyticsDeviceTablet:
		return true
	default:
		return false
	}
}

type SessionAnalyticsFilters struct {
	From       time.Time
	To         time.Time
	Page       int
	PageSize   int
	SessionID  string
	Path       string
	DeviceType SessionAnalyticsDeviceType
}
