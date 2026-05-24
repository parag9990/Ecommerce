package domain

import (
	"strings"
	"time"
)

type SellerSettings struct {
	SellerID       string
	ReturnPolicy   string
	ShippingPolicy string
	SupportEmail   string
	Settings       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s SellerSettings) Normalized() SellerSettings {
	s.SellerID = strings.TrimSpace(s.SellerID)
	s.ReturnPolicy = strings.TrimSpace(s.ReturnPolicy)
	s.ShippingPolicy = strings.TrimSpace(s.ShippingPolicy)
	s.SupportEmail = strings.TrimSpace(s.SupportEmail)
	if s.Settings == nil {
		s.Settings = map[string]any{}
	}
	if !s.CreatedAt.IsZero() {
		s.CreatedAt = s.CreatedAt.UTC()
	}
	if !s.UpdatedAt.IsZero() {
		s.UpdatedAt = s.UpdatedAt.UTC()
	}
	return s
}

func DefaultSellerSettings(sellerID string) SellerSettings {
	return SellerSettings{
		SellerID: strings.TrimSpace(sellerID),
		Settings: map[string]any{},
	}.Normalized()
}
