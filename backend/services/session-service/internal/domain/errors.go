package domain

import "errors"

var (
	ErrInvalidSession         = errors.New("invalid session model")
	ErrInvalidSessionEvent    = errors.New("invalid session event")
	ErrInvalidJourney         = errors.New("invalid journey model")
	ErrInvalidHeatmap         = errors.New("invalid heatmap model")
	ErrInvalidAnalytics       = errors.New("invalid analytics model")
	ErrInvalidRetentionPolicy = errors.New("invalid retention policy")
	ErrInvalidActiveSession   = errors.New("invalid active session")
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionEventNotFound   = errors.New("session event not found")
	ErrJourneyNotFound        = errors.New("journey not found")
	ErrHeatmapNotFound        = errors.New("heatmap not found")
	ErrNilSession             = errors.New("session is nil")
)
