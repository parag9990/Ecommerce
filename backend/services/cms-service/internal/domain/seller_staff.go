package domain

import "time"

type SellerStaff struct {
	StaffID   string
	SellerID  string
	UserID    string
	Role      Role
	Status    StaffStatus
	InvitedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s SellerStaff) Active() bool {
	return s.Status.IsActive()
}
