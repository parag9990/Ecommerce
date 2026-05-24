package domain

import "strings"

type StaffStatus string

const (
	StaffStatusInvited  StaffStatus = "invited"
	StaffStatusActive   StaffStatus = "active"
	StaffStatusDisabled StaffStatus = "disabled"
	StaffStatusRevoked  StaffStatus = "revoked"
)

type ActorContext struct {
	UserID      string
	SessionID   string
	SellerID    string
	TenantID    string
	RequestID   string
	Roles       []Role
	StaffStatus StaffStatus
}

func (s StaffStatus) Normalized() StaffStatus {
	return StaffStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s StaffStatus) IsActive() bool {
	return s.Normalized() == StaffStatusActive
}

func (a ActorContext) Normalized() ActorContext {
	a.UserID = strings.TrimSpace(a.UserID)
	a.SessionID = strings.TrimSpace(a.SessionID)
	a.SellerID = strings.TrimSpace(a.SellerID)
	a.TenantID = strings.TrimSpace(a.TenantID)
	a.RequestID = strings.TrimSpace(a.RequestID)
	values := make([]string, 0, len(a.Roles))
	for _, role := range a.Roles {
		values = append(values, role.String())
	}
	a.Roles = NormalizeRoles(values)
	a.StaffStatus = a.StaffStatus.Normalized()
	return a
}

func (a ActorContext) Authenticated() bool {
	a = a.Normalized()
	return a.UserID != ""
}
